package interactors

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// ElrondProxyStub -
type ElrondProxyStub struct {
	GetNetworkConfigCalled func(ctx context.Context) (*data.NetworkConfig, error)
	SendTransactionCalled  func(ctx context.Context, transaction *data.Transaction) (string, error)
	SendTransactionsCalled func(ctx context.Context, txs []*data.Transaction) ([]string, error)
	ExecuteVMQueryCalled   func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
	GetAccountCalled       func(ctx context.Context, address core.AddressHandler) (*data.Account, error)
}

// GetNetworkConfig -
func (eps *ElrondProxyStub) GetNetworkConfig(ctx context.Context) (*data.NetworkConfig, error) {
	if eps.GetNetworkConfigCalled != nil {
		return eps.GetNetworkConfigCalled(ctx)
	}

	return &data.NetworkConfig{}, nil
}

// SendTransaction -
func (eps *ElrondProxyStub) SendTransaction(ctx context.Context, transaction *data.Transaction) (string, error) {
	if eps.SendTransactionCalled != nil {
		return eps.SendTransactionCalled(ctx, transaction)
	}

	return "", nil
}

// SendTransactions -
func (eps *ElrondProxyStub) SendTransactions(ctx context.Context, txs []*data.Transaction) ([]string, error) {
	if eps.SendTransactionCalled != nil {
		return eps.SendTransactionsCalled(ctx, txs)
	}

	return make([]string, 0), nil
}

// ExecuteVMQuery -
func (eps *ElrondProxyStub) ExecuteVMQuery(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
	if eps.ExecuteVMQueryCalled != nil {
		return eps.ExecuteVMQueryCalled(ctx, vmRequest)
	}

	return &data.VmValuesResponseData{}, nil
}

// GetAccount -
func (eps *ElrondProxyStub) GetAccount(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
	if eps.GetAccountCalled != nil {
		return eps.GetAccountCalled(ctx, address)
	}

	return &data.Account{}, nil
}

// IsInterfaceNil -
func (eps *ElrondProxyStub) IsInterfaceNil() bool {
	return eps == nil
}
