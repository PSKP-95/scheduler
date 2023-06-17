package hooks

import (
	"context"
	"database/sql"
	"time"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/util"
)

type Executor struct {
	config util.Config
	store  db.Store
	hooks  map[string]Hook
	exChan chan Message
	Logger *util.Log
}

func NewExecutor(config util.Config, store db.Store, exChan chan Message, logger *util.Log) (*Executor, error) {
	ex := &Executor{
		config: config,
		store:  store,
		hooks:  getHooks(),
		exChan: exChan,
		Logger: logger,
	}

	return ex, nil
}

func (ex *Executor) GetHooks() map[string]Hook {
	return ex.hooks
}

func (ex *Executor) Submit(msg Message) {
	ex.exChan <- msg
}

func (ex *Executor) Execute() {
	for {
		msg := <-ex.exChan
		switch msg.Type {
		case TRIGGER:
			ex.createHistoryForOccurence(&msg)
			go ex.hooks[msg.Schedule.Hook].Perform(msg, ex.exChan)
		case SCHEDULED:
			ex.createHistoryForOccurence(&msg)
			go ex.hooks[msg.Schedule.Hook].Perform(msg, ex.exChan)
			ex.createNewOccurence(msg.Schedule)
		case SUCCESS:
			ex.store.UpdateStatusAndDetails(context.Background(),
				db.UpdateStatusAndDetailsParams{
					OccurenceID: msg.Occurence.ID,
					Status:      db.StatusSuccess,
					Details:     msg.Details,
				},
			)
			ex.store.DeleteOccurence(context.Background(), msg.Occurence.ID)
		case FAILED:
			ex.store.UpdateStatusAndDetails(context.Background(),
				db.UpdateStatusAndDetailsParams{
					OccurenceID: msg.Occurence.ID,
					Status:      db.StatusFailure,
					Details:     msg.Details,
				},
			)
			err := ex.store.DeleteOccurence(context.Background(), msg.Occurence.ID)
			if err != nil {
				ex.Logger.ErrorLog.Fatalln(err)
			}
		}
	}
}

func (ex *Executor) createNewOccurence(schedule db.Schedule) {
	nextOccurence, err := util.CalculateNextOccurence(schedule.Cron)
	if err != nil {
		ex.Logger.ErrorLog.Fatalln(err)
	}

	occurenceParams := db.CreateOccurenceParams{
		Schedule: schedule.ID,
		Manual:   false,
		Status:   db.StatusPending,
		Occurence: sql.NullTime{
			Time:  nextOccurence,
			Valid: true,
		},
	}

	_, err = ex.store.CreateOccurence(context.Background(), occurenceParams)

	if err != nil {
		ex.Logger.ErrorLog.Fatalln(err)
	}
}

func (ex *Executor) createHistoryForOccurence(msg *Message) {
	schedule, err := ex.store.GetSchedule(context.Background(), msg.Occurence.Schedule)
	if err != nil {
		ex.Logger.ErrorLog.Fatalln(err)
	}
	msg.Schedule = schedule
	historyParam := getHistoryParam(schedule, msg.Occurence)
	_, err = ex.store.CreateHistory(context.Background(), historyParam)
	if err != nil {
		ex.Logger.ErrorLog.Fatalln(err)
	}
}

func getHistoryParam(schedule db.Schedule, occurence db.NextOccurence) db.CreateHistoryParams {
	historyParam := db.CreateHistoryParams{
		OccurenceID: occurence.ID,
		Schedule:    schedule.ID,
		Status:      db.StatusRunning,
		Manual:      occurence.Manual,
		Details:     "",
		ScheduledAt: occurence.Occurence.Time,
		StartedAt:   time.Now(),
	}

	return historyParam
}
