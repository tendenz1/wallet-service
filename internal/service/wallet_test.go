package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"wallet/internal/model"
	"wallet/internal/service"
	servicemocks "wallet/internal/service/mocks"
)

func TestWalletService_GetBalance(t *testing.T) {
	t.Parallel()

	walletID := uuid.New()
	repositoryError := errors.New("repository error")
	tests := []struct {
		name        string
		setupMock   func(*servicemocks.MockWalletRepository)
		wantBalance int64
		wantErr     error
	}{
		{
			name: "success",
			setupMock: func(repository *servicemocks.MockWalletRepository) {
				repository.EXPECT().GetBalance(gomock.Any(), walletID).Return(int64(1250), nil)
			},
			wantBalance: 1250,
		},
		{
			name: "repository error",
			setupMock: func(repository *servicemocks.MockWalletRepository) {
				repository.EXPECT().GetBalance(gomock.Any(), walletID).Return(int64(0), repositoryError)
			},
			wantErr: repositoryError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			controller := gomock.NewController(t)
			repository := servicemocks.NewMockWalletRepository(controller)
			tt.setupMock(repository)
			walletService := service.NewWalletService(repository)

			balance, err := walletService.GetBalance(context.Background(), walletID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantBalance, balance)
		})
	}
}

func TestWalletService_ApplyOperation(t *testing.T) {
	t.Parallel()

	walletID := uuid.New()
	repositoryError := errors.New("repository error")
	tests := []struct {
		name        string
		request     model.WalletOperationRequest
		setupMock   func(*servicemocks.MockWalletRepository)
		wantBalance int64
		wantErr     error
	}{
		{
			name:    "successful deposit",
			request: model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationDeposit, Amount: 100},
			setupMock: func(repository *servicemocks.MockWalletRepository) {
				repository.EXPECT().Deposit(gomock.Any(), walletID, int64(100)).Return(int64(1100), nil)
			},
			wantBalance: 1100,
		},
		{
			name:    "successful withdraw",
			request: model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationWithdraw, Amount: 100},
			setupMock: func(repository *servicemocks.MockWalletRepository) {
				repository.EXPECT().Withdraw(gomock.Any(), walletID, int64(100)).Return(int64(900), nil)
			},
			wantBalance: 900,
		},
		{
			name:    "invalid amount",
			request: model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationDeposit, Amount: -1},
			wantErr: model.ErrInvalidAmount,
		},
		{
			name:    "invalid operation",
			request: model.WalletOperationRequest{WalletID: walletID, OperationType: "TRANSFER", Amount: 100},
			wantErr: model.ErrInvalidOperation,
		},
		{
			name:    "repository error",
			request: model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationDeposit, Amount: 100},
			setupMock: func(repository *servicemocks.MockWalletRepository) {
				repository.EXPECT().Deposit(gomock.Any(), walletID, int64(100)).Return(int64(0), repositoryError)
			},
			wantErr: repositoryError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			controller := gomock.NewController(t)
			repository := servicemocks.NewMockWalletRepository(controller)
			if tt.setupMock != nil {
				tt.setupMock(repository)
			}
			walletService := service.NewWalletService(repository)

			balance, err := walletService.ApplyOperation(context.Background(), tt.request)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantBalance, balance)
		})
	}
}
