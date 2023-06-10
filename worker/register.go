package worker

import (
	"context"

	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/PSKP-95/schedular/hooks"
	"github.com/PSKP-95/schedular/util"
	"github.com/google/uuid"
)

type Worker struct {
	id       uuid.UUID
	config   util.Config
	store    db.Store
	executor *hooks.Executor
}

func NewWorker(config util.Config, store db.Store, executor *hooks.Executor) (*Worker, error) {
	worker := &Worker{
		id:       uuid.New(),
		config:   config,
		store:    store,
		executor: executor,
	}

	return worker, nil
}

func (worker *Worker) Register() error {
	_, err := worker.store.CreateWorker(context.Background(), worker.id)

	if err != nil {
		return err
	}

	return nil
}
