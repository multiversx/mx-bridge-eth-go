package bridgeV2

import (
	"context"
	"errors"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
)

var errNotImplemented = errors.New("not implemented")

// ElrondClientStub -
type ElrondClientStub struct {
	GetPendingCalled                               func(ctx context.Context) (*clients.TransferBatch, error)
	GetCurrentBatchAsDataBytesCalled               func(ctx context.Context) ([][]byte, error)
	WasProposedTransferCalled                      func(ctx context.Context, batch *clients.TransferBatch) (bool, error)
	QuorumReachedCalled                            func(ctx context.Context, actionID uint64) (bool, error)
	WasExecutedCalled                              func(ctx context.Context, actionID uint64) (bool, error)
	GetActionIDForProposeTransferCalled            func(ctx context.Context, batch *clients.TransferBatch) (uint64, error)
	WasProposedSetStatusCalled                     func(ctx context.Context, batch *clients.TransferBatch) (bool, error)
	GetTransactionsStatusesCalled                  func(ctx context.Context, batchID uint64) ([]byte, error)
	GetActionIDForSetStatusOnPendingTransferCalled func(ctx context.Context, batch *clients.TransferBatch) (uint64, error)
	GetLastExecutedEthBatchIDCalled                func(ctx context.Context) (uint64, error)
	GetLastExecutedEthTxIDCalled                   func(ctx context.Context) (uint64, error)
	ProposeSetStatusCalled                         func(ctx context.Context, batch *clients.TransferBatch) (string, error)
	ResolveNewDepositsCalled                       func(ctx context.Context, batch *clients.TransferBatch) error
	ProposeTransferCalled                          func(ctx context.Context, batch *clients.TransferBatch) (string, error)
	SignCalled                                     func(ctx context.Context, actionID uint64) (string, error)
	WasSignedCalled                                func(ctx context.Context, actionID uint64) (bool, error)
	PerformActionCalled                            func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error)
	GetMaxNumberOfRetriesOnQuorumReachedCalled     func() uint64
	CloseCalled                                    func() error
}

// GetPending -
func (stub *ElrondClientStub) GetPending(ctx context.Context) (*clients.TransferBatch, error) {
	if stub.GetPendingCalled != nil {
		return stub.GetPendingCalled(ctx)
	}

	return nil, errNotImplemented
}

// GetCurrentBatchAsDataBytes -
func (stub *ElrondClientStub) GetCurrentBatchAsDataBytes(ctx context.Context) ([][]byte, error) {
	if stub.GetCurrentBatchAsDataBytesCalled != nil {
		return stub.GetCurrentBatchAsDataBytesCalled(ctx)
	}

	return make([][]byte, 0), nil
}

// WasProposedTransfer -
func (stub *ElrondClientStub) WasProposedTransfer(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
	if stub.WasProposedTransferCalled != nil {
		return stub.WasProposedTransferCalled(ctx, batch)
	}

	return false, nil
}

// QuorumReached -
func (stub *ElrondClientStub) QuorumReached(ctx context.Context, actionID uint64) (bool, error) {
	if stub.QuorumReachedCalled != nil {
		return stub.QuorumReachedCalled(ctx, actionID)
	}

	return false, nil
}

// WasExecuted -
func (stub *ElrondClientStub) WasExecuted(ctx context.Context, actionID uint64) (bool, error) {
	if stub.WasExecutedCalled != nil {
		return stub.WasExecutedCalled(ctx, actionID)
	}

	return false, nil
}

// GetActionIDForProposeTransfer -
func (stub *ElrondClientStub) GetActionIDForProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
	if stub.GetActionIDForProposeTransferCalled != nil {
		return stub.GetActionIDForProposeTransferCalled(ctx, batch)
	}

	return 0, nil
}

// WasProposedSetStatus -
func (stub *ElrondClientStub) WasProposedSetStatus(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
	if stub.WasProposedSetStatusCalled != nil {
		return stub.WasProposedSetStatusCalled(ctx, batch)
	}

	return false, nil
}

// GetTransactionsStatuses -
func (stub *ElrondClientStub) GetTransactionsStatuses(ctx context.Context, batchID uint64) ([]byte, error) {
	if stub.GetTransactionsStatusesCalled != nil {
		return stub.GetTransactionsStatusesCalled(ctx, batchID)
	}

	return make([]byte, 0), nil
}

// GetActionIDForSetStatusOnPendingTransfer -
func (stub *ElrondClientStub) GetActionIDForSetStatusOnPendingTransfer(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
	if stub.GetActionIDForSetStatusOnPendingTransferCalled != nil {
		return stub.GetActionIDForSetStatusOnPendingTransferCalled(ctx, batch)
	}

	return 0, nil
}

// GetLastExecutedEthBatchID -
func (stub *ElrondClientStub) GetLastExecutedEthBatchID(ctx context.Context) (uint64, error) {
	if stub.GetLastExecutedEthBatchIDCalled != nil {
		return stub.GetLastExecutedEthBatchIDCalled(ctx)
	}

	return 0, nil
}

// GetLastExecutedEthTxID -
func (stub *ElrondClientStub) GetLastExecutedEthTxID(ctx context.Context) (uint64, error) {
	if stub.GetLastExecutedEthTxIDCalled != nil {
		return stub.GetLastExecutedEthTxIDCalled(ctx)
	}

	return 0, nil
}

// ProposeSetStatus -
func (stub *ElrondClientStub) ProposeSetStatus(ctx context.Context, batch *clients.TransferBatch) (string, error) {
	if stub.ProposeSetStatusCalled != nil {
		return stub.ProposeSetStatusCalled(ctx, batch)
	}

	return "", nil
}

// ResolveNewDeposits -
func (stub *ElrondClientStub) ResolveNewDeposits(ctx context.Context, batch *clients.TransferBatch) error {
	if stub.ResolveNewDepositsCalled != nil {
		return stub.ResolveNewDepositsCalled(ctx, batch)
	}

	return nil
}

// ProposeTransfer -
func (stub *ElrondClientStub) ProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (string, error) {
	if stub.ProposeTransferCalled != nil {
		return stub.ProposeTransferCalled(ctx, batch)
	}

	return "", nil
}

// Sign -
func (stub *ElrondClientStub) Sign(ctx context.Context, actionID uint64) (string, error) {
	if stub.SignCalled != nil {
		return stub.SignCalled(ctx, actionID)
	}

	return "", nil
}

// WasSigned -
func (stub *ElrondClientStub) WasSigned(ctx context.Context, actionID uint64) (bool, error) {
	if stub.WasSignedCalled != nil {
		return stub.WasSignedCalled(ctx, actionID)
	}

	return false, nil
}

// PerformAction -
func (stub *ElrondClientStub) PerformAction(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
	if stub.PerformActionCalled != nil {
		return stub.PerformActionCalled(ctx, actionID, batch)
	}

	return "", nil
}

// GetMaxNumberOfRetriesOnQuorumReached -
func (stub *ElrondClientStub) GetMaxNumberOfRetriesOnQuorumReached() uint64 {
	if stub.GetMaxNumberOfRetriesOnQuorumReachedCalled != nil {
		return stub.GetMaxNumberOfRetriesOnQuorumReachedCalled()
	}

	return 0
}

// Close -
func (stub *ElrondClientStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *ElrondClientStub) IsInterfaceNil() bool {
	return stub == nil
}
