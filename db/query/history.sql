-- name: CreateHistory :one
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
) RETURNING *;

-- name: ListHistory :many
SELECT * FROM history
WHERE schedule = $1
ORDER BY scehduled_at
LIMIT $2
OFFSET $3;

-- name: UpdateStatusAndDetails :one
UPDATE history 
SET details = $2, status = $3, completed_at = now()
WHERE occurence_id = $1 RETURNING *;
