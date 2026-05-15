package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fuel-points/internal/auth"
	"fuel-points/internal/config"
	"fuel-points/internal/handlers"
	"fuel-points/internal/middleware"
	"fuel-points/internal/repository"
	"fuel-points/internal/services"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	//_ "fuel-points/docs"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Подключаемся к SQLite
	db, err := config.NewSQLiteDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Создаем таблицы если их нет
	err = runMigrations(db)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Инициализируем репозитории
	userRepo := repository.NewUserRepository(db)
	kpiRepo := repository.NewKPIRepository(db)
	resultRepo := repository.NewResultRepository(db)
	levelRepo := repository.NewLevelRepository(db)

	// Инициализируем JWT менеджер (токен живет 24 часа)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 24*time.Hour)

	// Инициализируем middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Создаем калькулятор баллов
	scoreCalculator := services.NewScoreCalculator(
		kpiRepo,
		resultRepo,
		userRepo,
		levelRepo,
	)

	// Инициализируем handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	userHandler := handlers.NewUserHandler(userRepo)
	kpiHandler := handlers.NewKPIHandler(kpiRepo)
	resultHandler := handlers.NewResultHandler(
		resultRepo,
		kpiRepo,
		userRepo,
		levelRepo,
		scoreCalculator,
	)
	levelHandler := handlers.NewLevelHandler(levelRepo, userRepo)
	uploadHandler := handlers.NewUploadHandler("./uploads", 10<<20) // 10MB max

	// Создаем роутер
	router := mux.NewRouter()

	// Глобальные middleware
	router.Use(middleware.CORS)
	router.Use(middleware.Logging)

	// Публичные маршруты (без авторизации)
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")
	router.HandleFunc("/api/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Swagger документация
	/*
		router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(

			httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
			httpSwagger.DeepLinking(true),
			httpSwagger.DocExpansion("none"),
			httpSwagger.DomID("swagger-ui"),
		))
	*/

	// Защищенные API маршруты
	api := router.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware.Authenticate) // Все маршруты в api требуют авторизации

	// Маршруты для работы с пользователями (доступны всем авторизованным)
	api.HandleFunc("/users/me", authHandler.GetCurrentUser).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/refresh", authHandler.RefreshToken).Methods("POST", "OPTIONS")
	api.HandleFunc("/tms", userHandler.GetTMsForExpert).Methods("GET", "OPTIONS")

	// Маршруты для KPI (доступны всем авторизованным)
	api.HandleFunc("/kpi/categories", kpiHandler.GetAllCategories).Methods("GET", "OPTIONS")
	api.HandleFunc("/kpi/indicators", kpiHandler.GetAllIndicators).Methods("GET", "OPTIONS")
	api.HandleFunc("/kpi/categories/{category}/indicators", kpiHandler.GetIndicatorsByCategory).Methods("GET", "OPTIONS")

	// Маршруты для результатов (доступны всем авторизованным)
	api.HandleFunc("/results/my", resultHandler.GetMyResults).Methods("GET", "OPTIONS")              // для ТМ
	api.HandleFunc("/results/user/{userId}", resultHandler.GetUserResults).Methods("GET", "OPTIONS") // для эксперта/координатора
	api.HandleFunc("/results/user/{userId}/yearly", resultHandler.GetYearlySummary).Methods("GET", "OPTIONS")

	// Маршруты для уровней (доступны всем авторизованным)
	api.HandleFunc("/levels", levelHandler.GetAllLevels).Methods("GET", "OPTIONS")
	api.HandleFunc("/levels/user/{userId}", levelHandler.GetUserCurrentLevel).Methods("GET", "OPTIONS")
	api.HandleFunc("/levels/user/{userId}/history", levelHandler.GetUserLevelHistory).Methods("GET", "OPTIONS")

	// Маршруты для загрузки файлов (доступны всем авторизованным)
	api.HandleFunc("/upload", uploadHandler.UploadFile).Methods("POST", "OPTIONS")
	api.HandleFunc("/upload/{type}/{filename}", uploadHandler.DownloadFile).Methods("GET", "OPTIONS")

	// Маршруты для координатора (требуется роль coordinator)
	// Управление пользователями
	api.Handle("/users",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(userHandler.GetAllUsers),
		),
	).Methods("GET", "OPTIONS")

	api.Handle("/users/register",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(authHandler.Register),
		),
	).Methods("POST", "OPTIONS")

	api.Handle("/users/{id}",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(userHandler.GetUserByID),
		),
	).Methods("GET", "OPTIONS")

	api.Handle("/users/{id}",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(userHandler.UpdateUser),
		),
	).Methods("PUT", "OPTIONS")

	api.Handle("/users/{id}",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(userHandler.DeleteUser),
		),
	).Methods("DELETE", "OPTIONS")

	// Управление KPI (только координатор)
	api.Handle("/kpi/indicators",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(kpiHandler.CreateIndicator),
		),
	).Methods("POST", "OPTIONS")

	api.Handle("/kpi/indicators/{id}",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(kpiHandler.UpdateIndicator),
		),
	).Methods("PUT", "OPTIONS")

	api.Handle("/kpi/indicators/{id}",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(kpiHandler.DeleteIndicator),
		),
	).Methods("DELETE", "OPTIONS")

	// Управление файлами (удаление только для координатора)
	api.Handle("/upload/{type}/{filename}",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(uploadHandler.DeleteFile),
		),
	).Methods("DELETE", "OPTIONS")
	// Ввод результатов (требуется роль expert или coordinator)
	api.Handle("/results/enter",
		authMiddleware.RequireRole("expert", "coordinator")(
			http.HandlerFunc(resultHandler.EnterResults),
		),
	).Methods("POST", "OPTIONS")

	api.Handle("/results/{id}/confirm",
		authMiddleware.RequireRole("expert", "coordinator")(
			http.HandlerFunc(resultHandler.ConfirmResults),
		),
	).Methods("POST", "OPTIONS")

	// Получение всех результатов за период (только координатор)
	api.Handle("/results",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(resultHandler.GetMonthlyResults),
		),
	).Methods("GET", "OPTIONS")

	api.HandleFunc("/results/{id}/indicators", resultHandler.GetDetailedResults).Methods("GET", "OPTIONS")
	// Получение результата по ID
	api.HandleFunc("/results/{id}", resultHandler.GetResultByID).Methods("GET", "OPTIONS")
	// Обновление результата
	api.Handle("/results/{id}",
		authMiddleware.RequireRole("expert", "coordinator")(
			http.HandlerFunc(resultHandler.UpdateResult),
		),
	).Methods("PUT", "OPTIONS")

	// Отклонение результата
	api.Handle("/results/{id}/reject",
		authMiddleware.RequireRole("coordinator")(
			http.HandlerFunc(resultHandler.RejectResults),
		),
	).Methods("POST", "OPTIONS")

	// После создания других handlers
	dbHandler := handlers.NewDBHandler(cfg.DBPath)

	router.HandleFunc("/api/db/download", dbHandler.DownloadDB).Methods("GET")

	// Статические файлы (для загрузок) - доступны публично
	router.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))),
	)

	// Настраиваем HTTP сервер
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		log.Printf("Available endpoints:")
		log.Printf("  GET  /health")
		log.Printf("  GET  /swagger/")
		log.Printf("  POST /api/login")
		log.Printf("  GET  /api/users/me (protected)")
		log.Printf("  POST /api/users/refresh (protected)")
		log.Printf("  GET  /api/tms (protected)")
		log.Printf("  GET  /api/kpi/categories (protected)")
		log.Printf("  GET  /api/kpi/indicators (protected)")
		log.Printf("  GET  /api/results/my (protected) - для ТМ")
		log.Printf("  GET  /api/levels (protected)")
		log.Printf("  POST /api/upload (protected)")
		log.Printf("  POST /api/results/enter (expert/coordinator)")
		log.Printf("  GET  /api/users (coordinator only)")
		log.Printf("  POST /api/users/register (coordinator only)")
		log.Printf("  POST /api/kpi/indicators (coordinator only)")
		log.Printf("  GET  /api/results (coordinator only)")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func runMigrations(db *sqlx.DB) error {
	// Читаем файл миграции
	migrationSQL, err := os.ReadFile("migrations/001_create_tables.sqlite.sql")
	if err != nil {
		return err
	}

	// Выполняем миграцию
	_, err = db.Exec(string(migrationSQL))
	return err
}
