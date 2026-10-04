package model

import (
	"time"

	"github.com/google/uuid"
)

// Environment represents an environment in the system.
type Environment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid();index" json:"id"` // Unique identifier for the environment
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`                         // Name of the environment
	Description string     `gorm:"type:varchar(255)" json:"description"`                           // Description of the environment
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`                               // Timestamp when the environment was created
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                               // Timestamp when the environment was last updated
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`                     // Foreign key referencing the associated project
	Project     Project    `json:"project"`                                                        // The associated project object
	Variables   []Variable `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"variables"` // List of variables associated with the environment
}
