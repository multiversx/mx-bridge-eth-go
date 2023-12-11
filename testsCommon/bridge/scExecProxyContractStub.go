package bridge

import (
	"context"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// SCExecProxyContractStub -
type SCExecProxyContractStub struct {
	FilterLogsCalled func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

// FilterLogs -
func (stub *SCExecProxyContractStub) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if stub.FilterLogsCalled != nil {
		return stub.FilterLogsCalled(ctx, q)
	}

	return []types.Log{}, nil
}
