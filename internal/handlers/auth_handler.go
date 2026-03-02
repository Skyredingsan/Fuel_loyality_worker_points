package handlers

import (
	"encoding/json"
	"fuel-points/internal/auth"
	"fuel-points/internal/middleware"
	"fuel-points/internal/models"
	"fuel-points/internal/repository"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
	validate   *validator.Validate
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		validate:   validator.New(),
	}
}

// Login - обработчик входа в систему
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	// Декодируем JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Валидируем запрос
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ищем пользователя по email
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if !h.userRepo.CheckPassword(user, req.Password) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Генерируем JWT токен
	token, err := h.jwtManager.Generate(user.ID, user.Email, string(user.Role))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ (без хеша пароля)
	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Register - обработчик регистрации (только для координатора)
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем, не существует ли уже пользователь с таким email
	existingUser, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Создаем пользователя
	user, err := h.userRepo.Create(&req)
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// GetCurrentUser - получение информации о текущем пользователе
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста (установлено middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// RefreshToken - обновление токена
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Получаем текущего пользователя из контекста
	userID, _ := r.Context().Value(middleware.UserIDKey).(int)
	email, _ := r.Context().Value(middleware.UserEmailKey).(string)
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)

	// Генерируем новый токен
	token, err := h.jwtManager.Generate(userID, email, role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
