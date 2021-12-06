package bridgeV2

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type EthToElrondBridgeStub struct {
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

func (stub *EthToElrondBridgeStub) GetLogger() logger.Logger {
	if stub.GetLoggerCalled != nil {
		return stub.GetLoggerCalled()
	}
	return nil
}

func (stub *EthToElrondBridgeStub) MyTurnAsLeader() bool {
	if stub.MyTurnAsLeaderCalled != nil {
		return stub.MyTurnAsLeaderCalled()
	}
	return false
}

func (stub *EthToElrondBridgeStub) GetAndStoreActionID(ctx context.Context) (uint64, error) {
	if stub.GetAndStoreActionIDCalled != nil {
		return stub.GetAndStoreActionIDCalled(ctx)
	}
	return 0, notImplemented
}

func (stub *EthToElrondBridgeStub) GetStoredActionID() uint64 {
	if stub.GetStoredActionIDCalled != nil {
		return stub.GetStoredActionIDCalled()
	}
	return 0
}

func (stub *EthToElrondBridgeStub) GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error {
	if stub.GetAndStoreBatchFromEthereumCalled != nil {
		return stub.GetAndStoreBatchFromEthereumCalled(ctx, nonce)
	}
	return notImplemented
}

func (stub *EthToElrondBridgeStub) GetStoredBatch() *clients.TransferBatch {
	if stub.GetStoredBatchCalled != nil {
		return stub.GetStoredBatchCalled()
	}
	return nil
}

func (stub *EthToElrondBridgeStub) GetLastExecutedEthBatchIDFromElrond(ctx context.Context) (uint64, error) {
	if stub.GetLastExecutedEthBatchIDFromElrondCalled != nil {
		return stub.GetLastExecutedEthBatchIDFromElrondCalled(ctx)
	}
	return 0, notImplemented
}

func (stub *EthToElrondBridgeStub) VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error {
	if stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled != nil {
		return stub.VerifyLastDepositNonceExecutedOnEthereumBatchCalled(ctx)
	}
	return notImplemented
}

func (stub *EthToElrondBridgeStub) WasTransferProposedOnElrond(ctx context.Context) (bool, error) {
	if stub.WasTransferProposedOnElrondCalled != nil {
		return stub.WasTransferProposedOnElrondCalled(ctx)
	}
	return false, notImplemented
}

func (stub *EthToElrondBridgeStub) ProposeTransferOnElrond(ctx context.Context) error {
	if stub.ProposeTransferOnElrondCalled != nil {
		return stub.ProposeTransferOnElrondCalled(ctx)
	}
	return notImplemented
}

func (stub *EthToElrondBridgeStub) WasProposedTransferSigned(ctx context.Context) (bool, error) {
	if stub.WasProposedTransferSignedCalled != nil {
		return stub.WasProposedTransferSignedCalled(ctx)
	}
	return false, notImplemented
}

func (stub *EthToElrondBridgeStub) SignProposedTransfer(ctx context.Context) error {
	if stub.SignProposedTransferCalled != nil {
		return stub.SignProposedTransferCalled(ctx)
	}
	return notImplemented
}

func (stub *EthToElrondBridgeStub) IsQuorumReached(ctx context.Context) (bool, error) {
	if stub.IsQuorumReachedCalled != nil {
		return stub.IsQuorumReached(ctx)
	}
	return false, notImplemented
}

func (stub *EthToElrondBridgeStub) WasActionIDPerformed(ctx context.Context) (bool, error) {
	if stub.WasActionIDPerformedCalled != nil {
		return stub.WasActionIDPerformed(ctx)
	}
	return false, notImplemented
}

func (stub *EthToElrondBridgeStub) PerformActionID(ctx context.Context) error {
	if stub.PerformActionIDCalled != nil {
		return stub.PerformActionID(ctx)
	}
	return notImplemented
}
