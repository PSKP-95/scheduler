package main

import (
	"database/sql"
	"log"

	"github.com/PSKP-95/schedular/api"
	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/PSKP-95/schedular/hooks"
	"github.com/PSKP-95/schedular/util"
	"github.com/PSKP-95/schedular/worker"
	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	executorChan := make(chan hooks.Message)

	store := db.New(conn)
	executor, err := hooks.NewExecutor(config, store, executorChan)
	worker, err := worker.NewWorker(config, store, executor)

	if err != nil {
		log.Fatal("something wrong while creating executor: ", err)
	}

	server, err := api.NewServer(config, store, executor, worker)

	if err != nil {
		log.Fatal("something wrong while creating new server: ", err)
	}

	// register worker
	worker.Register()

	// start worker
	go worker.Work()

	// start executor
	go executor.Execute()

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot start server: ", err)
	}
}
