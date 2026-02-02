package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns          = 25
	defaultMinConns          = 5
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = time.Minute * 30
	defaultHealthCheckPeriod = time.Minute
	defaultConnectTimeout    = time.Second * 5
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string

	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

type Database struct {
	Pool *pgxpool.Pool
	cfg  *Config
}

func New(ctx context.Context, cfg *Config) (*Database, error) {
	cfg.setDefaults()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to parse config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	poolConfig.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to create pool: %w", err)
	}

	db := &Database{
		Pool: pool,
		cfg:  cfg,
	}

	if err := db.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: failed to ping: %w", err)
	}

	slog.Info("postgres: connected successfully",
		slog.String("host", cfg.Host),
		slog.String("port", cfg.Port),
		slog.String("database", cfg.DBName),
	)

	return db, nil
}

func (c *Config) setDefaults() {
	if c.MaxConns == 0 {
		c.MaxConns = defaultMaxConns
	}
	if c.MinConns == 0 {
		c.MinConns = defaultMinConns
	}
	if c.MaxConnLifetime == 0 {
		c.MaxConnLifetime = defaultMaxConnLifetime
	}
	if c.MaxConnIdleTime == 0 {
		c.MaxConnIdleTime = defaultMaxConnIdleTime
	}
	if c.HealthCheckPeriod == 0 {
		c.HealthCheckPeriod = defaultHealthCheckPeriod
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = defaultConnectTimeout
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
}

func (d *Database) Ping(ctx context.Context) error {
	return d.Pool.Ping(ctx)
}

func (d *Database) Close() {
	if d.Pool != nil {
		d.Pool.Close()
		slog.Info("postgres: connection closed")
	}
}

func (d *Database) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := d.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: health check failed: %w", err)
	}

	return nil
}

func (d *Database) Stats() *pgxpool.Stat {
	return d.Pool.Stat()
}
