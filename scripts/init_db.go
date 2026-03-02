package main

import (
	"log"
	"os"

	"fuel-points/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := config.LoadConfig()

	// Удаляем существующую БД если нужно пересоздать
	if _, err := os.Stat(cfg.DBPath); err == nil {
		log.Printf("Removing existing database: %s", cfg.DBPath)
		os.Remove(cfg.DBPath)
	}

	// Создаем новую БД
	db, err := sqlx.Connect("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatal("Failed to create database:", err)
	}
	defer db.Close()

	// Включаем внешние ключи
	db.Exec("PRAGMA foreign_keys = ON")

	// Читаем и выполняем миграции
	migrationSQL, err := os.ReadFile("migrations/001_create_tables.sqlite.sql")
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("Database initialized successfully!")
}
