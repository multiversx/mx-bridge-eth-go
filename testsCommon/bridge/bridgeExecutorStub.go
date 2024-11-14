package bridge

import (
	"context"
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
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
	GetLastExecutedEthBatchIDFromMultiversXCalled              func(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatchCalled        func(ctx context.Context) error
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
	GetAndStoreBatchFromEthereumCalled                         func(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereumCalled                       func(ctx context.Context) (bool, error)
	SignTransferOnEthereumCalled                               func() error
	PerformTransferOnEthereumCalled                            func(ctx context.Context) error
	ProcessQuorumReachedOnEthereumCalled                       func(ctx context.Context) (bool, error)
	WaitForTransferConfirmationCalled                          func(ctx context.Context)
	WaitAndReturnFinalBatchStatusesCalled                      func(ctx context.Context) []byte
	GetBatchStatusesFromEthereumCalled                         func(ctx context.Context) ([]byte, error)
	ProcessMaxQuorumRetriesOnEthereumCalled                    func() bool
	ResetRetriesCountOnEthereumCalled                          func()
	ClearStoredP2PSignaturesForEthereumCalled                  func()
	CheckMultiversXClientAvailabilityCalled                    func(ctx context.Context) error
	CheckEthereumClientAvailabilityCalled                      func(ctx context.Context) error
	CheckAvailableTokensCalled                                 func(ctx context.Context, ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error
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
	return nil, notImplemented
}

// StoreBatchFromMultiversX -
func (stub *BridgeExecutorStub) StoreBatchFromMultiversX(batch *bridgeCore.TransferBatch) error {
	stub.incrementFunctionCounter()
	if stub.StoreBatchFromMultiversXCalled != nil {
		return stub.StoreBatchFromMultiversXCalled(batch)
	}
	return notImplemented
}

// GetStoredBatch -
func (stub *BridgeExecutorStub) GetStoredBatch() *bridgeCore.TransferBatch {
	stub.incrementFunctionCounter()
	if stub.GetStoredBatchCalled != nil {
		return stub.GetStoredBatchCalled()
	}
	return nil
}

// GetLastExecutedEthBatchIDFromMultiversX -
func (stub *BridgeExecutorStub) GetLastExecutedEthBatchIDFromMultiversX(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetLastExecutedEthBatchIDFromMultiversXCalled != nil {
		return stub.GetLastExecutedEthBatchIDFromMultiversXCalled(ctx)
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

// GetAndStoreActionIDForProposeTransferOnMultiversX -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeTransferOnMultiversX(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeTransferOnMultiversXCalled != nil {
		return stub.GetAndStoreActionIDForProposeTransferOnMultiversXCalled(ctx)
	}
	return 0, notImplemented
}

// GetAndStoreActionIDForProposeSetStatusFromMultiversX -
func (stub *BridgeExecutorStub) GetAndStoreActionIDForProposeSetStatusFromMultiversX(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDForProposeSetStatusFromMultiversXCalled != nil {
		return stub.GetAndStoreActionIDForProposeSetStatusFromMultiversXCalled(ctx)
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

// WasTransferProposedOnMultiversX -
func (stub *BridgeExecutorStub) WasTransferProposedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferProposedOnMultiversXCalled != nil {
		return stub.WasTransferProposedOnMultiversXCalled(ctx)
	}
	return false, notImplemented
}

// ProposeTransferOnMultiversX -
func (stub *BridgeExecutorStub) ProposeTransferOnMultiversX(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeTransferOnMultiversXCalled != nil {
		return stub.ProposeTransferOnMultiversXCalled(ctx)
	}
	return notImplemented
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
	return false, notImplemented
}

// ProposeSetStatusOnMultiversX -
func (stub *BridgeExecutorStub) ProposeSetStatusOnMultiversX(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeSetStatusOnMultiversXCalled != nil {
		return stub.ProposeSetStatusOnMultiversXCalled(ctx)
	}
	return notImplemented
}

// WasActionSignedOnMultiversX -
func (stub *BridgeExecutorStub) WasActionSignedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionSignedOnMultiversXCalled != nil {
		return stub.WasActionSignedOnMultiversXCalled(ctx)
	}
	return false, notImplemented
}

// SignActionOnMultiversX -
func (stub *BridgeExecutorStub) SignActionOnMultiversX(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignActionOnMultiversXCalled != nil {
		return stub.SignActionOnMultiversXCalled(ctx)
	}
	return notImplemented
}

// ProcessQuorumReachedOnMultiversX -
func (stub *BridgeExecutorStub) ProcessQuorumReachedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.ProcessQuorumReachedOnMultiversXCalled != nil {
		return stub.ProcessQuorumReachedOnMultiversXCalled(ctx)
	}
	return false, notImplemented
}

// WasActionPerformedOnMultiversX -
func (stub *BridgeExecutorStub) WasActionPerformedOnMultiversX(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionPerformedOnMultiversXCalled != nil {
		return stub.WasActionPerformedOnMultiversXCalled(ctx)
	}
	return false, notImplemented
}

// PerformActionOnMultiversX -
func (stub *BridgeExecutorStub) PerformActionOnMultiversX(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformActionOnMultiversXCalled != nil {
		return stub.PerformActionOnMultiversXCalled(ctx)
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

// ProcessMaxQuorumRetriesOnEthereum -
func (stub *BridgeExecutorStub) ProcessMaxQuorumRetriesOnEthereum() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxQuorumRetriesOnEthereumCalled != nil {
		return stub.ProcessMaxQuorumRetriesOnEthereumCalled()
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

// CheckMultiversXClientAvailability -
func (stub *BridgeExecutorStub) CheckMultiversXClientAvailability(ctx context.Context) error {
	if stub.CheckMultiversXClientAvailabilityCalled != nil {
		return stub.CheckMultiversXClientAvailabilityCalled(ctx)
	}
	return notImplemented
}

// CheckEthereumClientAvailability -
func (stub *BridgeExecutorStub) CheckEthereumClientAvailability(ctx context.Context) error {
	if stub.CheckEthereumClientAvailabilityCalled != nil {
		return stub.CheckEthereumClientAvailabilityCalled(ctx)
	}
	return notImplemented
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
func (stub *BridgeExecutorStub) CheckAvailableTokens(ctx context.Context, ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	if stub.CheckAvailableTokensCalled != nil {
		return stub.CheckAvailableTokensCalled(ctx, ethTokens, mvxTokens, amounts, direction)
	}

	return nil
}
