package models

import (
	"time"
)

type UserRole string

const (
	RoleTM          UserRole = "tm"
	RoleExpert      UserRole = "expert"
	RoleCoordinator UserRole = "coordinator"
)

type User struct {
	ID           int       `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"` // "-" скрывает поле в JSON
	Role         UserRole  `db:"role" json:"role"`
	FIO          string    `db:"fio" json:"fio"`
	ClusterName  *string   `db:"cluster_name" json:"cluster_name,omitempty"`
	AZSCount     int       `db:"azs_count" json:"azs_count"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Для создания пользователя (без ID и дат)
type CreateUserRequest struct {
	Email       string   `json:"email" validate:"required,email"`
	Password    string   `json:"password" validate:"required,min=6"`
	Role        UserRole `json:"role" validate:"required,oneof=tm expert coordinator"`
	FIO         string   `json:"fio" validate:"required"`
	ClusterName *string  `json:"cluster_name"`
	AZSCount    int      `json:"azs_count"`
}

// Для обновления пользователя
type UpdateUserRequest struct {
	Email       *string   `json:"email,omitempty" validate:"omitempty,email"`
	Password    *string   `json:"password,omitempty" validate:"omitempty,min=6"`
	Role        *UserRole `json:"role,omitempty" validate:"omitempty,oneof=tm expert coordinator"`
	FIO         *string   `json:"fio,omitempty"`
	ClusterName *string   `json:"cluster_name,omitempty"`
	AZSCount    *int      `json:"azs_count,omitempty"`
}

// Для логина
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
