package model

import "github.com/google/uuid"

// VariableVersion represents a version of a variable in the system.
type VariableVersion struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid();index" json:"id"` // Unique identifier for the variable version
	Value      string    `gorm:"type:text;not null" json:"value"`                                // Value of the variable version
	CreatedAt  int64     `gorm:"autoCreateTime" json:"created_at"`                               // Timestamp when the variable version was created
	UpdatedAt  int64     `gorm:"autoUpdateTime" json:"updated_at"`                               // Timestamp when the variable version was last updated
	VariableID uuid.UUID `gorm:"type:uuid;not null;index" json:"variable_id"`                    // Foreign key referencing the associated variable
	Variable   Variable  `json:"variable"`                                                       // The associated variable object
}
