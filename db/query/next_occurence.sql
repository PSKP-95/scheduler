-- name: CreateOccurence :one
INSERT INTO next_occurence (
  schedule,
  worker,
  manual,
  status,
  occurence,
  last_updated
) VALUES (
  $1, $2, $3, $4, $5, now()
) RETURNING *;

-- name: GetOccurence :one
SELECT * FROM next_occurence
WHERE id = $1 LIMIT 1;

-- name: DeleteOccurence :exec
DELETE FROM next_occurence
WHERE id = $1;

-- name: UnassignedWorkInFuture :exec
UPDATE next_occurence 
SET worker = $1
WHERE occurence < (CURRENT_TIMESTAMP + INTERVAL '300' SECOND) and worker IS NULL;

-- name: MyExpiredWork :many
SELECT * FROM next_occurence
WHERE occurence < CURRENT_TIMESTAMP and worker = $1
ORDER BY occurence;
