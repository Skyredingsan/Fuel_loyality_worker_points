package repository

import (
	"database/sql"
	"fmt"

	"fuel-points/internal/models"

	"github.com/jmoiron/sqlx"
)

type KPIRepository struct {
	db *sqlx.DB
}

func NewKPIRepository(db *sqlx.DB) *KPIRepository {
	return &KPIRepository{db: db}
}

// GetAllCategories - получение всех категорий KPI
func (r *KPIRepository) GetAllCategories() ([]models.KPICategory, error) {
	var categories []models.KPICategory
	query := `SELECT id, name, code, description, created_at FROM kpi_categories ORDER BY code`

	err := r.db.Select(&categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

// GetCategoryByCode - получение категории по коду (ПМ, ОЭК, ЭКЛ, КБ)
func (r *KPIRepository) GetCategoryByCode(code string) (*models.KPICategory, error) {
	var category models.KPICategory
	query := `SELECT id, name, code, description, created_at FROM kpi_categories WHERE code = $1`

	err := r.db.Get(&category, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// CreateIndicator - создание нового показателя KPI
func (r *KPIRepository) CreateIndicator(req *models.CreateIndicatorRequest) (*models.KPIIndicator, error) {
	// Получаем ID категории по коду
	category, err := r.GetCategoryByCode(req.CategoryCode)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, fmt.Errorf("category with code %s not found", req.CategoryCode)
	}

	indicator := &models.KPIIndicator{
		CategoryID:    category.ID,
		Code:          req.Code,
		Name:          req.Name,
		Description:   req.Description,
		Unit:          req.Unit,
		IndicatorType: req.IndicatorType,
		BaseValue:     req.BaseValue,
		BaseWeight:    req.BaseWeight,
		ExtraWeight:   req.ExtraWeight,
		PenaltyWeight: req.PenaltyWeight,
	}

	query := `
        INSERT INTO kpi_indicators (
            category_id, code, name, description, unit, indicator_type,
            base_value, base_weight, extra_weight, penalty_weight
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, created_at
    `

	err = r.db.QueryRow(
		query,
		indicator.CategoryID, indicator.Code, indicator.Name, indicator.Description,
		indicator.Unit, indicator.IndicatorType, indicator.BaseValue, indicator.BaseWeight,
		indicator.ExtraWeight, indicator.PenaltyWeight,
	).Scan(&indicator.ID, &indicator.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create indicator: %w", err)
	}

	return indicator, nil
}

// GetAllIndicators - получение всех показателей KPI
func (r *KPIRepository) GetAllIndicators() ([]models.KPIIndicator, error) {
	var indicators []models.KPIIndicator
	query := `
        SELECT i.*, c.code as category_code, c.name as category_name
        FROM kpi_indicators i
        JOIN kpi_categories c ON i.category_id = c.id
        ORDER BY c.code, i.code
    `

	err := r.db.Select(&indicators, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get indicators: %w", err)
	}

	return indicators, nil
}

// GetIndicatorsByCategory - получение показателей по категории
func (r *KPIRepository) GetIndicatorsByCategory(categoryCode string) ([]models.KPIIndicator, error) {
	var indicators []models.KPIIndicator
	query := `
        SELECT i.*, c.code as category_code, c.name as category_name
        FROM kpi_indicators i
        JOIN kpi_categories c ON i.category_id = c.id
        WHERE c.code = $1
        ORDER BY i.code
    `

	err := r.db.Select(&indicators, query, categoryCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get indicators by category: %w", err)
	}

	return indicators, nil
}

// GetIndicatorByCode - получение показателя по коду
func (r *KPIRepository) GetIndicatorByCode(code string) (*models.KPIIndicator, error) {
	var indicator models.KPIIndicator
	query := `SELECT * FROM kpi_indicators WHERE code = $1`

	err := r.db.Get(&indicator, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get indicator: %w", err)
	}

	return &indicator, nil
}

// UpdateIndicator - обновление показателя
func (r *KPIRepository) UpdateIndicator(id int, req *models.CreateIndicatorRequest) (*models.KPIIndicator, error) {
	// Получаем ID категории по коду
	category, err := r.GetCategoryByCode(req.CategoryCode)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, fmt.Errorf("category with code %s not found", req.CategoryCode)
	}

	query := `
        UPDATE kpi_indicators 
        SET category_id = $1, code = $2, name = $3, description = $4, 
            unit = $5, indicator_type = $6, base_value = $7, base_weight = $8,
            extra_weight = $9, penalty_weight = $10
        WHERE id = $11
        RETURNING id, category_id, code, name, description, unit, 
                  indicator_type, base_value, base_weight, extra_weight, 
                  penalty_weight, created_at
    `

	var indicator models.KPIIndicator
	err = r.db.QueryRow(
		query,
		category.ID, req.Code, req.Name, req.Description, req.Unit,
		req.IndicatorType, req.BaseValue, req.BaseWeight,
		req.ExtraWeight, req.PenaltyWeight, id,
	).Scan(
		&indicator.ID, &indicator.CategoryID, &indicator.Code, &indicator.Name,
		&indicator.Description, &indicator.Unit, &indicator.IndicatorType,
		&indicator.BaseValue, &indicator.BaseWeight, &indicator.ExtraWeight,
		&indicator.PenaltyWeight, &indicator.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update indicator: %w", err)
	}

	return &indicator, nil
}

// DeleteIndicator - удаление показателя
func (r *KPIRepository) DeleteIndicator(id int) error {
	query := `DELETE FROM kpi_indicators WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete indicator: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("indicator not found")
	}

	return nil
}
