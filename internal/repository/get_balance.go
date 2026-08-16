package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wallet/internal/model"
)

func (r *PostgresRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	const query = `
		SELECT balance
		FROM wallets
		WHERE id = @wallet_id
	`
	var balance int64
	if err := r.pool.QueryRow(
		ctx,
		query,
		pgx.NamedArgs{
			"wallet_id": walletID,
		},
	).Scan(&balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, model.ErrWalletNotFound
		}
		return 0, fmt.Errorf("query wallet balance: %w", err)
	}
	return balance, nil
}
