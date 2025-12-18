package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"auth/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Issuer      string
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	ResetOTPTTL time.Duration
}

type Service struct {
	users   domain.UserRepository
	refresh domain.RefreshRepository
	resets  domain.PasswordResetRepository
	cfg     Config
}

func New(users domain.UserRepository, refresh domain.RefreshRepository, resets domain.PasswordResetRepository, cfg Config) *Service {
	return &Service{users: users, refresh: refresh, resets: resets, cfg: cfg}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) Register(ctx context.Context, email, password string) (TokenPair, error) {
	email = normEmail(email)
	if err := validatePassword(password); err != nil {
		return TokenPair{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return TokenPair{}, err
	}
	u, err := s.users.Create(ctx, email, string(hash))
	if err != nil {
		return TokenPair{}, err
	}
	return s.issuePair(ctx, u.ID)
}

func (s *Service) Login(ctx context.Context, email, password string) (TokenPair, error) {
	email = normEmail(email)
	u, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return TokenPair{}, domain.ErrInvalidCreds
	}
	if u.IsBlocked {
		return TokenPair{}, domain.ErrBlockedUser
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return TokenPair{}, domain.ErrInvalidCreds
	}
	_ = s.users.SetLastLogin(ctx, u.ID, time.Now().UTC())
	return s.issuePair(ctx, u.ID)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	if refreshToken == "" {
		return TokenPair{}, domain.ErrInvalidRefresh
	}
	rs, err := s.refresh.ByHashActive(ctx, sha256sum(refreshToken), time.Now())
	if err != nil {
		return TokenPair{}, domain.ErrInvalidRefresh
	}
	u, err := s.users.ByID(ctx, rs.UserID)
	if err != nil || u.IsBlocked {
		_ = s.refresh.RevokeAllForUser(ctx, rs.UserID)
		if err == nil && u.IsBlocked {
			return TokenPair{}, domain.ErrBlockedUser
		}
		return TokenPair{}, domain.ErrInvalidRefresh
	}
	_ = s.refresh.RevokeByID(ctx, rs.ID) // rotation
	return s.issuePair(ctx, rs.UserID)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	rs, err := s.refresh.ByHashActive(ctx, sha256sum(refreshToken), time.Now())
	if err != nil {
		return nil
	}
	return s.refresh.RevokeByID(ctx, rs.ID)
}

func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.refresh.RevokeAllForUser(ctx, userID)
}

func (s *Service) StartPasswordResetOTP(ctx context.Context, email string) (string, error) {
	email = normEmail(email)
	u, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return "", nil
	}
	code, err := randomString(6)
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	_, err = s.resets.CreateOTP(ctx, u.ID, string(hash), time.Now().Add(s.cfg.ResetOTPTTL))
	if err != nil {
		return "", err
	}
	return code, nil
}

func (s *Service) ConfirmPasswordResetOTP(ctx context.Context, email, code, newPassword string) error {
	email = normEmail(email)
	u, err := s.users.ByEmail(ctx, email)
	if err != nil {
		return domain.ErrInvalidOTP
	}
	pr, err := s.resets.FindValidByUser(ctx, u.ID, time.Now())
	if err != nil {
		return domain.ErrInvalidOTP
	}
	if bcrypt.CompareHashAndPassword([]byte(pr.OTPHash), []byte(code)) != nil {
		return domain.ErrInvalidOTP
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	newHash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err := s.users.UpdatePassword(ctx, u.ID, string(newHash)); err != nil {
		return err
	}
	_ = s.resets.MarkUsed(ctx, pr.ID)
	_ = s.refresh.RevokeAllForUser(ctx, u.ID)
	return nil
}

func (s *Service) ValidateAccess(ctx context.Context, access string) (uuid.UUID, error) {
	claims, err := s.parseAccess(access)
	if err != nil {
		return uuid.Nil, err
	}

	subStr, _ := claims["sub"].(string)
	id, err := uuid.Parse(subStr)
	if err != nil {
		return uuid.Nil, err
	}

	u, err := s.users.ByID(ctx, id)
	if err != nil {
		return uuid.Nil, domain.ErrInvalidCreds
	}
	if u.IsBlocked {
		return uuid.Nil, domain.ErrInvalidCreds
	}

	return u.ID, nil
}

func (s *Service) issuePair(ctx context.Context, userID uuid.UUID) (TokenPair, error) {
	rawRefresh, err := randomString(32)
	if err != nil {
		return TokenPair{}, err
	}
	_, err = s.refresh.Create(ctx, userID, sha256sum(rawRefresh), nil, nil, time.Now().Add(s.cfg.RefreshTTL))
	if err != nil {
		return TokenPair{}, err
	}
	access, err := s.signAccess(userID)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: rawRefresh}, nil
}

func (s *Service) signAccess(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"typ": "access",
		"iss": s.cfg.Issuer,
		"iat": now.Unix(),
		"exp": now.Add(s.cfg.AccessTTL).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) parseAccess(token string) (jwt.MapClaims, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("bad alg")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !t.Valid {
		return nil, domain.ErrInvalidCreds
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || claims["typ"] != "access" {
		return nil, domain.ErrInvalidCreds
	}
	return claims, nil
}

func normEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validatePassword(pw string) error {
	if len(pw) < 8 || len(pw) > 72 {
		return errors.New("password length must be 8..72")
	}
	return nil
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b), err
}

func sha256sum(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}
