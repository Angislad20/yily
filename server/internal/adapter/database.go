package adapter

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewDB initializes and returns a new GORM database connection using SQLite.
func NewDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate()

	return db
}
