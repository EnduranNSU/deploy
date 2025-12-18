package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"auth/internal/adapter/out/postgres/gen"
	"auth/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

/* ========== агрегатор ========== */

type Repositories struct {
	User    domain.UserRepository
	Refresh domain.RefreshRepository
	Reset   domain.PasswordResetRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	q := gen.New(db)
	return &Repositories{
		User:    &userRepo{q: q},
		Refresh: &refreshRepo{q: q},
		Reset:   &resetRepo{q: q},
	}
}

/* ========== helpers ========== */

func ptrTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		t := nt.Time
		return &t
	}
	return nil
}

func optString(v any) *string {
	switch x := v.(type) {
	case sql.NullString:
		if x.Valid {
			s := x.String
			return &s
		}
	case string:
		if x != "" {
			return &x
		}
	}
	return nil
}

func mapNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func isUniqueViolation(err error) bool {
	var e1 *pq.Error
	if errors.As(err, &e1) && string(e1.Code) == "23505" {
		return true
	}
	var e2 *pgconn.PgError
	return errors.As(err, &e2) && e2.Code == "23505"
}

/* ====================== users ====================== */

type userRepo struct{ q gen.Querier }

func (r *userRepo) Create(ctx context.Context, email, passwordHash string) (domain.User, error) {
	u, err := r.q.CreateUser(ctx, gen.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if isUniqueViolation(err) {
			log.Warn().
				Str("operation", "users.Create").
				Str("email", email).
				Msg("email already exists")
			return domain.User{}, domain.ErrAlreadyExists
		}
		log.Error().
			Err(err).
			Str("operation", "users.Create").
			Str("email", email).
			Msg("failed to create user")
		return domain.User{}, err
	}

	log.Debug().
		Str("operation", "users.Create").
		Str("user_id", u.ID.String()).
		Str("email", u.Email).
		Time("created_at", u.CreatedAt).
		Msg("user created")

	return domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsBlocked:    u.IsBlocked,
		LastLoginAt:  ptrTime(u.LastLoginAt),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}, nil
}

func (r *userRepo) ByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().
				Str("operation", "users.ByEmail").
				Str("email", email).
				Msg("user not found")
		} else {
			log.Error().
				Err(err).
				Str("operation", "users.ByEmail").
				Str("email", email).
				Msg("failed to get user by email")
		}
		return domain.User{}, mapNotFound(err)
	}

	log.Debug().
		Str("operation", "users.ByEmail").
		Str("user_id", u.ID.String()).
		Str("email", u.Email).
		Msg("user fetched")

	return domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsBlocked:    u.IsBlocked,
		LastLoginAt:  ptrTime(u.LastLoginAt),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}, nil
}

func (r *userRepo) ByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().
				Str("operation", "users.ByID").
				Str("user_id", id.String()).
				Msg("user not found")
		} else {
			log.Error().
				Err(err).
				Str("operation", "users.ByID").
				Str("user_id", id.String()).
				Msg("failed to get user by id")
		}
		return domain.User{}, mapNotFound(err)
	}

	log.Debug().
		Str("operation", "users.ByID").
		Str("user_id", u.ID.String()).
		Str("email", u.Email).
		Msg("user fetched")

	return domain.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		IsBlocked:    u.IsBlocked,
		LastLoginAt:  ptrTime(u.LastLoginAt),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}, nil
}

func (r *userRepo) UpdatePassword(ctx context.Context, id uuid.UUID, newHash string) error {
	if err := r.q.UpdatePassword(ctx, gen.UpdatePasswordParams{
		ID:           id,
		PasswordHash: newHash,
	}); err != nil {
		log.Error().
			Err(err).
			Str("operation", "users.UpdatePassword").
			Str("user_id", id.String()).
			Msg("failed to update password")
		return err
	}

	log.Debug().
		Str("operation", "users.UpdatePassword").
		Str("user_id", id.String()).
		Msg("password updated")
	return nil
}

func (r *userRepo) SetLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	if err := r.q.SetLastLogin(ctx, gen.SetLastLoginParams{
		ID:          id,
		LastLoginAt: sql.NullTime{Time: t, Valid: true},
	}); err != nil {
		log.Error().
			Err(err).
			Str("operation", "users.SetLastLogin").
			Str("user_id", id.String()).
			Msg("failed to set last_login_at")
		return err
	}

	log.Debug().
		Str("operation", "users.SetLastLogin").
		Str("user_id", id.String()).
		Time("last_login_at", t).
		Msg("last_login_at set")
	return nil
}

func (r *userRepo) SetBlocked(ctx context.Context, id uuid.UUID, blocked bool) error {
	if err := r.q.SetBlocked(ctx, gen.SetBlockedParams{
		ID:        id,
		IsBlocked: blocked,
	}); err != nil {
		log.Error().
			Err(err).
			Str("operation", "users.SetBlocked").
			Str("user_id", id.String()).
			Bool("blocked", blocked).
			Msg("failed to set is_blocked")
		return err
	}

	log.Debug().
		Str("operation", "users.SetBlocked").
		Str("user_id", id.String()).
		Bool("blocked", blocked).
		Msg("is_blocked set")
	return nil
}

/* ================== refresh_sessions ================== */

type refreshRepo struct{ q gen.Querier }

func (r *refreshRepo) Create(ctx context.Context, userID uuid.UUID, tokenHash []byte, ua, ip *string, exp time.Time) (domain.RefreshSession, error) {
	rs, err := r.q.CreateRefreshSession(ctx, gen.CreateRefreshSessionParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: exp,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "refresh.Create").
			Str("user_id", userID.String()).
			Time("expires_at", exp).
			Msg("failed to create refresh session")
		return domain.RefreshSession{}, err
	}

	ev := log.Debug().
		Str("operation", "refresh.Create").
		Str("session_id", rs.ID.String()).
		Str("user_id", rs.UserID.String()).
		Time("expires_at", rs.ExpiresAt)
	if uaV := optString(rs.UserAgent); uaV != nil {
		ev = ev.Str("user_agent", *uaV)
	}
	if ipV := optString(rs.Ip); ipV != nil {
		ev = ev.Str("ip", *ipV)
	}
	ev.Msg("refresh session created")

	return domain.RefreshSession{
		ID:        rs.ID,
		UserID:    rs.UserID,
		TokenHash: rs.TokenHash,
		UserAgent: optString(rs.UserAgent),
		IP:        optString(rs.Ip),
		CreatedAt: rs.CreatedAt,
		ExpiresAt: rs.ExpiresAt,
		RevokedAt: ptrTime(rs.RevokedAt),
	}, nil
}

func (r *refreshRepo) ByHashActive(ctx context.Context, tokenHash []byte, _ time.Time) (domain.RefreshSession, error) {
	rs, err := r.q.GetRefreshByHashActive(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().
				Str("operation", "refresh.ByHashActive").
				Msg("active refresh not found")
		} else {
			log.Error().
				Err(err).
				Str("operation", "refresh.ByHashActive").
				Msg("failed to get refresh by hash")
		}
		return domain.RefreshSession{}, mapNotFound(err)
	}

	ev := log.Debug().
		Str("operation", "refresh.ByHashActive").
		Str("session_id", rs.ID.String()).
		Str("user_id", rs.UserID.String()).
		Time("expires_at", rs.ExpiresAt)
	if uaV := optString(rs.UserAgent); uaV != nil {
		ev = ev.Str("user_agent", *uaV)
	}
	if ipV := optString(rs.Ip); ipV != nil {
		ev = ev.Str("ip", *ipV)
	}
	ev.Msg("refresh session fetched")

	return domain.RefreshSession{
		ID:        rs.ID,
		UserID:    rs.UserID,
		TokenHash: rs.TokenHash,
		UserAgent: optString(rs.UserAgent),
		IP:        optString(rs.Ip),
		CreatedAt: rs.CreatedAt,
		ExpiresAt: rs.ExpiresAt,
		RevokedAt: ptrTime(rs.RevokedAt),
	}, nil
}

func (r *refreshRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	if err := r.q.RevokeRefreshByID(ctx, id); err != nil {
		log.Error().
			Err(err).
			Str("operation", "refresh.RevokeByID").
			Str("session_id", id.String()).
			Msg("failed to revoke refresh session")
		return err
	}

	log.Debug().
		Str("operation", "refresh.RevokeByID").
		Str("session_id", id.String()).
		Msg("refresh session revoked")
	return nil
}

func (r *refreshRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.RevokeAllRefreshForUser(ctx, userID); err != nil {
		log.Error().
			Err(err).
			Str("operation", "refresh.RevokeAllForUser").
			Str("user_id", userID.String()).
			Msg("failed to revoke all refresh sessions for user")
		return err
	}

	log.Debug().
		Str("operation", "refresh.RevokeAllForUser").
		Str("user_id", userID.String()).
		Msg("all refresh sessions revoked for user")
	return nil
}

/* ================= password_resets (OTP) ================= */

type resetRepo struct{ q gen.Querier }

func (r *resetRepo) CreateOTP(ctx context.Context, userID uuid.UUID, otpHash string, exp time.Time) (domain.PasswordReset, error) {
	pr, err := r.q.CreatePasswordResetOTP(ctx, gen.CreatePasswordResetOTPParams{
		UserID:    userID,
		OtpHash:   otpHash,
		ExpiresAt: exp,
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("operation", "reset.CreateOTP").
			Str("user_id", userID.String()).
			Time("expires_at", exp).
			Msg("failed to create password reset")
		return domain.PasswordReset{}, err
	}

	log.Debug().
		Str("operation", "reset.CreateOTP").
		Str("reset_id", pr.ID.String()).
		Str("user_id", pr.UserID.String()).
		Time("expires_at", pr.ExpiresAt).
		Msg("password reset created")

	return domain.PasswordReset{
		ID:        pr.ID,
		UserID:    pr.UserID,
		OTPHash:   pr.OtpHash,
		ExpiresAt: pr.ExpiresAt,
		UsedAt:    ptrTime(pr.UsedAt),
	}, nil
}

func (r *resetRepo) FindValidByUser(ctx context.Context, userID uuid.UUID, _ time.Time) (domain.PasswordReset, error) {
	pr, err := r.q.FindValidPasswordResetByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Warn().
				Str("operation", "reset.FindValidByUser").
				Str("user_id", userID.String()).
				Msg("valid reset not found")
		} else {
			log.Error().
				Err(err).
				Str("operation", "reset.FindValidByUser").
				Str("user_id", userID.String()).
				Msg("failed to find valid password reset")
		}
		return domain.PasswordReset{}, mapNotFound(err)
	}

	log.Debug().
		Str("operation", "reset.FindValidByUser").
		Str("reset_id", pr.ID.String()).
		Str("user_id", pr.UserID.String()).
		Time("expires_at", pr.ExpiresAt).
		Msg("valid password reset fetched")

	return domain.PasswordReset{
		ID:        pr.ID,
		UserID:    pr.UserID,
		OTPHash:   pr.OtpHash,
		ExpiresAt: pr.ExpiresAt,
		UsedAt:    ptrTime(pr.UsedAt),
	}, nil
}

func (r *resetRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	if err := r.q.MarkPasswordResetUsed(ctx, id); err != nil {
		log.Error().
			Err(err).
			Str("operation", "reset.MarkUsed").
			Str("reset_id", id.String()).
			Msg("failed to mark reset as used")
		return err
	}

	log.Debug().
		Str("operation", "reset.MarkUsed").
		Str("reset_id", id.String()).
		Msg("password reset marked as used")
	return nil
}
