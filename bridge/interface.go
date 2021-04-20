package bridge

import (
	"context"
)

type Bridge interface {
	GetPendingDepositTransaction(context.Context) *DepositTransaction
	Propose(context.Context, *DepositTransaction)
	WasProposed(context.Context, *DepositTransaction) bool
	WasExecuted(context.Context, *DepositTransaction) bool
	Sign(context.Context, *DepositTransaction)
	Execute(context.Context, *DepositTransaction) (string, error)
	SignersCount(context.Context, *DepositTransaction) uint
}
