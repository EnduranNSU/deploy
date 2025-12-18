-- ===== users =====
-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, password_hash, is_blocked, last_login_at, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, is_blocked, last_login_at, created_at, updated_at
FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, is_blocked, last_login_at, created_at, updated_at
FROM users WHERE id = $1;

-- name: UpdatePassword :exec
UPDATE users SET password_hash = $2, updated_at = now()
WHERE id = $1;

-- name: SetLastLogin :exec
UPDATE users SET last_login_at = $2, updated_at = now()
WHERE id = $1;

-- name: SetBlocked :exec
UPDATE users SET is_blocked = $2, updated_at = now()
WHERE id = $1;

-- ===== refresh_sessions =====
-- name: CreateRefreshSession :one
INSERT INTO refresh_sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token_hash, user_agent, ip, created_at, expires_at, revoked_at;

-- name: GetRefreshByHashActive :one
SELECT id, user_id, token_hash, user_agent, ip, created_at, expires_at, revoked_at
FROM refresh_sessions
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND now() < expires_at;

-- name: RevokeRefreshByID :exec
UPDATE refresh_sessions
SET revoked_at = now()
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeAllRefreshForUser :exec
UPDATE refresh_sessions
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- ===== password_resets (OTP) =====
-- name: CreatePasswordResetOTP :one
INSERT INTO password_resets (user_id, otp_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, otp_hash, expires_at, used_at;

-- name: FindValidPasswordResetByUser :one
SELECT id, user_id, otp_hash, expires_at, used_at
FROM password_resets
WHERE user_id = $1
  AND used_at IS NULL
  AND now() < expires_at
ORDER BY expires_at DESC
LIMIT 1;

-- name: MarkPasswordResetUsed :exec
UPDATE password_resets
SET used_at = now()
WHERE id = $1 AND used_at IS NULL;
