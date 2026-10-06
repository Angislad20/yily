package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yannick2009/yily/internal/domain/model"
	"gorm.io/gorm"
)

var (
	ErrVariableVersionNotFound = errors.New("variable latest version not found") // ErrNotFound is returned when a record is not found in the database.
)

// VariableVersionRepository defines the interface for interacting with variable version data in the database.
type VariableVersionRepository interface {
	GetCurrentActive(ctx context.Context, variableID uuid.UUID) (*model.VariableVersion, error) // GetCurrentActive retrieves the current active variable version for a given variable ID.
	Create(ctx context.Context, variableVersion *model.VariableVersion) error                   // Create creates a new variable version in the database.
	SetActive(ctx context.Context, variableVersionID uuid.UUID, active bool) error              // Update updates an existing variable version in the database.
	DeactivateAll(ctx context.Context, variableID uuid.UUID) error                              // DeactivateAll deactivates all active versions of a variable.
	GetAll(ctx context.Context, variableID uuid.UUID) ([]model.VariableVersion, error)          // GetAll retrieves all variable versions from the database.
}

// variableVersionRepository is a struct that implements the VariableVersionRepository interface.
type variableVersionRepository struct {
	DB *gorm.DB
}

// NewVariableVersionRepository creates a new instance of variableVersionRepository with the provided GORM database connection.
func NewVariableVersionRepository(db *gorm.DB) VariableVersionRepository {
	return &variableVersionRepository{DB: db}
}

func (r *variableVersionRepository) Create(ctx context.Context, variableVersion *model.VariableVersion) error {
	var maxVersion int
	err := r.DB.Model(&model.VariableVersion{}).
		Where("variable_id = ?", variableVersion.VariableID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	if err != nil {
		return err
	}
	variableVersion.Version = maxVersion + 1
	return gorm.G[model.VariableVersion](r.DB).Create(ctx, variableVersion)
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
