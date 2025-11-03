package bridge

import (
	"context"
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"sync"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// BridgeExecutorStub -
type BridgeExecutorStub struct {
	functionCalledCounter map[string]int
	mutExecutor           sync.RWMutex
	fullPath              string

	PrintInfoCalled                                            func(logLevel logger.LogLevel, message string, extras ...interface{})
	MyTurnAsLeaderCalled                                       func() bool
	GetBatchFromMultiversXCalled                               func(ctx context.Context) (*bridgeCore.TransferBatch, error)
	StoreBatchFromMultiversXCalled                             func(batch *bridgeCore.TransferBatch) error
	GetStoredBatchCalled                                       func() *bridgeCore.TransferBatch
	GetLastExecutedPeerBatchIDFromMultiversXCalled             func(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnPeerBatchCalled            func(ctx context.Context) error
	GetAndStoreActionIDForProposeTransferOnMultiversXCalled    func(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromMultiversXCalled func(ctx context.Context) (uint64, error)
	GetStoredActionIDCalled                                    func() uint64
	WasTransferProposedOnMultiversXCalled                      func(ctx context.Context) (bool, error)
	ProposeTransferOnMultiversXCalled                          func(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnMultiversXCalled   func() bool
	ResetRetriesOnWasTransferProposedOnMultiversXCalled        func()
	WasSetStatusProposedOnMultiversXCalled                     func(ctx context.Context) (bool, error)
	ProposeSetStatusOnMultiversXCalled                         func(ctx context.Context) error
	WasActionSignedOnMultiversXCalled                          func(ctx context.Context) (bool, error)
	SignActionOnMultiversXCalled                               func(ctx context.Context) error
	ProcessQuorumReachedOnMultiversXCalled                     func(ctx context.Context) (bool, error)
	WasActionPerformedOnMultiversXCalled                       func(ctx context.Context) (bool, error)
	PerformActionOnMultiversXCalled                            func(ctx context.Context) error
	ResolveNewDepositsStatusesCalled                           func(numDeposits uint64)
	ProcessMaxQuorumRetriesOnMultiversXCalled                  func() bool
	ResetRetriesCountOnMultiversXCalled                        func()
	GetAndStoreBatchFromPeerChainCalled                        func(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnPeerChainCalled                      func(ctx context.Context) (bool, error)
	SignTransferOnPeerChainCalled                              func() error
	PerformTransferOnPeerChainCalled                           func(ctx context.Context) error
	ProcessQuorumReachedOnPeerChainCalled                      func(ctx context.Context) (bool, error)
	WaitForTransferConfirmationCalled                          func(ctx context.Context)
	WaitAndReturnFinalBatchStatusesCalled                      func(ctx context.Context) []byte
	GetBatchStatusesFromPeerChainCalled                        func(ctx context.Context) ([]byte, error)
	ProcessMaxQuorumRetriesOnPeerChainCalled                   func() bool
	ResetRetriesCountOnPeerChainCalled                         func()
	ClearStoredP2PSignaturesForPeerChainCalled                 func()
	CheckMultiversXClientAvailabilityCalled                    func(ctx context.Context) error
	CheckPeerClientAvailabilityCalled                          func(ctx context.Context) error
	CheckAvailableTokensCalled                                 func(ctx context.Context, peerTokens [][]byte, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error
}

// NewBridgeExecutorStub creates a new BridgeExecutorStub instance
func NewBridgeExecutorStub() *BridgeExecutorStub {
	return &BridgeExecutorStub{
		functionCalledCounter: make(map[string]int),
		fullPath:              "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge.(*BridgeExecutorStub).",
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

// GetBatchFromMultiversX -
func (stub *BridgeExecutorStub) GetBatchFromMultiversX(ctx context.Context) (*bridgeCore.TransferBatch, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchFromMultiversXCalled != nil {
		return stub.GetBatchFromMultiversXCalled(ctx)
	}
	return nil, errNotImplemented
}

// StoreBatchFromMultiversX -
func (stub *BridgeExecutorStub) StoreBatchFromMultiversX(batch *bridgeCore.TransferBatch) error {
	stub.incrementFunctionCounter()
	if stub.StoreBatchFromMultiversXCalled != nil {
		return stub.StoreBatchFromMultiversXCalled(batch)
	}
	return errNotImplemented
}

// GetStoredBatch -
func (stub *BridgeExecutorStub) GetStoredBatch() *bridgeCore.TransferBatch {
	stub.incrementFunctionCounter()
	if stub.GetStoredBatchCalled != nil {
		return stub.GetStoredBatchCalled()
	}
	return nil
}

// GetLastExecutedPeerBatchIDFromMultiversX -
func (stub *BridgeExecutorStub) GetLastExecutedPeerBatchIDFromMultiversX(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetLastExecutedPeerBatchIDFromMultiversXCalled != nil {
		return stub.GetLastExecutedPeerBatchIDFromMultiversXCalled(ctx)
	}
	return 0, errNotImplemented
}

// VerifyLastDepositNonceExecutedOnPeerBatch -
func (stub *BridgeExecutorStub) VerifyLastDepositNonceExecutedOnPeerBatch(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.VerifyLastDepositNonceExecutedOnPeerBatchCalled != nil {
		return stub.VerifyLastDepositNonceExecutedOnPeerBatchCalled(ctx)
	}
	return errNotImplemented
}

// GetAndStoreActionIDForProposeTransferOnMultiversX -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeTransferOnMultiversX(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeTransferOnMultiversXCalled != nil {
		return stub.GetAndStoreActionIDForProposeTransferOnMultiversXCalled(ctx)
	}
	return 0, errNotImplemented
}

// GetAndStoreActionIDForProposeSetStatusFromMultiversX -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeSetStatusFromMultiversX(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeSetStatusFromMultiversXCalled != nil {
		return stub.GetAndStoreActionIDForProposeSetStatusFromMultiversXCalled(ctx)
	}
	return 0, errNotImplemented
}

// GetStoredActionID -
func (stub *BridgeExecutorStub) GetStoredActionID() uint64 {
	stub.incrementFunctionCounter()
	if stub.GetStoredActionIDCalled != nil {
		return stub.GetStoredActionIDCalled()
	}
	return 0
}

// WasTransferProposedOnMultiversX -
func (stub *BridgeExecutorStub) WasTransferProposedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferProposedOnMultiversXCalled != nil {
		return stub.WasTransferProposedOnMultiversXCalled(ctx)
	}
	return false, errNotImplemented
}

// ProposeTransferOnMultiversX -
func (stub *BridgeExecutorStub) ProposeTransferOnMultiversX(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeTransferOnMultiversXCalled != nil {
		return stub.ProposeTransferOnMultiversXCalled(ctx)
	}
	return errNotImplemented
}

// ProcessMaxRetriesOnWasTransferProposedOnMultiversX -
func (stub *BridgeExecutorStub) ProcessMaxRetriesOnWasTransferProposedOnMultiversX() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnWasTransferProposedOnMultiversXCalled != nil {
		return stub.ProcessMaxRetriesOnWasTransferProposedOnMultiversXCalled()
	}
	return false
}

// ResetRetriesOnWasTransferProposedOnMultiversX -
func (stub *BridgeExecutorStub) ResetRetriesOnWasTransferProposedOnMultiversX() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesOnWasTransferProposedOnMultiversXCalled != nil {
		stub.ResetRetriesOnWasTransferProposedOnMultiversXCalled()
	}
}

// WasSetStatusProposedOnMultiversX -
func (stub *BridgeExecutorStub) WasSetStatusProposedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasSetStatusProposedOnMultiversXCalled != nil {
		return stub.WasSetStatusProposedOnMultiversXCalled(ctx)
	}
	return false, errNotImplemented
}

// ProposeSetStatusOnMultiversX -
func (stub *BridgeExecutorStub) ProposeSetStatusOnMultiversX(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeSetStatusOnMultiversXCalled != nil {
		return stub.ProposeSetStatusOnMultiversXCalled(ctx)
	}
	return errNotImplemented
}

// WasActionSignedOnMultiversX -
func (stub *BridgeExecutorStub) WasActionSignedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionSignedOnMultiversXCalled != nil {
		return stub.WasActionSignedOnMultiversXCalled(ctx)
	}
	return false, errNotImplemented
}

// SignActionOnMultiversX -
func (stub *BridgeExecutorStub) SignActionOnMultiversX(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignActionOnMultiversXCalled != nil {
		return stub.SignActionOnMultiversXCalled(ctx)
	}
	return errNotImplemented
}

// ProcessQuorumReachedOnMultiversX -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnMultiversXCalled != nil {
		return stub.ProcessQuorumReachedOnMultiversXCalled(ctx)
	}
	return false, errNotImplemented
}

// WasActionPerformedOnMultiversX -
func (stub *BridgeExecutorStub) WasActionPerformedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionPerformedOnMultiversXCalled != nil {
		return stub.WasActionPerformedOnMultiversXCalled(ctx)
	}
	return false, errNotImplemented
}

// PerformActionOnMultiversX -
func (stub *BridgeExecutorStub) PerformActionOnMultiversX(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformActionOnMultiversXCalled != nil {
		return stub.PerformActionOnMultiversXCalled(ctx)
	}
	return errNotImplemented
}

// ResolveNewDepositsStatuses -
func (stub *BridgeExecutorStub) ResolveNewDepositsStatuses(numDeposits uint64) {
	stub.incrementFunctionCounter()
	if stub.ResolveNewDepositsStatusesCalled != nil {
		stub.ResolveNewDepositsStatusesCalled(numDeposits)
	}
}

// ProcessMaxQuorumRetriesOnMultiversX -
func (stub *BridgeExecutorStub) ProcessMaxQuorumRetriesOnMultiversX() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxQuorumRetriesOnMultiversXCalled != nil {
		return stub.ProcessMaxQuorumRetriesOnMultiversXCalled()
	}
	return false
}

// ResetRetriesCountOnMultiversX -
func (stub *BridgeExecutorStub) ResetRetriesCountOnMultiversX() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnMultiversXCalled != nil {
		stub.ResetRetriesCountOnMultiversXCalled()
	}
}

// GetAndStoreBatchFromPeerChain -
func (stub *BridgeExecutorStub) GetAndStoreBatchFromPeerChain(ctx context.Context, nonce uint64) error {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreBatchFromPeerChainCalled != nil {
		return stub.GetAndStoreBatchFromPeerChainCalled(ctx, nonce)
	}
	return errNotImplemented
}

// WasTransferPerformedOnPeerChain -
func (stub *BridgeExecutorStub) WasTransferPerformedOnPeerChain(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferPerformedOnPeerChainCalled != nil {
		return stub.WasTransferPerformedOnPeerChainCalled(ctx)
	}
	return false, errNotImplemented
}

// SignTransferOnPeerChain -
func (stub *BridgeExecutorStub) SignTransferOnPeerChain() error {
	stub.incrementFunctionCounter()
	if stub.SignTransferOnPeerChainCalled != nil {
		return stub.SignTransferOnPeerChainCalled()
	}
	return errNotImplemented
}

// PerformTransferOnPeerChain -
func (stub *BridgeExecutorStub) PerformTransferOnPeerChain(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformTransferOnPeerChainCalled != nil {
		return stub.PerformTransferOnPeerChainCalled(ctx)
	}
	return errNotImplemented
}

// ProcessQuorumReachedOnPeerChain -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnPeerChain(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnPeerChainCalled != nil {
		return stub.ProcessQuorumReachedOnPeerChainCalled(ctx)
	}
	return false, errNotImplemented
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

// GetBatchStatusesFromPeerChain -
func (stub *BridgeExecutorStub) GetBatchStatusesFromPeerChain(ctx context.Context) ([]byte, error) {
	stub.incrementFunctionCounter()
	if stub.GetBatchStatusesFromPeerChainCalled != nil {
		return stub.GetBatchStatusesFromPeerChainCalled(ctx)
	}
	return nil, errNotImplemented
}

// ProcessMaxQuorumRetriesOnPeerChain -
func (stub *BridgeExecutorStub) ProcessMaxQuorumRetriesOnPeerChain() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxQuorumRetriesOnPeerChainCalled != nil {
		return stub.ProcessMaxQuorumRetriesOnPeerChainCalled()
	}
	return false
}

// ResetRetriesCountOnPeerChain -
func (stub *BridgeExecutorStub) ResetRetriesCountOnPeerChain() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnPeerChainCalled != nil {
		stub.ResetRetriesCountOnPeerChainCalled()
	}
}

// ClearStoredP2PSignaturesForPeerChain -
func (stub *BridgeExecutorStub) ClearStoredP2PSignaturesForPeerChain() {
	stub.incrementFunctionCounter()
	if stub.ClearStoredP2PSignaturesForPeerChainCalled != nil {
		stub.ClearStoredP2PSignaturesForPeerChainCalled()
	}
}

// CheckMultiversXClientAvailability -
func (stub *BridgeExecutorStub) CheckMultiversXClientAvailability(ctx context.Context) error {
	if stub.CheckMultiversXClientAvailabilityCalled != nil {
		return stub.CheckMultiversXClientAvailabilityCalled(ctx)
	}
	return errNotImplemented
}

// CheckPeerClientAvailability -
func (stub *BridgeExecutorStub) CheckPeerClientAvailability(ctx context.Context) error {
	if stub.CheckPeerClientAvailabilityCalled != nil {
		return stub.CheckPeerClientAvailabilityCalled(ctx)
	}
	return errNotImplemented
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

// CheckAvailableTokens -
func (stub *BridgeExecutorStub) CheckAvailableTokens(ctx context.Context, peerTokens [][]byte, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	if stub.CheckAvailableTokensCalled != nil {
		return stub.CheckAvailableTokensCalled(ctx, peerTokens, mvxTokens, amounts, direction)
	}

	return nil
}
