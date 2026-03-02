package models

import (
	"time"
)

type ResultStatus string

const (
	StatusDraft     ResultStatus = "draft"
	StatusConfirmed ResultStatus = "confirmed"
)

type MonthlyResult struct {
	ID        int          `db:"id" json:"id"`
	UserID    int          `db:"user_id" json:"user_id"`
	ExpertID  *int         `db:"expert_id" json:"expert_id"`
	Period    time.Time    `db:"period" json:"period"` // первый день месяца
	Status    ResultStatus `db:"status" json:"status"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`

	// Joined fields (не в БД)
	User   *User `db:"-" json:"user,omitempty"`
	Expert *User `db:"-" json:"expert,omitempty"`
}

type IndicatorResult struct {
	ID                    int       `db:"id" json:"id"`
	MonthlyResultID       int       `db:"monthly_result_id" json:"monthly_result_id"`
	IndicatorID           int       `db:"indicator_id" json:"indicator_id"`
	FactValue             *float64  `db:"fact_value" json:"fact_value"`
	CalculatedPoints      int       `db:"calculated_points" json:"calculated_points"`
	SupportingDocumentURL *string   `db:"supporting_document_url" json:"supporting_document_url,omitempty"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`

	// Вложенная структура для индикатора (используем db:"prefix.*" для маппинга)
	Indicator *KPIIndicator `db:"indicator" json:"indicator,omitempty"`
}

// Для ввода результатов экспертом
type EnterResultRequest struct {
	UserID           int                    `json:"user_id" validate:"required"`
	Period           string                 `json:"period" validate:"required"`       // формат: 2024-05
	IndicatorResults []IndicatorResultInput `json:"results" validate:"required,dive"` // Обрати внимание: json:"results" но поле IndicatorResults
}

type IndicatorResultInput struct {
	IndicatorCode string   `json:"indicator_code" validate:"required"`
	FactValue     *float64 `json:"fact_value"`
	DocumentURL   *string  `json:"document_url,omitempty"`
}

// Итоговый результат по категории для отображения в своде
type CategorySummary struct {
	CategoryCode  string `json:"category_code"`
	CategoryName  string `json:"category_name"`
	BasePoints    int    `json:"base_points"`
	ExtraPoints   int    `json:"extra_points"`
	PenaltyPoints int    `json:"penalty_points"`
	TotalPoints   int    `json:"total_points"`
}

type FullResultSummary struct {
	UserID          int               `json:"user_id"`
	UserFIO         string            `json:"user_fio"`
	Period          string            `json:"period"`
	Categories      []CategorySummary `json:"categories"`
	DetailedResults []IndicatorResult `json:"detailed_results,omitempty"` // Добавлено
	TotalPoints     int               `json:"total_points"`
	Level           *Level            `json:"level,omitempty"`
}
