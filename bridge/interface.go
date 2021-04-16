package bridge

import (
	"context"
	"math/big"
)

type SafeTxChan chan *DepositTransaction

type Bridge interface {
	GetTransactions(context.Context, *big.Int, SafeTxChan)

	Bridge(*DepositTransaction) (string, error)
}
