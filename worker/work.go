package worker

import (
	"context"
	"fmt"
	"time"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/google/uuid"
)

func (worker *Worker) Work() {
	periodicTicker := time.NewTicker(time.Duration(worker.config.WorkPollTimeout) * time.Second)
	defer periodicTicker.Stop()

	alarmTicker := time.NewTicker(15 * time.Second)

loop:
	for {
		select {
		case <-periodicTicker.C:
			alarmTicker.Stop()
			worker.removeDeadBodies()
			worker.punchCard()
			worker.getNewWork()
			worker.poll(alarmTicker)

		case <-alarmTicker.C:
			alarmTicker.Stop()
			worker.poll(alarmTicker)

		case <-worker.killSwitch:
			// perform suicide
			err := worker.store.DeleteWorker(context.Background(), worker.id)
			if err != nil {
				worker.Logger.ErrorLog.Println("Failed to perform suicide. exiting without suicide.")
			}
			break loop
		}
	}
	fmt.Println("Graceful shutdown of worker.")
	worker.killSwitch <- struct{}{}
}

func (worker *Worker) getNewWork() {
	params := db.AssignUnassignedWorkParams{
		Worker: uuid.NullUUID{
			UUID:  worker.id,
			Valid: true,
		},
		Column2: worker.config.WorkLookAheadSec,
	}

	err := worker.store.AssignUnassignedWork(context.Background(), params)
	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}
}

func (worker *Worker) poll(ticker *time.Ticker) {
	expiredOccurence, err := worker.store.MyExpiredWork(context.Background(), uuid.NullUUID{
		UUID:  worker.id,
		Valid: true,
	})
	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}

	worker.submitBulkWorkToExecutor(expiredOccurence)

	nextTime, err := worker.store.GetNextImmediateWork(context.Background(), uuid.NullUUID{
		UUID:  worker.id,
		Valid: true,
	})
	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}

	until := time.Until(nextTime)
	if until > 0 {
		ticker.Reset(until)
	}
}

func (worker *Worker) submitBulkWorkToExecutor(expiredOccurences []db.NextOccurence) {
	for _, v := range expiredOccurences {
		msg := hooks.Message{
			Occurence: v,
			Type:      hooks.SCHEDULED,
		}
		worker.Logger.InfoLog.Println(msg)
		worker.executor.Submit(msg)
	}
}

func (worker *Worker) removeDeadBodies() {
	err := worker.store.RemoveDeadWorkers(context.Background())
	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}
}

func (worker *Worker) punchCard() {
	err := worker.store.ProveLiveliness(context.Background(), worker.id)
	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}
}
