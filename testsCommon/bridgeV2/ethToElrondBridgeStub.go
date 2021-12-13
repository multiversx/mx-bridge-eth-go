package bridgeV2

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var fullPath = "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2.(*EthToElrondBridgeStub)."

type EthToElrondBridgeStub struct {
	functionCalledCounter map[string]int
	mutExecutor           sync.RWMutex

	GetLoggerCalled      func() logger.Logger
	MyTurnAsLeaderCalled func() bool

	GetAndStoreActionIDFromElrondCalled                 func(ctx context.Context) (uint64, error)
	GetStoredActionIDCalled                             func() uint64
	GetAndStoreBatchFromEthereumCalled                  func(ctx context.Context, nonce uint64) error
	GetStoredBatchCalled                                func() *clients.TransferBatch
	GetLastExecutedEthBatchIDFromElrondCalled           func(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatchCalled func(ctx context.Context) error
	WasTransferProposedOnElrondCalled                   func(ctx context.Context) (bool, error)
	ProposeTransferOnElrondCalled                       func(ctx context.Context) error
	WasProposedTransferSignedOnElrondCalled             func(ctx context.Context) (bool, error)
	SignProposedTransferOnElrondCalled                  func(ctx context.Context) error
	IsQuorumReachedOnElrondCalled                       func(ctx context.Context) (bool, error)
	WasActionIDPerformedOnElrondCalled                  func(ctx context.Context) (bool, error)
	PerformActionIDOnElrondCalled                       func(ctx context.Context) error
	ProcessMaxRetriesOnElrondCalled                     func() bool
	ResetRetriesCountOnElrondCalled                     func()
}

// NewEthToElrondBridgeStub creates a new EthToElrondBridgeStub instance
func NewEthToElrondBridgeStub() *EthToElrondBridgeStub {
	return &EthToElrondBridgeStub{
		functionCalledCounter: make(map[string]int),
	}
}

// GetLogger -
func (stub *EthToElrondBridgeStub) GetLogger() logger.Logger {
	stub.incrementFunctionCounter()
	if stub.GetLoggerCalled != nil {
		return stub.GetLoggerCalled()
	}
	return nil
}

// MyTurnAsLeader -
func (stub *EthToElrondBridgeStub) MyTurnAsLeader() bool {
	stub.incrementFunctionCounter()
	if stub.MyTurnAsLeaderCalled != nil {
		return stub.MyTurnAsLeaderCalled()
	}
	return false
}

// GetAndStoreActionIDFromElrond -
func (stub *EthToElrondBridgeStub) GetAndStoreActionIDFromElrond(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDFromElrondCalled != nil {
		return stub.GetAndStoreActionIDFromElrondCalled(ctx)
	}
	return 0, notImplemented
}

// GetStoredActionID -
func (stub *EthToElrondBridgeStub) GetStoredActionID() uint64 {
	stub.incrementFunctionCounter()
	if stub.GetStoredActionIDCalled != nil {
		return stub.GetStoredActionIDCalled()
	}
	return 0
}

// GetAndStoreBatchFromEthereum -
func (stub *EthToElrondBridgeStub) GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreBatchFromEthereumCalled != nil {
		return stub.GetAndStoreBatchFromEthereumCalled(ctx, nonce)
	}
	return notImplemented
}

// GetStoredBatch -
func (stub *EthToElrondBridgeStub) GetStoredBatch() *clients.TransferBatch {
	stub.incrementFunctionCounter()
	if stub.GetStoredBatchCalled != nil {
		return stub.GetStoredBatchCalled()
	}
	return nil
}

// GetLastExecutedEthBatchIDFromElrond -
func (stub *EthToElrondBridgeStub) GetLastExecutedEthBatchIDFromElrond(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetLastExecutedEthBatchIDFromElrondCalled != nil {
		return stub.GetLastExecutedEthBatchIDFromElrondCalled(ctx)
	}
	return 0, notImplemented
}

// VerifyLastDepositNonceExecutedOnEthereumBatch -
func (stub *EthToElrondBridgeStub) VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled != nil {
		return stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled(ctx)
	}
	return notImplemented
}

// WasTransferProposedOnElrond -
func (stub *EthToElrondBridgeStub) WasTransferProposedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasTransferProposedOnElrondCalled != nil {
		return stub.WasTransferProposedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// ProposeTransferOnElrond -
func (stub *EthToElrondBridgeStub) ProposeTransferOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.ProposeTransferOnElrondCalled != nil {
		return stub.ProposeTransferOnElrondCalled(ctx)
	}
	return notImplemented
}

// WasProposedTransferSignedOnElrond -
func (stub *EthToElrondBridgeStub) WasProposedTransferSignedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasProposedTransferSignedOnElrondCalled != nil {
		return stub.WasProposedTransferSignedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// SignProposedTransferOnElrond -
func (stub *EthToElrondBridgeStub) SignProposedTransferOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignProposedTransferOnElrondCalled != nil {
		return stub.SignProposedTransferOnElrondCalled(ctx)
	}
	return notImplemented
}

// IsQuorumReachedOnElrond -
func (stub *EthToElrondBridgeStub) IsQuorumReachedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.IsQuorumReachedOnElrondCalled != nil {
		return stub.IsQuorumReachedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// WasActionIDPerformedOnElrond -
func (stub *EthToElrondBridgeStub) WasActionIDPerformedOnElrond(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionIDPerformedOnElrondCalled != nil {
		return stub.WasActionIDPerformedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

// PerformActionIDOnElrond -
func (stub *EthToElrondBridgeStub) PerformActionIDOnElrond(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformActionIDOnElrondCalled != nil {
		return stub.PerformActionIDOnElrondCalled(ctx)
	}
	return notImplemented
}

// ProcessMaxRetriesOnElrond -
func (stub *EthToElrondBridgeStub) ProcessMaxRetriesOnElrond() bool {
	stub.incrementFunctionCounter()
	if stub.ProcessMaxRetriesOnElrondCalled != nil {
		return stub.ProcessMaxRetriesOnElrondCalled()
	}
	return false
}

// ResetRetriesCountOnElrond -
func (stub *EthToElrondBridgeStub) ResetRetriesCountOnElrond() {
	stub.incrementFunctionCounter()
	if stub.ResetRetriesCountOnElrondCalled != nil {
		stub.ResetRetriesCountOnElrondCalled()
	}
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (stub *EthToElrondBridgeStub) incrementFunctionCounter() {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	stub.functionCalledCounter[strings.ReplaceAll(runtime.FuncForPC(pc).Name(), fullPath, "")]++
}

// GetFunctionCounter returns the called counter of a given function
func (stub *EthToElrondBridgeStub) GetFunctionCounter(function string) int {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	return stub.functionCalledCounter[function]
}

// IsInterfaceNil -
func (stub *EthToElrondBridgeStub) IsInterfaceNil() bool {
	return stub == nil
}
