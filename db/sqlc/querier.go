// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Querier interface {
	AssignUnassignedWork(ctx context.Context, arg AssignUnassignedWorkParams) error
	CreateHistory(ctx context.Context, arg CreateHistoryParams) (History, error)
	CreateOccurence(ctx context.Context, arg CreateOccurenceParams) (NextOccurence, error)
	CreateSchedule(ctx context.Context, arg CreateScheduleParams) (Schedule, error)
	CreateWorker(ctx context.Context, id uuid.UUID) (PunchCard, error)
	DeleteOccurence(ctx context.Context, id int32) error
	DeleteSchedule(ctx context.Context, id uuid.UUID) error
	DeleteWorker(ctx context.Context, id uuid.UUID) error
	GetNextImmediateWork(ctx context.Context, worker uuid.NullUUID) (time.Time, error)
	GetOccurence(ctx context.Context, id int32) (NextOccurence, error)
	GetSchedule(ctx context.Context, id uuid.UUID) (Schedule, error)
	GetWorker(ctx context.Context, id uuid.UUID) (PunchCard, error)
	ListHistory(ctx context.Context, arg ListHistoryParams) ([]ListHistoryRow, error)
	ListSchedules(ctx context.Context, arg ListSchedulesParams) ([]ListSchedulesRow, error)
	ListWorkers(ctx context.Context, arg ListWorkersParams) ([]PunchCard, error)
	MyExpiredWork(ctx context.Context, worker uuid.NullUUID) ([]NextOccurence, error)
	ProveLiveliness(ctx context.Context, id uuid.UUID) error
	RemoveDeadWorkers(ctx context.Context) error
	UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Schedule, error)
	UpdateStatusAndDetails(ctx context.Context, arg UpdateStatusAndDetailsParams) (History, error)
}

var _ Querier = (*Queries)(nil)
