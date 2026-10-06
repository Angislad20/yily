package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yannick2009/yily/internal/domain/model"
	"github.com/yannick2009/yily/internal/repository"
)

// TagService defines the interface for tag business logic.
type TagService interface {
	Create(ctx context.Context, name string) (*model.Tag, error) // Create creates a new tag.
	Delete(ctx context.Context, tagID uuid.UUID) error           // Delete deletes a tag.
}

// tagService implements TagService.
type tagService struct {
	tagRepository repository.TagRepository
}

// NewTagService creates a new tagService with the provided TagRepository.
func NewTagService(tagRepository repository.TagRepository) TagService {
	return &tagService{tagRepository}
}

// Create creates a new tag.
func (s *tagService) Create(ctx context.Context, name string) (*model.Tag, error) {
	tag := &model.Tag{Name: name}
	if err := s.tagRepository.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

// Delete deletes a tag.
func (s *tagService) Delete(ctx context.Context, tagID uuid.UUID) error {
	return s.tagRepository.Delete(ctx, tagID)
}
