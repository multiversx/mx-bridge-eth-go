package bridge

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
)

// EthereumClientStub -
type EthereumClientStub struct {
	GetBatchCalled                         func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error)
	WasExecutedCalled                      func(ctx context.Context, batchID uint64) (bool, error)
	GenerateMessageHashCalled              func(batch *clients.TransferBatch) (common.Hash, error)
	BroadcastSignatureForMessageHashCalled func(msgHash common.Hash)
	ExecuteTransferCalled                  func(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error)
	CheckClientAvailabilityCalled          func(ctx context.Context) error
	GetTransactionsStatusesCalled          func(ctx context.Context, batchId uint64) ([]byte, error)
	GetQuorumSizeCalled                    func(ctx context.Context) (*big.Int, error)
	IsQuorumReachedCalled                  func(ctx context.Context, msgHash common.Hash) (bool, error)
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
func (stub *EthereumClientStub) GenerateMessageHash(batch *clients.TransferBatch) (common.Hash, error) {
	if stub.GenerateMessageHashCalled != nil {
		return stub.GenerateMessageHashCalled(batch)
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
func (stub *EthereumClientStub) ExecuteTransfer(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error) {
	if stub.ExecuteTransferCalled != nil {
		return stub.ExecuteTransferCalled(ctx, msgHash, batch, quorum)
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

// IsInterfaceNil -
func (stub *EthereumClientStub) IsInterfaceNil() bool {
	return stub == nil
}
