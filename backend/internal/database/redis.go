package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Fixsbreaker/event-hub/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func ConnectRedis(cfg *config.Config) error {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	Rdb = redis.NewClient(&redis.Options{
		Addr:        addr,
		Password:    "", // no password set
		DB:          0,  // use default DB
		DialTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	return nil
}
