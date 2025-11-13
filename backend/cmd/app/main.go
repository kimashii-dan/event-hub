package main

import (
	"log"
	"os"

	"github.com/Fixsbreaker/event-hub/backend/internal/config"
	"github.com/Fixsbreaker/event-hub/backend/internal/db"
	"github.com/Fixsbreaker/event-hub/backend/internal/handler"
	"github.com/Fixsbreaker/event-hub/backend/internal/repository"
	"github.com/Fixsbreaker/event-hub/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("docker/.env"); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	config := config.Load()

	db.Connect()

	r := gin.Default()

	userRepo := repository.NewUserRepository(db.DB)

	authService := service.NewAuthService(userRepo, config.JWTSecret, config.JWTExpirationTime)

	handler.NewAuthHandler(r, authService)

	port := config.ServerPort
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}
