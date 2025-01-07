package db

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"testing"

	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
	"go.uber.org/zap/zaptest"
)

func TestInitRedis(t *testing.T) {
	mockRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start mock Redis server: %v", err)
	}
	defer mockRedis.Close()

	cfg := &config.Config{
		Redis: config.Redis{
			URL: mockRedis.Addr(),
		},
	}

	logger := zaptest.NewLogger(t)

	rdb := InitRedis(cfg, logger)
	if rdb == nil {
		t.Fatalf("InitRedis returned nil")
	}

	ctx := context.Background()
	err = rdb.Set(ctx, "test_key", "test_value", 0).Err()
	if err != nil {
		t.Errorf("failed to set key in Redis: %v", err)
	}

	val, err := rdb.Get(ctx, "test_key").Result()
	if err != nil {
		t.Errorf("failed to get key from Redis: %v", err)
	}

	if val != "test_value" {
		t.Errorf("unexpected value from Redis: got %v, want %v", val, "test_value")
	}
}
