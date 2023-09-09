package hooks

import (
	"context"
	"sync"
	"time"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/mlog"
)

type Executor struct {
	store  db.Store
	hooks  map[string]Hook
	exChan chan Message
	Logger *mlog.Log
	wg     *sync.WaitGroup
}

func NewExecutor(store db.Store, exChan chan Message, logger *mlog.Log, wg *sync.WaitGroup) (*Executor, error) {
	ex := &Executor{
		store:  store,
		hooks:  getHooks(),
		exChan: exChan,
		Logger: logger,
		wg:     wg,
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
			ex.Logger.InfoLog.Println("TRIGGER: ", msg.Occurence)
			ex.createHistoryForOccurence(&msg)
			ex.wg.Add(1)
			go ex.hooks[msg.Schedule.Hook].Perform(msg, ex.exChan, ex.Logger)
		case SCHEDULED:
			ex.Logger.InfoLog.Println("SCHEDULED: ", msg.Occurence)
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

			ex.wg.Add(1)
			go ex.hooks[msg.Schedule.Hook].Perform(msg, ex.exChan, ex.Logger)
		case SUCCESS:
			ex.Logger.InfoLog.Println("SUCCESS: ", msg.Occurence)
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
			ex.wg.Done()
		case FAILED:
			ex.Logger.InfoLog.Println("FAILED: ", msg.Occurence)
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
			ex.wg.Done()
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
