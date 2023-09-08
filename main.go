package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/PSKP-95/scheduler/api"
	"github.com/PSKP-95/scheduler/config"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/PSKP-95/scheduler/util"
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

	logger := &util.Log{
		InfoLog:  infoLog,
		ErrorLog: errorLog,
	}

	executorChan := make(chan hooks.Message)

	store := db.NewStore(conn)

	executor, err := hooks.NewExecutor(store, executorChan, logger)
	if err != nil {
		errorLog.Fatal("something wrong while creating executor: ", err)
	}

	worker, err := worker.NewWorker(workerConfig, store, executor, logger)
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

	err = server.Start(serverConfig.ServerAddress)
	if err != nil {
		errorLog.Fatal("Cannot start server: ", err)
	}
}
