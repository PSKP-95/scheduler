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

func (worker *Worker) Work() {
	periodicTicker := time.NewTicker(time.Duration(worker.config.WorkPollTimeout) * time.Second)
	longPeriodicTicker := time.NewTicker(time.Duration(worker.config.WorkPollTimeout) * 6 * time.Second)
	defer periodicTicker.Stop()
	defer longPeriodicTicker.Stop()

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

		case <-longPeriodicTicker.C:
			worker.createOccurrenceForValidSchedules()

		case <-alarmTicker.C:
			alarmTicker.Stop()
			worker.poll(alarmTicker)

		case <-worker.killSwitch:
			// perform suicide
			err := worker.store.DeleteWorker(context.Background(), worker.id)
			if err != nil {
				log.Error().Err(err).Msg("Failed to perform suicide. exiting without suicide.")
			}
			break loop
		}
	}
	log.Info().Msg("Graceful shutdown of worker.")
	worker.killSwitch <- struct{}{}
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
		log.Error().Err(err).Msg("")
	}
}

func (worker *Worker) poll(ticker *time.Ticker) {
	expiredOccurence, err := worker.store.MyExpiredWork(context.Background(), uuid.NullUUID{
		UUID:  worker.id,
		Valid: true,
	})
	if err != nil {
		log.Error().Err(err).Msg("")
	}

	worker.submitBulkWorkToExecutor(expiredOccurence)

	nextTime, err := worker.store.GetNextImmediateWork(context.Background(), uuid.NullUUID{
		UUID:  worker.id,
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

func (worker *Worker) submitBulkWorkToExecutor(expiredOccurences []db.NextOccurence) {
	for _, v := range expiredOccurences {
		msg := hooks.Message{
			Occurence: v,
			Type:      hooks.SCHEDULED,
		}
		log.Info().Msgf("%v", msg)
		worker.executor.Submit(msg)
	}
}

func (worker *Worker) removeDeadBodies() {
	err := worker.store.RemoveDeadWorkers(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Error while removing dead workers.")
	}
}

func (worker *Worker) punchCard() {
	err := worker.store.ProveLiveliness(context.Background(), worker.id)
	if err != nil {
		log.Error().Err(err).Msg("Error while punching card.")
	}
}
