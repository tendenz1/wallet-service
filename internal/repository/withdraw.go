package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wallet/internal/model"
)

func (r *PostgresRepository) Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) (int64, error) {
	const query = `
		UPDATE wallets
		SET balance = balance - @amount
		WHERE id = @wallet_id
		  AND balance >= @amount
		RETURNING balance
	`
	var balance int64
	if err := r.pool.QueryRow(
		ctx,
		query,
		pgx.NamedArgs{
			"wallet_id": walletID,
			"amount":    amount,
		},
	).Scan(&balance); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("withdraw funds: %w", err)
		}
		exists, existsErr := r.walletExists(ctx, walletID)
		if existsErr != nil {
			return 0, existsErr
		}
		if !exists {
			return 0, model.ErrWalletNotFound
		}
		return 0, model.ErrInsufficientFunds
	}
	return balance, nil
}

func (r *PostgresRepository) walletExists(ctx context.Context, walletID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM wallets
			WHERE id = @wallet_id
		)
	`
	var exists bool
	if err := r.pool.QueryRow(
		ctx,
		query,
		pgx.NamedArgs{
			"wallet_id": walletID,
		},
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check wallet existence: %w", err)
	}
	return exists, nil
}
