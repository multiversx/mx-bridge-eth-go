package elrondToEth

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// ElrondToEthBridge defines the operations for the Elrond to Ethereum half-bridge
type ElrondToEthBridge interface {
	GetLogger() logger.Logger
	MyTurnAsLeader() bool

	GetBatchFromElrond(ctx context.Context) (*clients.TransferBatch, error)
	StoreBatchFromElrond(ctx context.Context, batch *clients.TransferBatch) error
	GetStoredBatchFromElrond() *clients.TransferBatch

	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	SignTransferOnEthereum(ctx context.Context) error
	IsQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	PerformTransferOnEthereum(ctx context.Context) error

	GetAndStoreActionIDForSetStatusFromElrond(ctx context.Context) (uint64, error)
	GetStoredActionIDForSetStatus() uint64
	ResolveNewDepositsStatuses(numDeposits uint64)
	GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error)
	WasSetStatusProposedOnElrond(ctx context.Context) (bool, error)
	ProposeSetStatusOnElrond(ctx context.Context) error
	WasProposedSetStatusSignedOnElrond(ctx context.Context) (bool, error)
	SignProposedSetStatusOnElrond(ctx context.Context) error
	ProcessQuorumReachedOnElrond(ctx context.Context) (bool, error)
	WasSetStatusPerformedOnElrond(ctx context.Context) (bool, error)
	PerformSetStatusOnElrond(ctx context.Context) error

	WaitForTransferConfirmation(ctx context.Context)

	ProcessMaxRetriesOnElrond() bool
	ResetRetriesCountOnElrond()
	ProcessMaxRetriesOnEthereum() bool
	ResetRetriesCountOnEthereum()

	IsInterfaceNil() bool
}
