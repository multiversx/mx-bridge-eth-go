package bridge

import (
	"context"
	ethmultiversx "github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
)

// EthereumClientStub -
type EthereumClientStub struct {
	GetBatchCalled                         func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error)
	WasExecutedCalled                      func(ctx context.Context, batchID uint64) (bool, error)
	GenerateMessageHashCalled              func(batch *ethmultiversx.ArgListsBatch, batchID uint64) (common.Hash, error)
	BroadcastSignatureForMessageHashCalled func(msgHash common.Hash)
	ExecuteTransferCalled                  func(ctx context.Context, msgHash common.Hash, batch *ethmultiversx.ArgListsBatch, batchId uint64, quorum int) (string, error)
	CheckClientAvailabilityCalled          func(ctx context.Context) error
	GetTransactionsStatusesCalled          func(ctx context.Context, batchId uint64) ([]byte, error)
	GetQuorumSizeCalled                    func(ctx context.Context) (*big.Int, error)
	IsQuorumReachedCalled                  func(ctx context.Context, msgHash common.Hash) (bool, error)
	CheckRequiredBalanceCalled             func(ctx context.Context, erc20Address common.Address, value *big.Int) error
	TokenMintedBalancesCalled              func(ctx context.Context, token common.Address) (*big.Int, error)
	WhitelistedTokensMintBurnCalled        func(ctx context.Context, token common.Address) (bool, error)
}

// GetBatch -
func (stub *EthereumClientStub) GetBatch(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchCalled(ctx, nonce)
	}

	return nil, errNotImplemented
}

// WasExecuted -
func (stub *EthereumClientStub) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
	if stub.WasExecutedCalled != nil {
		return stub.WasExecutedCalled(ctx, batchID)
	}

	return false, errNotImplemented
}

// GenerateMessageHash -
func (stub *EthereumClientStub) GenerateMessageHash(batch *ethmultiversx.ArgListsBatch, batchID uint64) (common.Hash, error) {
	if stub.GenerateMessageHashCalled != nil {
		return stub.GenerateMessageHashCalled(batch, batchID)
	}

	return common.Hash{}, errNotImplemented
}

// BroadcastSignatureForMessageHash -
func (stub *EthereumClientStub) BroadcastSignatureForMessageHash(msgHash common.Hash) {
	if stub.BroadcastSignatureForMessageHashCalled != nil {
		stub.BroadcastSignatureForMessageHashCalled(msgHash)
	}
}

// ExecuteTransfer -
func (stub *EthereumClientStub) ExecuteTransfer(ctx context.Context, msgHash common.Hash, batch *ethmultiversx.ArgListsBatch, batchId uint64, quorum int) (string, error) {
	if stub.ExecuteTransferCalled != nil {
		return stub.ExecuteTransferCalled(ctx, msgHash, batch, batchId, quorum)
	}

	return "", errNotImplemented
}

// CheckClientAvailability -
func (stub *EthereumClientStub) CheckClientAvailability(ctx context.Context) error {
	if stub.CheckClientAvailabilityCalled != nil {
		return stub.CheckClientAvailabilityCalled(ctx)
	}

	return nil
}

// GetTransactionsStatuses -
func (stub *EthereumClientStub) GetTransactionsStatuses(ctx context.Context, batchId uint64) ([]byte, error) {
	if stub.GetTransactionsStatusesCalled != nil {
		return stub.GetTransactionsStatusesCalled(ctx, batchId)
	}

	return nil, errNotImplemented
}

// GetQuorumSize -
func (stub *EthereumClientStub) GetQuorumSize(ctx context.Context) (*big.Int, error) {
	if stub.GetQuorumSizeCalled != nil {
		return stub.GetQuorumSizeCalled(ctx)
	}

	return nil, errNotImplemented
}

// IsQuorumReached -
func (stub *EthereumClientStub) IsQuorumReached(ctx context.Context, msgHash common.Hash) (bool, error) {
	if stub.IsQuorumReachedCalled != nil {
		return stub.IsQuorumReachedCalled(ctx, msgHash)
	}

	return false, errNotImplemented
}

// CheckRequiredBalance -
func (stub *EthereumClientStub) CheckRequiredBalance(ctx context.Context, erc20Address common.Address, value *big.Int) error {
	if stub.CheckRequiredBalanceCalled != nil {
		return stub.CheckRequiredBalanceCalled(ctx, erc20Address, value)
	}

	return nil
}

// TokenMintedBalances -
func (stub *EthereumClientStub) TokenMintedBalances(ctx context.Context, token common.Address) (*big.Int, error) {
	if stub.TokenMintedBalancesCalled != nil {
		return stub.TokenMintedBalancesCalled(ctx, token)
	}

	return nil, nil
}

// WhitelistedTokensMintBurn -
func (stub *EthereumClientStub) WhitelistedTokensMintBurn(ctx context.Context, token common.Address) (bool, error) {
	if stub.WhitelistedTokensMintBurnCalled != nil {
		return stub.WhitelistedTokensMintBurnCalled(ctx, token)
	}

	return false, nil
}

// IsInterfaceNil -
func (stub *EthereumClientStub) IsInterfaceNil() bool {
	return stub == nil
}
