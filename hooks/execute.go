package hooks

import (
	"context"
	"sync"
	"time"

	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/rs/zerolog/log"
)

type Executor struct {
	store  db.Store
	hooks  map[string]Hook
	exChan chan Message
	wg     *sync.WaitGroup
}

func NewExecutor(store db.Store, exChan chan Message, wg *sync.WaitGroup) (*Executor, error) {
	ex := &Executor{
		store:  store,
		hooks:  getHooks(),
		exChan: exChan,
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
			log.Info().Msgf("TRIGGER: Occurrence Id: %d, Occurence: %s, Schedule: %s", msg.Occurence.ID, msg.Occurence.Occurence.Time, msg.Occurence.Schedule)
			ex.createHistoryForOccurence(&msg)
			ex.wg.Add(1)
			go ex.hooks[msg.Schedule.Hook].Perform(msg, ex.exChan)
		case SCHEDULED:
			log.Info().Msgf("SCHEDULED: Occurrence Id: %d, Occurence: %s, Schedule: %s", msg.Occurence.ID, msg.Occurence.Occurence.Time, msg.Occurence.Schedule)
			if msg.Occurence.Status == db.StatusRunning {
				log.Info().Msg("Occurence already running.")
				continue
			}

			schedule, err := ex.store.GetSchedule(context.Background(), msg.Occurence.Schedule)
			if err != nil {
				log.Error().Err(err).Msg("Error while getting schedule.")
			}

			if !schedule.Active || msg.Occurence.Occurence.Time.After(schedule.Till) {
				_ = ex.store.DeleteOccurence(context.Background(), msg.Occurence.ID)

				continue
			}

			msg.Schedule = schedule

			err = ex.store.UpdateHistoryAndOccurence(context.Background(), msg.Schedule, msg.Occurence)
			if err != nil {
				log.Info().Err(err).Msgf("Schedule: %v, Occurence: %v", schedule.ID, msg.Occurence.ID)
			}

			ex.wg.Add(1)
			go ex.hooks[msg.Schedule.Hook].Perform(msg, ex.exChan)
		case SUCCESS:
			log.Info().Msgf("SUCCESS: Occurrence Id: %d, Occurence: %s, Schedule: %v", msg.Occurence.ID, msg.Occurence.Occurence.Time, msg.Occurence.Schedule)
			params := db.UpdateHistoryAndDeleteOccurenceParams{
				Schedule:  msg.Schedule,
				Occurence: msg.Occurence,
				Details:   msg.Details,
				Status:    db.StatusSuccess,
			}

			err := ex.store.UpdateHistoryAndDeleteOccurence(context.Background(), params)
			if err != nil {
				log.Error().Err(err).Msg("error while updating history & deleting occurence")
			}
			ex.wg.Done()
		case FAILED:
			log.Info().Msgf("FAILED: Occurrence Id: %d, Occurence: %s, Schedule: %v", msg.Occurence.ID, msg.Occurence.Occurence.Time, msg.Occurence.Schedule)
			params := db.UpdateHistoryAndDeleteOccurenceParams{
				Schedule:  msg.Schedule,
				Occurence: msg.Occurence,
				Details:   msg.Details,
				Status:    db.StatusFailure,
			}

			err := ex.store.UpdateHistoryAndDeleteOccurence(context.Background(), params)
			if err != nil {
				log.Error().Err(err).Msg("error while updating history & deleting occurence")
			}
			ex.wg.Done()
		}
	}
}

func (ex *Executor) createHistoryForOccurence(msg *Message) {
	schedule, err := ex.store.GetSchedule(context.Background(), msg.Occurence.Schedule)
	if err != nil {
		log.Error().Err(err).Msg("Error while getting schedule")
	}

	msg.Schedule = schedule
	historyParam := getHistoryParam(schedule, msg.Occurence)

	_, err = ex.store.CreateHistory(context.Background(), historyParam)
	if err != nil {
		log.Error().Err(err).Msg("Error while creating history.")
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
