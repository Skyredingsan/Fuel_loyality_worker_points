package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Определяем путь к БД (в папке data)
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/fuel-points.db"
	}

	// Создаём папку data, если её нет
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatal("❌ Ошибка создания папки data:", err)
	}

	log.Printf("📁 Подключаюсь к БД: %s", dbPath)

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		log.Fatal("❌ Ошибка подключения к БД:", err)
	}
	defer db.Close()

	// Проверяем, есть ли уже KPI
	var count int
	db.Get(&count, "SELECT COUNT(*) FROM kpi_indicators")
	if count > 0 {
		log.Printf("⏩ KPI уже загружены (%d записей), пропускаем", count)
		return
	}

	log.Println("📊 Начинаю загрузку KPI показателей...")

	// ==================== КАТЕГОРИИ ====================
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
				log.Printf("❌ Ошибка создания категории %s: %v", cat.Code, err)
			} else {
				log.Printf("✅ Создана категория: %s (%s)", cat.Name, cat.Code)
			}
		}
	}

	// ==================== ПОКАЗАТЕЛИ ====================
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
		// ПМ - Продажи и маржа
		{"ПМ", "ПМ1", "Средний процент выполнения плана по продажам топлива", "При достижении 90% ставится 1, иначе 0", "%", "base", float64Ptr(90), intPtr(50), nil, nil},
		{"ПМ", "ПМ2", "Средний процент выполнения плана по маржинальности", "При достижении 90% ставится 1, иначе 0", "%", "base", float64Ptr(90), intPtr(50), nil, nil},
		{"ПМ", "ПМ3", "Средний процент выполнения плана продаж сопутствующих товаров", "При достижении 90% ставится 1, иначе 0", "%", "base", float64Ptr(90), intPtr(50), nil, nil},
		{"ПМ", "ДПМ1", "Перевыполнение плана по продажам топлива", "За каждый % сверх нормы", "%", "extra", nil, nil, intPtr(10), nil},
		{"ПМ", "ДПМ2", "Перевыполнение плана по продажам СТ", "За каждый % сверх нормы", "%", "extra", nil, nil, intPtr(10), nil},
		{"ПМ", "ДПМ3", "Перевыполнение плана по маржинальности", "За каждый % сверх нормы", "%", "extra", nil, nil, intPtr(10), nil},

		// ОЭК - Операционная эффективность
		{"ОЭК", "ОЭК1", "Средний балл по оценкам тайного покупателя", "При достижении 95% ставится 1, иначе 0", "%", "base", float64Ptr(95), intPtr(20), nil, nil},
		{"ОЭК", "ОЭК2", "Соблюдение стандартов работы", "При отсутствии штрафов ставится 1, иначе 0", "шт", "base", float64Ptr(0), intPtr(20), nil, nil},
		{"ОЭК", "ДОЭК", "Количество АЗС с оценкой тайного покупателя выше 95%", "За каждую АЗС с превышающим показателем", "шт", "extra", nil, nil, intPtr(5), nil},
		{"ОЭК", "ШОЭК", "Наличие штрафов со стороны проверяющих органов", "За каждый штраф", "шт", "penalty", nil, nil, nil, intPtr(-5)},

		// ЭКЛ - Эффективность команды
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

		// КБ - Культура безопасности
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
			log.Printf("❌ Ошибка создания %s: %v", ind.Code, err)
		} else {
			log.Printf("✅ Создан показатель: %s (%s)", ind.Name, ind.Code)
		}
	}

	// Уровни
	levels := []struct {
		Name             string
		MinPointsPerYear int
		Privileges       string
	}{
		{"Специалист Трассы", 0, `{"bonus": "Стандартный пакет мотивации"}`},
		{"Тактик Магистрали", 4321, `{"bonus": "Доплата 20% к окладу"}`},
		{"Стратег Гран-при", 4321, `{"bonus": "Доплата 50% к окладу", "prize": "Поездка на Кубу"}`},
	}

	for _, lvl := range levels {
		var exists int
		db.Get(&exists, "SELECT COUNT(*) FROM levels WHERE name = $1", lvl.Name)
		if exists == 0 {
			_, err := db.Exec(`
				INSERT INTO levels (name, min_points_per_year, privileges)
				VALUES ($1, $2, $3)
			`, lvl.Name, lvl.MinPointsPerYear, lvl.Privileges)
			if err != nil {
				log.Printf("❌ Ошибка создания уровня %s: %v", lvl.Name, err)
			} else {
				log.Printf("✅ Создан уровень: %s", lvl.Name)
			}
		}
	}

	log.Println("✅ Загрузка KPI завершена!")
}

func float64Ptr(f float64) *float64 { return &f }
func intPtr(i int) *int             { return &i }
