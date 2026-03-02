package repository

import (
	"database/sql"
	"fmt"
	"time"

	"fuel-points/internal/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Создание пользователя
func (r *UserRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		FIO:          req.FIO,
		ClusterName:  req.ClusterName,
		AZSCount:     req.AZSCount,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
        INSERT INTO users (email, password_hash, role, fio, cluster_name, azs_count, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `

	err = r.db.QueryRow(
		query,
		user.Email, user.PasswordHash, user.Role, user.FIO,
		user.ClusterName, user.AZSCount, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Получение пользователя по email (для логина)
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User

	query := `SELECT id, email, password_hash, role, fio, cluster_name, azs_count, created_at, updated_at 
              FROM users WHERE email = $1`

	err := r.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // пользователь не найден
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Получение пользователя по ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var user models.User

	query := `SELECT id, email, password_hash, role, fio, cluster_name, azs_count, created_at, updated_at 
              FROM users WHERE id = $1`

	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// Получение всех пользователей с фильтром по роли
func (r *UserRepository) GetAll(role *models.UserRole) ([]models.User, error) {
	var users []models.User
	var query string
	var args []interface{}

	if role != nil {
		query = `SELECT id, email, role, fio, cluster_name, azs_count, created_at, updated_at 
                FROM users WHERE role = $1 ORDER BY fio`
		args = append(args, *role)
	} else {
		query = `SELECT id, email, role, fio, cluster_name, azs_count, created_at, updated_at 
                FROM users ORDER BY role, fio`
	}

	err := r.db.Select(&users, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return users, nil
}

// Получение ТМ для эксперта
func (r *UserRepository) GetTMsForExpert() ([]models.User, error) {
	var users []models.User

	query := `SELECT id, email, role, fio, cluster_name, azs_count, created_at, updated_at 
              FROM users WHERE role = 'tm' ORDER BY fio`

	err := r.db.Select(&users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get TMs: %w", err)
	}

	return users, nil
}

// Проверка пароля
func (r *UserRepository) CheckPassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

// Обновление пользователя
func (r *UserRepository) Update(id int, req *models.UpdateUserRequest) (*models.User, error) {
	// Сначала получаем текущего пользователя
	user, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Обновляем поля
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.FIO != nil {
		user.FIO = *req.FIO
	}
	if req.ClusterName != nil {
		user.ClusterName = req.ClusterName
	}
	if req.AZSCount != nil {
		user.AZSCount = *req.AZSCount
	}
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	user.UpdatedAt = time.Now()

	query := `UPDATE users 
              SET email = $1, password_hash = $2, role = $3, fio = $4, 
                  cluster_name = $5, azs_count = $6, updated_at = $7
              WHERE id = $8`

	_, err = r.db.Exec(query,
		user.Email, user.PasswordHash, user.Role, user.FIO,
		user.ClusterName, user.AZSCount, user.UpdatedAt, user.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// Удаление пользователя
func (r *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
