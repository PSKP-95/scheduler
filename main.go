package main

import (
	"database/sql"
	"log"

	"github.com/PSKP-95/schedular/api"
	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/PSKP-95/schedular/util"
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

	store := db.New(conn)
	server, err := api.NewServer(config, store)

	if err != nil {
		log.Fatal("something wrong while creating new server: ", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("Cannot start server: ", err)
	}
}
