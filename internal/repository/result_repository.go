package repository

import (
	"database/sql"
	"fmt"
	"time"

	"fuel-points/internal/models"

	"github.com/jmoiron/sqlx"
)

type ResultRepository struct {
	db *sqlx.DB
}

func NewResultRepository(db *sqlx.DB) *ResultRepository {
	return &ResultRepository{db: db}
}

// CreateMonthlyResult - создание записи о ежемесячных результатах
func (r *ResultRepository) CreateMonthlyResult(userID int, expertID int, period time.Time) (*models.MonthlyResult, error) {
	result := &models.MonthlyResult{
		UserID:    userID,
		ExpertID:  &expertID,
		Period:    period,
		Status:    models.StatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
        INSERT INTO monthly_results (user_id, expert_id, period, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT(user_id, period) DO UPDATE SET
            expert_id = excluded.expert_id,
            status = excluded.status,
            updated_at = excluded.updated_at
        RETURNING id
    `

	err := r.db.QueryRow(
		query,
		result.UserID, result.ExpertID, result.Period, result.Status,
		result.CreatedAt, result.UpdatedAt,
	).Scan(&result.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create monthly result: %w", err)
	}

	return result, nil
}

// GetMonthlyResult - получение результатов за месяц по пользователю
func (r *ResultRepository) GetMonthlyResult(userID int, period time.Time) (*models.MonthlyResult, error) {
	var result models.MonthlyResult
	query := `SELECT id, user_id, expert_id, period, status, created_at, updated_at 
              FROM monthly_results WHERE user_id = $1 AND period = $2`

	err := r.db.Get(&result, query, userID, period)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get monthly result: %w", err)
	}

	return &result, nil
}

// GetMonthlyResultsByPeriod - получение всех результатов за период
func (r *ResultRepository) GetMonthlyResultsByPeriod(period time.Time) ([]models.MonthlyResult, error) {
	var results []models.MonthlyResult
	query := `SELECT id, user_id, expert_id, period, status, created_at, updated_at 
              FROM monthly_results WHERE period = $1 ORDER BY user_id`

	err := r.db.Select(&results, query, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly results: %w", err)
	}

	return results, nil
}

// SaveIndicatorResult - сохранение результата по показателю
func (r *ResultRepository) SaveIndicatorResult(
	monthlyResultID int,
	indicatorID int,
	factValue *float64,
	calculatedPoints int,
	documentURL *string,
) (*models.IndicatorResult, error) {

	result := &models.IndicatorResult{
		MonthlyResultID:       monthlyResultID,
		IndicatorID:           indicatorID,
		FactValue:             factValue,
		CalculatedPoints:      calculatedPoints,
		SupportingDocumentURL: documentURL,
		CreatedAt:             time.Now(),
	}

	query := `
        INSERT INTO indicator_results (
            monthly_result_id, indicator_id, fact_value, 
            calculated_points, supporting_document_url, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT(monthly_result_id, indicator_id) DO UPDATE SET
            fact_value = excluded.fact_value,
            calculated_points = excluded.calculated_points,
            supporting_document_url = excluded.supporting_document_url
        RETURNING id
    `

	err := r.db.QueryRow(
		query,
		result.MonthlyResultID, result.IndicatorID, result.FactValue,
		result.CalculatedPoints, result.SupportingDocumentURL, result.CreatedAt,
	).Scan(&result.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to save indicator result: %w", err)
	}

	return result, nil
}

// GetIndicatorResults - получение всех результатов для monthly_result
func (r *ResultRepository) GetIndicatorResults(monthlyResultID int) ([]models.IndicatorResult, error) {
	var results []models.IndicatorResult
	query := `
        SELECT ir.*, 
               i.id as "indicator.id", 
               i.category_id as "indicator.category_id", 
               i.code as "indicator.code", 
               i.name as "indicator.name",
               i.description as "indicator.description", 
               i.unit as "indicator.unit", 
               i.indicator_type as "indicator.indicator_type",
               i.base_value as "indicator.base_value", 
               i.base_weight as "indicator.base_weight",
               i.extra_weight as "indicator.extra_weight", 
               i.penalty_weight as "indicator.penalty_weight",
               i.created_at as "indicator.created_at",
               c.code as "indicator.category_code", 
               c.name as "indicator.category_name"
        FROM indicator_results ir
        JOIN kpi_indicators i ON ir.indicator_id = i.id
        JOIN kpi_categories c ON i.category_id = c.id
        WHERE ir.monthly_result_id = $1
        ORDER BY i.code
    `

	err := r.db.Select(&results, query, monthlyResultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get indicator results: %w", err)
	}

	return results, nil
}

// ConfirmMonthlyResult - подтверждение результатов (перевод в статус confirmed)
func (r *ResultRepository) ConfirmMonthlyResult(monthlyResultID int) error {
	query := `UPDATE monthly_results SET status = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.Exec(query, models.StatusConfirmed, time.Now(), monthlyResultID)
	if err != nil {
		return fmt.Errorf("failed to confirm monthly result: %w", err)
	}

	return nil
}

// GetUserResultsForYear - получение всех результатов пользователя за год
func (r *ResultRepository) GetUserResultsForYear(userID int, year int) ([]models.MonthlyResult, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	var results []models.MonthlyResult
	query := `
        SELECT id, user_id, expert_id, period, status, created_at, updated_at 
        FROM monthly_results 
        WHERE user_id = $1 AND period BETWEEN $2 AND $3 AND status = 'confirmed'
        ORDER BY period
    `

	err := r.db.Select(&results, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user results for year: %w", err)
	}

	return results, nil
}

// GetTotalPointsForYear - получение суммы баллов пользователя за год
func (r *ResultRepository) GetTotalPointsForYear(userID int, year int) (int, error) {
	query := `
        SELECT COALESCE(SUM(ir.calculated_points), 0)
        FROM indicator_results ir
        JOIN monthly_results mr ON ir.monthly_result_id = mr.id
        WHERE mr.user_id = $1 
          AND strftime('%Y', mr.period) = $2
          AND mr.status = 'confirmed'
    `

	var total int
	err := r.db.Get(&total, query, userID, fmt.Sprintf("%d", year))
	if err != nil {
		return 0, fmt.Errorf("failed to get total points for year: %w", err)
	}

	return total, nil
}

// GetMonthlyResultByID - получение результата по ID
func (r *ResultRepository) GetMonthlyResultByID(id int) (*models.MonthlyResult, error) {
	var result models.MonthlyResult
	query := `SELECT id, user_id, expert_id, period, status, created_at, updated_at 
              FROM monthly_results WHERE id = $1`

	err := r.db.Get(&result, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get monthly result by id: %w", err)
	}

	return &result, nil
}

// DeleteMonthlyResult - удаление результата (для отклонения)
func (r *ResultRepository) DeleteMonthlyResult(id int) error {
	query := `DELETE FROM monthly_results WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete monthly result: %w", err)
	}

	return nil
}

// DeleteIndicatorResults - удаление всех результатов по monthly_result_id
func (r *ResultRepository) DeleteIndicatorResults(monthlyResultID int) error {
	query := `DELETE FROM indicator_results WHERE monthly_result_id = $1`

	_, err := r.db.Exec(query, monthlyResultID)
	if err != nil {
		return fmt.Errorf("failed to delete indicator results: %w", err)
	}

	return nil
}
