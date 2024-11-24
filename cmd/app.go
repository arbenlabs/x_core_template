package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"x/core/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Driver string

const (
	PGXDriver Driver = "pgx"
)

// Application models/domains to be migrated as db tables
var models []interface{}

// Connect to postgres database
func ConnectToDB(ctx context.Context, cfg config.Database) (*gorm.DB, error) {
	dbDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	sqlDB, err := sql.Open(string(PGXDriver), dbDSN)
	if err != nil {
		return nil, fmt.Errorf("error opening sql connection")
	}

	// Check the sql connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging sql connection: %v", err)
	}

	// Set max connections
	if cfg.MaxConnections > 0 {
		sqlDB.SetMaxOpenConns(int(cfg.MaxConnections))
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error opening gorm connection")
	}

	handleMigrations(gormDB, models)
	return gormDB, nil
}

// Automatically handle database migrations
func handleMigrations(db *gorm.DB, models []interface{}) error {
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("error during auto migration for %v", model)
		}
	}

	return nil
}
