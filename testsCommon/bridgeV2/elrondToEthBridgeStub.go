package bridgeV2

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	v2 "github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type ElrondToEthBridgeStub struct {
	functionCalledCounter map[string]int
	mutExecutor           sync.RWMutex
	fullPath              string

	GetLoggerCalled      func() logger.Logger
	MyTurnAsLeaderCalled func() bool

	GetBatchFromElrondCalled       func(ctx context.Context) (*clients.TransferBatch, error)
	StoreBatchFromElrondCalled     func(ctx context.Context, batch *clients.TransferBatch) error
	GetStoredBatchFromElrondCalled func() *clients.TransferBatch

	WasTransferPerformedOnEthereumCalled func(ctx context.Context) (bool, error)
	SignTransferOnEthereumCalled         func(ctx context.Context) error
	IsQuorumReachedOnEthereumCalled      func(ctx context.Context) (bool, error)
	PerformTransferOnEthereumCalled      func(ctx context.Context) error

	GetAndStoreActionIDForSetStatusFromElrondCalled func(ctx context.Context) (uint64, error)
	GetStoredActionIDForSetStatusCalled             func() uint64
	ResolveNewDpositsStatusesCalled                 func(ctx context.Context, numDeposits uint64) error
	GetBatchStatusesFromEthereumCalled              func(ctx context.Context) ([]byte, error)
	WasSetStatusProposedOnElrondCalled              func(ctx context.Context) (bool, error)
	ProposeSetStatusOnElrondCalled                  func(ctx context.Context) error
	WasProposedSetStatusSignedOnElrondCalled        func(ctx context.Context) (bool, error)
	SignProposedSetStatusOnElrondCalled             func(ctx context.Context) error
	IsQuorumReachedOnElrondCalled                   func(ctx context.Context) (bool, error)
	WasSetStatusPerformedOnElrondCalled             func(ctx context.Context) (bool, error)
	PerformSetStatusOnElrondCalled                  func(ctx context.Context) error

	WaitForTransferConfirmationCalled func(ctx context.Context)

	ProcessMaxRetriesOnElrondCalled   func() bool
	ResetRetriesCountOnElrondCalled   func()
	ProcessMaxRetriesOnEthereumCalled func() bool
	ResetRetriesCountOnEthereumCalled func()
}

// NewElrondToEthBridgeStub creates a new ElrondToEthBridgeStub instance
func NewElrondToEthBridgeStub() *ElrondToEthBridgeStub {
	return &ElrondToEthBridgeStub{
		functionCalledCounter: make(map[string]int),
		fullPath:              "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2.(*ElrondToEthBridgeStub).",
	}
}

// GetLogger -
func (stub *ElrondToEthBridgeStub) GetLogger() logger.Logger {
	stub.incrementFunctionCounter()
	if stub.GetLoggerCalled != nil {
		return stub.GetLoggerCalled()
	}
	return nil
}

// MyTurnAsLeader -
func (stub *ElrondToEthBridgeStub) MyTurnAsLeader() bool {
	stub.incrementFunctionCounter()
	if stub.MyTurnAsLeaderCalled != nil {
		return stub.MyTurnAsLeaderCalled()
	}
	return false
}

// GetBatchFromElrond -
func (stub *ElrondToEthBridgeStub) GetBatchFromElrond(ctx context.Context) (*clients.TransferBatch, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchFromElrondCalled != nil {
		return stub.GetBatchFromElrondCalled(ctx)
	}
	return nil, notImplemented
}

// StoreBatchFromElrond -
func (stub *ElrondToEthBridgeStub) StoreBatchFromElrond(ctx context.Context, batch *clients.TransferBatch) error {
	stub.incrementFunctionCounter()
	if stub.StoreBatchFromElrondCalled != nil {
		return stub.StoreBatchFromElrondCalled(ctx, batch)
	}
	return notImplemented
}

// GetStoredBatchFromElrond -
func (stub *ElrondToEthBridgeStub) GetStoredBatchFromElrond() *clients.TransferBatch {
	stub.incrementFunctionCounter()
	if stub.GetStoredBatchFromElrondCalled != nil {
		return stub.GetStoredBatchFromElrondCalled()
	}
	return nil
}

// WasTransferPerformedOnEthereum -
func (stub *ElrondToEthBridgeStub) WasTransferPerformedOnEthereum(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferPerformedOnEthereumCalled != nil {
		return stub.WasTransferPerformedOnEthereumCalled(ctx)
	}
	return false, notImplemented
}

// SignTransferOnEthereum -
func (stub *ElrondToEthBridgeStub) SignTransferOnEthereum(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignTransferOnEthereumCalled != nil {
		return stub.SignTransferOnEthereumCalled(ctx)
	}
	return notImplemented
}

// IsQuorumReachedOnEthereum -
func (stub *ElrondToEthBridgeStub) IsQuorumReachedOnEthereum(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.IsQuorumReachedOnEthereumCalled != nil {
		return stub.IsQuorumReachedOnEthereumCalled(ctx)
	}
	return false, notImplemented
}

// PerformTransferOnEthereum -
func (stub *ElrondToEthBridgeStub) PerformTransferOnEthereum(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformTransferOnEthereumCalled != nil {
		return stub.PerformTransferOnEthereumCalled(ctx)
	}
	return notImplemented
}

// GetAndStoreActionIDForSetStatusFromElrond -
func (stub *ElrondToEthBridgeStub) GetAndStoreActionIDForSetStatusFromElrond(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForSetStatusFromElrondCalled != nil {
		return stub.GetAndStoreActionIDForSetStatusFromElrondCalled(ctx)
	}
	return v2.InvalidActionID, notImplemented
}

// GetStoredActionIDForSetStatus -
func (stub *ElrondToEthBridgeStub) GetStoredActionIDForSetStatus() uint64 {
	stub.incrementFunctionCounter()
	if stub.GetStoredActionIDForSetStatusCalled != nil {
		return stub.GetStoredActionIDForSetStatusCalled()
	}
	return v2.InvalidActionID
}

// ResolveNewDpositsStatuses -
func (stub *ElrondToEthBridgeStub) ResolveNewDpositsStatuses(ctx context.Context, numDeposits uint64) error {
	stub.incrementFunctionCounter()
	if stub.ResolveNewDpositsStatusesCalled != nil {
		return stub.ResolveNewDpositsStatusesCalled(ctx, numDeposits)
	}
	return notImplemented
}

// GetBatchStatusesFromEthereum -
func (stub *ElrondToEthBridgeStub) GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchStatusesFromEthereumCalled != nil {
		return stub.GetBatchStatusesFromEthereumCalled(ctx)
	}
	return nil, notImplemented
}

// WasSetStatusProposedOnElrond -
func (stub *ElrondToEthBridgeStub) WasSetStatusProposedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasSetStatusProposedOnElrondCalled != nil {
		return stub.WasSetStatusProposedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// ProposeSetStatusOnElrond -
func (stub *ElrondToEthBridgeStub) ProposeSetStatusOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeSetStatusOnElrondCalled != nil {
		return stub.ProposeSetStatusOnElrondCalled(ctx)
	}
	return notImplemented
}

// WasProposedSetStatusSignedOnElrond -
func (stub *ElrondToEthBridgeStub) WasProposedSetStatusSignedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasProposedSetStatusSignedOnElrondCalled != nil {
		return stub.WasProposedSetStatusSignedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// SignProposedSetStatusOnElrond -
func (stub *ElrondToEthBridgeStub) SignProposedSetStatusOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignProposedSetStatusOnElrondCalled != nil {
		return stub.SignProposedSetStatusOnElrondCalled(ctx)
	}
	return notImplemented
}

// IsQuorumReachedOnElrond -
func (stub *ElrondToEthBridgeStub) IsQuorumReachedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.IsQuorumReachedOnElrondCalled != nil {
		return stub.IsQuorumReachedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// WasSetStatusPerformedOnElrond -
func (stub *ElrondToEthBridgeStub) WasSetStatusPerformedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasSetStatusPerformedOnElrondCalled != nil {
		return stub.WasSetStatusPerformedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// PerformSetStatusOnElrond -
func (stub *ElrondToEthBridgeStub) PerformSetStatusOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformSetStatusOnElrondCalled != nil {
		return stub.PerformSetStatusOnElrondCalled(ctx)
	}
	return notImplemented
}

// WaitForTransferConfirmation -
func (stub *ElrondToEthBridgeStub) WaitForTransferConfirmation(ctx context.Context) {
	stub.incrementFunctionCounter()
	if stub.WaitForTransferConfirmationCalled != nil {
		stub.WaitForTransferConfirmationCalled(ctx)
	}
}

// ProcessMaxRetriesOnElrond -
func (stub *ElrondToEthBridgeStub) ProcessMaxRetriesOnElrond() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnElrondCalled != nil {
		return stub.ProcessMaxRetriesOnElrondCalled()
	}
	return false
}

// ResetRetriesCountOnElrond -
func (stub *ElrondToEthBridgeStub) ResetRetriesCountOnElrond() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnElrondCalled != nil {
		stub.ResetRetriesCountOnElrondCalled()
	}
}

// ProcessMaxRetriesOnEthereum -
func (stub *ElrondToEthBridgeStub) ProcessMaxRetriesOnEthereum() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnEthereumCalled != nil {
		return stub.ProcessMaxRetriesOnEthereumCalled()
	}
	return false
}

// ResetRetriesCountOnEthereum -
func (stub *ElrondToEthBridgeStub) ResetRetriesCountOnEthereum() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnEthereumCalled != nil {
		stub.ResetRetriesCountOnEthereumCalled()
	}
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (stub *ElrondToEthBridgeStub) incrementFunctionCounter() {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	stub.functionCalledCounter[strings.ReplaceAll(runtime.FuncForPC(pc).Name(), stub.fullPath, "")]++
}

// GetFunctionCounter returns the called counter of a given function
func (stub *ElrondToEthBridgeStub) GetFunctionCounter(function string) int {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	return stub.functionCalledCounter[function]
}

// IsInterfaceNil -
func (stub *ElrondToEthBridgeStub) IsInterfaceNil() bool {
	return stub == nil
}
