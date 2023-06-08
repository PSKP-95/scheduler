-- name: CreateHistory :one
INSERT INTO history (
  schedule,
  status,
  details,
  scehduled_at,
  started_at,
  completed_at
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListHistory :many
SELECT * FROM history
WHERE schedule = $1
ORDER BY scehduled_at
LIMIT $2
OFFSET $3;
