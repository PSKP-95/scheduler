package worker

import (
	"context"

	"github.com/PSKP-95/scheduler/config"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/google/uuid"
)

type Worker struct {
	id         uuid.UUID
	config     config.WorkerConfig
	store      db.Store
	executor   *hooks.Executor
	killSwitch chan struct{}
}

func NewWorker(config config.WorkerConfig, store db.Store, executor *hooks.Executor, killSwitch chan struct{}) (*Worker, error) {
	worker := &Worker{
		id:         uuid.New(),
		config:     config,
		store:      store,
		executor:   executor,
		killSwitch: killSwitch,
	}

	return worker, nil
}

func (worker *Worker) GetWorkerId() uuid.UUID {
	return worker.id
}

func (worker *Worker) Register() error {
	_, err := worker.store.CreateWorker(context.Background(), worker.id)

	if err != nil {
		return err
	}

	return nil
}
