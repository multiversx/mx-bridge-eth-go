package bridge

import (
	"context"
)

type Bridge interface {
	GetPendingDepositTransaction(context.Context) *DepositTransaction
	Propose(*DepositTransaction)
	WasProposed(*DepositTransaction) bool
	WasExecuted(*DepositTransaction) bool
	Sign(*DepositTransaction)
	Execute(*DepositTransaction) (string, error)
	SignersCount(*DepositTransaction) uint
}
