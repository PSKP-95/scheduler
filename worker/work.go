package worker

import (
	"context"
	"fmt"
	"time"

	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/PSKP-95/schedular/hooks"
	"github.com/google/uuid"
)

func (worker *Worker) Work() {
	for range time.Tick(10 * time.Second) {
		fmt.Println("Hi from worker")
		worker.removeDeadBodies()
		worker.punchCard()
		worker.checkForWork()
		worker.doWork()
	}
}

func (worker *Worker) removeDeadBodies() {
	err := worker.store.RemoveDeadWorkers(context.Background())
	worker.Logger.ErrorLog.Println(err)
}

func (worker *Worker) punchCard() {
	err := worker.store.ProveLiveliness(context.Background(), worker.id)
	worker.Logger.ErrorLog.Println(err)
}

func (worker *Worker) checkForWork() {
	params := db.UnassignedWorkInFutureParams{
		Worker: uuid.NullUUID{
			UUID:  worker.id,
			Valid: true,
		},
		Column2: worker.config.WorkLookAheadSec,
	}
	err := worker.store.UnassignedWorkInFuture(context.Background(), params)

	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}
}

func (worker *Worker) doWork() {
	work, err := worker.store.MyExpiredWork(context.Background(), uuid.NullUUID{
		UUID:  worker.GetWorkerId(),
		Valid: true,
	})

	worker.Logger.InfoLog.Println(work)

	if err != nil {
		worker.Logger.ErrorLog.Println(err)
		return
	}

	for _, v := range work {
		msg := hooks.Message{
			Occurence: v,
			Type:      hooks.SCHEDULED,
		}
		worker.Logger.InfoLog.Println(msg)
		worker.executor.Submit(msg)
	}
}
