package services

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/yannick2009/yily/internal/crypto"
	"github.com/yannick2009/yily/internal/domain/model"
	"github.com/yannick2009/yily/internal/repository"
)

const (
	DefaultSessionTTL         = 24 * time.Hour // DefaultSessionTTL is the default maximum session lifetime.
	DefaultSessionIdleTimeout = 2 * time.Hour  // DefaultSessionIdleTimeout is the default maximum session inactivity duration.
	MinPasswordLength         = 12             // MinPasswordLength is the minimum Super Admin password length.
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrWeakPassword       = errors.New("password must be at least 12 characters long")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionRevoked     = errors.New("session revoked")
)

type authUserContextKey struct{}
type authSessionContextKey struct{}

// WithAuthenticatedIdentity stores the authenticated User and Session in the context.
func WithAuthenticatedIdentity(ctx context.Context, user *model.User, session *model.Session) context.Context {
	ctx = context.WithValue(ctx, authUserContextKey{}, user)
	ctx = context.WithValue(ctx, authSessionContextKey{}, session)
	return ctx
}

// AuthenticatedUserFromContext retrieves the authenticated User from the context.
func AuthenticatedUserFromContext(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(authUserContextKey{}).(*model.User)
	return user, ok && user != nil
}

// AuthenticatedSessionFromContext retrieves the authenticated Session from the context.
func AuthenticatedSessionFromContext(ctx context.Context) (*model.Session, bool) {
	session, ok := ctx.Value(authSessionContextKey{}).(*model.Session)
	return session, ok && session != nil
}

// AuthService defines the interface for authentication and session management.
type AuthService interface {
	Setup(ctx context.Context, email, password string) (*model.User, error)                                                     // Setup creates the initial Super Admin account.
	Login(ctx context.Context, email, password, existingRawToken string) (string, *model.Session, *model.User, error)           // Login authenticates the Super Admin and issues a new session.
	ValidateSession(ctx context.Context, rawToken string) (*model.User, *model.Session, error)                                  // ValidateSession verifies a raw session token and returns its identity.
	Logout(ctx context.Context, rawToken string) error                                                                          // Logout revokes a session by its raw token.
}

// AuthConfig holds optional timing parameters for AuthService.
type AuthConfig struct {
	SessionTTL  time.Duration
	IdleTimeout time.Duration
	NowFunc     func() time.Time
}

// authService implements AuthService.
type authService struct {
	userRepository    repository.UserRepository
	sessionRepository repository.SessionRepository
	sessionTTL        time.Duration
	idleTimeout       time.Duration
	nowFunc           func() time.Time

	dummyOnce sync.Once
	dummyHash string
}

// NewAuthService creates a new AuthService with default session settings.
func NewAuthService(userRepository repository.UserRepository, sessionRepository repository.SessionRepository) AuthService {
	return NewAuthServiceWithConfig(userRepository, sessionRepository, AuthConfig{})
}

// NewAuthServiceWithConfig creates a new AuthService with custom timing settings.
func NewAuthServiceWithConfig(userRepository repository.UserRepository, sessionRepository repository.SessionRepository, cfg AuthConfig) AuthService {
	ttl := cfg.SessionTTL
	if ttl <= 0 {
		ttl = DefaultSessionTTL
	}
	idle := cfg.IdleTimeout
	if idle <= 0 {
		idle = DefaultSessionIdleTimeout
	}
	nowFn := cfg.NowFunc
	if nowFn == nil {
		nowFn = time.Now
	}

	return &authService{
		userRepository:    userRepository,
		sessionRepository: sessionRepository,
		sessionTTL:        ttl,
		idleTimeout:       idle,
		nowFunc:           nowFn,
	}
}

// getDummyHash lazily initializes a dummy Argon2id hash to mitigate timing enumeration.
func (s *authService) getDummyHash() string {
	s.dummyOnce.Do(func() {
		h, err := crypto.HashPassword("dummy-password-for-constant-time-mitigation")
		if err == nil {
			s.dummyHash = h
		}
	})
	return s.dummyHash
}

// Setup creates the initial Super Admin account with a lowercase email and Argon2id hash.
func (s *authService) Setup(ctx context.Context, email, password string) (*model.User, error) {
	normalizedEmail, err := normalizeAndValidateEmail(email)
	if err != nil {
		return nil, err
	}

	if len(password) < MinPasswordLength {
		return nil, ErrWeakPassword
	}

	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        normalizedEmail,
		PasswordHash: passwordHash,
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates the Super Admin, rotates any prior session, and stores the new SHA-256 session hash.
func (s *authService) Login(ctx context.Context, email, password, existingRawToken string) (string, *model.Session, *model.User, error) {
	normalizedEmail := normalizeEmail(email)
	if normalizedEmail == "" || password == "" {
		return "", nil, nil, ErrInvalidCredentials
	}

	user, err := s.userRepository.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			if dummy := s.getDummyHash(); dummy != "" {
				_, _ = crypto.VerifyPassword(password, dummy)
			}
			return "", nil, nil, ErrInvalidCredentials
		}
		return "", nil, nil, err
	}

	valid, err := crypto.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return "", nil, nil, ErrInvalidCredentials
	}

	now := s.nowFunc().UTC()

	if trimmedToken := strings.TrimSpace(existingRawToken); trimmedToken != "" {
		oldHash := crypto.HashToken(trimmedToken)
		_ = s.sessionRepository.RevokeByTokenHash(ctx, oldHash, now)
	}

	rawToken, tokenHash, err := crypto.GenerateSessionToken()
	if err != nil {
		return "", nil, nil, err
	}

	session := &model.Session{
		UserID:         user.ID,
		TokenHash:      tokenHash,
		CreatedAt:      now,
		LastActivityAt: now,
		ExpiresAt:      now.Add(s.sessionTTL),
	}

	if err := s.sessionRepository.Create(ctx, session); err != nil {
		return "", nil, nil, err
	}

	return rawToken, session, user, nil
}

// ValidateSession verifies a raw token's SHA-256 hash, checks revocation and expiration, and refreshes activity.
func (s *authService) ValidateSession(ctx context.Context, rawToken string) (*model.User, *model.Session, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return nil, nil, ErrUnauthorized
	}

	tokenHash := crypto.HashToken(rawToken)
	session, err := s.sessionRepository.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, nil, ErrUnauthorized
		}
		return nil, nil, err
	}

	if session.IsRevoked() {
		return nil, nil, ErrSessionRevoked
	}

	now := s.nowFunc().UTC()
	if session.IsExpired(now, s.idleTimeout) {
		return nil, nil, ErrSessionExpired
	}

	user, err := s.userRepository.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrUnauthorized
		}
		return nil, nil, err
	}

	session.LastActivityAt = now
	_ = s.sessionRepository.TouchActivity(ctx, session.ID, now)

	return user, session, nil
}

// Logout revokes the session matching the SHA-256 hash of the raw token.
func (s *authService) Logout(ctx context.Context, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return ErrUnauthorized
	}

	now := s.nowFunc().UTC()
	tokenHash := crypto.HashToken(rawToken)

	if err := s.sessionRepository.RevokeByTokenHash(ctx, tokenHash, now); err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return ErrUnauthorized
		}
		return err
	}
	return nil
}

// normalizeEmail trims whitespace and lowercases an email address.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// normalizeAndValidateEmail normalizes an email address to lowercase and validates its format.
func normalizeAndValidateEmail(email string) (string, error) {
	normalized := normalizeEmail(email)
	if normalized == "" {
		return "", ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized || !strings.Contains(normalized, "@") {
		return "", ErrInvalidEmail
	}

	parts := strings.Split(normalized, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", ErrInvalidEmail
	}

	return normalized, nil
}
