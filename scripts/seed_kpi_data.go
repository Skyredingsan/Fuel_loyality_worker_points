package main

import (
	"log"

	"fuel-points/internal/config"
	"fuel-points/internal/models"
	"fuel-points/internal/repository"

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

	kpiRepo := repository.NewKPIRepository(db)

	// Создаем показатели KPI на основе Excel файла

	// ПМ - Продажи и маржа
	indicators := []models.CreateIndicatorRequest{
		// Базовые показатели ПМ
		{
			Code:          "ПМ1",
			Name:          "Средний процент выполнения плана по продажам топлива",
			CategoryCode:  "ПМ",
			Description:   "При достижении 90% ставится 1, иначе 0",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(90),
			BaseWeight:    intPtr(50),
		},
		{
			Code:          "ПМ2",
			Name:          "Средний процент выполнения плана по маржинальности",
			CategoryCode:  "ПМ",
			Description:   "При достижении 90% ставится 1, иначе 0",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(90),
			BaseWeight:    intPtr(50),
		},
		{
			Code:          "ПМ3",
			Name:          "Средний процент выполнения плана продаж сопутствующих товаров",
			CategoryCode:  "ПМ",
			Description:   "При достижении 90% ставится 1, иначе 0",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(90),
			BaseWeight:    intPtr(50),
		},
		// Дополнительные баллы ПМ
		{
			Code:          "ДПМ1",
			Name:          "Перевыполнение плана по продажам топлива",
			CategoryCode:  "ПМ",
			Description:   "За каждый % сверх нормы",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(10),
		},
		{
			Code:          "ДПМ2",
			Name:          "Перевыполнение плана по продажам СТ",
			CategoryCode:  "ПМ",
			Description:   "За каждый % сверх нормы",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(10),
		},
		{
			Code:          "ДПМ3",
			Name:          "Перевыполнение плана по маржинальности",
			CategoryCode:  "ПМ",
			Description:   "За каждый % сверх нормы",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(10),
		},

		// ОЭК - Операционная эффективность и качество
		{
			Code:          "ОЭК1",
			Name:          "Средний балл по оценкам тайного покупателя",
			CategoryCode:  "ОЭК",
			Description:   "При достижении 95% ставится 1, иначе 0",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(95),
			BaseWeight:    intPtr(20),
		},
		{
			Code:          "ОЭК2",
			Name:          "Соблюдение стандартов работы (нулевой уровень штрафов)",
			CategoryCode:  "ОЭК",
			Description:   "При отсутствии штрафов ставится 1, иначе 0",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(0),
			BaseWeight:    intPtr(20),
		},
		{
			Code:          "ДОЭК",
			Name:          "Количество АЗС с оценкой тайного покупателя выше 95%",
			CategoryCode:  "ОЭК",
			Description:   "За каждую АЗС с превышающим показателем",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(5),
		},
		{
			Code:          "ШОЭК",
			Name:          "Наличие штрафов со стороны проверяющих органов",
			CategoryCode:  "ОЭК",
			Description:   "Штрафные баллы за каждый штраф",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypePenalty,
			PenaltyWeight: intPtr(-5),
		},

		// ЭКЛ - Эффективность команды и лидерство
		{
			Code:          "ЭКЛ1",
			Name:          "Текучесть кадров",
			CategoryCode:  "ЭКЛ",
			Description:   "Не выше 15%",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(15),
			BaseWeight:    intPtr(20),
		},
		{
			Code:          "ЭКЛ2",
			Name:          "Процент укомплектованности",
			CategoryCode:  "ЭКЛ",
			Description:   "Не ниже 85%",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(85),
			BaseWeight:    intPtr(30),
		},
		{
			Code:          "ЭКЛ3",
			Name:          "Отсутствие нарушений трудовой дисциплины",
			CategoryCode:  "ЭКЛ",
			Description:   "0 нарушений",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(0),
			BaseWeight:    intPtr(30),
		},
		{
			Code:          "ЭКЛ4",
			Name:          "Показатель оценки ПТК персонала",
			CategoryCode:  "ЭКЛ",
			Description:   "Не ниже 70%",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(70),
			BaseWeight:    intPtr(20),
		},
		{
			Code:          "ЭКЛ5",
			Name:          "Выполнение плана по корпоративному обучению",
			CategoryCode:  "ЭКЛ",
			Description:   "Не ниже 90%",
			Unit:          "%",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(90),
			BaseWeight:    intPtr(20),
		},
		{
			Code:          "ДЭКЛ1",
			Name:          "Количество сотрудников с оценкой ПТК выше 90%",
			CategoryCode:  "ЭКЛ",
			Description:   "За каждого сотрудника",
			Unit:          "чел",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(2),
		},
		{
			Code:          "ДЭКЛ2",
			Name:          "Наличие сотрудников со статусом 'Опыт'",
			CategoryCode:  "ЭКЛ",
			Description:   "За каждого сотрудника",
			Unit:          "чел",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(5),
		},
		{
			Code:          "ДЭКЛ3",
			Name:          "Наличие сотрудников со статусом 'Эксперт'",
			CategoryCode:  "ЭКЛ",
			Description:   "За каждого сотрудника",
			Unit:          "чел",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(10),
		},
		{
			Code:          "ДЭКЛ4",
			Name:          "Победители конкурса ЛПП (ЛОК этап)",
			CategoryCode:  "ЭКЛ",
			Description:   "За каждого сотрудника",
			Unit:          "чел",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(5),
		},
		{
			Code:          "ДЭКЛ5",
			Name:          "Победители конкурса ЛПП (Финальный этап)",
			CategoryCode:  "ЭКЛ",
			Description:   "За каждого сотрудника",
			Unit:          "чел",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(10),
		},
		{
			Code:          "ШЭКЛ",
			Name:          "Наличие дисциплинарных взысканий",
			CategoryCode:  "ЭКЛ",
			Description:   "За каждый случай",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypePenalty,
			PenaltyWeight: intPtr(-10),
		},

		// КБ - Культура безопасности
		{
			Code:          "КБ1",
			Name:          "Отсутствие травматизма",
			CategoryCode:  "КБ",
			Description:   "0 случаев",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(0),
			BaseWeight:    intPtr(30),
		},
		{
			Code:          "КБ2",
			Name:          "Отсутствие нарушений ПБОТОС",
			CategoryCode:  "КБ",
			Description:   "0 нарушений",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypeBase,
			BaseValue:     float64Ptr(0),
			BaseWeight:    intPtr(20),
		},
		{
			Code:          "ДКБ1",
			Name:          "Инициативы по устранению опасных ситуаций",
			CategoryCode:  "КБ",
			Description:   "За каждую подтвержденную инициативу",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(10),
		},
		{
			Code:          "ДКБ2",
			Name:          "Номинация 'За безопасный труд' (ЛОК этап)",
			CategoryCode:  "КБ",
			Description:   "За каждого номинанта",
			Unit:          "чел",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(5),
		},
		{
			Code:          "ДКБ3",
			Name:          "Номинация 'За безопасный труд' (Финальный этап)",
			CategoryCode:  "КБ",
			Description:   "За каждого номинанта",
			Unit:          "чел",
			IndicatorType: models.IndicatorTypeExtra,
			ExtraWeight:   intPtr(10),
		},
		{
			Code:          "ШКБ1",
			Name:          "Случаи травматизма",
			CategoryCode:  "КБ",
			Description:   "За каждый случай",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypePenalty,
			PenaltyWeight: intPtr(-20),
		},
		{
			Code:          "ШКБ2",
			Name:          "Нарушения ПБОТОС",
			CategoryCode:  "КБ",
			Description:   "За каждый случай",
			Unit:          "шт",
			IndicatorType: models.IndicatorTypePenalty,
			PenaltyWeight: intPtr(-10),
		},
	}

	for _, ind := range indicators {
		_, err := kpiRepo.CreateIndicator(&ind)
		if err != nil {
			log.Printf("Error creating indicator %s: %v", ind.Code, err)
		} else {
			log.Printf("Created indicator: %s", ind.Code)
		}
	}

	log.Println("KPI data seeding completed!")
}

func float64Ptr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
