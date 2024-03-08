package steps

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// Executor defines a generic bridge interface able to handle both halves of the bridge
type Executor interface {
	PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{})
	MyTurnAsLeader() bool

	GetBatchFromMultiversX(ctx context.Context) (*clients.TransferBatch, error)
	StoreBatchFromMultiversX(batch *clients.TransferBatch) error
	GetStoredBatch() *clients.TransferBatch

	GetLastExecutedEthBatchIDFromMultiversX(ctx context.Context) (uint64, error)
	VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error

	GetAndStoreActionIDForProposeTransferOnMultiversX(ctx context.Context) (uint64, error)
	GetAndStoreActionIDForProposeSetStatusFromMultiversX(ctx context.Context) (uint64, error)
	GetStoredActionID() uint64

	WasTransferProposedOnMultiversX(ctx context.Context) (bool, error)
	ProposeTransferOnMultiversX(ctx context.Context) error
	ProcessMaxRetriesOnWasTransferProposedOnMultiversX() bool
	ResetRetriesOnWasTransferProposedOnMultiversX()

	WasSetStatusProposedOnMultiversX(ctx context.Context) (bool, error)
	ProposeSetStatusOnMultiversX(ctx context.Context) error

	WasActionSignedOnMultiversX(ctx context.Context) (bool, error)
	SignActionOnMultiversX(ctx context.Context) error

	ProcessQuorumReachedOnMultiversX(ctx context.Context) (bool, error)
	WasActionPerformedOnMultiversX(ctx context.Context) (bool, error)
	PerformActionOnMultiversX(ctx context.Context) error
	ResolveNewDepositsStatuses(numDeposits uint64)

	ProcessMaxQuorumRetriesOnMultiversX() bool
	ResetRetriesCountOnMultiversX()

	GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error
	WasTransferPerformedOnEthereum(ctx context.Context) (bool, error)
	SignTransferOnEthereum() error
	PerformTransferOnEthereum(ctx context.Context) error
	ProcessQuorumReachedOnEthereum(ctx context.Context) (bool, error)
	WaitForTransferConfirmation(ctx context.Context)
	WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte
	GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error)

	ProcessMaxQuorumRetriesOnEthereum() bool
	ResetRetriesCountOnEthereum()
	ClearStoredP2PSignaturesForEthereum()

	ValidateBatch(ctx context.Context, batch *clients.TransferBatch) (bool, error)
	CheckMultiversXClientAvailability(ctx context.Context) error
	CheckEthereumClientAvailability(ctx context.Context) error
	CheckAvailableTokens(ctx context.Context, ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error

	IsInterfaceNil() bool
}
