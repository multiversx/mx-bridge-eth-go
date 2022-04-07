package clients

import "math/big"

// GasHandler defines the component able to fetch the current gas price
type GasHandler interface {
	GetCurrentGasPrice() (*big.Int, error)
	IsInterfaceNil() bool
}

// BatchValidator defines the operations for a component that can verify a batch
type BatchValidator interface {
	ValidateBatch(batch string) (bool, error)
	IsInterfaceNil() bool
}
