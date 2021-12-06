package bridgeV2

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ContractBackendStub struct {
	CodeAtCalled              func(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	CallContractCalled        func(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	HeaderByNumberCalled      func(ctx context.Context, number *big.Int) (*types.Header, error)
	PendingCodeAtCalled       func(ctx context.Context, account common.Address) ([]byte, error)
	PendingNonceAtCalled      func(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPriceCalled     func(ctx context.Context) (*big.Int, error)
	SuggestGasTipCapCalled    func(ctx context.Context) (*big.Int, error)
	EstimateGasCalled         func(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)
	SendTransactionCalled     func(ctx context.Context, tx *types.Transaction) error
	FilterLogsCalled          func(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogsCalled func(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
}

func (stub *ContractBackendStub) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	if stub.CodeAtCalled != nil {
		return stub.CodeAtCalled(ctx, contract, blockNumber)
	}
	return nil, notImplemented
}

func (stub *ContractBackendStub) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	if stub.CallContractCalled != nil {
		return stub.CallContractCalled(ctx, call, blockNumber)
	}
	return nil, notImplemented
}

func (stub *ContractBackendStub) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	if stub.HeaderByNumberCalled != nil {
		return stub.HeaderByNumberCalled(ctx, number)
	}
	return nil, notImplemented
}

func (stub *ContractBackendStub) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	if stub.PendingCodeAtCalled != nil {
		return stub.PendingCodeAtCalled(ctx, account)
	}
	return nil, notImplemented
}

func (stub *ContractBackendStub) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	if stub.PendingNonceAtCalled != nil {
		return stub.PendingNonceAtCalled(ctx, account)
	}
	return 0, notImplemented
}

func (stub *ContractBackendStub) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	if stub.SuggestGasPriceCalled != nil {
		return stub.SuggestGasPriceCalled(ctx)
	}
	return nil, notImplemented
}

func (stub *ContractBackendStub) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	if stub.SuggestGasTipCapCalled != nil {
		return stub.SuggestGasTipCapCalled(ctx)
	}
	return nil, notImplemented
}

func (stub *ContractBackendStub) EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error) {
	if stub.EstimateGasCalled != nil {
		return stub.EstimateGasCalled(ctx, call)
	}
	return 0, notImplemented
}

func (stub *ContractBackendStub) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	if stub.SendTransactionCalled != nil {
		return stub.SendTransactionCalled(ctx, tx)
	}
	return notImplemented
}

func (stub *ContractBackendStub) FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error) {
	if stub.FilterLogsCalled != nil {
		return stub.FilterLogsCalled(ctx, query)
	}
	return nil, notImplemented
}

func (stub *ContractBackendStub) SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	if stub.SubscribeFilterLogsCalled != nil {
		return stub.SubscribeFilterLogsCalled(ctx, query, ch)
	}
	return nil, notImplemented
}
