package services

import (
	"fmt"
	"math"
	"time"

	"fuel-points/internal/models"
	"fuel-points/internal/repository"
)

type ScoreCalculator struct {
	kpiRepo    *repository.KPIRepository
	resultRepo *repository.ResultRepository
	userRepo   *repository.UserRepository
	levelRepo  *repository.LevelRepository
}

func NewScoreCalculator(
	kpiRepo *repository.KPIRepository,
	resultRepo *repository.ResultRepository,
	userRepo *repository.UserRepository,
	levelRepo *repository.LevelRepository,
) *ScoreCalculator {
	return &ScoreCalculator{
		kpiRepo:    kpiRepo,
		resultRepo: resultRepo,
		userRepo:   userRepo,
		levelRepo:  levelRepo,
	}
}

// CalculateBasePoints - расчет базовых баллов
// Формула: если факт >= базового_значения то вес * 1 иначе 0
func (s *ScoreCalculator) CalculateBasePoints(indicator *models.KPIIndicator, factValue float64) int {
	if indicator.IndicatorType != models.IndicatorTypeBase {
		return 0
	}

	if indicator.BaseValue == nil || indicator.BaseWeight == nil {
		return 0
	}

	// Проверяем достижение критерия
	if factValue >= *indicator.BaseValue {
		return *indicator.BaseWeight * 1
	}

	return 0
}

// CalculateExtraPoints - расчет дополнительных баллов
// Формула: факт * вес_дополнительного
func (s *ScoreCalculator) CalculateExtraPoints(indicator *models.KPIIndicator, factValue float64) int {
	if indicator.IndicatorType != models.IndicatorTypeExtra {
		return 0
	}

	if indicator.ExtraWeight == nil {
		return 0
	}

	// Для процентов (ДПМ1, ДПМ2, ДПМ3) - умножаем % перевыполнения на вес
	// Для штук (ДОЭК, ДЭКЛ1) - умножаем количество на вес
	return int(math.Round(factValue * float64(*indicator.ExtraWeight)))
}

// CalculatePenaltyPoints - расчет штрафных баллов
// Формула: факт * вес_штрафного (отрицательные баллы)
func (s *ScoreCalculator) CalculatePenaltyPoints(indicator *models.KPIIndicator, factValue float64) int {
	if indicator.IndicatorType != models.IndicatorTypePenalty {
		return 0
	}

	if indicator.PenaltyWeight == nil {
		return 0
	}

	// Штрафные баллы отрицательные
	return int(math.Round(factValue * float64(*indicator.PenaltyWeight)))
}

// CalculateAllPoints - расчет всех баллов для показателя
func (s *ScoreCalculator) CalculateAllPoints(indicator *models.KPIIndicator, factValue *float64) int {
	if factValue == nil {
		return 0
	}

	switch indicator.IndicatorType {
	case models.IndicatorTypeBase:
		return s.CalculateBasePoints(indicator, *factValue)
	case models.IndicatorTypeExtra:
		return s.CalculateExtraPoints(indicator, *factValue)
	case models.IndicatorTypePenalty:
		return s.CalculatePenaltyPoints(indicator, *factValue)
	default:
		return 0
	}
}

// ProcessMonthlyResults - обработка ежемесячных результатов
func (s *ScoreCalculator) ProcessMonthlyResults(
	userID int,
	expertID int,
	period time.Time,
	results []models.IndicatorResultInput,
) (*models.MonthlyResult, error) {

	// Создаем или получаем monthly_result
	monthlyResult, err := s.resultRepo.CreateMonthlyResult(userID, expertID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to create monthly result: %w", err)
	}

	// Получаем все показатели для проверки
	allIndicators, err := s.kpiRepo.GetAllIndicators()
	if err != nil {
		return nil, fmt.Errorf("failed to get indicators: %w", err)
	}

	// Создаем map для быстрого доступа
	indicatorMap := make(map[string]*models.KPIIndicator)
	for _, ind := range allIndicators {
		indicatorMap[ind.Code] = &ind
	}

	// Обрабатываем каждый результат
	for _, input := range results {
		indicator, exists := indicatorMap[input.IndicatorCode]
		if !exists {
			return nil, fmt.Errorf("indicator %s not found", input.IndicatorCode)
		}

		// Рассчитываем баллы
		points := s.CalculateAllPoints(indicator, input.FactValue)

		// Сохраняем результат
		_, err = s.resultRepo.SaveIndicatorResult(
			monthlyResult.ID,
			indicator.ID,
			input.FactValue,
			points,
			input.DocumentURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to save indicator result: %w", err)
		}
	}

	return monthlyResult, nil
}

// GetCategorySummary - получение сводки по категории
func (s *ScoreCalculator) GetCategorySummary(
	monthlyResultID int,
	categoryCode string,
) (*models.CategorySummary, error) {

	// Получаем все результаты
	results, err := s.resultRepo.GetIndicatorResults(monthlyResultID)
	if err != nil {
		return nil, err
	}

	summary := &models.CategorySummary{
		CategoryCode: categoryCode,
	}

	// Суммируем баллы по типу
	for _, result := range results {
		if result.Indicator.CategoryCode != categoryCode {
			continue
		}

		switch result.Indicator.IndicatorType {
		case models.IndicatorTypeBase:
			summary.BasePoints += result.CalculatedPoints
		case models.IndicatorTypeExtra:
			summary.ExtraPoints += result.CalculatedPoints
		case models.IndicatorTypePenalty:
			summary.PenaltyPoints += result.CalculatedPoints
		}
	}

	summary.TotalPoints = summary.BasePoints + summary.ExtraPoints + summary.PenaltyPoints

	return summary, nil
}

// GetFullSummary - получение полного сводного отчета с детализацией
func (s *ScoreCalculator) GetFullSummary(userID int, period time.Time) (*models.FullResultSummary, error) {
	// Получаем monthly result
	monthlyResult, err := s.resultRepo.GetMonthlyResult(userID, period)
	if err != nil {
		return nil, err
	}
	if monthlyResult == nil {
		return nil, fmt.Errorf("no results found for user %d in period %s", userID, period.Format("2006-01"))
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Получаем детальные результаты по показателям
	indicatorResults, err := s.resultRepo.GetIndicatorResults(monthlyResult.ID)
	if err != nil {
		return nil, err
	}

	summary := &models.FullResultSummary{
		UserID:          userID,
		UserFIO:         user.FIO,
		Period:          period.Format("2006-01"),
		Categories:      []models.CategorySummary{},
		DetailedResults: indicatorResults, // Добавляем детальные результаты
	}

	// Получаем все категории
	categories, err := s.kpiRepo.GetAllCategories()
	if err != nil {
		return nil, err
	}

	// Для каждой категории собираем сводку
	totalPoints := 0
	for _, cat := range categories {
		catSummary, err := s.GetCategorySummary(monthlyResult.ID, cat.Code)
		if err != nil {
			return nil, err
		}

		if catSummary != nil {
			catSummary.CategoryName = cat.Name
			summary.Categories = append(summary.Categories, *catSummary)
			totalPoints += catSummary.TotalPoints
		}
	}

	summary.TotalPoints = totalPoints

	// Определяем уровень пользователя по итогам года
	year := period.Year()
	yearlyPoints, err := s.resultRepo.GetTotalPointsForYear(userID, year)
	if err != nil {
		return nil, err
	}

	level, err := s.levelRepo.GetLevelByPoints(yearlyPoints)
	if err != nil {
		return nil, err
	}
	summary.Level = level

	return summary, nil
}
