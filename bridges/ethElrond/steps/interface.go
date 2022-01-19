package steps

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// Executor defines a generic bridge interface able to handle both halves of the bridge
type Executor interface {
	PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{})
	MyTurnAsLeader() bool

	GetBatchFromElrond(ctx context.Context) (*clients.TransferBatch, error)
	StoreBatchFromElrond(batch *clients.TransferBatch) error
	GetStoredBatch() *clients.TransferBatch
	GetLastExecutedEthBatchIDFromElrond(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error

	GetAndStoreActionIDForProposeTransferOnElrond(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromElrond(ctx context.Context) (uint64, error)
	GetStoredActionID() uint64

	WasTransferProposedOnElrond(ctx context.Context) (bool, error)
	ProposeTransferOnElrond(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnElrond() bool
	ResetRetriesOnWasTransferProposedOnElrond()

	WasSetStatusProposedOnElrond(ctx context.Context) (bool, error)
	ProposeSetStatusOnElrond(ctx context.Context) error

	WasActionSignedOnElrond(ctx context.Context) (bool, error)
	SignActionOnElrond(ctx context.Context) error

	ProcessQuorumReachedOnElrond(ctx context.Context) (bool, error)
	WasActionPerformedOnElrond(ctx context.Context) (bool, error)
	PerformActionOnElrond(ctx context.Context) error
	ResolveNewDepositsStatuses(numDeposits uint64)

	ProcessMaxRetriesOnElrond() bool
	ResetRetriesCountOnElrond()

	GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	SignTransferOnEthereum() error
	PerformTransferOnEthereum(ctx context.Context) error
	ProcessQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	WaitForTransferConfirmation(ctx context.Context)
	WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte
	GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error)

	ProcessMaxRetriesOnEthereum() bool
	ResetRetriesCountOnEthereum()
	ClearStoredP2PSignaturesForEthereum()

	IsInterfaceNil() bool
}
