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
