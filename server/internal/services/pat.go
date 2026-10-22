package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yannick2009/yily/internal/crypto"
	"github.com/yannick2009/yily/internal/domain/model"
	"github.com/yannick2009/yily/internal/repository"
)

const (
	DefaultPATTTL = 30 * 24 * time.Hour // DefaultPATTTL is the default Personal Access Token lifetime.
)

var (
	ErrPATExpired           = errors.New("personal access token expired")
	ErrPATRevoked           = errors.New("personal access token revoked")
	ErrInvalidPATName       = errors.New("invalid personal access token name")
	ErrInvalidPATExpiration = errors.New("invalid personal access token expiration")
)

type authPATContextKey struct{}

// WithAuthenticatedPAT stores the authenticated User and PersonalAccessToken in the context.
func WithAuthenticatedPAT(ctx context.Context, user *model.User, pat *model.PersonalAccessToken) context.Context {
	ctx = context.WithValue(ctx, authUserContextKey{}, user)
	ctx = context.WithValue(ctx, authPATContextKey{}, pat)
	return ctx
}

// AuthenticatedPATFromContext retrieves the authenticated PersonalAccessToken from the context.
func AuthenticatedPATFromContext(ctx context.Context) (*model.PersonalAccessToken, bool) {
	pat, ok := ctx.Value(authPATContextKey{}).(*model.PersonalAccessToken)
	return pat, ok && pat != nil
}

// PATService defines the interface for Personal Access Token management and validation.
type PATService interface {
	Create(ctx context.Context, userID uuid.UUID, name string, ttl time.Duration) (string, *model.PersonalAccessToken, error) // Create generates and persists a new PAT for the user.
	Validate(ctx context.Context, rawToken string) (*model.User, *model.PersonalAccessToken, error)                           // Validate verifies a raw PAT and returns its owner and metadata.
	GetAll(ctx context.Context, userID uuid.UUID) ([]model.PersonalAccessToken, error)                                        // GetAll lists all PATs belonging to a user.
	GetByID(ctx context.Context, userID uuid.UUID, patID uuid.UUID) (*model.PersonalAccessToken, error)                       // GetByID retrieves a single PAT belonging to a user.
	Revoke(ctx context.Context, userID uuid.UUID, patID uuid.UUID) error                                                      // Revoke marks a user's PAT as revoked by its ID.
	RevokeByToken(ctx context.Context, rawToken string) error                                                                 // RevokeByToken marks a PAT as revoked by its raw token.
}

// PATConfig holds optional timing parameters for PATService.
type PATConfig struct {
	DefaultTTL time.Duration
	NowFunc    func() time.Time
}

// patService implements PATService.
type patService struct {
	userRepository repository.UserRepository
	patRepository  repository.PATRepository
	defaultTTL     time.Duration
	nowFunc        func() time.Time
}

// NewPATService creates a new PATService with default settings.
func NewPATService(userRepository repository.UserRepository, patRepository repository.PATRepository) PATService {
	return NewPATServiceWithConfig(userRepository, patRepository, PATConfig{})
}

// NewPATServiceWithConfig creates a new PATService with custom timing settings.
func NewPATServiceWithConfig(userRepository repository.UserRepository, patRepository repository.PATRepository, cfg PATConfig) PATService {
	ttl := cfg.DefaultTTL
	if ttl <= 0 {
		ttl = DefaultPATTTL
	}
	nowFn := cfg.NowFunc
	if nowFn == nil {
		nowFn = time.Now
	}

	return &patService{
		userRepository: userRepository,
		patRepository:  patRepository,
		defaultTTL:     ttl,
		nowFunc:        nowFn,
	}
}

// Create generates a 256-bit PAT, stores only its SHA-256 hash, and returns the raw token once.
func (s *patService) Create(ctx context.Context, userID uuid.UUID, name string, ttl time.Duration) (string, *model.PersonalAccessToken, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return "", nil, ErrInvalidPATName
	}

	if ttl < 0 {
		return "", nil, ErrInvalidPATExpiration
	}
	if ttl == 0 {
		ttl = s.defaultTTL
	}

	if _, err := s.userRepository.GetByID(ctx, userID); err != nil {
		return "", nil, err
	}

	rawToken, tokenHash, err := crypto.GeneratePAT()
	if err != nil {
		return "", nil, err
	}

	now := s.nowFunc().UTC()
	pat := &model.PersonalAccessToken{
		UserID:    userID,
		Name:      trimmedName,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	if err := s.patRepository.Create(ctx, pat); err != nil {
		return "", nil, err
	}

	return rawToken, pat, nil
}

// Validate checks a raw PAT's format, SHA-256 hash, revocation, and expiration.
func (s *patService) Validate(ctx context.Context, rawToken string) (*model.User, *model.PersonalAccessToken, error) {
	rawToken = strings.TrimSpace(rawToken)
	if !strings.HasPrefix(rawToken, crypto.PATPrefix) || len(rawToken) <= len(crypto.PATPrefix) {
		return nil, nil, ErrUnauthorized
	}

	tokenHash := crypto.HashToken(rawToken)
	pat, err := s.patRepository.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrPATNotFound) {
			return nil, nil, ErrUnauthorized
		}
		return nil, nil, err
	}

	if pat.IsRevoked() {
		return nil, nil, ErrPATRevoked
	}

	now := s.nowFunc().UTC()
	if pat.IsExpired(now) {
		return nil, nil, ErrPATExpired
	}

	user, err := s.userRepository.GetByID(ctx, pat.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, ErrUnauthorized
		}
		return nil, nil, err
	}

	return user, pat, nil
}

// GetAll retrieves all PATs belonging to a user without exposing raw tokens or hashes.
func (s *patService) GetAll(ctx context.Context, userID uuid.UUID) ([]model.PersonalAccessToken, error) {
	return s.patRepository.GetAllByUserID(ctx, userID)
}

// GetByID retrieves a single PAT belonging to a user by ID.
func (s *patService) GetByID(ctx context.Context, userID uuid.UUID, patID uuid.UUID) (*model.PersonalAccessToken, error) {
	return s.patRepository.GetByID(ctx, patID, userID)
}

// Revoke marks a user's PAT as revoked by its ID.
func (s *patService) Revoke(ctx context.Context, userID uuid.UUID, patID uuid.UUID) error {
	now := s.nowFunc().UTC()
	return s.patRepository.RevokeByID(ctx, patID, userID, now)
}

// RevokeByToken marks a PAT as revoked by its raw token.
func (s *patService) RevokeByToken(ctx context.Context, rawToken string) error {
	rawToken = strings.TrimSpace(rawToken)
	if !strings.HasPrefix(rawToken, crypto.PATPrefix) || len(rawToken) <= len(crypto.PATPrefix) {
		return ErrUnauthorized
	}

	now := s.nowFunc().UTC()
	tokenHash := crypto.HashToken(rawToken)
	if err := s.patRepository.RevokeByTokenHash(ctx, tokenHash, now); err != nil {
		if errors.Is(err, repository.ErrPATNotFound) {
			return ErrUnauthorized
		}
		return err
	}
	return nil
}

