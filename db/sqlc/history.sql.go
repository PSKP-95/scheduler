// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: history.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createHistory = `-- name: CreateHistory :one
INSERT INTO history (
  occurence_id,
  schedule,
  status,
  details,
  manual,
  scheduled_at,
  started_at,
  completed_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING occurence_id, schedule, status, details, manual, scheduled_at, started_at, completed_at
`

type CreateHistoryParams struct {
	OccurenceID int32     `json:"occurence_id"`
	Schedule    uuid.UUID `json:"schedule"`
	Status      Status    `json:"status"`
	Details     string    `json:"details"`
	Manual      bool      `json:"manual"`
	ScheduledAt time.Time `json:"scheduled_at"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

func (q *Queries) CreateHistory(ctx context.Context, arg CreateHistoryParams) (History, error) {
	row := q.db.QueryRowContext(ctx, createHistory,
		arg.OccurenceID,
		arg.Schedule,
		arg.Status,
		arg.Details,
		arg.Manual,
		arg.ScheduledAt,
		arg.StartedAt,
		arg.CompletedAt,
	)
	var i History
	err := row.Scan(
		&i.OccurenceID,
		&i.Schedule,
		&i.Status,
		&i.Details,
		&i.Manual,
		&i.ScheduledAt,
		&i.StartedAt,
		&i.CompletedAt,
	)
	return i, err
}

const listHistory = `-- name: ListHistory :many
SELECT occurence_id, schedule, status, details, manual, scheduled_at, started_at, completed_at FROM history
WHERE schedule = $1
ORDER BY scehduled_at
LIMIT $2
OFFSET $3
`

type ListHistoryParams struct {
	Schedule uuid.UUID `json:"schedule"`
	Limit    int32     `json:"limit"`
	Offset   int32     `json:"offset"`
}

func (q *Queries) ListHistory(ctx context.Context, arg ListHistoryParams) ([]History, error) {
	rows, err := q.db.QueryContext(ctx, listHistory, arg.Schedule, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []History{}
	for rows.Next() {
		var i History
		if err := rows.Scan(
			&i.OccurenceID,
			&i.Schedule,
			&i.Status,
			&i.Details,
			&i.Manual,
			&i.ScheduledAt,
			&i.StartedAt,
			&i.CompletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateStatusAndDetails = `-- name: UpdateStatusAndDetails :one
UPDATE history 
SET details = $2, status = $3, completed_at = now()
WHERE occurence_id = $1 RETURNING occurence_id, schedule, status, details, manual, scheduled_at, started_at, completed_at
`

type UpdateStatusAndDetailsParams struct {
	OccurenceID int32  `json:"occurence_id"`
	Details     string `json:"details"`
	Status      Status `json:"status"`
}

func (q *Queries) UpdateStatusAndDetails(ctx context.Context, arg UpdateStatusAndDetailsParams) (History, error) {
	row := q.db.QueryRowContext(ctx, updateStatusAndDetails, arg.OccurenceID, arg.Details, arg.Status)
	var i History
	err := row.Scan(
		&i.OccurenceID,
		&i.Schedule,
		&i.Status,
		&i.Details,
		&i.Manual,
		&i.ScheduledAt,
		&i.StartedAt,
		&i.CompletedAt,
	)
	return i, err
}
