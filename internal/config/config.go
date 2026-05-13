package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath     string
	JWTSecret  string
	ServerPort string
}

func LoadConfig() *Config {
	// Загружаем .env файл (опционально, на Render не обязательно)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Получаем DB_PATH, с поддержкой Render Persistent Disk
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// На Render диск монтируется в /opt/render/project/src/data
		// Но для локальной разработки используем ./data
		if os.Getenv("RENDER") == "true" {
			dbPath = "/opt/render/project/src/data/fuel-points.db"
		} else {
			// Локальная разработка
			dbPath = "./data/fuel-points.db"
		}
	}

	// Убеждаемся, что папка существует
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Printf("Warning: failed to create data directory: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "development-secret-key-do-not-use-in-production"
		log.Println("WARNING: Using default JWT_SECRET. Set this in production!")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	return &Config{
		DBPath:     dbPath,
		JWTSecret:  jwtSecret,
		ServerPort: serverPort,
	}
}
