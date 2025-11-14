package main

import (
	"log"
	"os"

	"github.com/Fixsbreaker/event-hub/backend/internal/config"
	"github.com/Fixsbreaker/event-hub/backend/internal/db"
	"github.com/Fixsbreaker/event-hub/backend/internal/handler"
	"github.com/Fixsbreaker/event-hub/backend/internal/middleware"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// load env
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("docker/.env"); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	cfg := config.Load()

	// set Gin mode
	gin.SetMode(cfg.GinMode)

	// connect DB
	db.Connect()

	// init router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	// repositories
	userRepo := repository.NewUserRepository(db.DB)
	eventRepo := repository.NewEventRepository(db.DB)

	// services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpirationTime)
	eventService := service.NewEventService(eventRepo)

	// public routes
	handler.NewAuthHandler(r, authService)

	// protected routes group (на будущее, но уже готово для middleware)
	api := r.Group("/api")
	api.Use(middleware.Auth(cfg.JWTSecret))

	// тут позже: registrationHandler, eventHandler и т.п.
	// handler.NewEventHandler(api, eventService)

	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
