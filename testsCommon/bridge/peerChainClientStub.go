package bridge

import (
	"context"
	"math/big"

	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
)

// PeerChainClientStub -
type PeerChainClientStub struct {
	GetBatchCalled                         func(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error)
	WasExecutedCalled                      func(ctx context.Context, batchID uint64) (bool, error)
	GenerateMessageHashCalled              func(batch *batchProcessor.ArgListsBatch, batchID uint64) ([]byte, error)
	BroadcastSignatureForMessageHashCalled func(msgHash []byte)
	ExecuteTransferCalled                  func(ctx context.Context, msgHash []byte, batch *batchProcessor.ArgListsBatch, batchID uint64, quorum int) (string, error)
	CheckClientAvailabilityCalled          func(ctx context.Context) error
	GetTransactionsStatusesCalled          func(ctx context.Context, batchID uint64) ([]byte, error)
	GetQuorumSizeCalled                    func(ctx context.Context) (*big.Int, error)
	IsQuorumReachedCalled                  func(ctx context.Context, msgHash []byte) (bool, error)
	GetBatchSCMetadataCalled               func(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error)
	CheckRequiredBalanceCalled             func(ctx context.Context, token []byte, value *big.Int) error
	TotalBalancesCalled                    func(ctx context.Context, token []byte) (*big.Int, error)
	MintBalancesCalled                     func(ctx context.Context, token []byte) (*big.Int, error)
	BurnBalancesCalled                     func(ctx context.Context, token []byte) (*big.Int, error)
	MintBurnTokensCalled                   func(ctx context.Context, token []byte) (bool, error)
	NativeTokensCalled                     func(ctx context.Context, token []byte) (bool, error)
	WhitelistedTokensCalled                func(ctx context.Context, token []byte) (bool, error)
}

// GetBatch -
func (stub *PeerChainClientStub) GetBatch(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchCalled(ctx, nonce)
	}
	return nil, false, errNotImplemented
}

// WasExecuted -
func (stub *PeerChainClientStub) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
	if stub.WasExecutedCalled != nil {
		return stub.WasExecutedCalled(ctx, batchID)
	}
	return false, errNotImplemented
}

// GenerateMessageHash -
func (stub *PeerChainClientStub) GenerateMessageHash(batch *batchProcessor.ArgListsBatch, batchID uint64) ([]byte, error) {
	if stub.GenerateMessageHashCalled != nil {
		return stub.GenerateMessageHashCalled(batch, batchID)
	}
	return nil, errNotImplemented
}

// BroadcastSignatureForMessageHash -
func (stub *PeerChainClientStub) BroadcastSignatureForMessageHash(msgHash []byte) {
	if stub.BroadcastSignatureForMessageHashCalled != nil {
		stub.BroadcastSignatureForMessageHashCalled(msgHash)
	}
}

// ExecuteTransfer -
func (stub *PeerChainClientStub) ExecuteTransfer(ctx context.Context, msgHash []byte, batch *batchProcessor.ArgListsBatch, batchID uint64, quorum int) (string, error) {
	if stub.ExecuteTransferCalled != nil {
		return stub.ExecuteTransferCalled(ctx, msgHash, batch, batchID, quorum)
	}
	return "", errNotImplemented
}

// CheckClientAvailability -
func (stub *PeerChainClientStub) CheckClientAvailability(ctx context.Context) error {
	if stub.CheckClientAvailabilityCalled != nil {
		return stub.CheckClientAvailabilityCalled(ctx)
	}
	return errNotImplemented
}

// GetTransactionsStatuses -
func (stub *PeerChainClientStub) GetTransactionsStatuses(ctx context.Context, batchID uint64) ([]byte, error) {
	if stub.GetTransactionsStatusesCalled != nil {
		return stub.GetTransactionsStatusesCalled(ctx, batchID)
	}
	return nil, errNotImplemented
}

// GetQuorumSize -
func (stub *PeerChainClientStub) GetQuorumSize(ctx context.Context) (*big.Int, error) {
	if stub.GetQuorumSizeCalled != nil {
		return stub.GetQuorumSizeCalled(ctx)
	}
	return nil, errNotImplemented
}

// IsQuorumReached -
func (stub *PeerChainClientStub) IsQuorumReached(ctx context.Context, msgHash []byte) (bool, error) {
	if stub.IsQuorumReachedCalled != nil {
		return stub.IsQuorumReachedCalled(ctx, msgHash)
	}
	return false, errNotImplemented
}

// GetBatchSCMetadata -
func (stub *PeerChainClientStub) GetBatchSCMetadata(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
	if stub.GetBatchSCMetadataCalled != nil {
		return stub.GetBatchSCMetadataCalled(ctx, nonce, blockNumber)
	}
	return nil, errNotImplemented
}

// CheckRequiredBalance -
func (stub *PeerChainClientStub) CheckRequiredBalance(ctx context.Context, token []byte, value *big.Int) error {
	if stub.CheckRequiredBalanceCalled != nil {
		return stub.CheckRequiredBalanceCalled(ctx, token, value)
	}
	return errNotImplemented
}

// TotalBalances -
func (stub *PeerChainClientStub) TotalBalances(ctx context.Context, token []byte) (*big.Int, error) {
	if stub.TotalBalancesCalled != nil {
		return stub.TotalBalancesCalled(ctx, token)
	}
	return nil, errNotImplemented
}

// MintBalances -
func (stub *PeerChainClientStub) MintBalances(ctx context.Context, token []byte) (*big.Int, error) {
	if stub.MintBalancesCalled != nil {
		return stub.MintBalancesCalled(ctx, token)
	}
	return nil, errNotImplemented
}

// BurnBalances -
func (stub *PeerChainClientStub) BurnBalances(ctx context.Context, token []byte) (*big.Int, error) {
	if stub.BurnBalancesCalled != nil {
		return stub.BurnBalancesCalled(ctx, token)
	}
	return nil, errNotImplemented
}

// MintBurnTokens -
func (stub *PeerChainClientStub) MintBurnTokens(ctx context.Context, token []byte) (bool, error) {
	if stub.MintBurnTokensCalled != nil {
		return stub.MintBurnTokensCalled(ctx, token)
	}
	return false, errNotImplemented
}

// NativeTokens -
func (stub *PeerChainClientStub) NativeTokens(ctx context.Context, token []byte) (bool, error) {
	if stub.NativeTokensCalled != nil {
		return stub.NativeTokensCalled(ctx, token)
	}
	return false, errNotImplemented
}

// WhitelistedTokens -
func (stub *PeerChainClientStub) WhitelistedTokens(ctx context.Context, token []byte) (bool, error) {
	if stub.WhitelistedTokensCalled != nil {
		return stub.WhitelistedTokensCalled(ctx, token)
	}
	return false, errNotImplemented
}

// IsInterfaceNil -
func (stub *PeerChainClientStub) IsInterfaceNil() bool {
	return stub == nil
}
