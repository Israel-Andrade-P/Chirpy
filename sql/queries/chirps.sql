-- name: CreateChirp :one
INSERT INTO chirps (body, created_at, updated_at, user_id)
VALUES (
    $1, NOW(), NOW(), $2
)
RETURNING *;