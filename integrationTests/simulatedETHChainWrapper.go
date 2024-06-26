package integrationTests

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
)

const simulatedETHChainID = 1337

type simulatedETHChainWrapper struct {
	*backends.SimulatedBackend
}

// NewSimulatedETHChainWrapper returns a new instance of simulatedETHChainWrapper
func NewSimulatedETHChainWrapper(simulatedBackend *backends.SimulatedBackend) *simulatedETHChainWrapper {
	return &simulatedETHChainWrapper{
		simulatedBackend,
	}
}

// BlockNumber returns the current block number
func (wrapper *simulatedETHChainWrapper) BlockNumber(_ context.Context) (uint64, error) {
	return wrapper.Blockchain().CurrentBlock().Number.Uint64(), nil
}

// ChainID returns the default chain id for the simulated backend
func (wrapper *simulatedETHChainWrapper) ChainID(_ context.Context) (*big.Int, error) {
	return big.NewInt(simulatedETHChainID), nil
}
