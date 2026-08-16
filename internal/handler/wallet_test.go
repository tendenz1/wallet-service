package handler_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"wallet/internal/handler"
	handlermocks "wallet/internal/handler/mocks"
	"wallet/internal/model"
)

func TestWalletHandler_ApplyOperation(t *testing.T) {
	t.Parallel()

	walletID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	serviceError := errors.New("service error")
	tests := []struct {
		name       string
		body       string
		setupMock  func(*handlermocks.MockWalletService)
		wantStatus int
		wantJSON   string
	}{
		{
			name: "success",
			body: `{"walletId":"00000000-0000-0000-0000-000000000001","operationType":"DEPOSIT","amount":100}`,
			setupMock: func(service *handlermocks.MockWalletService) {
				request := model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationDeposit, Amount: 100}
				service.EXPECT().ApplyOperation(gomock.Any(), request).Return(int64(1100), nil)
			},
			wantStatus: http.StatusOK,
			wantJSON:   `{"walletId":"00000000-0000-0000-0000-000000000001","balance":1100}`,
		},
		{
			name:       "malformed JSON",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantJSON:   `{"error":"invalid request body"}`,
		},
		{
			name:       "invalid wallet UUID",
			body:       `{"walletId":"invalid","operationType":"DEPOSIT","amount":100}`,
			wantStatus: http.StatusBadRequest,
			wantJSON:   `{"error":"invalid walletId"}`,
		},
		{
			name: "invalid amount",
			body: `{"walletId":"00000000-0000-0000-0000-000000000001","operationType":"DEPOSIT","amount":0}`,
			setupMock: func(service *handlermocks.MockWalletService) {
				request := model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationDeposit, Amount: 0}
				service.EXPECT().ApplyOperation(gomock.Any(), request).Return(int64(0), model.ErrInvalidAmount)
			},
			wantStatus: http.StatusBadRequest,
			wantJSON:   `{"error":"amount must be greater than zero"}`,
		},
		{
			name: "wallet not found",
			body: `{"walletId":"00000000-0000-0000-0000-000000000001","operationType":"DEPOSIT","amount":100}`,
			setupMock: func(service *handlermocks.MockWalletService) {
				request := model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationDeposit, Amount: 100}
				service.EXPECT().ApplyOperation(gomock.Any(), request).Return(int64(0), model.ErrWalletNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantJSON:   `{"error":"wallet not found"}`,
		},
		{
			name: "insufficient funds",
			body: `{"walletId":"00000000-0000-0000-0000-000000000001","operationType":"WITHDRAW","amount":100}`,
			setupMock: func(service *handlermocks.MockWalletService) {
				request := model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationWithdraw, Amount: 100}
				service.EXPECT().ApplyOperation(gomock.Any(), request).Return(int64(0), model.ErrInsufficientFunds)
			},
			wantStatus: http.StatusConflict,
			wantJSON:   `{"error":"insufficient funds"}`,
		},
		{
			name: "unexpected service error",
			body: `{"walletId":"00000000-0000-0000-0000-000000000001","operationType":"DEPOSIT","amount":100}`,
			setupMock: func(service *handlermocks.MockWalletService) {
				request := model.WalletOperationRequest{WalletID: walletID, OperationType: model.OperationDeposit, Amount: 100}
				service.EXPECT().ApplyOperation(gomock.Any(), request).Return(int64(0), serviceError)
			},
			wantStatus: http.StatusInternalServerError,
			wantJSON:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			controller := gomock.NewController(t)
			service := handlermocks.NewMockWalletService(controller)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", strings.NewReader(tt.body))

			newTestHandler(service).Router().ServeHTTP(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.JSONEq(t, tt.wantJSON, recorder.Body.String())
		})
	}
}

func TestWalletHandler_GetBalance(t *testing.T) {
	t.Parallel()

	walletID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	serviceError := errors.New("service error")
	tests := []struct {
		name       string
		path       string
		setupMock  func(*handlermocks.MockWalletService)
		wantStatus int
		wantJSON   string
	}{
		{
			name: "success",
			path: "/api/v1/wallets/" + walletID.String(),
			setupMock: func(service *handlermocks.MockWalletService) {
				service.EXPECT().GetBalance(gomock.Any(), walletID).Return(int64(1200), nil)
			},
			wantStatus: http.StatusOK,
			wantJSON:   `{"walletId":"00000000-0000-0000-0000-000000000001","balance":1200}`,
		},
		{
			name:       "invalid UUID",
			path:       "/api/v1/wallets/invalid",
			wantStatus: http.StatusBadRequest,
			wantJSON:   `{"error":"invalid wallet UUID"}`,
		},
		{
			name: "wallet not found",
			path: "/api/v1/wallets/" + walletID.String(),
			setupMock: func(service *handlermocks.MockWalletService) {
				service.EXPECT().GetBalance(gomock.Any(), walletID).Return(int64(0), model.ErrWalletNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantJSON:   `{"error":"wallet not found"}`,
		},
		{
			name: "unexpected service error",
			path: "/api/v1/wallets/" + walletID.String(),
			setupMock: func(service *handlermocks.MockWalletService) {
				service.EXPECT().GetBalance(gomock.Any(), walletID).Return(int64(0), serviceError)
			},
			wantStatus: http.StatusInternalServerError,
			wantJSON:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			controller := gomock.NewController(t)
			service := handlermocks.NewMockWalletService(controller)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)

			newTestHandler(service).Router().ServeHTTP(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.JSONEq(t, tt.wantJSON, recorder.Body.String())
		})
	}
}

func newTestHandler(service handler.WalletService) *handler.WalletHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return handler.NewWalletHandler(service, logger)
}
