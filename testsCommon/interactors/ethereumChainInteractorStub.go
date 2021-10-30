package interactors

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

// EthereumChainInteractorStub -
type EthereumChainInteractorStub struct {
	GetRelayersCalled func(ctx context.Context) ([]common.Address, error)
}

// GetRelayers -
func (stub *EthereumChainInteractorStub) GetRelayers(ctx context.Context) ([]common.Address, error) {
	if stub.GetRelayersCalled != nil {
		return stub.GetRelayersCalled(ctx)
	}

	return nil, nil
}

// IsInterfaceNil -
func (stub *EthereumChainInteractorStub) IsInterfaceNil() bool {
	return stub == nil
}
