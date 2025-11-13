package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/Fixsbreaker/event-hub/backend/internal/db"
)

func main() {

	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load("docker/.env"); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	db.Connect()

	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "EventHub initialized!",
		})
	})

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}
