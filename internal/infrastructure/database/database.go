package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheAmirhosssein/class-manager/config"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

const migrationsDir = "migrations"

func Open(cfg config.Database) (*sql.DB, error) {
	if cfg.Driver != "" && cfg.Driver != "sqlite" {
		return nil, fmt.Errorf("unsupported database driver %q (only sqlite is configured)", cfg.Driver)
	}
	if cfg.Path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return db, nil
}

func MigrateUp(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.Up(db, migrationsDir)
}

func MigrateDown(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	return goose.Down(db, migrationsDir)
}
