package repository

import (
	"database/sql"
	"fmt"

	"fuel-points/internal/models"

	"github.com/jmoiron/sqlx"
)

type LevelRepository struct {
	db *sqlx.DB
}

func NewLevelRepository(db *sqlx.DB) *LevelRepository {
	return &LevelRepository{db: db}
}

// GetAllLevels - получение всех уровней
func (r *LevelRepository) GetAllLevels() ([]models.Level, error) {
	var levels []models.Level
	query := `SELECT id, name, min_points_per_year, privileges, created_at FROM levels ORDER BY min_points_per_year`

	err := r.db.Select(&levels, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get levels: %w", err)
	}

	// Распарсиваем JSON привилегий
	for i := range levels {
		if err := levels[i].UnmarshalPrivileges(); err != nil {
			return nil, fmt.Errorf("failed to unmarshal privileges: %w", err)
		}
	}

	return levels, nil
}

// GetLevelByPoints - получение уровня по количеству баллов
func (r *LevelRepository) GetLevelByPoints(points int) (*models.Level, error) {
	var level models.Level
	query := `
        SELECT id, name, min_points_per_year, privileges, created_at 
        FROM levels 
        WHERE min_points_per_year <= $1 
        ORDER BY min_points_per_year DESC 
        LIMIT 1
    `

	err := r.db.Get(&level, query, points)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если нет подходящего уровня, возвращаем самый первый (минимальный)
			return r.GetLowestLevel()
		}
		return nil, fmt.Errorf("failed to get level by points: %w", err)
	}

	if err := level.UnmarshalPrivileges(); err != nil {
		return nil, err
	}

	return &level, nil
}

// GetLowestLevel - получение самого низкого уровня
func (r *LevelRepository) GetLowestLevel() (*models.Level, error) {
	var level models.Level
	query := `SELECT id, name, min_points_per_year, privileges, created_at FROM levels ORDER BY min_points_per_year LIMIT 1`

	err := r.db.Get(&level, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get lowest level: %w", err)
	}

	if err := level.UnmarshalPrivileges(); err != nil {
		return nil, err
	}

	return &level, nil
}

// AssignLevelToUser - присвоение уровня пользователю (запись в историю)
func (r *LevelRepository) AssignLevelToUser(userID int, levelID int, pointsYear int) error {
	query := `
        INSERT INTO user_level_history (user_id, level_id, assigned_at, points_year)
        VALUES ($1, $2, date('now'), $3)
    `

	_, err := r.db.Exec(query, userID, levelID, pointsYear)
	if err != nil {
		return fmt.Errorf("failed to assign level to user: %w", err)
	}

	return nil
}

// GetUserCurrentLevel - получение текущего уровня пользователя
func (r *LevelRepository) GetUserCurrentLevel(userID int) (*models.Level, error) {
	var level models.Level
	query := `
        SELECT l.id, l.name, l.min_points_per_year, l.privileges, l.created_at
        FROM levels l
        JOIN user_level_history ulh ON l.id = ulh.level_id
        WHERE ulh.user_id = $1
        ORDER BY ulh.assigned_at DESC
        LIMIT 1
    `

	err := r.db.Get(&level, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user current level: %w", err)
	}

	if err := level.UnmarshalPrivileges(); err != nil {
		return nil, err
	}

	return &level, nil
}

// GetUserLevelHistory - получение истории уровней пользователя
func (r *LevelRepository) GetUserLevelHistory(userID int) ([]models.UserLevelHistory, error) {
	var history []models.UserLevelHistory
	query := `
        SELECT ulh.id, ulh.user_id, ulh.level_id, ulh.assigned_at, ulh.points_year, ulh.created_at,
               l.name as level_name, l.min_points_per_year, l.privileges
        FROM user_level_history ulh
        JOIN levels l ON ulh.level_id = l.id
        WHERE ulh.user_id = $1
        ORDER BY ulh.assigned_at DESC
    `

	err := r.db.Select(&history, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user level history: %w", err)
	}

	// Распарсиваем JSON для каждого уровня в истории
	for i := range history {
		if history[i].Level != nil {
			if err := history[i].Level.UnmarshalPrivileges(); err != nil {
				return nil, err
			}
		}
	}

	return history, nil
}
