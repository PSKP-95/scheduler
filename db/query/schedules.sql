-- name: CreateSchedule :one
INSERT INTO schedules (
  id,
  cron,
  hook,
  owner,
  active,
  till,
  last_modified
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;


-- name: GetSchedule :one
SELECT * FROM schedules
WHERE id = $1 LIMIT 1;

-- name: ListSchedules :many
SELECT * FROM schedules
WHERE owner = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: DeleteSchedule :exec
DELETE FROM schedules
WHERE id = $1;
