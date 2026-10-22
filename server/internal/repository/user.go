package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yannick2009/yily/internal/domain/model"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("super admin user already exists")
)

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error                // Create persists the Super Admin user.
	GetByEmail(ctx context.Context, email string) (*model.User, error) // GetByEmail retrieves a user by normalized email.
	GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error) // GetByID retrieves a user by ID.
}

// userRepository implements UserRepository.
type userRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{DB: db}
}

// Create persists the single Super Admin account in the database.
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.User{}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrUserAlreadyExists
		}

		if err := gorm.G[model.User](tx).Create(ctx, user); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrUserAlreadyExists
			}
			return err
		}
		return nil
	})
}

// GetByEmail retrieves a user by normalized email.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := gorm.G[model.User](r.DB).Where("email = ?", email).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByID retrieves a user by ID.
func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := gorm.G[model.User](r.DB).Where("id = ?", userID).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}
