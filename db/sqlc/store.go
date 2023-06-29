package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Querier
	ScheduleOpPeriodic(ctx context.Context, scheduleOpsParams ScheduleOpParams) (ScheduleOpResult, error)
	ScheduleOpAlarm(ctx context.Context, scheduleOpsParams ScheduleOpParams) (ScheduleOpResult, error)
}

// store provides all functions to execute db queries and transactions
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

type ScheduleOpParams struct {
	Wid          uuid.UUID
	LookAheadSec string
}

type ScheduleOpResult struct {
	ExpiredNextOccurences []NextOccurence
	NextWork              time.Time
}

func (store *SQLStore) ScheduleOpPeriodic(ctx context.Context, scheduleOpsParams ScheduleOpParams) (ScheduleOpResult, error) {
	var result ScheduleOpResult
	err := store.execTx(ctx, func(q *Queries) error {
		params := AssignUnassignedWorkParams{
			Worker: uuid.NullUUID{
				UUID:  scheduleOpsParams.Wid,
				Valid: true,
			},
			Column2: scheduleOpsParams.LookAheadSec,
		}

		err := store.AssignUnassignedWork(context.Background(), params)
		if err != nil {
			return err
		}

		newWork, err := store.MyExpiredWork(context.Background(), uuid.NullUUID{
			UUID:  scheduleOpsParams.Wid,
			Valid: true,
		})
		if err != nil {
			return err
		}
		result.ExpiredNextOccurences = append(result.ExpiredNextOccurences, newWork...)

		result.NextWork, err = store.GetNextImmediateWork(context.Background(), uuid.NullUUID{
			UUID:  scheduleOpsParams.Wid,
			Valid: true,
		})
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}

func (store *SQLStore) ScheduleOpAlarm(ctx context.Context, scheduleOpsParams ScheduleOpParams) (ScheduleOpResult, error) {
	var result ScheduleOpResult
	err := store.execTx(ctx, func(q *Queries) error {
		newWork, err := store.MyExpiredWork(context.Background(), uuid.NullUUID{
			UUID:  scheduleOpsParams.Wid,
			Valid: true,
		})
		if err != nil {
			return err
		}
		result.ExpiredNextOccurences = append(result.ExpiredNextOccurences, newWork...)

		result.NextWork, err = store.GetNextImmediateWork(context.Background(), uuid.NullUUID{
			UUID:  scheduleOpsParams.Wid,
			Valid: true,
		})
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}
