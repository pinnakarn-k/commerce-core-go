package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pinnakarn-k/commerce-core-go/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
	log  *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) (*Postgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.DBMaxConns)
	poolConfig.MinConns = int32(cfg.DBMinConns)
	poolConfig.MaxConnLifetime = cfg.DBConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.DBConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	logger.Info("postgres connected",
		"max_conns", cfg.DBMaxConns,
		"min_conns", cfg.DBMinConns,
	)

	return &Postgres{
		Pool: pool,
		log:  logger,
	}, nil
}

func (p *Postgres) Close() {
	if p == nil || p.Pool == nil {
		return
	}

	if p.log != nil {
		p.log.Info("closing postgres connection")
	}

	p.Pool.Close()

	if p.log != nil {
		p.log.Info("postgres connection closed")
	}
}
