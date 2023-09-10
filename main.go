package main

import (
	"database/sql"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/PSKP-95/scheduler/api"
	"github.com/PSKP-95/scheduler/config"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/PSKP-95/scheduler/worker"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimestampFieldName = "timestamp"

	log.Logger = zerolog.New(os.Stdout).With().Str("application", "scheduler").Timestamp().Caller().Logger()

	dbConfig, serverConfig, workerConfig, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	conn, err := sql.Open(dbConfig.Driver, dbConfig.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to db")
	}

	executorChan := make(chan hooks.Message)
	workerKillSwitch := make(chan struct{})
	var wg sync.WaitGroup

	store := db.NewStore(conn)

	executor, err := hooks.NewExecutor(store, executorChan, &wg)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to db")
	}

	worker, err := worker.NewWorker(workerConfig, store, executor, workerKillSwitch)
	if err != nil {
		log.Fatal().Err(err).Msg("something wrong while creating worker")
	}

	server, err := api.NewServer(serverConfig, store, executor, worker)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to db")
	}

	// register worker
	err = worker.Register()
	if err != nil {
		log.Fatal().Err(err).Msg("Error while registering worker")
	}

	// start worker
	go worker.Work()

	// start executor
	go executor.Execute()

	go func() {
		err = server.Start(serverConfig.ServerAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("Cannot start server")
		}
	}()

	// graceful shutdown of webserver and db connection.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Info().Msg("Running cleanup tasks.")

	// close server so no new requests allowed
	_ = server.Shutdown()

	// stop worker
	workerKillSwitch <- struct{}{}
	<-workerKillSwitch

	// stop executor.
	log.Info().Msg("Waiting for hooks to complete.")
	wg.Wait()
	log.Info().Msg("Hooks completed.")

	// close db
	_ = store.Close()

	log.Info().Msg("Graceful shutdown done.")

}
