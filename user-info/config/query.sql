-- name: CreateUserInfo :exec
INSERT INTO user_info (weight, height, date, age, user_id)
VALUES ($1, $2, $3, $4, $5);

-- name: GetLatestUserInfoByUserID :one
SELECT id, weight, height, date, age, user_id
FROM user_info 
WHERE user_id = $1
ORDER BY date DESC, id DESC
LIMIT 1;

-- name: GetAllUserInfoByUserID :many
SELECT id, weight, height, date, age, user_id
FROM user_info 
WHERE user_id = $1
ORDER BY date DESC, id DESC;