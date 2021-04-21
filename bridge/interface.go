package bridge

import (
	"context"
)

type Bridge interface {
	GetPendingDepositTransaction(context.Context) *DepositTransaction
	ProposeTransfer(context.Context, *DepositTransaction)
	ProposeSetStatusSuccessOnPendingTransfer(context.Context)
	ProposeSetStatusFailedOnPendingTransfer(context.Context)
	WasProposedTransfer(context.Context, Nonce) bool
	GetActionIdForProposeTransfer(context.Context, Nonce) ActionId
	WasProposedSetStatusSuccessOnPendingTransfer(context.Context) bool
	WasProposedSetStatusFailedOnPendingTransfer(context.Context) bool
	GetActionIdForSetStatusOnPendingTransfer(context.Context) ActionId
	WasExecuted(context.Context, ActionId) bool
	Sign(context.Context, ActionId)
	Execute(context.Context, ActionId) (string, error)
	SignersCount(context.Context, ActionId) uint
}
