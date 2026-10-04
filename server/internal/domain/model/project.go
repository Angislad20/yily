package model

import (
	"time"

	"github.com/google/uuid"
)

// Project represents a project in the system.
type Project struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid();index" json:"id"`    // Unique identifier for the project
	Name         string        `gorm:"type:varchar(255);not null" json:"name"`                            // Name of the project
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"created_at"`                                  // Timestamp when the project was created
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updated_at"`                                  // Timestamp when the project was last updated
	Environments []Environment `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"environments"` // List of environments associated with the project
}
