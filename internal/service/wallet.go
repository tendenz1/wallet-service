package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wallet/internal/model"
)

type WalletRepository interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
	Deposit(ctx context.Context, walletID uuid.UUID, amount int64) (int64, error)
	Withdraw(ctx context.Context, walletID uuid.UUID, amount int64) (int64, error)
}

type WalletService struct {
	repository WalletRepository
}

func NewWalletService(repository WalletRepository) *WalletService {
	return &WalletService{repository: repository}
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	balance, err := s.repository.GetBalance(ctx, walletID)
	if err != nil {
		return 0, fmt.Errorf("get wallet balance: %w", err)
	}
	return balance, nil
}

func (s *WalletService) ApplyOperation(ctx context.Context, request model.WalletOperationRequest) (int64, error) {
	if request.Amount <= 0 {
		return 0, model.ErrInvalidAmount
	}
	var (
		balance int64
		err     error
	)
	switch request.OperationType {
	case model.OperationDeposit:
		balance, err = s.repository.Deposit(ctx, request.WalletID, request.Amount)
	case model.OperationWithdraw:
		balance, err = s.repository.Withdraw(ctx, request.WalletID, request.Amount)
	default:
		return 0, model.ErrInvalidOperation
	}
	if err != nil {
		return 0, fmt.Errorf("apply %s operation: %w", request.OperationType, err)
	}
	return balance, nil
}
