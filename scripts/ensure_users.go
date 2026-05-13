package main

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Настройка подключения
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/fuel-points.db"
	}
	log.Printf("Подключаюсь к БД: %s", dbPath)

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer db.Close()

	// 2. Функция безопасного добавления
	addUser := func(email, password, role, fio, cluster string, azsCount int) {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Ошибка хэша для %s: %v", email, err)
			return
		}

		var exists bool
		db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email)
		if exists {
			log.Printf("Пользователь %s уже существует", email)
			return
		}

		_, err = db.Exec(`
			INSERT INTO users (email, password_hash, role, fio, cluster_name, azs_count, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, datetime('now'), datetime('now'))
		`, email, string(hashed), role, fio, cluster, azsCount)

		if err != nil {
			log.Printf("Ошибка создания %s: %v", email, err)
		} else {
			log.Printf("✅ Создан пользователь: %s (%s)", fio, role)
		}
	}

	// 3. Добавляем нужных пользователей
	addUser("VelichkinaSV@fuel.ru", "SVVelichkina123!@#", "coordinator", "Величкина Светлана Владимировна", "Центральный офис", 0)
	addUser("ValeevDI@fuel.ru", "29Dek2001!", "coordinator", "Валеев Денис Игоревич", "Центральный офис", 0)

	log.Println("✅ Готово!")
}
