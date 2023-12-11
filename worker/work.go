package worker

import (
	"context"
	"database/sql"
	"time"

	"github.com/PSKP-95/scheduler/cron"
	db "github.com/PSKP-95/scheduler/db/sqlc"
	"github.com/PSKP-95/scheduler/hooks"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (w *Worker) Work() {
	periodicTicker := time.NewTicker(time.Duration(w.config.WorkPollTimeout) * time.Second)
	longPeriodicTicker := time.NewTicker(time.Duration(w.config.WorkPollTimeout) * 6 * time.Second)
	defer periodicTicker.Stop()
	defer longPeriodicTicker.Stop()

	alarmTicker := time.NewTicker(15 * time.Second)

loop:
	for {
		select {
		case <-periodicTicker.C:
			alarmTicker.Stop()
			w.removeDeadBodies()
			w.punchCard()
			w.getNewWork()
			w.poll(alarmTicker)

		case <-longPeriodicTicker.C:
			w.createOccurrenceForValidSchedules()

		case <-alarmTicker.C:
			alarmTicker.Stop()
			w.poll(alarmTicker)

		case <-w.killSwitch:
			// perform suicide
			err := w.store.DeleteWorker(context.Background(), w.id)
			if err != nil {
				log.Error().Err(err).Msg("Failed to perform suicide. exiting without suicide.")
			}

			break loop
		}
	}
	log.Info().Msg("Graceful shutdown of w.")
	w.killSwitch <- struct{}{}
}

func (w *Worker) createOccurrenceForValidSchedules() {
	schedules, err := w.store.ValidSchedulesWithoutOccurence(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("")

		return
	}

	for _, schedule := range schedules {
		nextOccurence, err := cron.CalculateNextOccurence(schedule.Cron)
		if err != nil {
			log.Error().Err(err).Msg("")

			continue
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

		_, err = w.store.CreateOccurence(context.Background(), occurenceParams)
		if err != nil {
			log.Error().Err(err).Msg("")
		}
	}
}

func (w *Worker) getNewWork() {
	params := db.AssignUnassignedWorkParams{
		Worker: uuid.NullUUID{
			UUID:  w.id,
			Valid: true,
		},
		Column2: w.config.WorkLookAheadSec,
	}

	err := w.store.AssignUnassignedWork(context.Background(), params)
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}

func (w *Worker) poll(ticker *time.Ticker) {
	expiredOccurence, err := w.store.MyExpiredWork(context.Background(), uuid.NullUUID{
		UUID:  w.id,
		Valid: true,
	})
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	w.submitBulkWorkToExecutor(expiredOccurence)

	nextTime, err := w.store.GetNextImmediateWork(context.Background(), uuid.NullUUID{
		UUID:  w.id,
		Valid: true,
	})
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	until := time.Until(nextTime)
	if until > 0 {
		ticker.Reset(until)
	}
}

func (w *Worker) submitBulkWorkToExecutor(expiredOccurences []db.NextOccurence) {
	for _, v := range expiredOccurences {
		msg := hooks.Message{
			Occurence: v,
			Type:      hooks.SCHEDULED,
		}
		log.Info().Msgf("%v", msg)
		w.executor.Submit(msg)
	}
}

func (w *Worker) removeDeadBodies() {
	err := w.store.RemoveDeadWorkers(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error while removing dead workers.")
	}
}

func (w *Worker) punchCard() {
	err := w.store.ProveLiveliness(context.Background(), w.id)
	if err != nil {
		log.Error().Err(err).Msg("Error while punching card.")
	}
}
