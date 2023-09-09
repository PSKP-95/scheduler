package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/PSKP-95/scheduler/api"
	"github.com/PSKP-95/scheduler/config"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/PSKP-95/scheduler/mlog"
	"github.com/PSKP-95/scheduler/worker"
	_ "github.com/lib/pq"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	dbConfig, serverConfig, workerConfig, err := config.LoadConfig(".")
	if err != nil {
		errorLog.Fatal("cannot load config: ", err)
	}

	conn, err := sql.Open(dbConfig.Driver, dbConfig.URL)
	if err != nil {
		errorLog.Fatal("Cannot connect to db: ", err)
	}

	logger := &mlog.Log{
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}

	executorChan := make(chan hooks.Message)
	workerKillSwitch := make(chan struct{})
	var wg sync.WaitGroup

	store := db.NewStore(conn)

	executor, err := hooks.NewExecutor(store, executorChan, logger, &wg)
	if err != nil {
		errorLog.Fatal("something wrong while creating executor: ", err)
	}

	worker, err := worker.NewWorker(workerConfig, store, executor, logger, workerKillSwitch)
	if err != nil {
		errorLog.Fatal("something wrong while creating worker: ", err)
	}

	server, err := api.NewServer(serverConfig, store, executor, worker, logger)
	if err != nil {
		errorLog.Fatal("something wrong while creating new server: ", err)
	}

	// register worker
	err = worker.Register()
	if err != nil {
		errorLog.Fatal("Error while registering worker: ", err)
	}

	// start worker
	go worker.Work()

	// start executor
	go executor.Execute()

	go func() {
		err = server.Start(serverConfig.ServerAddress)
		if err != nil {
			errorLog.Fatal("Cannot start server: ", err)
		}
	}()

	// graceful shutdown of webserver and db connection.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Println("Running cleanup tasks.")

	// close server so no new requests allowed
	_ = server.Shutdown()

	// stop worker
	workerKillSwitch <- struct{}{}
	<-workerKillSwitch

	// stop executor.
	fmt.Println("Waiting for hooks to complete.")
	wg.Wait()
	fmt.Println("Hooks completed.")

	// close db
	_ = store.Close()

	fmt.Println("Graceful shutdown done.")
}
