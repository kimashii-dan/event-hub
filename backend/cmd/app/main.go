package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// start server
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Stop Worker Pool gracefully
	log.Println("Stopping worker pool...")
	notifWorkerPool.Stop()

	log.Println("Server exiting")
}
