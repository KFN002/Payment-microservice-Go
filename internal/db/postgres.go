package db

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
	"gitlab.crja72.ru/gospec/go8/payment/internal/utils"
	"go.uber.org/zap"
)

func BuildPostgresDSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DB, cfg.Postgres.SSLMode)
}

// NewPostgres Подключение к постгрес
func NewPostgres(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*pgxpool.Pool, error) {
	dsn := BuildPostgresDSN(cfg)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse pool config: %v", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create connection pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Database ping failed: %v", err)
	}
	logger.Info("Succesfully connected to db")
	return pool, nil
}

// MigratePostgres миграции для постгрес
func MigratePostgres(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger, migrations fs.FS) error {
	goose.SetLogger(utils.GooseZapLogger(logger))
	goose.SetBaseFS(migrations)
	goose.SetDialect("postgres")
	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("Migration failed: %v", err)
	}
	logger.Info("Successfully applied migrations")
	return nil
}
