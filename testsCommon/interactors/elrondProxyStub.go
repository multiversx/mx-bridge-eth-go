package interactors

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// ElrondProxyStub -
type ElrondProxyStub struct {
	GetNetworkConfigCalled  func(ctx context.Context) (*data.NetworkConfig, error)
	SendTransactionCalled   func(ctx context.Context, transaction *data.Transaction) (string, error)
	SendTransactionsCalled  func(ctx context.Context, txs []*data.Transaction) ([]string, error)
	ExecuteVMQueryCalled    func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
	GetAccountCalled        func(ctx context.Context, address core.AddressHandler) (*data.Account, error)
	GetNetworkStatusCalled  func(ctx context.Context, shardID uint32) (*data.NetworkStatus, error)
	GetShardOfAddressCalled func(ctx context.Context, bech32Address string) (uint32, error)
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

// GetNetworkStatus -
func (eps *ElrondProxyStub) GetNetworkStatus(ctx context.Context, shardID uint32) (*data.NetworkStatus, error) {
	if eps.GetNetworkStatusCalled != nil {
		return eps.GetNetworkStatusCalled(ctx, shardID)
	}

	return nil, fmt.Errorf("not implemented")
}

// GetShardOfAddress -
func (eps *ElrondProxyStub) GetShardOfAddress(ctx context.Context, bech32Address string) (uint32, error) {
	if eps.GetShardOfAddressCalled != nil {
		return eps.GetShardOfAddressCalled(ctx, bech32Address)
	}

	return 0, fmt.Errorf("not implemented")
}

// IsInterfaceNil -
func (eps *ElrondProxyStub) IsInterfaceNil() bool {
	return eps == nil
}
