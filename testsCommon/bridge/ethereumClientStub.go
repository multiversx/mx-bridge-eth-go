package bridge

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	bridgeCommon "github.com/multiversx/mx-bridge-eth-go/common"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
)

// EthereumClientStub -
type EthereumClientStub struct {
	GetBatchCalled                         func(ctx context.Context, nonce uint64) (*bridgeCommon.TransferBatch, bool, error)
	WasExecutedCalled                      func(ctx context.Context, batchID uint64) (bool, error)
	GenerateMessageHashCalled              func(batch *batchProcessor.ArgListsBatch, batchID uint64) (common.Hash, error)
	BroadcastSignatureForMessageHashCalled func(msgHash common.Hash)
	ExecuteTransferCalled                  func(ctx context.Context, msgHash common.Hash, batch *batchProcessor.ArgListsBatch, batchId uint64, quorum int) (string, error)
	CheckClientAvailabilityCalled          func(ctx context.Context) error
	GetTransactionsStatusesCalled          func(ctx context.Context, batchId uint64) ([]byte, error)
	GetQuorumSizeCalled                    func(ctx context.Context) (*big.Int, error)
	IsQuorumReachedCalled                  func(ctx context.Context, msgHash common.Hash) (bool, error)
	GetBatchSCMetadataCalled               func(ctx context.Context, nonce uint64) ([]*contract.ERC20SafeERC20SCDeposit, error)
	CheckRequiredBalanceCalled             func(ctx context.Context, erc20Address common.Address, value *big.Int) error
	TotalBalancesCalled                    func(ctx context.Context, account common.Address) (*big.Int, error)
	MintBalancesCalled                     func(ctx context.Context, account common.Address) (*big.Int, error)
	BurnBalancesCalled                     func(ctx context.Context, account common.Address) (*big.Int, error)
	MintBurnTokensCalled                   func(ctx context.Context, account common.Address) (bool, error)
	NativeTokensCalled                     func(ctx context.Context, account common.Address) (bool, error)
	WhitelistedTokensCalled                func(ctx context.Context, account common.Address) (bool, error)
}

// GetBatch -
func (stub *EthereumClientStub) GetBatch(ctx context.Context, nonce uint64) (*bridgeCommon.TransferBatch, bool, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchCalled(ctx, nonce)
	}

	return nil, false, errNotImplemented
}

// WasExecuted -
func (stub *EthereumClientStub) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
	if stub.WasExecutedCalled != nil {
		return stub.WasExecutedCalled(ctx, batchID)
	}

	return false, errNotImplemented
}

// GenerateMessageHash -
func (stub *EthereumClientStub) GenerateMessageHash(batch *batchProcessor.ArgListsBatch, batchID uint64) (common.Hash, error) {
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
func (stub *EthereumClientStub) ExecuteTransfer(ctx context.Context, msgHash common.Hash, batch *batchProcessor.ArgListsBatch, batchId uint64, quorum int) (string, error) {
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

	return errNotImplemented
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

// GetBatchSCMetadata -
func (stub *EthereumClientStub) GetBatchSCMetadata(ctx context.Context, nonce uint64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
	if stub.GetBatchSCMetadataCalled != nil {
		return stub.GetBatchSCMetadataCalled(ctx, nonce)
	}

	return []*contract.ERC20SafeERC20SCDeposit{}, errNotImplemented
}

// CheckRequiredBalance -
func (stub *EthereumClientStub) CheckRequiredBalance(ctx context.Context, erc20Address common.Address, value *big.Int) error {
	if stub.CheckRequiredBalanceCalled != nil {
		return stub.CheckRequiredBalanceCalled(ctx, erc20Address, value)
	}

	return errNotImplemented
}

// TotalBalances -
func (stub *EthereumClientStub) TotalBalances(ctx context.Context, account common.Address) (*big.Int, error) {
	if stub.TotalBalancesCalled != nil {
		return stub.TotalBalancesCalled(ctx, account)
	}

	return nil, errNotImplemented
}

// MintBalances -
func (stub *EthereumClientStub) MintBalances(ctx context.Context, account common.Address) (*big.Int, error) {
	if stub.MintBalancesCalled != nil {
		return stub.MintBalancesCalled(ctx, account)
	}

	return nil, errNotImplemented
}

// BurnBalances -
func (stub *EthereumClientStub) BurnBalances(ctx context.Context, account common.Address) (*big.Int, error) {
	if stub.BurnBalancesCalled != nil {
		return stub.BurnBalancesCalled(ctx, account)
	}

	return nil, errNotImplemented
}

// MintBurnTokens -
func (stub *EthereumClientStub) MintBurnTokens(ctx context.Context, account common.Address) (bool, error) {
	if stub.MintBurnTokensCalled != nil {
		return stub.MintBurnTokensCalled(ctx, account)
	}

	return false, errNotImplemented
}

// NativeTokens -
func (stub *EthereumClientStub) NativeTokens(ctx context.Context, account common.Address) (bool, error) {
	if stub.NativeTokensCalled != nil {
		return stub.NativeTokensCalled(ctx, account)
	}

	return false, errNotImplemented
}

// WhitelistedTokens -
func (stub *EthereumClientStub) WhitelistedTokens(ctx context.Context, account common.Address) (bool, error) {
	if stub.WhitelistedTokensCalled != nil {
		return stub.WhitelistedTokensCalled(ctx, account)
	}

	return false, errNotImplemented
}

// IsInterfaceNil -
func (stub *EthereumClientStub) IsInterfaceNil() bool {
	return stub == nil
}
