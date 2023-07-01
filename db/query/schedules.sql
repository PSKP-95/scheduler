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
  $1, $2, $3, $4, $5, $6, now()
) RETURNING *;


-- name: GetSchedule :one
SELECT * FROM schedules
WHERE id = $1 LIMIT 1;

-- name: ListSchedules :many
SELECT *, COUNT(*) OVER () AS total_records FROM schedules
WHERE owner = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: DeleteSchedule :exec
DELETE FROM schedules
WHERE id = $1;

-- name: UpdateAccount :one
UPDATE schedules 
SET cron = $2, hook = $3, active = $4, till = $5, last_modified = now()
WHERE id = $1 RETURNING *;
