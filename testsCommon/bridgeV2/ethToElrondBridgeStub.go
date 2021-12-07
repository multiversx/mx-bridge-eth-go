package bridgeV2

import (
	"context"
	"fmt"
	"runtime"
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

	GetAndStoreActionIDCalled          func(ctx context.Context) (uint64, error)
	GetStoredActionIDCalled            func() uint64
	GetAndStoreBatchFromEthereumCalled func(ctx context.Context, nonce uint64) error
	GetStoredBatchCalled               func() *clients.TransferBatch

	GetLastExecutedEthBatchIDFromElrondCalled           func(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatchCalled func(ctx context.Context) error
	WasTransferProposedOnElrondCalled                   func(ctx context.Context) (bool, error)
	ProposeTransferOnElrondCalled                       func(ctx context.Context) error
	WasProposedTransferSignedCalled                     func(ctx context.Context) (bool, error)
	SignProposedTransferCalled                          func(ctx context.Context) error
	IsQuorumReachedCalled                               func(ctx context.Context) (bool, error)
	WasActionIDPerformedCalled                          func(ctx context.Context) (bool, error)
	PerformActionIDCalled                               func(ctx context.Context) error
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

// GetAndStoreActionID -
func (stub *EthToElrondBridgeStub) GetAndStoreActionID(ctx context.Context) (uint64, error) {
	stub.incrementFunctionCounter()
	if stub.GetAndStoreActionIDCalled != nil {
		return stub.GetAndStoreActionIDCalled(ctx)
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

// WasProposedTransferSigned -
func (stub *EthToElrondBridgeStub) WasProposedTransferSigned(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasProposedTransferSignedCalled != nil {
		return stub.WasProposedTransferSignedCalled(ctx)
	}
	return false, notImplemented
}

// SignProposedTransfer -
func (stub *EthToElrondBridgeStub) SignProposedTransfer(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.SignProposedTransferCalled != nil {
		return stub.SignProposedTransferCalled(ctx)
	}
	return notImplemented
}

// IsQuorumReached -
func (stub *EthToElrondBridgeStub) IsQuorumReached(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.IsQuorumReachedCalled != nil {
		return stub.IsQuorumReachedCalled(ctx)
	}
	return false, notImplemented
}

// WasActionIDPerformed -
func (stub *EthToElrondBridgeStub) WasActionIDPerformed(ctx context.Context) (bool, error) {
	stub.incrementFunctionCounter()
	if stub.WasActionIDPerformedCalled != nil {
		return stub.WasActionIDPerformedCalled(ctx)
	}
	return false, notImplemented
}

// PerformActionID -
func (stub *EthToElrondBridgeStub) PerformActionID(ctx context.Context) error {
	stub.incrementFunctionCounter()
	if stub.PerformActionIDCalled != nil {
		return stub.PerformActionIDCalled(ctx)
	}
	return notImplemented
}

// -------- helper functions

// incrementFunctionCounter increments the counter for the function that called it
func (stub *EthToElrondBridgeStub) incrementFunctionCounter() {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	pc, _, _, _ := runtime.Caller(1)
	fmt.Printf("BridgeExecutorMock: called %s\n", runtime.FuncForPC(pc).Name())
	stub.functionCalledCounter[runtime.FuncForPC(pc).Name()]++
}

// GetFunctionCounter returns the called counter of a given function
func (stub *EthToElrondBridgeStub) GetFunctionCounter(function string) int {
	stub.mutExecutor.Lock()
	defer stub.mutExecutor.Unlock()

	return stub.functionCalledCounter[fullPath+function]
}

// IsInterfaceNil -
func (stub *EthToElrondBridgeStub) IsInterfaceNil() bool {
	return stub == nil
}
