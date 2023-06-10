// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: punch_card.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createWorker = `-- name: CreateWorker :one
INSERT INTO punch_card (
  id,
  last_punch
) VALUES (
  $1, now()
) RETURNING id, last_punch, created_at
`

func (q *Queries) CreateWorker(ctx context.Context, id uuid.UUID) (PunchCard, error) {
	row := q.db.QueryRowContext(ctx, createWorker, id)
	var i PunchCard
	err := row.Scan(&i.ID, &i.LastPunch, &i.CreatedAt)
	return i, err
}

const deleteWorker = `-- name: DeleteWorker :exec
DELETE FROM punch_card
WHERE id = $1
`

func (q *Queries) DeleteWorker(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteWorker, id)
	return err
}

const getWorker = `-- name: GetWorker :one
SELECT id, last_punch, created_at FROM punch_card
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetWorker(ctx context.Context, id uuid.UUID) (PunchCard, error) {
	row := q.db.QueryRowContext(ctx, getWorker, id)
	var i PunchCard
	err := row.Scan(&i.ID, &i.LastPunch, &i.CreatedAt)
	return i, err
}

const listWorkers = `-- name: ListWorkers :many
SELECT id, last_punch, created_at FROM punch_card
ORDER BY id
LIMIT $1
OFFSET $2
`

type ListWorkersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListWorkers(ctx context.Context, arg ListWorkersParams) ([]PunchCard, error) {
	rows, err := q.db.QueryContext(ctx, listWorkers, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PunchCard{}
	for rows.Next() {
		var i PunchCard
		if err := rows.Scan(&i.ID, &i.LastPunch, &i.CreatedAt); err != nil {
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

const proveLiveliness = `-- name: ProveLiveliness :exec
UPDATE punch_card 
SET last_punch = now()
WHERE id = $1
`

func (q *Queries) ProveLiveliness(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, proveLiveliness, id)
	return err
}

const removeDeadWorkers = `-- name: RemoveDeadWorkers :exec
DELETE FROM punch_card
WHERE last_punch < (CURRENT_TIMESTAMP - INTERVAL '30' SECOND)
`

func (q *Queries) RemoveDeadWorkers(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, removeDeadWorkers)
	return err
}
