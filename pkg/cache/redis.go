package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/wind7vn/fnb_be/pkg/common/logger"
	"github.com/wind7vn/fnb_be/pkg/config"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {
	if config.AppConfig.RedisURL == "" {
		logger.Log.Warn("Redis URL is empty. Real-time features will be disabled.")
		return
	}

	opt, err := redis.ParseURL(config.AppConfig.RedisURL)
	if err != nil {
		logger.Log.Sugar().Fatalf("Failed to parse Redis URL: %v", err)
	}

	RedisClient = redis.NewClient(opt)

	_, err = RedisClient.Ping(Ctx).Result()
	if err != nil {
		logger.Log.Sugar().Fatalf("Failed to connect to Redis: %v", err)
	}

	logger.Log.Info("Connected to Redis System successfully!")
}
