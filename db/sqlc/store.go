package db

import (
	"database/sql"
)

type Store interface {
	Querier
}

// store provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db: db,
	}
}
