package clients

import (
	"context"
	"math/big"
)

// GasHandler defines the component able to fetch the current gas price
type GasHandler interface {
	GetCurrentGasPrice() (*big.Int, error)
	IsInterfaceNil() bool
}

// BatchValidator defines the operations for a component that can verify a batch
type BatchValidator interface {
	ValidateBatch(ctx context.Context, batch *TransferBatch) (bool, error)
	IsInterfaceNil() bool
}
