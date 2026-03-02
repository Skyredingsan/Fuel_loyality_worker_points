package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fuel-points/internal/repository"

	"github.com/gorilla/mux"
)

type LevelHandler struct {
	levelRepo *repository.LevelRepository
	userRepo  *repository.UserRepository
}

func NewLevelHandler(levelRepo *repository.LevelRepository, userRepo *repository.UserRepository) *LevelHandler {
	return &LevelHandler{
		levelRepo: levelRepo,
		userRepo:  userRepo,
	}
}

// GetAllLevels - получение всех уровней
func (h *LevelHandler) GetAllLevels(w http.ResponseWriter, r *http.Request) {
	levels, err := h.levelRepo.GetAllLevels()
	if err != nil {
		http.Error(w, "Failed to get levels: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(levels)
}

// GetUserCurrentLevel - получение текущего уровня пользователя
func (h *LevelHandler) GetUserCurrentLevel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	level, err := h.levelRepo.GetUserCurrentLevel(userID)
	if err != nil {
		http.Error(w, "Failed to get user level: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if level == nil {
		// Если уровень не найден, берем минимальный
		level, err = h.levelRepo.GetLowestLevel()
		if err != nil {
			http.Error(w, "Failed to get lowest level: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(level)
}

// GetUserLevelHistory - получение истории уровней пользователя
func (h *LevelHandler) GetUserLevelHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	history, err := h.levelRepo.GetUserLevelHistory(userID)
	if err != nil {
		http.Error(w, "Failed to get level history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
