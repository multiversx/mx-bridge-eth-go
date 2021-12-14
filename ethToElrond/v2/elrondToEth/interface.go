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

	GetAndStoreActionIDFromElrond(ctx context.Context) (uint64, error)
	GetStoredActionID() uint64
	GetAndStoreBatchFromElrond(ctx context.Context, nonce uint64) error
	GetStoredBatch() *clients.TransferBatch

	WasTransferProposedOnEthereum(ctx context.Context) (bool, error)
	ProposeTransferOnEthereum(ctx context.Context) error
	WasProposedTransferSignedOnEthereum(ctx context.Context) (bool, error)
	SignProposedTransferOnEthereum(ctx context.Context) error
	IsQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	PerformTransferOnEthereum(ctx context.Context) error

	WasSetStatusProposedOnElrond(ctx context.Context) (bool, error)
	ProposeSetStatusOnElrond(ctx context.Context) error
	WasProposedSetStatusSignedOnElrond(ctx context.Context) (bool, error)
	SignProposedSetStatusOnElrond(ctx context.Context) error
	IsQuorumReachedOnElrond(ctx context.Context) (bool, error)
	WasSetStatusPerformedOnElrond(ctx context.Context) (bool, error)
	PerformSetStatusOnElrond(ctx context.Context) error

	ProcessMaxRetriesOnElrond() bool
	ResetRetriesCountOnElrond()

	IsInterfaceNil() bool
}
