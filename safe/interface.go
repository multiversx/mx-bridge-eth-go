package safe

import (
	"context"
	"math/big"
)

type SafeTxChan chan *DepositTransaction

type Safe interface {
	GetTransactions(context.Context, *big.Int, SafeTxChan)

	Bridge(*DepositTransaction)
}
