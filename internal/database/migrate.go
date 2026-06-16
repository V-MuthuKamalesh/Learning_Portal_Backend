package database

import (
	"github.com/collegeassess/backend/internal/models"
	"gorm.io/gorm"
)

// Migrate ensures the pgcrypto extension exists and auto-migrates all models.
func Migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return err
	}
	return db.AutoMigrate(models.AllModels()...)
}
