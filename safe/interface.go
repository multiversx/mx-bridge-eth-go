package safe

import "context"

type Safe interface {
	GetTransactions(context.Context, uint64) chan *DepositTransaction
}
