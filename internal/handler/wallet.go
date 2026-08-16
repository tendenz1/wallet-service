package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"wallet/internal/model"
)

type WalletService interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
	ApplyOperation(ctx context.Context, request model.WalletOperationRequest) (int64, error)
}

type WalletHandler struct {
	service WalletService
	logger  *slog.Logger
}

type operationRequest struct {
	WalletID      string              `json:"walletId"`
	OperationType model.OperationType `json:"operationType"`
	Amount        int64               `json:"amount"`
}

type walletResponse struct {
	WalletID uuid.UUID `json:"walletId"`
	Balance  int64     `json:"balance"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewWalletHandler(service WalletService, logger *slog.Logger) *WalletHandler {
	return &WalletHandler{service: service, logger: logger}
}

func (h *WalletHandler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/wallet", h.applyOperation)
	mux.HandleFunc("GET /api/v1/wallets/{walletID}", h.getBalance)
	return mux
}

func (h *WalletHandler) applyOperation(w http.ResponseWriter, r *http.Request) {
	var body operationRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid walletId")
		return
	}
	balance, err := h.service.ApplyOperation(r.Context(), model.WalletOperationRequest{
		WalletID:      walletID,
		OperationType: body.OperationType,
		Amount:        body.Amount,
	})
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, walletResponse{WalletID: walletID, Balance: balance})
}

func (h *WalletHandler) getBalance(w http.ResponseWriter, r *http.Request) {
	walletID, err := uuid.Parse(r.PathValue("walletID"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid wallet UUID")
		return
	}

	balance, err := h.service.GetBalance(r.Context(), walletID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, walletResponse{WalletID: walletID, Balance: balance})
}

func (h *WalletHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrInvalidAmount), errors.Is(err, model.ErrInvalidOperation):
		h.writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, model.ErrWalletNotFound):
		h.writeError(w, http.StatusNotFound, model.ErrWalletNotFound.Error())
	case errors.Is(err, model.ErrInsufficientFunds):
		h.writeError(w, http.StatusConflict, model.ErrInsufficientFunds.Error())
	default:
		h.logger.Error("unexpected wallet service error", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *WalletHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, errorResponse{Error: message})
}

func (h *WalletHandler) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		h.logger.Error("encode HTTP response", "error", err)
	}
}
