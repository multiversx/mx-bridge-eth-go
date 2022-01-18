package bridge

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// BridgeExecutorStub -
type BridgeExecutorStub struct {
	functionCalledCounter map[string]int
	mutExecutor           sync.RWMutex
	fullPath              string

	PrintInfoCalled                                        func(logLevel logger.LogLevel, message string, extras ...interface{})
	MyTurnAsLeaderCalled                                   func() bool
	GetBatchFromElrondCalled                               func(ctx context.Context) (*clients.TransferBatch, error)
	StoreBatchFromElrondCalled                             func(batch *clients.TransferBatch) error
	GetStoredBatchCalled                                   func() *clients.TransferBatch
	GetLastExecutedEthBatchIDFromElrondCalled              func(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatchCalled    func(ctx context.Context) error
	GetAndStoreActionIDForProposeTransferOnElrondCalled    func(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromElrondCalled func(ctx context.Context) (uint64, error)
	GetStoredActionIDCalled                                func() uint64
	WasTransferProposedOnElrondCalled                      func(ctx context.Context) (bool, error)
	ProposeTransferOnElrondCalled                          func(ctx context.Context) error
	WasSetStatusProposedOnElrondCalled                     func(ctx context.Context) (bool, error)
	ProposeSetStatusOnElrondCalled                         func(ctx context.Context) error
	WasActionSignedOnElrondCalled                          func(ctx context.Context) (bool, error)
	SignActionOnElrondCalled                               func(ctx context.Context) error
	ProcessQuorumReachedOnElrondCalled                     func(ctx context.Context) (bool, error)
	WasActionPerformedOnElrondCalled                       func(ctx context.Context) (bool, error)
	PerformActionOnElrondCalled                            func(ctx context.Context) error
	ResolveNewDepositsStatusesCalled                       func(numDeposits uint64)
	ProcessMaxRetriesOnElrondCalled                        func() bool
	ResetRetriesCountOnElrondCalled                        func()
	GetAndStoreBatchFromEthereumCalled                     func(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereumCalled                   func(ctx context.Context) (bool, error)
	SignTransferOnEthereumCalled                           func() error
	PerformTransferOnEthereumCalled                        func(ctx context.Context) error
	ProcessQuorumReachedOnEthereumCalled                   func(ctx context.Context) (bool, error)
	WaitForTransferConfirmationCalled                      func(ctx context.Context)
	WaitAndReturnFinalBatchStatusesCalled                  func(ctx context.Context) []byte
	GetBatchStatusesFromEthereumCalled                     func(ctx context.Context) ([]byte, error)
	ProcessMaxRetriesOnEthereumCalled                      func() bool
	ResetRetriesCountOnEthereumCalled                      func()
	ClearStoredP2PSignaturesForEthereumCalled              func()
}

// NewBridgeExecutorStub creates a new BridgeExecutorStub instance
func NewBridgeExecutorStub() *BridgeExecutorStub {
	return &BridgeExecutorStub{
		functionCalledCounter: make(map[string]int),
		fullPath:              "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge.(*BridgeExecutorStub).",
	}
}

// PrintInfo -
func (stub *BridgeExecutorStub) PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{}) {
	stub.incrementFunctionCounter()
	if stub.PrintInfoCalled != nil {
		stub.PrintInfoCalled(logLevel, message, extras...)
	}
}

// MyTurnAsLeader -
func (stub *BridgeExecutorStub) MyTurnAsLeader() bool {
	stub.incrementFunctionCounter()
	if stub.MyTurnAsLeaderCalled != nil {
		return stub.MyTurnAsLeaderCalled()
	}
	return false
}

// GetBatchFromElrond -
func (stub *BridgeExecutorStub) GetBatchFromElrond(ctx context.Context) (*clients.TransferBatch, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchFromElrondCalled != nil {
		return stub.GetBatchFromElrondCalled(ctx)
	}
	return nil, notImplemented
}

// StoreBatchFromElrond -
func (stub *BridgeExecutorStub) StoreBatchFromElrond(batch *clients.TransferBatch) error {
	stub.incrementFunctionCounter()
	if stub.StoreBatchFromElrondCalled != nil {
		return stub.StoreBatchFromElrondCalled(batch)
	}
	return notImplemented
}

// GetStoredBatch -
func (stub *BridgeExecutorStub) GetStoredBatch() *clients.TransferBatch {
	stub.incrementFunctionCounter()
	if stub.GetStoredBatchCalled != nil {
		return stub.GetStoredBatchCalled()
	}
	return nil
}

// GetLastExecutedEthBatchIDFromElrond -
func (stub *BridgeExecutorStub) GetLastExecutedEthBatchIDFromElrond(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetLastExecutedEthBatchIDFromElrondCalled != nil {
		return stub.GetLastExecutedEthBatchIDFromElrondCalled(ctx)
	}
	return 0, notImplemented
}

// VerifyLastDepositNonceExecutedOnEthereumBatch -
func (stub *BridgeExecutorStub) VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled != nil {
		return stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled(ctx)
	}
	return notImplemented
}

// GetAndStoreActionIDForProposeTransferOnElrond -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeTransferOnElrond(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeTransferOnElrondCalled != nil {
		return stub.GetAndStoreActionIDForProposeTransferOnElrondCalled(ctx)
	}
	return 0, notImplemented
}

// GetAndStoreActionIDForProposeSetStatusFromElrond -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeSetStatusFromElrond(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeSetStatusFromElrondCalled != nil {
		return stub.GetAndStoreActionIDForProposeSetStatusFromElrondCalled(ctx)
	}
	return 0, notImplemented
}

// GetStoredActionID -
func (stub *BridgeExecutorStub) GetStoredActionID() uint64 {
	stub.incrementFunctionCounter()
	if stub.GetStoredActionIDCalled != nil {
		return stub.GetStoredActionIDCalled()
	}
	return 0
}

// WasTransferProposedOnElrond -
func (stub *BridgeExecutorStub) WasTransferProposedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferProposedOnElrondCalled != nil {
		return stub.WasTransferProposedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// ProposeTransferOnElrond -
func (stub *BridgeExecutorStub) ProposeTransferOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeTransferOnElrondCalled != nil {
		return stub.ProposeTransferOnElrondCalled(ctx)
	}
	return notImplemented
}

// WasSetStatusProposedOnElrond -
func (stub *BridgeExecutorStub) WasSetStatusProposedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasSetStatusProposedOnElrondCalled != nil {
		return stub.WasSetStatusProposedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// ProposeSetStatusOnElrond -
func (stub *BridgeExecutorStub) ProposeSetStatusOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeSetStatusOnElrondCalled != nil {
		return stub.ProposeSetStatusOnElrondCalled(ctx)
	}
	return notImplemented
}

// WasActionSignedOnElrond -
func (stub *BridgeExecutorStub) WasActionSignedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionSignedOnElrondCalled != nil {
		return stub.WasActionSignedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// SignActionOnElrond -
func (stub *BridgeExecutorStub) SignActionOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignActionOnElrondCalled != nil {
		return stub.SignActionOnElrondCalled(ctx)
	}
	return notImplemented
}

// ProcessQuorumReachedOnElrond -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnElrondCalled != nil {
		return stub.ProcessQuorumReachedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// WasActionPerformedOnElrond -
func (stub *BridgeExecutorStub) WasActionPerformedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionPerformedOnElrondCalled != nil {
		return stub.WasActionPerformedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// PerformActionOnElrond -
func (stub *BridgeExecutorStub) PerformActionOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformActionOnElrondCalled != nil {
		return stub.PerformActionOnElrondCalled(ctx)
	}
	return notImplemented
}

// ResolveNewDepositsStatuses -
func (stub *BridgeExecutorStub) ResolveNewDepositsStatuses(numDeposits uint64) {
	stub.incrementFunctionCounter()
	if stub.ResolveNewDepositsStatusesCalled != nil {
		stub.ResolveNewDepositsStatusesCalled(numDeposits)
	}
}

// ProcessMaxRetriesOnElrond -
func (stub *BridgeExecutorStub) ProcessMaxRetriesOnElrond() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnElrondCalled != nil {
		return stub.ProcessMaxRetriesOnElrondCalled()
	}
	return false
}

// ResetRetriesCountOnElrond -
func (stub *BridgeExecutorStub) ResetRetriesCountOnElrond() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnElrondCalled != nil {
		stub.ResetRetriesCountOnElrondCalled()
	}
}

// GetAndStoreBatchFromEthereum -
func (stub *BridgeExecutorStub) GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreBatchFromEthereumCalled != nil {
		return stub.GetAndStoreBatchFromEthereumCalled(ctx, nonce)
	}
	return notImplemented
}

// WasTransferPerformedOnEthereum -
func (stub *BridgeExecutorStub) WasTransferPerformedOnEthereum(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferPerformedOnEthereumCalled != nil {
		return stub.WasTransferPerformedOnEthereumCalled(ctx)
	}
	return false, notImplemented
}

// SignTransferOnEthereum -
func (stub *BridgeExecutorStub) SignTransferOnEthereum() error {
	stub.incrementFunctionCounter()
	if stub.SignTransferOnEthereumCalled != nil {
		return stub.SignTransferOnEthereumCalled()
	}
	return notImplemented
}

// PerformTransferOnEthereum -
func (stub *BridgeExecutorStub) PerformTransferOnEthereum(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformTransferOnEthereumCalled != nil {
		return stub.PerformTransferOnEthereumCalled(ctx)
	}
	return notImplemented
}

// ProcessQuorumReachedOnEthereum -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnEthereum(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnEthereumCalled != nil {
		return stub.ProcessQuorumReachedOnEthereumCalled(ctx)
	}
	return false, notImplemented
}

// WaitForTransferConfirmation -
func (stub *BridgeExecutorStub) WaitForTransferConfirmation(ctx context.Context) {
	stub.incrementFunctionCounter()
	if stub.WaitForTransferConfirmationCalled != nil {
		stub.WaitForTransferConfirmationCalled(ctx)
	}
}

// WaitAndReturnFinalBatchStatuses -
func (stub *BridgeExecutorStub) WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte {
	stub.incrementFunctionCounter()
	if stub.WaitAndReturnFinalBatchStatusesCalled != nil {
		return stub.WaitAndReturnFinalBatchStatusesCalled(ctx)
	}
	return nil
}

// GetBatchStatusesFromEthereum -
func (stub *BridgeExecutorStub) GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchStatusesFromEthereumCalled != nil {
		return stub.GetBatchStatusesFromEthereumCalled(ctx)
	}
	return nil, notImplemented
}

// ProcessMaxRetriesOnEthereum -
func (stub *BridgeExecutorStub) ProcessMaxRetriesOnEthereum() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnEthereumCalled != nil {
		return stub.ProcessMaxRetriesOnEthereumCalled()
	}
	return false
}

// ResetRetriesCountOnEthereum -
func (stub *BridgeExecutorStub) ResetRetriesCountOnEthereum() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnEthereumCalled != nil {
		stub.ResetRetriesCountOnEthereumCalled()
	}
}

// ClearStoredP2PSignaturesForEthereum -
func (stub *BridgeExecutorStub) ClearStoredP2PSignaturesForEthereum() {
	stub.incrementFunctionCounter()
	if stub.ClearStoredP2PSignaturesForEthereumCalled != nil {
		stub.ClearStoredP2PSignaturesForEthereumCalled()
	}
}

// IsInterfaceNil -
func (stub *BridgeExecutorStub) IsInterfaceNil() bool {
	return stub == nil
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (stub *BridgeExecutorStub) incrementFunctionCounter() {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	stub.functionCalledCounter[strings.ReplaceAll(runtime.FuncForPC(pc).Name(), stub.fullPath, "")]++
}

// GetFunctionCounter returns the called counter of a given function
func (stub *BridgeExecutorStub) GetFunctionCounter(function string) int {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	return stub.functionCalledCounter[function]
}
