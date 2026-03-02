package config

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func NewSQLiteDB(cfg *Config) (*sqlx.DB, error) {
	// SQLite не требует host/port/user/password, только путь к файлу
	db, err := sqlx.Connect("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Настройка прагм для производительности
	db.Exec("PRAGMA foreign_keys = ON")
	db.Exec("PRAGMA journal_mode = WAL")   // Write-Ahead Logging для лучшей конкурентности
	db.Exec("PRAGMA synchronous = NORMAL") // Баланс скорости и безопасности

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to SQLite database: %s", cfg.DBPath)
	return db, nil
}
