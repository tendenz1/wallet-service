package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

const (
	minPortNumber = 1
	maxPortNumber = 65535
)

type Config struct {
	AppPort         int
	ShutdownTimeout time.Duration
	Postgres        PostgresConfig
}

type PostgresConfig struct {
	Host           string
	Port           int
	Database       string
	User           string
	Password       string
	MaxConns       int32
	ConnectTimeout time.Duration
}

func Load() (Config, error) {
	v := viper.New()
	v.SetConfigFile("config.env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("POSTGRES_HOST", "localhost")
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_DB", "wallet")
	v.SetDefault("POSTGRES_USER", "wallet")
	v.SetDefault("POSTGRES_PASSWORD", "wallet")
	v.SetDefault("DB_MAX_CONNS", 20)
	v.SetDefault("DB_CONNECT_TIMEOUT", "10s")
	v.SetDefault("SHUTDOWN_TIMEOUT", "10s")

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !os.IsNotExist(err) && !errors.As(err, &notFound) {
			return Config{}, fmt.Errorf("read config.env: %w", err)
		}
	}
	cfg := Config{
		AppPort:         v.GetInt("APP_PORT"),
		ShutdownTimeout: v.GetDuration("SHUTDOWN_TIMEOUT"),

		Postgres: PostgresConfig{
			Host:           v.GetString("POSTGRES_HOST"),
			Port:           v.GetInt("POSTGRES_PORT"),
			Database:       v.GetString("POSTGRES_DB"),
			User:           v.GetString("POSTGRES_USER"),
			Password:       v.GetString("POSTGRES_PASSWORD"),
			MaxConns:       v.GetInt32("DB_MAX_CONNS"),
			ConnectTimeout: v.GetDuration("DB_CONNECT_TIMEOUT"),
		},
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Address() string {
	return ":" + strconv.Itoa(c.AppPort)
}

func (c PostgresConfig) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   c.Database,
	}
	u.RawQuery = "sslmode=disable"
	return u.String()
}

func (c Config) validate() error {
	if c.AppPort < minPortNumber || c.AppPort > maxPortNumber {
		return fmt.Errorf("APP_PORT must be between %d and %d", minPortNumber, maxPortNumber)
	}
	if c.Postgres.Host == "" || c.Postgres.Database == "" || c.Postgres.User == "" {
		return fmt.Errorf("PostgreSQL host, database, and user must not be empty")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be greater than zero")
	}
	if c.Postgres.ConnectTimeout <= 0 {
		return fmt.Errorf("DB_CONNECT_TIMEOUT must be greater than zero")
	}
	if c.Postgres.Port < minPortNumber || c.Postgres.Port > maxPortNumber {
		return fmt.Errorf("POSTGRES_PORT must be between %d and %d", minPortNumber, maxPortNumber)
	}
	if c.Postgres.MaxConns < 1 {
		return fmt.Errorf("DB_MAX_CONNS must be greater than zero")
	}
	return nil
}
