package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	ServerPort string
	GinMode    string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// JWT
	JWTSecret          string
	JWTExpirationHours int
	JWTExpirationTime  time.Duration

	// Redis (optional for now)
	RedisHost string
	RedisPort string
}

func Load() *Config {
	jwtExpHours := getEnvAsInt("JWT_EXPIRATION_HOURS", 24)

	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		GinMode:    getEnv("GIN_MODE", "debug"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "eventhub"),

		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpirationHours: jwtExpHours,
		JWTExpirationTime:  time.Duration(jwtExpHours) * time.Hour,

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
