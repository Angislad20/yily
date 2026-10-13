package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yannick2009/yily/internal/domain/model"
	"gorm.io/gorm"
)

var (
	ErrVariableVersionNotFound = errors.New("variable latest version not found")
	ErrDuplicateVersion        = errors.New("duplicate variable version")
)

// EncryptFn encrypts a value using the assigned version number.
type EncryptFn func(version int) (encryptedValue string, err error)

// VariableVersionRepository defines the interface for interacting with variable version data.
type VariableVersionRepository interface {
	GetCurrentActive(ctx context.Context, variableID uuid.UUID) (*model.VariableVersion, error)                          // GetCurrentActive retrieves the current active version.
	CreateWithEncryption(ctx context.Context, variableID uuid.UUID, encryptFn EncryptFn) (*model.VariableVersion, error) // CreateWithEncryption creates and encrypts a version.
	SetActive(ctx context.Context, variableVersionID uuid.UUID, active bool) error                                       // SetActive updates the active status.
	DeactivateAll(ctx context.Context, variableID uuid.UUID) error                                                       // DeactivateAll deactivates all versions.
	GetAll(ctx context.Context, variableID uuid.UUID) ([]model.VariableVersion, error)                                   // GetAll retrieves all versions.
}

// variableVersionRepository is a struct that implements the VariableVersionRepository interface.
type variableVersionRepository struct {
	DB *gorm.DB
}

// NewVariableVersionRepository creates a new instance of variableVersionRepository.
func NewVariableVersionRepository(db *gorm.DB) VariableVersionRepository {
	return &variableVersionRepository{DB: db}
}

// CreateWithEncryption computes the next version, encrypts via callback, and persists the version atomically.
func (r *variableVersionRepository) CreateWithEncryption(ctx context.Context, variableID uuid.UUID, encryptFn EncryptFn) (*model.VariableVersion, error) {
	var createdVersion model.VariableVersion
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var maxVersion int
		err := tx.Model(&model.VariableVersion{}).
			Where("variable_id = ?", variableID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&maxVersion).Error
		if err != nil {
			return err
		}

		nextVersion := maxVersion + 1
		encryptedValue, err := encryptFn(nextVersion)
		if err != nil {
			return err
		}

		createdVersion = model.VariableVersion{
			VariableID: variableID,
			Version:    nextVersion,
			Value:      encryptedValue,
			Active:     true,
		}

		if err := gorm.G[model.VariableVersion](tx).Create(ctx, &createdVersion); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrDuplicateVersion
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &createdVersion, nil
}

// SetActive updates the active status of a variable version in the database.
func (r *variableVersionRepository) SetActive(ctx context.Context, variableVersionID uuid.UUID, active bool) error {
	_, err := gorm.G[model.VariableVersion](r.DB).Where("id = ?", variableVersionID).Update(ctx, "active", active)
	return err
}

// DeactivateAll deactivates all active versions of a variable.
func (r *variableVersionRepository) DeactivateAll(ctx context.Context, variableID uuid.UUID) error {
	_, err := gorm.G[model.VariableVersion](r.DB).Where("variable_id = ? AND active = ?", variableID, true).Update(ctx, "active", false)
	return err
}

// GetAll retrieves all variable versions from the database.
func (r *variableVersionRepository) GetAll(ctx context.Context, variableID uuid.UUID) ([]model.VariableVersion, error) {
	variableVersions, err := gorm.G[model.VariableVersion](r.DB).Where("variable_id = ?", variableID).Order("version DESC").Find(ctx)
	if err != nil {
		return nil, err
	}
	return variableVersions, nil
}

// GetCurrentActive retrieves the current active variable version for a given variable ID.
func (r *variableVersionRepository) GetCurrentActive(ctx context.Context, variableID uuid.UUID) (*model.VariableVersion, error) {
	variableVersion, err := gorm.G[model.VariableVersion](r.DB).Where("variable_id = ? AND active = ?", variableID, true).Order("version DESC").First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVariableVersionNotFound
		}
		return nil, err
	}
	return &variableVersion, nil
}
