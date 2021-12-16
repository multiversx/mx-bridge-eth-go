package bridge

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// Executor defines a generic bridge interface able to handle both halves of the bridge
type Executor interface {
	GetLogger() logger.Logger
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

	WasSetStatusProposedOnElrond(ctx context.Context) (bool, error)
	ProposeSetStatusOnElrond(ctx context.Context) error

	WasActionSignedOnElrond(ctx context.Context) (bool, error)
	SignActionOnElrond(ctx context.Context) error

	IsQuorumReachedOnElrond(ctx context.Context) (bool, error)
	WasActionPerformedOnElrond(ctx context.Context) (bool, error)
	PerformActionOnElrond(ctx context.Context) error
	ResolveNewDepositsStatuses(numDeposits uint64)

	ProcessMaxRetriesOnElrond() bool
	ResetRetriesCountOnElrond()

	GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	SignTransferOnEthereum() error
	PerformTransferOnEthereum(ctx context.Context) error
	IsQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	WaitForTransferConfirmation(ctx context.Context)
	GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error)

	ProcessMaxRetriesOnEthereum() bool
	ResetRetriesCountOnEthereum()

	IsInterfaceNil() bool
}
