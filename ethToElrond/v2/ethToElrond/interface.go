package ethToElrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// EthToElrondBridge defines the operations for the Ethereum to Elrond half-bridge
type EthToElrondBridge interface {
	GetLogger() logger.Logger
	MyTurnAsLeader() bool

	GetAndStoreActionIDFromElrond(ctx context.Context) (uint64, error)
	GetStoredActionID() uint64
	GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error
	GetStoredBatch() *clients.TransferBatch

	GetLastExecutedEthBatchIDFromElrond(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error
	WasTransferProposedOnElrond(ctx context.Context) (bool, error)
	ProposeTransferOnElrond(ctx context.Context) error
	WasProposedTransferSignedOnElrond(ctx context.Context) (bool, error)
	SignProposedTransferOnElrond(ctx context.Context) error
	IsQuorumReachedOnElrond(ctx context.Context) (bool, error)
	WasActionIDPerformedOnElrond(ctx context.Context) (bool, error)
	PerformActionIDOnElrond(ctx context.Context) error

	IsMaxRetriesReachedOnElrond() bool
	ResetRetriesCountOnElrond()
	IsMaxRetriesReachedOnEthereum() bool
	ResetRetriesCountOnEthereum()

	IsInterfaceNil() bool
}
