package worker

import (
	"context"
	"time"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
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
			worker.periodicPoll(alarmTicker)

		case <-alarmTicker.C:
			worker.alarmPoll(alarmTicker)
		}
	}
}

func (worker *Worker) periodicPoll(ticker *time.Ticker) {
	scheduleOpParams := db.ScheduleOpParams{
		Wid:          worker.id,
		LookAheadSec: worker.config.WorkLookAheadSec,
	}
	scheduleOpResult, err := worker.store.ScheduleOpPeriodic(context.Background(), scheduleOpParams)
	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}

	until := time.Until(scheduleOpResult.NextWork)
	if until > 0 {
		ticker.Reset(until)
	}

	for _, v := range scheduleOpResult.ExpiredNextOccurences {
		msg := hooks.Message{
			Occurence: v,
			Type:      hooks.SCHEDULED,
		}
		worker.Logger.InfoLog.Println(msg)
		worker.executor.Submit(msg)
	}
}

func (worker *Worker) alarmPoll(ticker *time.Ticker) {
	scheduleOpParams := db.ScheduleOpParams{
		Wid:          worker.id,
		LookAheadSec: worker.config.WorkLookAheadSec,
	}
	scheduleOpResult, err := worker.store.ScheduleOpAlarm(context.Background(), scheduleOpParams)
	if err != nil {
		worker.Logger.ErrorLog.Println(err)
	}

	until := time.Until(scheduleOpResult.NextWork)
	if until > 0 {
		ticker.Reset(until)
	}

	for _, v := range scheduleOpResult.ExpiredNextOccurences {
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
	worker.Logger.ErrorLog.Println(err)
}

func (worker *Worker) punchCard() {
	err := worker.store.ProveLiveliness(context.Background(), worker.id)
	worker.Logger.ErrorLog.Println(err)
}
