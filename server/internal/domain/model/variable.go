package model

import (
	"time"

	"github.com/google/uuid"
)

// Variable represents a variable in the system.
type Variable struct {
	ID            uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid();index" json:"id"` // Unique identifier for the variable
	Name          string            `gorm:"type:varchar(255);not null" json:"name"`                         // Name of the variable
	Value         string            `gorm:"type:varchar(255)" json:"value"`                                 // Value of the variable
	CreatedAt     time.Time         `gorm:"autoCreateTime" json:"created_at"`                               // Timestamp when the variable was created
	UpdatedAt     time.Time         `gorm:"autoUpdateTime" json:"updated_at"`                               // Timestamp when the variable was last updated
	EnvironmentID uuid.UUID         `gorm:"type:uuid;not null;index" json:"environment_id"`                 // Foreign key referencing the associated environment
	Environment   Environment       `json:"environment"`                                                    // The associated environment object
	Versions      []VariableVersion `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"versions"`  // List of variable versions associated with the variable
}
