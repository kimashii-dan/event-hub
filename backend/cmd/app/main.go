package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Fixsbreaker/event-hub/backend/internal/cache"
	"github.com/Fixsbreaker/event-hub/backend/internal/config"
	"github.com/Fixsbreaker/event-hub/backend/internal/database"
	"github.com/Fixsbreaker/event-hub/backend/internal/handler"
	"github.com/Fixsbreaker/event-hub/backend/internal/middleware"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/Fixsbreaker/event-hub/backend/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("docker/.env"); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	// конфиг
	cfg := config.Load()

	// подключение к БД + миграции
	dbConn := database.Connect(cfg)

	// Redis
	if err := database.ConnectRedis(cfg); err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		// We might want to continue without cache or fail. For now, we continue but cache operations will fail/error out if not handled?
		// Actually cache.NewRedisCache takes the client. If connection failed Rdb might be nil or we should handle it.
		// database.ConnectRedis assigns to global Rdb.
	}
	redisCache := cache.NewRedisCache(database.Rdb)

	r := gin.Default()

	// глобальный логгер запросов
	r.Use(middleware.Logger())

	// простой health-check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "EventHub initialized!",
		})
	})

	// auth

	userRepo := repository.NewUserRepository(dbConn)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpirationTime)
	handler.NewAuthHandler(r, authService)

	// events

	eventRepo := repository.NewEventRepository(dbConn)
	eventService := service.NewEventService(eventRepo, redisCache)
	authMW := middleware.Auth(cfg.JWTSecret)

	handler.NewEventHandler(r, eventService, authMW)

	// registrations

	regRepo := repository.NewRegistrationRepository(dbConn)
	regService := service.NewRegistrationService(regRepo, eventRepo)
	handler.NewRegistrationHandler(r, regService, authMW)

	// users
	userService := service.NewUserService(userRepo)
	handler.NewUserHandler(r, userService, authMW)

	// notifications

	// Create Worker Pool for notifications (5 workers, buffer 100)
	notifWorkerPool := worker.NewWorkerPool(5, 100)
	notifWorkerPool.Start()
	defer notifWorkerPool.Stop() // Cleanup on exit

	notificationRepo := repository.NewNotificationRepository(dbConn)
	notificationService := service.NewNotificationService(notificationRepo, notifWorkerPool)
	handler.NewNotificationHandler(r, notificationService, authMW)

	// start server

	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
