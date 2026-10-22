package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yannick2009/yily/internal/domain/model"
	"gorm.io/gorm"
)

var (
	ErrPATNotFound = errors.New("personal access token not found")
)

// PATRepository defines the interface for Personal Access Token persistence.
type PATRepository interface {
	Create(ctx context.Context, pat *model.PersonalAccessToken) error                                   // Create persists a new PAT.
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.PersonalAccessToken, error)           // GetByTokenHash retrieves a PAT by its SHA-256 hash.
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.PersonalAccessToken, error)    // GetByID retrieves a user's PAT by its ID.
	GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]model.PersonalAccessToken, error)          // GetAllByUserID retrieves all PATs belonging to a user.
	RevokeByID(ctx context.Context, id uuid.UUID, userID uuid.UUID, revokedAt time.Time) error          // RevokeByID marks a user's PAT as revoked by ID.
	RevokeByTokenHash(ctx context.Context, tokenHash string, revokedAt time.Time) error                 // RevokeByTokenHash marks a PAT as revoked by its SHA-256 hash.
}

// patRepository implements PATRepository.
type patRepository struct {
	DB *gorm.DB
}

// NewPATRepository creates a new PATRepository.
func NewPATRepository(db *gorm.DB) PATRepository {
	return &patRepository{DB: db}
}

// Create persists a new PAT in the database.
func (r *patRepository) Create(ctx context.Context, pat *model.PersonalAccessToken) error {
	return gorm.G[model.PersonalAccessToken](r.DB).Create(ctx, pat)
}

// GetByTokenHash retrieves a PAT by its SHA-256 token hash.
func (r *patRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.PersonalAccessToken, error) {
	pat, err := gorm.G[model.PersonalAccessToken](r.DB).Where("token_hash = ?", tokenHash).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPATNotFound
		}
		return nil, err
	}
	return &pat, nil
}

// GetByID retrieves a PAT by its ID and owning user ID.
func (r *patRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.PersonalAccessToken, error) {
	pat, err := gorm.G[model.PersonalAccessToken](r.DB).Where("id = ? AND user_id = ?", id, userID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPATNotFound
		}
		return nil, err
	}
	return &pat, nil
}

// GetAllByUserID retrieves all PATs belonging to a user ordered by creation date descending.
func (r *patRepository) GetAllByUserID(ctx context.Context, userID uuid.UUID) ([]model.PersonalAccessToken, error) {
	return gorm.G[model.PersonalAccessToken](r.DB).Where("user_id = ?", userID).Order("created_at DESC").Find(ctx)
}

// RevokeByID marks an active PAT belonging to a user as revoked by its ID.
func (r *patRepository) RevokeByID(ctx context.Context, id uuid.UUID, userID uuid.UUID, revokedAt time.Time) error {
	res := r.DB.WithContext(ctx).
		Model(&model.PersonalAccessToken{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", id, userID).
		Update("revoked_at", revokedAt)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := r.DB.WithContext(ctx).Model(&model.PersonalAccessToken{}).Where("id = ? AND user_id = ?", id, userID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrPATNotFound
		}
	}
	return nil
}

// RevokeByTokenHash marks an active PAT as revoked by its SHA-256 token hash.
func (r *patRepository) RevokeByTokenHash(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	res := r.DB.WithContext(ctx).
		Model(&model.PersonalAccessToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", revokedAt)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := r.DB.WithContext(ctx).Model(&model.PersonalAccessToken{}).Where("token_hash = ?", tokenHash).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrPATNotFound
		}
	}
	return nil
}

