package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"fuel-points/internal/middleware"
	"fuel-points/internal/models"
	"fuel-points/internal/repository"
	"fuel-points/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type ResultHandler struct {
	resultRepo *repository.ResultRepository
	kpiRepo    *repository.KPIRepository
	userRepo   *repository.UserRepository
	levelRepo  *repository.LevelRepository
	calculator *services.ScoreCalculator
	validate   *validator.Validate
}

func NewResultHandler(
	resultRepo *repository.ResultRepository,
	kpiRepo *repository.KPIRepository,
	userRepo *repository.UserRepository,
	levelRepo *repository.LevelRepository,
	calculator *services.ScoreCalculator,
) *ResultHandler {
	return &ResultHandler{
		resultRepo: resultRepo,
		kpiRepo:    kpiRepo,
		userRepo:   userRepo,
		levelRepo:  levelRepo,
		calculator: calculator,
		validate:   validator.New(),
	}
}

// EnterResults - ввод результатов за месяц (для эксперта)
func (h *ResultHandler) EnterResults(w http.ResponseWriter, r *http.Request) {
	var req models.EnterResultRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		http.Error(w, "Validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем ID эксперта из контекста
	expertID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Парсим период
	period, err := time.Parse("2006-01", req.Period)
	if err != nil {
		http.Error(w, "Invalid period format. Use YYYY-MM", http.StatusBadRequest)
		return
	}

	// Обрабатываем результаты через калькулятор - используем IndicatorResults
	monthlyResult, err := h.calculator.ProcessMonthlyResults(
		req.UserID,
		expertID,
		period,
		req.IndicatorResults, // Здесь должно быть IndicatorResults
	)
	if err != nil {
		http.Error(w, "Failed to process results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monthlyResult)
}

// ConfirmResults - подтверждение результатов (перевод в статус confirmed)
func (h *ResultHandler) ConfirmResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resultID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid result ID", http.StatusBadRequest)
		return
	}

	err = h.resultRepo.ConfirmMonthlyResult(resultID)
	if err != nil {
		http.Error(w, "Failed to confirm results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем результат для расчета уровня
	// TODO: Здесь можно добавить логику автоматического присвоения уровня
	// при достижении определенного количества баллов

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Results confirmed successfully",
	})
}

// GetUserResults - получение результатов пользователя за месяц
func (h *ResultHandler) GetUserResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		http.Error(w, "Period parameter is required (YYYY-MM)", http.StatusBadRequest)
		return
	}

	period, err := time.Parse("2006-01", periodStr)
	if err != nil {
		http.Error(w, "Invalid period format. Use YYYY-MM", http.StatusBadRequest)
		return
	}

	// Получаем сводный отчет
	summary, err := h.calculator.GetFullSummary(userID, period)
	if err != nil {
		http.Error(w, "Failed to get results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetMyResults - получение своих результатов (для ТМ)
func (h *ResultHandler) GetMyResults(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		// Если период не указан, берем текущий месяц
		periodStr = time.Now().Format("2006-01")
	}

	period, err := time.Parse("2006-01", periodStr)
	if err != nil {
		http.Error(w, "Invalid period format. Use YYYY-MM", http.StatusBadRequest)
		return
	}

	summary, err := h.calculator.GetFullSummary(userID, period)
	if err != nil {
		// Если нет результатов, возвращаем пустой объект, а не 500
		if err.Error() == fmt.Sprintf("no results found for user %d in period %s", userID, period.Format("2006-01")) {
			emptySummary := &models.FullResultSummary{
				UserID:      userID,
				UserFIO:     "",
				Period:      period.Format("2006-01"),
				Categories:  []models.CategorySummary{},
				TotalPoints: 0,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(emptySummary)
			return
		}
		http.Error(w, "Failed to get results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetMonthlyResults - получение всех результатов за период (для координатора)
func (h *ResultHandler) GetMonthlyResults(w http.ResponseWriter, r *http.Request) {
	periodStr := r.URL.Query().Get("period")
	if periodStr == "" {
		http.Error(w, "Period parameter is required (YYYY-MM)", http.StatusBadRequest)
		return
	}

	period, err := time.Parse("2006-01", periodStr)
	if err != nil {
		http.Error(w, "Invalid period format. Use YYYY-MM", http.StatusBadRequest)
		return
	}

	results, err := h.resultRepo.GetMonthlyResultsByPeriod(period)
	if err != nil {
		http.Error(w, "Failed to get results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Обогащаем данными пользователей
	var enrichedResults []map[string]interface{}
	for _, result := range results {
		user, _ := h.userRepo.GetByID(result.UserID)
		expert, _ := h.userRepo.GetByID(*result.ExpertID)

		enrichedResult := map[string]interface{}{
			"id":         result.ID,
			"period":     result.Period.Format("2006-01"),
			"status":     result.Status,
			"user":       user,
			"expert":     expert,
			"created_at": result.CreatedAt,
		}

		// Получаем результаты по показателям
		indicators, _ := h.resultRepo.GetIndicatorResults(result.ID)
		enrichedResult["indicators"] = indicators

		enrichedResults = append(enrichedResults, enrichedResult)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrichedResults)
}

// GetYearlySummary - получение годового summary для пользователя
func (h *ResultHandler) GetYearlySummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	// Получаем все результаты за год
	results, err := h.resultRepo.GetUserResultsForYear(userID, year)
	if err != nil {
		http.Error(w, "Failed to get yearly results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем общее количество баллов за год
	totalPoints, err := h.resultRepo.GetTotalPointsForYear(userID, year)
	if err != nil {
		http.Error(w, "Failed to calculate total points: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем уровень
	level, err := h.levelRepo.GetLevelByPoints(totalPoints)
	if err != nil {
		http.Error(w, "Failed to get level: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id":      userID,
		"year":         year,
		"total_points": totalPoints,
		"level":        level,
		"months":       results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDetailedResults - получение детальных результатов
func (h *ResultHandler) GetDetailedResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resultID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid result ID", http.StatusBadRequest)
		return
	}

	results, err := h.resultRepo.GetIndicatorResults(resultID)
	if err != nil {
		http.Error(w, "Failed to get detailed results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GetResultByID - получение результата по ID для редактирования
func (h *ResultHandler) GetResultByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resultID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid result ID", http.StatusBadRequest)
		return
	}

	// Получаем monthly result
	monthlyResult, err := h.resultRepo.GetMonthlyResultByID(resultID)
	if err != nil {
		http.Error(w, "Failed to get result: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if monthlyResult == nil {
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}

	// Получаем детальные результаты
	indicators, err := h.resultRepo.GetIndicatorResults(resultID)
	if err != nil {
		http.Error(w, "Failed to get indicators: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем информацию о пользователе
	user, err := h.userRepo.GetByID(monthlyResult.UserID)
	if err != nil {
		http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := map[string]interface{}{
		"id":         monthlyResult.ID,
		"user_id":    monthlyResult.UserID,
		"user":       user,
		"period":     monthlyResult.Period.Format("2006-01"),
		"status":     monthlyResult.Status,
		"indicators": indicators,
		"created_at": monthlyResult.CreatedAt,
		"updated_at": monthlyResult.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RejectResults - отклонение результатов
func (h *ResultHandler) RejectResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resultID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid result ID", http.StatusBadRequest)
		return
	}

	// Парсим тело запроса с причиной
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		http.Error(w, "Reason is required", http.StatusBadRequest)
		return
	}

	// Получаем результат
	monthlyResult, err := h.resultRepo.GetMonthlyResultByID(resultID)
	if err != nil {
		http.Error(w, "Failed to get result: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if monthlyResult == nil {
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}

	// Проверяем, что это черновик
	if monthlyResult.Status != models.StatusDraft {
		http.Error(w, "Only draft results can be rejected", http.StatusBadRequest)
		return
	}

	// Удаляем результаты
	err = h.resultRepo.DeleteMonthlyResult(resultID)
	if err != nil {
		http.Error(w, "Failed to reject results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Отправить уведомление эксперту с причиной отклонения

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Results rejected successfully",
		"reason":  req.Reason,
	})
}

// UpdateResult - обновление результатов (для черновиков)
func (h *ResultHandler) UpdateResult(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resultID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid result ID", http.StatusBadRequest)
		return
	}

	var req models.EnterResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if req.UserID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	if req.Period == "" {
		http.Error(w, "Period is required", http.StatusBadRequest)
		return
	}
	if len(req.IndicatorResults) == 0 { // Используем IndicatorResults
		http.Error(w, "At least one result is required", http.StatusBadRequest)
		return
	}

	// Получаем ID эксперта из контекста
	expertID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем существующий результат
	existingResult, err := h.resultRepo.GetMonthlyResultByID(resultID)
	if err != nil {
		http.Error(w, "Failed to get result: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if existingResult == nil {
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}

	// Проверяем, что это черновик
	if existingResult.Status != models.StatusDraft {
		http.Error(w, "Only draft results can be updated", http.StatusBadRequest)
		return
	}

	// Парсим период
	period, err := time.Parse("2006-01", req.Period)
	if err != nil {
		http.Error(w, "Invalid period format. Use YYYY-MM", http.StatusBadRequest)
		return
	}

	// Удаляем старые результаты
	err = h.resultRepo.DeleteIndicatorResults(resultID)
	if err != nil {
		http.Error(w, "Failed to delete old results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Создаем новые результаты - используем IndicatorResults
	_, err = h.calculator.ProcessMonthlyResults(req.UserID, expertID, period, req.IndicatorResults)
	if err != nil {
		http.Error(w, "Failed to process results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Results updated successfully",
	})
}
