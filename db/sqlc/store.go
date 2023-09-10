package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/PSKP-95/scheduler/cron"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Store interface {
	Querier
	UpdateHistoryAndOccurence(ctx context.Context, schedule Schedule, occurence NextOccurence) error
	UpdateHistoryAndDeleteOccurence(ctx context.Context, params UpdateHistoryAndDeleteOccurenceParams) error
	CreateScheduleAddNextOccurence(ctx context.Context, schedule CreateScheduleParams, occurence CreateOccurenceParams) (Schedule, error)
	Close() error
}

// store provides all functions to execute db queries and transactions.
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) Close() error {
	log.Info().Msg("Graceful shutdown of db.")
	return store.db.Close()
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)

	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb er: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit()
}

func (store *SQLStore) CreateScheduleAddNextOccurence(ctx context.Context, scheduleParams CreateScheduleParams, occurenceParams CreateOccurenceParams) (Schedule, error) {
	var sched Schedule

	err := store.execTx(ctx, func(q *Queries) error {
		schedule, err := store.CreateSchedule(ctx, scheduleParams)
		if err != nil {
			return err
		}

		occurenceParams.Schedule = schedule.ID

		_, err = store.CreateOccurence(ctx, occurenceParams)
		if err != nil {
			return err
		}

		sched = schedule

		return nil
	})

	return sched, err
}

func (store *SQLStore) UpdateHistoryAndOccurence(ctx context.Context, schedule Schedule, occurence NextOccurence) error {
	err := store.execTx(ctx, func(q *Queries) error {
		// put occurence in running state
		err := store.ChangeOccurenceStatus(context.Background(), ChangeOccurenceStatusParams{
			Status: StatusRunning,
			ID:     occurence.ID,
		})
		if err != nil {
			return err
		}

		historyParams := getHistoryParam(schedule, occurence)
		_, err = store.CreateHistory(context.Background(), historyParams)
		// if err != nil {
		// 	if pqErr, ok := err.(*pq.Error); ok {
		// 		if pqErr.Code.Name() != "unique_violation" {
		// 			return err
		// 		}
		// 	} else {
		// 		return err
		// 	}
		// }
		switch pqErr := err.(type) {
		case nil:
		case *pq.Error:
			if pqErr.Code.Name() != "unique_violation" {
				return err
			}
		default:
			return err
		}

		nextOccurence, err := cron.CalculateNextOccurence(schedule.Cron)
		if err != nil {
			return err
		}

		occurenceParams := CreateOccurenceParams{
			Schedule: schedule.ID,
			Manual:   false,
			Status:   StatusPending,
			Occurence: sql.NullTime{
				Time:  nextOccurence,
				Valid: true,
			},
		}

		_, err = store.CreateOccurence(context.Background(), occurenceParams)
		switch pqErr := err.(type) {
		case nil:
		case *pq.Error:
			if pqErr.Code.Name() != "unique_violation" {
				return err
			}
		default:
			return err
		}

		return nil
	})

	return err
}

type UpdateHistoryAndDeleteOccurenceParams struct {
	Schedule  Schedule
	Occurence NextOccurence
	Details   string
	Status    Status
}

func (store *SQLStore) UpdateHistoryAndDeleteOccurence(ctx context.Context, params UpdateHistoryAndDeleteOccurenceParams) error {

	err := store.execTx(ctx, func(q *Queries) error {
		_, err := store.UpdateStatusAndDetails(context.Background(),
			UpdateStatusAndDetailsParams{
				OccurenceID: params.Occurence.ID,
				Status:      params.Status,
				Details:     params.Details,
			},
		)
		if err != nil {
			return err
		}

		err = store.DeleteOccurence(context.Background(), params.Occurence.ID)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func getHistoryParam(schedule Schedule, occurence NextOccurence) CreateHistoryParams {
	historyParam := CreateHistoryParams{
		OccurenceID: occurence.ID,
		Schedule:    schedule.ID,
		Status:      StatusRunning,
		Manual:      occurence.Manual,
		Details:     "",
		ScheduledAt: occurence.Occurence.Time,
		StartedAt:   time.Now(),
	}

	return historyParam
}
