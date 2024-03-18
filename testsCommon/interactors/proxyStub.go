package interactors

import (
	"context"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// ProxyStub -
type ProxyStub struct {
	GetNetworkConfigCalled  func(ctx context.Context) (*data.NetworkConfig, error)
	SendTransactionCalled   func(ctx context.Context, transaction *transaction.FrontendTransaction) (string, error)
	SendTransactionsCalled  func(ctx context.Context, txs []*transaction.FrontendTransaction) ([]string, error)
	ExecuteVMQueryCalled    func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
	GetAccountCalled        func(ctx context.Context, address core.AddressHandler) (*data.Account, error)
	GetNetworkStatusCalled  func(ctx context.Context, shardID uint32) (*data.NetworkStatus, error)
	GetShardOfAddressCalled func(ctx context.Context, bech32Address string) (uint32, error)
	GetESDTTokenDataCalled  func(ctx context.Context, address core.AddressHandler, tokenIdentifier string, queryOptions api.AccountQueryOptions) (*data.ESDTFungibleTokenData, error)
}

// GetNetworkConfig -
func (eps *ProxyStub) GetNetworkConfig(ctx context.Context) (*data.NetworkConfig, error) {
	if eps.GetNetworkConfigCalled != nil {
		return eps.GetNetworkConfigCalled(ctx)
	}

	return &data.NetworkConfig{}, nil
}

// SendTransaction -
func (eps *ProxyStub) SendTransaction(ctx context.Context, transaction *transaction.FrontendTransaction) (string, error) {
	if eps.SendTransactionCalled != nil {
		return eps.SendTransactionCalled(ctx, transaction)
	}

	return "", nil
}

// SendTransactions -
func (eps *ProxyStub) SendTransactions(ctx context.Context, txs []*transaction.FrontendTransaction) ([]string, error) {
	if eps.SendTransactionCalled != nil {
		return eps.SendTransactionsCalled(ctx, txs)
	}

	return make([]string, 0), nil
}

// ExecuteVMQuery -
func (eps *ProxyStub) ExecuteVMQuery(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
	if eps.ExecuteVMQueryCalled != nil {
		return eps.ExecuteVMQueryCalled(ctx, vmRequest)
	}

	return &data.VmValuesResponseData{}, nil
}

// GetAccount -
func (eps *ProxyStub) GetAccount(ctx context.Context, address core.AddressHandler) (*data.Account, error) {
	if eps.GetAccountCalled != nil {
		return eps.GetAccountCalled(ctx, address)
	}

	return &data.Account{}, nil
}

// GetNetworkStatus -
func (eps *ProxyStub) GetNetworkStatus(ctx context.Context, shardID uint32) (*data.NetworkStatus, error) {
	if eps.GetNetworkStatusCalled != nil {
		return eps.GetNetworkStatusCalled(ctx, shardID)
	}

	return nil, fmt.Errorf("not implemented")
}

// GetShardOfAddress -
func (eps *ProxyStub) GetShardOfAddress(ctx context.Context, bech32Address string) (uint32, error) {
	if eps.GetShardOfAddressCalled != nil {
		return eps.GetShardOfAddressCalled(ctx, bech32Address)
	}

	return 0, fmt.Errorf("not implemented")
}

// GetESDTTokenData -
func (eps *ProxyStub) GetESDTTokenData(ctx context.Context, address core.AddressHandler, tokenIdentifier string, queryOptions api.AccountQueryOptions) (*data.ESDTFungibleTokenData, error) {
	if eps.GetESDTTokenDataCalled != nil {
		return eps.GetESDTTokenDataCalled(ctx, address, tokenIdentifier, queryOptions)
	}

	return &data.ESDTFungibleTokenData{}, fmt.Errorf("not implemented")
}

// IsInterfaceNil -
func (eps *ProxyStub) IsInterfaceNil() bool {
	return eps == nil
}
