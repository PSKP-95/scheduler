package hooks

import (
	"context"
	"fmt"
	"time"

	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/PSKP-95/schedular/util"
)

type Executor struct {
	config util.Config
	store  db.Store
	hooks  map[string]Hook
	exChan chan Message
}

func NewExecutor(config util.Config, store db.Store, exChan chan Message) (*Executor, error) {
	ex := &Executor{
		config: config,
		store:  store,
		hooks:  getHooks(),
		exChan: exChan,
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
			schedule, err := ex.store.GetSchedule(context.Background(), msg.Occurence.Schedule)
			if err != nil {
				fmt.Println("Processing failed")
			}
			msg.Schedule = schedule
			historyParam := getHistoryParam(schedule, msg.Occurence)
			_, err = ex.store.CreateHistory(context.Background(), historyParam)
			if err != nil {
				fmt.Println(err.Error())
			}
			go ex.hooks[schedule.Hook].Perform(msg, ex.exChan)
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
			ex.store.DeleteOccurence(context.Background(), msg.Occurence.ID)
		}
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
