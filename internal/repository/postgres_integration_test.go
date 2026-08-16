//go:build integration

package repository

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"wallet/internal/model"
)

func TestPostgresRepositoryOperations(t *testing.T) {
	pool := testPool(t)
	repository := NewPostgresRepository(pool)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	walletID := createTestWallet(t, pool, 1000)

	balance, err := repository.GetBalance(ctx, walletID)
	require.NoError(t, err)
	require.Equal(t, int64(1000), balance)

	balance, err = repository.Deposit(ctx, walletID, 250)
	require.NoError(t, err)
	require.Equal(t, int64(1250), balance)

	balance, err = repository.Withdraw(ctx, walletID, 200)
	require.NoError(t, err)
	require.Equal(t, int64(1050), balance)

	_, err = repository.Withdraw(ctx, walletID, 2000)
	require.ErrorIs(t, err, model.ErrInsufficientFunds)

	missingID := uuid.New()
	_, err = repository.GetBalance(ctx, missingID)
	require.ErrorIs(t, err, model.ErrWalletNotFound)
	_, err = repository.Deposit(ctx, missingID, 1)
	require.ErrorIs(t, err, model.ErrWalletNotFound)
	_, err = repository.Withdraw(ctx, missingID, 1)
	require.ErrorIs(t, err, model.ErrWalletNotFound)
}

func TestPostgresRepositoryConcurrentDeposits(t *testing.T) {
	pool := testPool(t)
	repository := NewPostgresRepository(pool)
	walletID := createTestWallet(t, pool, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	const (
		operations = 1000
		amount     = 100
	)
	start := make(chan struct{})
	errorsChannel := make(chan error, operations)
	var waitGroup sync.WaitGroup
	waitGroup.Add(operations)

	for range operations {
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := repository.Deposit(ctx, walletID, amount)
			errorsChannel <- err
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsChannel)

	for err := range errorsChannel {
		require.NoError(t, err)
	}
	balance, err := repository.GetBalance(ctx, walletID)
	require.NoError(t, err)
	require.Equal(t, int64(operations*amount), balance)
}

func TestPostgresRepositoryConcurrentWithdrawals(t *testing.T) {
	pool := testPool(t)
	repository := NewPostgresRepository(pool)
	walletID := createTestWallet(t, pool, 500)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	const operations = 1000
	start := make(chan struct{})
	errorsChannel := make(chan error, operations)
	var waitGroup sync.WaitGroup
	var successful atomic.Int64
	waitGroup.Add(operations)

	for range operations {
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := repository.Withdraw(ctx, walletID, 1)
			if err == nil {
				successful.Add(1)
				return
			}
			errorsChannel <- err
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsChannel)

	insufficient := 0
	for err := range errorsChannel {
		require.ErrorIs(t, err, model.ErrInsufficientFunds)
		insufficient++
	}
	require.Equal(t, int64(500), successful.Load())
	require.Equal(t, 500, insufficient)
	balance, err := repository.GetBalance(ctx, walletID)
	require.NoError(t, err)
	require.Equal(t, int64(0), balance)
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)
	poolConfig.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, pool.Ping(ctx))
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wallets (
			id UUID PRIMARY KEY,
			balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
		)`)
	require.NoError(t, err)
	return pool
}

func createTestWallet(t *testing.T, pool *pgxpool.Pool, balance int64) uuid.UUID {
	t.Helper()
	walletID := uuid.New()
	_, err := pool.Exec(context.Background(), `INSERT INTO wallets (id, balance) VALUES ($1, $2)`, walletID, balance)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM wallets WHERE id = $1`, walletID)
	})
	return walletID
}
