package main

import (
	"fuel-points/internal/config"
	"fuel-points/internal/models"
	"fuel-points/internal/repository"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sqlx.Connect("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Включаем внешние ключи
	db.Exec("PRAGMA foreign_keys = ON")

	userRepo := repository.NewUserRepository(db)

	// Создаем координатора
	coordinator := &models.CreateUserRequest{
		Email:    "coordinator@fuel.ru",
		Password: "password123",
		Role:     models.RoleCoordinator,
		FIO:      "Иванов Иван Иванович",
		AZSCount: 0,
	}

	_, err = userRepo.Create(coordinator)
	if err != nil {
		log.Printf("Coordinator may already exist: %v", err)
	} else {
		log.Println("Coordinator created")
	}

	// Создаем эксперта
	expert := &models.CreateUserRequest{
		Email:    "expert@fuel.ru",
		Password: "password123",
		Role:     models.RoleExpert,
		FIO:      "Петров Петр Петрович",
		AZSCount: 0,
	}

	_, err = userRepo.Create(expert)
	if err != nil {
		log.Printf("Expert may already exist: %v", err)
	} else {
		log.Println("Expert created")
	}

	// Создаем ТМ
	tms := []models.CreateUserRequest{
		{
			Email:       "tm1@fuel.ru",
			Password:    "password123",
			Role:        models.RoleTM,
			FIO:         "Сидоров Сидор Сидорович",
			ClusterName: stringPtr("Северный"),
			AZSCount:    5,
		},
		{
			Email:       "tm2@fuel.ru",
			Password:    "password123",
			Role:        models.RoleTM,
			FIO:         "Кузнецова Анна Михайловна",
			ClusterName: stringPtr("Южный"),
			AZSCount:    7,
		},
		{
			Email:       "tm3@fuel.ru",
			Password:    "password123",
			Role:        models.RoleTM,
			FIO:         "Новиков Алексей Дмитриевич",
			ClusterName: stringPtr("Западный"),
			AZSCount:    4,
		},
	}

	for _, tm := range tms {
		_, err := userRepo.Create(&tm)
		if err != nil {
			log.Printf("TM %s may already exist: %v", tm.Email, err)
		} else {
			log.Printf("TM %s created", tm.FIO)
		}
	}

	log.Println("Seeding completed!")
}

func stringPtr(s string) *string {
	return &s
}
