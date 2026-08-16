package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	setValidEnvironment(t)
	t.Setenv("APP_PORT", "9090")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("DB_MAX_CONNS", "15")
	t.Setenv("DB_CONNECT_TIMEOUT", "3s")
	t.Setenv("SHUTDOWN_TIMEOUT", "4s")

	cfg, err := Load()

	require.NoError(t, err)
	require.Equal(t, 9090, cfg.AppPort)
	require.Equal(t, 5433, cfg.Postgres.Port)
	require.Equal(t, int32(15), cfg.Postgres.MaxConns)
	require.Equal(t, 3*time.Second, cfg.Postgres.ConnectTimeout)
	require.Equal(t, 4*time.Second, cfg.ShutdownTimeout)
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*Config)
		wantError string
	}{
		{name: "valid config"},
		{
			name:      "app port below range",
			configure: func(cfg *Config) { cfg.AppPort = minPortNumber - 1 },
			wantError: "APP_PORT must be between 1 and 65535",
		},
		{
			name:      "app port above range",
			configure: func(cfg *Config) { cfg.AppPort = maxPortNumber + 1 },
			wantError: "APP_PORT must be between 1 and 65535",
		},
		{
			name:      "missing PostgreSQL host",
			configure: func(cfg *Config) { cfg.Postgres.Host = "" },
			wantError: "PostgreSQL host, database, and user must not be empty",
		},
		{
			name:      "PostgreSQL port below range",
			configure: func(cfg *Config) { cfg.Postgres.Port = minPortNumber - 1 },
			wantError: "POSTGRES_PORT must be between 1 and 65535",
		},
		{
			name:      "PostgreSQL port above range",
			configure: func(cfg *Config) { cfg.Postgres.Port = maxPortNumber + 1 },
			wantError: "POSTGRES_PORT must be between 1 and 65535",
		},
		{
			name:      "invalid maximum connections",
			configure: func(cfg *Config) { cfg.Postgres.MaxConns = 0 },
			wantError: "DB_MAX_CONNS must be greater than zero",
		},
		{
			name:      "invalid database connect timeout",
			configure: func(cfg *Config) { cfg.Postgres.ConnectTimeout = 0 },
			wantError: "DB_CONNECT_TIMEOUT must be greater than zero",
		},
		{
			name:      "invalid shutdown timeout",
			configure: func(cfg *Config) { cfg.ShutdownTimeout = 0 },
			wantError: "SHUTDOWN_TIMEOUT must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := validConfig()
			if tt.configure != nil {
				tt.configure(&cfg)
			}

			err := cfg.validate()

			if tt.wantError == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantError)
		})
	}
}

func TestPostgresConfig_DSN(t *testing.T) {
	t.Parallel()
	cfg := PostgresConfig{
		Host: "db.internal", Port: 5433, Database: "wallet db", User: "wallet-user", Password: "p@ss/word",
	}

	require.Equal(t, "postgres://wallet-user:p%40ss%2Fword@db.internal:5433/wallet%20db?sslmode=disable", cfg.DSN())
}

func validConfig() Config {
	return Config{
		AppPort:         8080,
		ShutdownTimeout: 10 * time.Second,
		Postgres: PostgresConfig{
			Host: "localhost", Port: 5432, Database: "wallet", User: "wallet", Password: "wallet",
			MaxConns: 20, ConnectTimeout: 10 * time.Second,
		},
	}
}

func setValidEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("APP_PORT", "8080")
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_DB", "wallet")
	t.Setenv("POSTGRES_USER", "wallet")
	t.Setenv("POSTGRES_PASSWORD", "wallet")
	t.Setenv("DB_MAX_CONNS", "20")
	t.Setenv("DB_CONNECT_TIMEOUT", "10s")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
}
