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
	// Загружаем .env файл (опционально)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Получаем DB_PATH из переменной окружения
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Определяем платформу по переменным окружения
		if os.Getenv("AMVERA") == "true" {
			// Amvera — постоянное хранилище
			dbPath = "/data/fuel-points.db"
			log.Println("Running on Amvera, using persistent storage at:", dbPath)
		} else if os.Getenv("RENDER") == "true" {
			// Render (ещё используется)
			dbPath = "/opt/render/project/src/data/fuel-points.db"
			log.Println("Running on Render, using:", dbPath)
		} else {
			// Локальная разработка
			dbPath = "./data/fuel-points.db"
			log.Println("Running locally, using:", dbPath)
		}
	}

	// Убеждаемся, что папка для БД существует
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Printf("Warning: failed to create data directory: %v", err)
	}

	// JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "development-secret-key-do-not-use-in-production"
		log.Println("WARNING: Using default JWT_SECRET. Set this in production!")
	}

	// Порт сервера
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Printf("Configuration loaded: DB_PATH=%s, PORT=%s", dbPath, serverPort)

	return &Config{
		DBPath:     dbPath,
		JWTSecret:  jwtSecret,
		ServerPort: serverPort,
	}
}
