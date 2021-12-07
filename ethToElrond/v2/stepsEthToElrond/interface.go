package stepsEthToElrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// EthToElrondBridge defines the operations for the Ethereum to Elrond half-bridge
type EthToElrondBridge interface {
	GetLogger() logger.Logger
	MyTurnAsLeader() bool

	GetAndStoreActionID(ctx context.Context) (uint64, error)
	GetStoredActionID() uint64
	GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error
	GetStoredBatch() *clients.TransferBatch

	GetLastExecutedEthBatchIDFromElrond(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error
	WasTransferProposedOnElrond(ctx context.Context) (bool, error)
	ProposeTransferOnElrond(ctx context.Context) error
	WasProposedTransferSigned(ctx context.Context) (bool, error)
	SignProposedTransfer(ctx context.Context) error
	IsQuorumReached(ctx context.Context) (bool, error)
	WasActionIDPerformed(ctx context.Context) (bool, error)
	PerformActionID(ctx context.Context) error

	IsInterfaceNil() bool
}
