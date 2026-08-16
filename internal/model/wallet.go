package model

import "github.com/google/uuid"

type OperationType string

const (
	OperationDeposit  OperationType = "DEPOSIT"
	OperationWithdraw OperationType = "WITHDRAW"
)

type WalletOperationRequest struct {
	WalletID      uuid.UUID
	OperationType OperationType
	Amount        int64
}
