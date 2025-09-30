-- name: GetJob :one
SELECT * FROM "job" WHERE "id" = $1;