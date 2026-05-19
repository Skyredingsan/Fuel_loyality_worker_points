package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fuel-points/internal/auth"
	"fuel-points/internal/config"
	"fuel-points/internal/handlers"
	"fuel-points/internal/middleware"
	"fuel-points/internal/repository"
	"fuel-points/internal/services"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	// _ "fuel-points/docs" // закомментировано, если нет swagger
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Подключаемся к SQLite
	db, err := config.NewSQLiteDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// СОЗДАЁМ ТАБЛИЦЫ И ДОБАВЛЯЕМ НАЧАЛЬНЫЕ ДАННЫЕ
	ensureDatabaseSetup(db)

	// Создаем таблицы если их нет (старая функция, оставляем для совместимости)
	err = runMigrations(db)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// ... остальной код main.go (репозитории, хендлеры, роутеры, сервер) ...
	// (весь ваш существующий код main.go идёт дальше без изменений)
}

// ensureDatabaseSetup — создаёт таблицы и добавляет начальные данные
func ensureDatabaseSetup(db *sqlx.DB) {
	log.Println("🔄 Checking database setup...")

	// 1. Создаём таблицы, если их нет
	migrationSQL, err := os.ReadFile("migrations/001_create_tables.sqlite.sql")
	if err != nil {
		log.Printf("⚠️ Warning: could not read migration file: %v", err)
	} else {
		_, err = db.Exec(string(migrationSQL))
		if err != nil {
			log.Printf("⚠️ Warning: could not apply migrations: %v", err)
		} else {
			log.Println("✅ Migrations applied successfully")
		}
	}

	// 2. Добавляем KPI показатели (если их нет)
	ensureKPI(db)

	// 3. Добавляем пользователей (если их нет)
	ensureUsers(db)

	log.Println("✅ Database setup complete")
}

// ensureKPI — добавляет KPI показатели, если их ещё нет
func ensureKPI(db *sqlx.DB) {
	// Проверяем, есть ли уже KPI
	var count int
	db.Get(&count, "SELECT COUNT(*) FROM kpi_indicators")
	if count > 0 {
		log.Printf("⏩ KPI already exist (%d records), skipping", count)
		return
	}

	log.Println("📊 Adding KPI indicators...")

	// Категории
	categories := []struct {
		Code        string
		Name        string
		Description string
	}{
		{"ПМ", "Продажи и маржа", "Показатели продаж топлива, СТ и маржинальности"},
		{"ОЭК", "Операционная эффективность и качество", "Тайный покупатель, аудиты, штрафы"},
		{"ЭКЛ", "Эффективность команды и лидерство", "Текучесть, обучение, дисциплина"},
		{"КБ", "Культура безопасности", "Травматизм, ПБОТОС, инициативы"},
	}

	for _, cat := range categories {
		var exists int
		db.Get(&exists, "SELECT COUNT(*) FROM kpi_categories WHERE code = $1", cat.Code)
		if exists == 0 {
			_, err := db.Exec(`
				INSERT INTO kpi_categories (code, name, description)
				VALUES ($1, $2, $3)
			`, cat.Code, cat.Name, cat.Description)
			if err != nil {
				log.Printf("❌ Error creating category %s: %v", cat.Code, err)
			} else {
				log.Printf("✅ Created category: %s (%s)", cat.Name, cat.Code)
			}
		}
	}

	// Показатели KPI (основные)
	indicators := []struct {
		CategoryCode  string
		Code          string
		Name          string
		Description   string
		Unit          string
		IndicatorType string
		BaseValue     *float64
		BaseWeight    *int
		ExtraWeight   *int
		PenaltyWeight *int
	}{
		// ПМ
		{"ПМ", "ПМ1", "Средний процент выполнения плана по продажам топлива", "При достижении 90% ставится 1, иначе 0", "%", "base", float64Ptr(90), intPtr(50), nil, nil},
		{"ПМ", "ПМ2", "Средний процент выполнения плана по маржинальности", "При достижении 90% ставится 1, иначе 0", "%", "base", float64Ptr(90), intPtr(50), nil, nil},
		{"ПМ", "ПМ3", "Средний процент выполнения плана продаж сопутствующих товаров", "При достижении 90% ставится 1, иначе 0", "%", "base", float64Ptr(90), intPtr(50), nil, nil},
		{"ПМ", "ДПМ1", "Перевыполнение плана по продажам топлива", "За каждый % сверх нормы", "%", "extra", nil, nil, intPtr(10), nil},
		{"ПМ", "ДПМ2", "Перевыполнение плана по продажам СТ", "За каждый % сверх нормы", "%", "extra", nil, nil, intPtr(10), nil},
		{"ПМ", "ДПМ3", "Перевыполнение плана по маржинальности", "За каждый % сверх нормы", "%", "extra", nil, nil, intPtr(10), nil},

		// ОЭК
		{"ОЭК", "ОЭК1", "Средний балл по оценкам тайного покупателя", "При достижении 95% ставится 1, иначе 0", "%", "base", float64Ptr(95), intPtr(20), nil, nil},
		{"ОЭК", "ОЭК2", "Соблюдение стандартов работы", "При отсутствии штрафов ставится 1, иначе 0", "шт", "base", float64Ptr(0), intPtr(20), nil, nil},
		{"ОЭК", "ДОЭК", "Количество АЗС с оценкой тайного покупателя выше 95%", "За каждую АЗС с превышающим показателем", "шт", "extra", nil, nil, intPtr(5), nil},
		{"ОЭК", "ШОЭК", "Наличие штрафов со стороны проверяющих органов", "За каждый штраф", "шт", "penalty", nil, nil, nil, intPtr(-5)},

		// ЭКЛ
		{"ЭКЛ", "ЭКЛ1", "Текучесть кадров", "Не выше 15%", "%", "base", float64Ptr(15), intPtr(20), nil, nil},
		{"ЭКЛ", "ЭКЛ2", "Процент укомплектованности", "Не ниже 85%", "%", "base", float64Ptr(85), intPtr(30), nil, nil},
		{"ЭКЛ", "ЭКЛ3", "Отсутствие нарушений трудовой дисциплины", "0 нарушений", "шт", "base", float64Ptr(0), intPtr(30), nil, nil},
		{"ЭКЛ", "ЭКЛ4", "Показатель оценки ПТК персонала", "Не ниже 70%", "%", "base", float64Ptr(70), intPtr(20), nil, nil},
		{"ЭКЛ", "ЭКЛ5", "Выполнение плана по корпоративному обучению", "Не ниже 90%", "%", "base", float64Ptr(90), intPtr(20), nil, nil},
		{"ЭКЛ", "ДЭКЛ1", "Количество сотрудников с оценкой ПТК выше 90%", "За каждого сотрудника", "чел", "extra", nil, nil, intPtr(2), nil},
		{"ЭКЛ", "ДЭКЛ2", "Наличие сотрудников со статусом 'Опыт'", "За каждого сотрудника", "чел", "extra", nil, nil, intPtr(5), nil},
		{"ЭКЛ", "ДЭКЛ3", "Наличие сотрудников со статусом 'Эксперт'", "За каждого сотрудника", "чел", "extra", nil, nil, intPtr(10), nil},
		{"ЭКЛ", "ДЭКЛ4", "Победители конкурса ЛПП (ЛОК этап)", "За каждого сотрудника", "чел", "extra", nil, nil, intPtr(5), nil},
		{"ЭКЛ", "ДЭКЛ5", "Победители конкурса ЛПП (Финальный этап)", "За каждого сотрудника", "чел", "extra", nil, nil, intPtr(10), nil},
		{"ЭКЛ", "ШЭКЛ", "Наличие дисциплинарных взысканий", "За каждый случай", "шт", "penalty", nil, nil, nil, intPtr(-10)},

		// КБ
		{"КБ", "КБ1", "Отсутствие травматизма", "0 случаев", "шт", "base", float64Ptr(0), intPtr(30), nil, nil},
		{"КБ", "КБ2", "Отсутствие нарушений ПБОТОС", "0 нарушений", "шт", "base", float64Ptr(0), intPtr(20), nil, nil},
		{"КБ", "ДКБ1", "Инициативы по устранению опасных ситуаций", "За каждую подтвержденную инициативу", "шт", "extra", nil, nil, intPtr(10), nil},
		{"КБ", "ДКБ2", "Номинация 'За безопасный труд' (ЛОК этап)", "За каждого номинанта", "чел", "extra", nil, nil, intPtr(5), nil},
		{"КБ", "ДКБ3", "Номинация 'За безопасный труд' (Финальный этап)", "За каждого номинанта", "чел", "extra", nil, nil, intPtr(10), nil},
		{"КБ", "ШКБ1", "Случаи травматизма", "За каждый случай", "шт", "penalty", nil, nil, nil, intPtr(-20)},
		{"КБ", "ШКБ2", "Нарушения ПБОТОС", "За каждый случай", "шт", "penalty", nil, nil, nil, intPtr(-10)},
	}

	for _, ind := range indicators {
		var exists int
		db.Get(&exists, "SELECT COUNT(*) FROM kpi_indicators WHERE code = $1", ind.Code)
		if exists > 0 {
			continue
		}

		var categoryID int
		db.Get(&categoryID, "SELECT id FROM kpi_categories WHERE code = $1", ind.CategoryCode)

		_, err := db.Exec(`
			INSERT INTO kpi_indicators (
				category_id, code, name, description, unit, indicator_type,
				base_value, base_weight, extra_weight, penalty_weight
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, categoryID, ind.Code, ind.Name, ind.Description, ind.Unit, ind.IndicatorType,
			ind.BaseValue, ind.BaseWeight, ind.ExtraWeight, ind.PenaltyWeight)

		if err != nil {
			log.Printf("❌ Error creating %s: %v", ind.Code, err)
		} else {
			log.Printf("✅ Created indicator: %s (%s)", ind.Name, ind.Code)
		}
	}

	log.Println("✅ KPI setup complete")
}

// ensureUsers — добавляет пользователей, если их нет
func ensureUsers(db *sqlx.DB) {
	var count int
	db.Get(&count, "SELECT COUNT(*) FROM users")
	if count > 0 {
		log.Printf("⏩ Users already exist (%d records), skipping", count)
		return
	}

	log.Println("👤 Adding default users...")

	users := []struct {
		Email       string
		Password    string
		Role        string
		FIO         string
		ClusterName string
		AZSCount    int
	}{
		{"VelichkinaSV@fuel.ru", "SVVelichkina123!@#", "coordinator", "Величкина Светлана Владимировна", "Центральный офис", 0},
		{"ValeevDI@fuel.ru", "29Dek2001!", "coordinator", "Валеев Денис Игоревич", "Центральный офис", 0},
	}

	for _, u := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("❌ Error hashing password for %s: %v", u.Email, err)
			continue
		}

		_, err = db.Exec(`
			INSERT INTO users (email, password_hash, role, fio, cluster_name, azs_count, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, datetime('now'), datetime('now'))
		`, u.Email, string(hashedPassword), u.Role, u.FIO, u.ClusterName, u.AZSCount)

		if err != nil {
			log.Printf("❌ Error creating user %s: %v", u.Email, err)
		} else {
			log.Printf("✅ Created user: %s (%s)", u.FIO, u.Role)
		}
	}

	log.Println("✅ Users setup complete")
}

func float64Ptr(f float64) *float64 { return &f }
func intPtr(i int) *int             { return &i }

func runMigrations(db *sqlx.DB) error {
	migrationSQL, err := os.ReadFile("migrations/001_create_tables.sqlite.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(migrationSQL))
	return err
}
