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
