package clients

import (
	"math/big"
)

// GasHandler defines the component able to fetch the current gas price
type GasHandler interface {
	GetCurrentGasPrice() (*big.Int, error)
	Close() error
	IsInterfaceNil() bool
}
