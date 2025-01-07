package db

import (
	"testing"

	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
)

func TestBuildPostgresDSN(t *testing.T) {
	cfg := &config.Config{
		Postgres: config.Postgres{
			Host:     "localhost",
			Port:     5432,
			User:     "user",
			Password: "password",
			DB:       "testdb",
			SSLMode:  "disable",
		},
	}

	dsn := BuildPostgresDSN(cfg)
	expected := "host=localhost port=5432 user=user password=password dbname=testdb sslmode=disable"
	if dsn != expected {
		t.Errorf("unexpected DSN: got %v, want %v", dsn, expected)
	}
}
