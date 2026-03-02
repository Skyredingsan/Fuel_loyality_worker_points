package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fuel-points/internal/models"
	"fuel-points/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type KPIHandler struct {
	kpiRepo  *repository.KPIRepository
	validate *validator.Validate
}

func NewKPIHandler(kpiRepo *repository.KPIRepository) *KPIHandler {
	return &KPIHandler{
		kpiRepo:  kpiRepo,
		validate: validator.New(),
	}
}

// GetAllCategories - получение всех категорий KPI
func (h *KPIHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.kpiRepo.GetAllCategories()
	if err != nil {
		http.Error(w, "Failed to get categories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// GetAllIndicators - получение всех показателей KPI
func (h *KPIHandler) GetAllIndicators(w http.ResponseWriter, r *http.Request) {
	indicators, err := h.kpiRepo.GetAllIndicators()
	if err != nil {
		http.Error(w, "Failed to get indicators: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indicators)
}

// GetIndicatorsByCategory - получение показателей по категории
func (h *KPIHandler) GetIndicatorsByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryCode := vars["category"]

	if categoryCode == "" {
		http.Error(w, "Category code is required", http.StatusBadRequest)
		return
	}

	indicators, err := h.kpiRepo.GetIndicatorsByCategory(categoryCode)
	if err != nil {
		http.Error(w, "Failed to get indicators: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indicators)
}

// CreateIndicator - создание нового показателя (только для координатора)
func (h *KPIHandler) CreateIndicator(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIndicatorRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	indicator, err := h.kpiRepo.CreateIndicator(&req)
	if err != nil {
		http.Error(w, "Failed to create indicator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(indicator)
}

// UpdateIndicator - обновление показателя (только для координатора)
func (h *KPIHandler) UpdateIndicator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid indicator ID", http.StatusBadRequest)
		return
	}

	var req models.CreateIndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	indicator, err := h.kpiRepo.UpdateIndicator(id, &req)
	if err != nil {
		http.Error(w, "Failed to update indicator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(indicator)
}

// DeleteIndicator - удаление показателя (только для координатора)
func (h *KPIHandler) DeleteIndicator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid indicator ID", http.StatusBadRequest)
		return
	}

	err = h.kpiRepo.DeleteIndicator(id)
	if err != nil {
		http.Error(w, "Failed to delete indicator: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
