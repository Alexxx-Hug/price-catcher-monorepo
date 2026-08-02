package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn) // почитай доку в функции, 17-18 строку можно тоже в GetDSN вынести
	if err != nil {
		return nil, fmt.Errorf("parse pgxpool config: %w", err)
	}

	cfg.MaxConns = 30 // вот это тоже должно быть в конфиге
	cfg.MinConns = 5  // вот это тоже должно быть в конфиге
	cfg.MaxConnLifetime = time.Minute * 30
	cfg.MaxConnIdleTime = time.Minute * 5

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
