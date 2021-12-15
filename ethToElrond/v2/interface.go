package v2

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ethereum/go-ethereum/common"
)

// Bridge defines a generic bridge interface able to handle both halves of the bridge
type Bridge interface {
	GetLogger() logger.Logger

	MyTurnAsLeaderOnElrond() bool
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
	ResolveNewDepositsStatuses(ctx context.Context, numDeposits uint64) error

	ProcessMaxRetriesOnElrond() bool
	ResetRetriesCountOnElrond()

	MyTurnAsLeaderOnEthereum() bool
	GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	SignTransferOnEthereum(ctx context.Context) error
	PerformTransferOnEthereum(ctx context.Context) error
	IsQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	WaitForTransferConfirmation(ctx context.Context)
	GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error)

	ProcessMaxRetriesOnEthereum() bool
	ResetRetriesCountOnEthereum()


	IsInterfaceNil() bool
}

// ElrondClient defines the behavior of the Elrond client able to communicate with the Elrond chain
type ElrondClient interface {
	GetPending(ctx context.Context) (*clients.TransferBatch, error)
	GetCurrentBatchAsDataBytes(ctx context.Context) ([][]byte, error)
	WasProposedTransfer(ctx context.Context, batch *clients.TransferBatch) (bool, error)
	QuorumReached(ctx context.Context, actionID uint64) (bool, error)
	WasExecuted(ctx context.Context, actionID uint64) (bool, error)
	GetActionIDForProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (uint64, error)
	WasProposedSetStatus(ctx context.Context, batch *clients.TransferBatch) (bool, error)
	GetTransactionsStatuses(ctx context.Context, batchID uint64) ([]byte, error)
	GetActionIDForSetStatusOnPendingTransfer(ctx context.Context, batch *clients.TransferBatch) (uint64, error)
	GetLastExecutedEthBatchID(ctx context.Context) (uint64, error)
	GetLastExecutedEthTxID(ctx context.Context) (uint64, error)

	ProposeSetStatus(ctx context.Context, batch *clients.TransferBatch) (string, error)
	ResolveNewDeposits(ctx context.Context, batch *clients.TransferBatch) error
	ProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (string, error)
	Sign(ctx context.Context, actionID uint64) (string, error)
	WasSigned(ctx context.Context, actionID uint64) (bool, error)
	PerformAction(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error)
	GetMaxNumberOfRetriesOnQuorumReached() uint64
	Close() error
	IsInterfaceNil() bool
}

// EthereumClient defines the behavior of the Ethereum client able to communicate with the Ethereum chain
type EthereumClient interface {
	GetBatch(ctx context.Context, nonce uint64) (*clients.TransferBatch, error)
	WasExecuted(ctx context.Context, batchID uint64) (bool, error)
	GenerateMessageHash(batch *clients.TransferBatch) (common.Hash, error)

	BroadcastSignatureForMessageHash(msgHash common.Hash)
	ExecuteTransfer(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error)
	GetMaxNumberOfRetriesOnQuorumReached() uint64
	GetQuorumSize(ctx context.Context) (*big.Int, error)
	IsQuorumReached(ctx context.Context, msgHash common.Hash,) (bool, error)
	IsInterfaceNil() bool
}

// TopologyProvider is able to manage the current relayers topology
type TopologyProvider interface {
	MyTurnAsLeader() bool
	IsInterfaceNil() bool
}
