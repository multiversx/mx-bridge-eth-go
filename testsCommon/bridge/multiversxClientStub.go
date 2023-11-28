package bridge

import (
	"context"
	"errors"

	"github.com/multiversx/mx-bridge-eth-go/clients"
)

var errNotImplemented = errors.New("not implemented")

// MultiversXClientStub -
type MultiversXClientStub struct {
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
	GetCurrentNonceCalled                          func(ctx context.Context) (uint64, error)
	ProposeSetStatusCalled                         func(ctx context.Context, batch *clients.TransferBatch) (string, error)
	ResolveNewDepositsCalled                       func(ctx context.Context, batch *clients.TransferBatch) error
	ProposeTransferCalled                          func(ctx context.Context, batch *clients.TransferBatch) (string, error)
	ProposeSCTransferCalled                        func(ctx context.Context, batch *clients.TransferBatch, metadata *clients.SCBatch) (string, error)
	SignCalled                                     func(ctx context.Context, actionID uint64) (string, error)
	WasSignedCalled                                func(ctx context.Context, actionID uint64) (bool, error)
	PerformActionCalled                            func(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error)
	CheckClientAvailabilityCalled                  func(ctx context.Context) error
	CloseCalled                                    func() error
}

// GetPending -
func (stub *MultiversXClientStub) GetPending(ctx context.Context) (*clients.TransferBatch, error) {
	if stub.GetPendingCalled != nil {
		return stub.GetPendingCalled(ctx)
	}

	return nil, errNotImplemented
}

// GetCurrentBatchAsDataBytes -
func (stub *MultiversXClientStub) GetCurrentBatchAsDataBytes(ctx context.Context) ([][]byte, error) {
	if stub.GetCurrentBatchAsDataBytesCalled != nil {
		return stub.GetCurrentBatchAsDataBytesCalled(ctx)
	}

	return make([][]byte, 0), nil
}

// WasProposedTransfer -
func (stub *MultiversXClientStub) WasProposedTransfer(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
	if stub.WasProposedTransferCalled != nil {
		return stub.WasProposedTransferCalled(ctx, batch)
	}

	return false, nil
}

// QuorumReached -
func (stub *MultiversXClientStub) QuorumReached(ctx context.Context, actionID uint64) (bool, error) {
	if stub.QuorumReachedCalled != nil {
		return stub.QuorumReachedCalled(ctx, actionID)
	}

	return false, nil
}

// WasExecuted -
func (stub *MultiversXClientStub) WasExecuted(ctx context.Context, actionID uint64) (bool, error) {
	if stub.WasExecutedCalled != nil {
		return stub.WasExecutedCalled(ctx, actionID)
	}

	return false, nil
}

// GetActionIDForProposeTransfer -
func (stub *MultiversXClientStub) GetActionIDForProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
	if stub.GetActionIDForProposeTransferCalled != nil {
		return stub.GetActionIDForProposeTransferCalled(ctx, batch)
	}

	return 0, nil
}

// WasProposedSetStatus -
func (stub *MultiversXClientStub) WasProposedSetStatus(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
	if stub.WasProposedSetStatusCalled != nil {
		return stub.WasProposedSetStatusCalled(ctx, batch)
	}

	return false, nil
}

// GetTransactionsStatuses -
func (stub *MultiversXClientStub) GetTransactionsStatuses(ctx context.Context, batchID uint64) ([]byte, error) {
	if stub.GetTransactionsStatusesCalled != nil {
		return stub.GetTransactionsStatusesCalled(ctx, batchID)
	}

	return make([]byte, 0), nil
}

// GetActionIDForSetStatusOnPendingTransfer -
func (stub *MultiversXClientStub) GetActionIDForSetStatusOnPendingTransfer(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
	if stub.GetActionIDForSetStatusOnPendingTransferCalled != nil {
		return stub.GetActionIDForSetStatusOnPendingTransferCalled(ctx, batch)
	}

	return 0, nil
}

// GetLastExecutedEthBatchID -
func (stub *MultiversXClientStub) GetLastExecutedEthBatchID(ctx context.Context) (uint64, error) {
	if stub.GetLastExecutedEthBatchIDCalled != nil {
		return stub.GetLastExecutedEthBatchIDCalled(ctx)
	}

	return 0, nil
}

// GetLastExecutedEthTxID -
func (stub *MultiversXClientStub) GetLastExecutedEthTxID(ctx context.Context) (uint64, error) {
	if stub.GetLastExecutedEthTxIDCalled != nil {
		return stub.GetLastExecutedEthTxIDCalled(ctx)
	}

	return 0, nil
}

// GetCurrentNonce -
func (stub *MultiversXClientStub) GetCurrentNonce(ctx context.Context) (uint64, error) {
	if stub.GetCurrentNonceCalled != nil {
		return stub.GetCurrentNonceCalled(ctx)
	}

	return 0, nil
}

// ProposeSetStatus -
func (stub *MultiversXClientStub) ProposeSetStatus(ctx context.Context, batch *clients.TransferBatch) (string, error) {
	if stub.ProposeSetStatusCalled != nil {
		return stub.ProposeSetStatusCalled(ctx, batch)
	}

	return "", nil
}

// ProposeTransfer -
func (stub *MultiversXClientStub) ProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (string, error) {
	if stub.ProposeTransferCalled != nil {
		return stub.ProposeTransferCalled(ctx, batch)
	}

	return "", nil
}

// ProposeSCTransfer -
func (stub *MultiversXClientStub) ProposeSCTransfer(ctx context.Context, batch *clients.TransferBatch, metadata *clients.SCBatch) (string, error) {
	if stub.ProposeSCTransferCalled != nil {
		return stub.ProposeSCTransferCalled(ctx, batch, metadata)
	}

	return "", nil
}

// Sign -
func (stub *MultiversXClientStub) Sign(ctx context.Context, actionID uint64) (string, error) {
	if stub.SignCalled != nil {
		return stub.SignCalled(ctx, actionID)
	}

	return "", nil
}

// WasSigned -
func (stub *MultiversXClientStub) WasSigned(ctx context.Context, actionID uint64) (bool, error) {
	if stub.WasSignedCalled != nil {
		return stub.WasSignedCalled(ctx, actionID)
	}

	return false, nil
}

// PerformAction -
func (stub *MultiversXClientStub) PerformAction(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
	if stub.PerformActionCalled != nil {
		return stub.PerformActionCalled(ctx, actionID, batch)
	}

	return "", nil
}

// CheckClientAvailability -
func (stub *MultiversXClientStub) CheckClientAvailability(ctx context.Context) error {
	if stub.CheckClientAvailabilityCalled != nil {
		return stub.CheckClientAvailabilityCalled(ctx)
	}

	return nil
}

// Close -
func (stub *MultiversXClientStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *MultiversXClientStub) IsInterfaceNil() bool {
	return stub == nil
}
