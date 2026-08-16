package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wallet/internal/model"
)

func (r *PostgresRepository) Deposit(ctx context.Context, walletID uuid.UUID, amount int64) (int64, error) {
	const query = `
		UPDATE wallets
		SET balance = balance + @amount
		WHERE id = @wallet_id
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
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, model.ErrWalletNotFound
		}
		return 0, fmt.Errorf("deposit funds: %w", err)
	}
	return balance, nil
}
