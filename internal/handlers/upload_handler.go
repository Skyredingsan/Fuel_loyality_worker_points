package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type UploadHandler struct {
	uploadDir string
	maxSize   int64 // максимальный размер файла в байтах
}

func NewUploadHandler(uploadDir string, maxSize int64) *UploadHandler {
	// Создаем директорию для загрузок, если её нет
	os.MkdirAll(uploadDir, 0755)

	return &UploadHandler{
		uploadDir: uploadDir,
		maxSize:   maxSize,
	}
}

// UploadFile - загрузка файла
func (h *UploadHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер загружаемого файла
	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)

	// Парсим multipart form
	err := r.ParseMultipartForm(h.maxSize)
	if err != nil {
		http.Error(w, "File too large. Max size: 10MB", http.StatusBadRequest)
		return
	}

	// Получаем файл из формы
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Получаем тип документа из формы
	docType := r.FormValue("type") // indicator_result, etc
	if docType == "" {
		docType = "general"
	}

	// Получаем ID связанной сущности
	entityID := r.FormValue("entity_id")

	// Валидация расширения файла
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
	}

	if !allowedExts[ext] {
		http.Error(w, "File type not allowed. Allowed: PDF, JPEG, PNG, DOC, DOCX, XLS, XLSX", http.StatusBadRequest)
		return
	}

	// Создаем уникальное имя файла
	filename := fmt.Sprintf("%d_%s_%s%s",
		time.Now().UnixNano(),
		docType,
		strings.ReplaceAll(entityID, "/", "_"),
		ext,
	)

	// Создаем поддиректорию по типу документа
	subDir := filepath.Join(h.uploadDir, docType)
	os.MkdirAll(subDir, 0755)

	// Полный путь к файлу
	filePath := filepath.Join(subDir, filename)

	// Создаем файл на диске
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Копируем содержимое
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем URL для доступа к файлу
	fileURL := fmt.Sprintf("/uploads/%s/%s", docType, filename)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"url":       fileURL,
		"filename":  header.Filename,
		"size":      fmt.Sprintf("%d", header.Size),
		"mime_type": header.Header.Get("Content-Type"),
	})
}

// DownloadFile - скачивание файла
func (h *UploadHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docType := vars["type"]
	filename := vars["filename"]

	// Безопасность: проверяем, что путь не выходит за пределы uploadDir
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.uploadDir, docType, filename)

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Отдаем файл
	http.ServeFile(w, r, filePath)
}

// DeleteFile - удаление файла
func (h *UploadHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docType := vars["type"]
	filename := vars["filename"]

	// Безопасность: проверяем, что путь не выходит за пределы uploadDir
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.uploadDir, docType, filename)

	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete file: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
