package worker

import (
	"context"
	"time"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/google/uuid"
)

func (worker *Worker) Work() {
	periodicTicker := time.NewTicker(10 * time.Second)
	defer periodicTicker.Stop()

	alarmTicker := time.NewTicker(15 * time.Second)

	for {
		select {
		case <-periodicTicker.C:
			worker.removeDeadBodies()
			worker.punchCard()
			worker.checkForWork()
			worker.doWork()
			worker.getImmediateWork(alarmTicker)

		case <-alarmTicker.C:
			worker.doWork()
			worker.getImmediateWork(alarmTicker)
		}
	}
}

func (worker *Worker) removeDeadBodies() {
	err := worker.store.RemoveDeadWorkers(context.Background())
	worker.Logger.ErrorLog.Println(err)
}

func (worker *Worker) getImmediateWork(ticker *time.Ticker) {
	nTime, err := worker.store.GetNextImmediateWork(context.Background(), uuid.NullUUID{
		UUID:  worker.GetWorkerId(),
		Valid: true,
	})

	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}
	until := time.Until(nTime)

	if until > 0 {
		ticker.Reset(until)
	}
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
