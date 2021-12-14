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

	GetAndStoreBatchFromElrond(ctx context.Context) error
	GetStoredBatch() *clients.TransferBatch

	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	SignTransferOnEthereum(ctx context.Context) error
	IsQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	PerformTransferOnEthereum(ctx context.Context) error

	GetAndStoreActionIDForSetStatusFromElrond(ctx context.Context) (uint64, error)
	GetStoredActionIDForSetStatus() uint64
	WasSetStatusProposedOnElrond(ctx context.Context) (bool, error)
	ProposeSetStatusOnElrond(ctx context.Context) error
	WasProposedSetStatusSignedOnElrond(ctx context.Context) (bool, error)
	SignProposedSetStatusOnElrond(ctx context.Context) error
	IsQuorumReachedOnElrond(ctx context.Context) (bool, error)
	WasSetStatusPerformedOnElrond(ctx context.Context) (bool, error)
	PerformSetStatusOnElrond(ctx context.Context) error

	ProcessMaxRetriesOnElrond() bool
	ResetRetriesCountOnElrond()
	ProcessMaxRetriesOnEthereum() bool
	ResetRetriesCountOnEthereum()

	IsInterfaceNil() bool
}
