package roleProvider

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// ElrondChainInteractor defines an Elrond client able to respond to VM queries
type ElrondChainInteractor interface {
	ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error)
	IsInterfaceNil() bool
}

// EthereumChainInteractor defines an Ethereum client able to respond to requests
type EthereumChainInteractor interface {
	GetRelayers(ctx context.Context) ([]common.Address, error)
	IsInterfaceNil() bool
}
