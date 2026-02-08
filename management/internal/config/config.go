package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/makhtech/management/internal/repository/postgres"
	"github.com/makhtech/management/pkg/directories"
)

type Config struct {
	Env      string         `json:"env"`
	GRPC     GRPCConfig     `json:"grpc"`
	Database DatabaseConfig `json:"repository"`
}

type GRPCConfig struct {
	//Address string `json:"address"`
	Port int
}

type DatabaseConfig struct {
	Host              string `json:"host"`
	Port              string `json:"port"`
	User              string `json:"user"`
	Password          string `json:"password"`
	DBName            string `json:"db_name"`
	SSLMode           string `json:"ssl_mode"`
	MaxConns          int32  `json:"max_conns"`
	MinConns          int32  `json:"min_conns"`
	MaxConnLifetime   string `json:"max_conn_lifetime"`
	MaxConnIdleTime   string `json:"max_conn_idle_time"`
	HealthCheckPeriod string `json:"health_check_period"`
	ConnectTimeout    string `json:"connect_timeout"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file does not exist: %s", path))
	}

	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("failed to read config file: %s", err))
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("failed to parse config file: %s", err))
	}

	slog.Info("config loaded", slog.String("path", path))

	return &cfg
}

func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
		if path == "" {
			path = filepath.Join(directories.FindDirectoryName("config"), "local.json")
		}
	}

	return path
}

func (c *DatabaseConfig) ToPostgresConfig() *postgres.Config {
	return &postgres.Config{
		Host:              c.Host,
		Port:              c.Port,
		User:              c.User,
		Password:          c.Password,
		DBName:            c.DBName,
		SSLMode:           c.SSLMode,
		MaxConns:          c.MaxConns,
		MinConns:          c.MinConns,
		MaxConnLifetime:   parseDuration(c.MaxConnLifetime, time.Hour),
		MaxConnIdleTime:   parseDuration(c.MaxConnIdleTime, time.Minute*30),
		HealthCheckPeriod: parseDuration(c.HealthCheckPeriod, time.Minute),
		ConnectTimeout:    parseDuration(c.ConnectTimeout, time.Second*5),
	}
}

func parseDuration(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultVal
	}
	return d
}
