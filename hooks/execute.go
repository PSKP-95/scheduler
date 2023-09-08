package hooks

import (
	"context"
	"time"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/util"
)

type Executor struct {
	store  db.Store
	hooks  map[string]Hook
	exChan chan Message
	Logger *util.Log
}

func NewExecutor(store db.Store, exChan chan Message, logger *util.Log) (*Executor, error) {
	ex := &Executor{
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
			schedule, err := ex.store.GetSchedule(context.Background(), msg.Occurence.Schedule)
			if err != nil {
				ex.Logger.ErrorLog.Println(err)
			}

			if !schedule.Active || msg.Occurence.Occurence.Time.After(schedule.Till) {
				_ = ex.store.DeleteOccurence(context.Background(), msg.Occurence.ID)
				continue
			}

			msg.Schedule = schedule

			err = ex.store.UpdateHistoryAndOccurence(context.Background(), msg.Schedule, msg.Occurence)
			if err != nil {
				ex.Logger.ErrorLog.Println(err)
				ex.Logger.ErrorLog.Println(schedule.ID, msg.Occurence.ID)
			}

			go ex.hooks[msg.Schedule.Hook].Perform(msg, ex.exChan)
		case SUCCESS:
			params := db.UpdateHistoryAndDeleteOccurenceParams{
				Schedule:  msg.Schedule,
				Occurence: msg.Occurence,
				Details:   msg.Details,
				Status:    db.StatusSuccess,
			}

			err := ex.store.UpdateHistoryAndDeleteOccurence(context.Background(), params)
			if err != nil {
				ex.Logger.ErrorLog.Println(err)
			}
		case FAILED:
			params := db.UpdateHistoryAndDeleteOccurenceParams{
				Schedule:  msg.Schedule,
				Occurence: msg.Occurence,
				Details:   msg.Details,
				Status:    db.StatusFailure,
			}

			err := ex.store.UpdateHistoryAndDeleteOccurence(context.Background(), params)
			if err != nil {
				ex.Logger.ErrorLog.Println(err)
			}
		}
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
		ex.Logger.ErrorLog.Println(err)
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
