package database

import (
	"fmt"
	"time"

	"github.com/collegeassess/backend/configs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewPostgres opens a pooled GORM connection to PostgreSQL.
func NewPostgres(cfg *configs.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	logLevel := gormlogger.Warn
	if cfg.IsProd() {
		logLevel = gormlogger.Error
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   gormlogger.Default.LogMode(logLevel),
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
