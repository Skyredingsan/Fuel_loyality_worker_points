package models

import (
	"encoding/json"
	"time"
)

type Level struct {
	ID               int    `db:"id" json:"id"`
	Name             string `db:"name" json:"name"` // Специалист Трассы, Тактик Магистрали, Стратег Гран-при
	MinPointsPerYear int    `db:"min_points_per_year" json:"min_points_per_year"`
	PrivilegesJSON   string `db:"privileges" json:"-"` // JSON as text in DB

	// Разбираем JSON при необходимости
	Privileges map[string]interface{} `db:"-" json:"privileges"`
	CreatedAt  time.Time              `db:"created_at" json:"created_at"`
}

// Для API
func (l *Level) UnmarshalPrivileges() error {
	if l.PrivilegesJSON == "" {
		l.Privileges = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal([]byte(l.PrivilegesJSON), &l.Privileges)
}

type UserLevelHistory struct {
	ID         int       `db:"id" json:"id"`
	UserID     int       `db:"user_id" json:"user_id"`
	LevelID    int       `db:"level_id" json:"level_id"`
	AssignedAt time.Time `db:"assigned_at" json:"assigned_at"`
	PointsYear int       `db:"points_year" json:"points_year"` // всего баллов за год на момент присвоения
	CreatedAt  time.Time `db:"created_at" json:"created_at"`

	// Joined
	Level *Level `db:"-" json:"level,omitempty"`
}
