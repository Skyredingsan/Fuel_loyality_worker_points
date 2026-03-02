package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fuel-points/internal/middleware"
	"fuel-points/internal/models"
	"fuel-points/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	userRepo *repository.UserRepository
	validate *validator.Validate
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		validate: validator.New(),
	}
}

// GetAllUsers - получение списка всех пользователей (для координатора)
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Получаем фильтр по роли из query параметра
	roleParam := r.URL.Query().Get("role")
	var role *models.UserRole
	if roleParam != "" {
		r := models.UserRole(roleParam)
		role = &r
	}

	users, err := h.userRepo.GetAll(role)
	if err != nil {
		http.Error(w, "Failed to get users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetUserByID - получение пользователя по ID
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(id)
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

// UpdateUser - обновление пользователя
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.Update(id, &req)
	if err != nil {
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// DeleteUser - удаление пользователя
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Не даем удалить самого себя
	currentUserID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if ok && currentUserID == id {
		http.Error(w, "Cannot delete yourself", http.StatusBadRequest)
		return
	}

	err = h.userRepo.Delete(id)
	if err != nil {
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTMsForExpert - получение списка ТМ для эксперта
func (h *UserHandler) GetTMsForExpert(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetTMsForExpert()
	if err != nil {
		http.Error(w, "Failed to get TMs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
