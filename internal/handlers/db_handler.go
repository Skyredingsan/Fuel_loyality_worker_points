package handlers

import (
	"net/http"
	"os"
)

type DBHandler struct {
	dbPath string
}

func NewDBHandler(dbPath string) *DBHandler {
	return &DBHandler{dbPath: dbPath}
}

// DownloadDB - скачивание файла базы данных (только для координатора)
func (h *DBHandler) DownloadDB(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что файл существует
	if _, err := os.Stat(h.dbPath); os.IsNotExist(err) {
		http.Error(w, "Database file not found", http.StatusNotFound)
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=fuel-points.db")

	http.ServeFile(w, r, h.dbPath)
}
