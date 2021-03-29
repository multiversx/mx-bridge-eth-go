package safe

import "context"

type SafeTxChan chan *DepositTransaction

type Safe interface {
	GetTransactions(context.Context, uint64) SafeTxChan
}
