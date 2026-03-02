package models

import (
	"time"
)

type IndicatorType string

const (
	IndicatorTypeBase    IndicatorType = "base"
	IndicatorTypeExtra   IndicatorType = "extra"
	IndicatorTypePenalty IndicatorType = "penalty"
)

type KPICategory struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Code        string    `db:"code" json:"code"` // ПМ, ОЭК, ЭКЛ, КБ
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type KPIIndicator struct {
	ID            int           `db:"id" json:"id"`
	CategoryID    int           `db:"category_id" json:"category_id"`
	Code          string        `db:"code" json:"code"` // ПМ1, ДПМ1, ШОЭК и т.д.
	Name          string        `db:"name" json:"name"`
	Description   string        `db:"description" json:"description"`
	Unit          string        `db:"unit" json:"unit"` // %, шт, чел
	IndicatorType IndicatorType `db:"indicator_type" json:"indicator_type"`

	// Для базовых показателей
	BaseValue  *float64 `db:"base_value" json:"base_value,omitempty"`
	BaseWeight *int     `db:"base_weight" json:"base_weight,omitempty"`

	// Для дополнительных
	ExtraWeight *int `db:"extra_weight" json:"extra_weight,omitempty"`

	// Для штрафных
	PenaltyWeight *int `db:"penalty_weight" json:"penalty_weight,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// Поля для join-запросов (не в БД)
	CategoryCode string `db:"category_code" json:"category_code,omitempty"`
	CategoryName string `db:"category_name" json:"category_name,omitempty"`
}

// Для создания/обновления показателей (эксперт/координатор)
type CreateIndicatorRequest struct {
	Code          string        `json:"code" validate:"required"`
	Name          string        `json:"name" validate:"required"`
	CategoryCode  string        `json:"category_code" validate:"required,oneof=ПМ ОЭК ЭКЛ КБ"`
	Description   string        `json:"description"`
	Unit          string        `json:"unit" validate:"required"`
	IndicatorType IndicatorType `json:"indicator_type" validate:"required,oneof=base extra penalty"`

	BaseValue     *float64 `json:"base_value"`
	BaseWeight    *int     `json:"base_weight"`
	ExtraWeight   *int     `json:"extra_weight"`
	PenaltyWeight *int     `json:"penalty_weight"`
}
