-- name: CreateWorker :one
INSERT INTO punch_card (
  id,
  last_punch
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetWorker :one
SELECT * FROM punch_card
WHERE id = $1 LIMIT 1;

-- name: ListWorkers :many
SELECT * FROM punch_card
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: DeleteWorker :exec
DELETE FROM punch_card
WHERE id = $1;

-- name: DeadWorkers :many
SELECT * FROM punch_card
WHERE last_punch < (CURRENT_TIMESTAMP - INTERVAL $1 SECOND); 
