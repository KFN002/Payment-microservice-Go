package db

import (
	"github.com/go-redis/redis/v8"
	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
	"go.uber.org/zap"
)

// InitRedis создание редиски
func InitRedis(cfg *config.Config, logger *zap.Logger) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.URL,
	})
	logger.Info("Redis connected")
	return rdb
}
