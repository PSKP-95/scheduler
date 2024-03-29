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

-- name: AssignUnassignedWork :exec
UPDATE next_occurence
SET worker = $1, status = 'pending'
WHERE occurence < (CURRENT_TIMESTAMP + $2 * INTERVAL '1 second') and worker IS NULL;

-- name: MyExpiredWork :many
SELECT * FROM next_occurence
WHERE occurence < CURRENT_TIMESTAMP and worker = $1 and status = 'pending'
ORDER BY occurence;

-- name: GetNextImmediateWork :one
SELECT MIN(occurence)::timestamptz FROM next_occurence
WHERE occurence > CURRENT_TIMESTAMP and worker = $1;

-- name: ChangeOccurenceStatus :exec
UPDATE next_occurence
SET status = $1
WHERE id = $2;
