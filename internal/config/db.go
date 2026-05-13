package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func NewSQLiteDB(cfg *Config) (*sqlx.DB, error) {
	// Получаем путь к БД (с поддержкой Render Persistent Disk)
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = "./data/fuel-points.db"
	}

	// Создаём папку для БД, если её нет
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	log.Printf("Connecting to database at: %s", dbPath)

	// Подключаемся к SQLite
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Настройка прагм для производительности и безопасности
	// Без этих настроек SQLite может работать медленно на Render
	queries := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",   // Write-Ahead Logging
		"PRAGMA synchronous = NORMAL", // Баланс скорости и безопасности
		"PRAGMA cache_size = -64000",  // 64MB кэша
		"PRAGMA busy_timeout = 5000",  // 5 секунд ожидания при блокировке
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Warning: failed to execute %s: %v", query, err)
		}
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to SQLite database: %s", dbPath)
	return db, nil
}
