package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yannick2009/yily/internal/domain/model"
	"gorm.io/gorm"
)

// TagRepository defines the interface for interacting with tag data in the database.
type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error  // Create creates a new tag in the database.
	Delete(ctx context.Context, tagID uuid.UUID) error // Delete deletes a tag from the database by its ID.
}

// tagRepository is a struct that implements the TagRepository interface.
type tagRepository struct {
	DB *gorm.DB
}

// NewTagRepository creates a new instance of tagRepository with the provided GORM database connection.
func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{DB: db}
}

// Create creates a new tag in the database.
func (r *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
	return gorm.G[model.Tag](r.DB).Create(ctx, tag)
}

// Delete deletes a tag from the database by its ID.
func (r *tagRepository) Delete(ctx context.Context, tagID uuid.UUID) error {
	_, err := gorm.G[model.Tag](r.DB).Where("id = ?", tagID).Delete(ctx)
	return err
}
