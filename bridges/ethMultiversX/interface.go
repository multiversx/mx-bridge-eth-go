package ethmultiversx

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
)

// MultiversXClient defines the behavior of the MultiversX client able to communicate with the MultiversX chain
type MultiversXClient interface {
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
	GetCurrentNonce(ctx context.Context) (uint64, error)

	ProposeSetStatus(ctx context.Context, batch *clients.TransferBatch) (string, error)
	ProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (string, error)
	Sign(ctx context.Context, actionID uint64) (string, error)
	WasSigned(ctx context.Context, actionID uint64) (bool, error)
	PerformAction(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error)
	CheckClientAvailability(ctx context.Context) error
	IsMintBurnToken(ctx context.Context, token []byte) (bool, error)
	IsNativeToken(ctx context.Context, token []byte) (bool, error)
	TotalBalances(ctx context.Context, token []byte) (*big.Int, error)
	MintBalances(ctx context.Context, token []byte) (*big.Int, error)
	BurnBalances(ctx context.Context, token []byte) (*big.Int, error)
	Close() error
	IsInterfaceNil() bool
}

// EthereumClient defines the behavior of the Ethereum client able to communicate with the Ethereum chain
type EthereumClient interface {
	GetBatch(ctx context.Context, nonce uint64) (*clients.TransferBatch, error)
	WasExecuted(ctx context.Context, batchID uint64) (bool, error)
	GenerateMessageHash(batch *batchProcessor.ArgListsBatch, batchId uint64) (common.Hash, error)

	BroadcastSignatureForMessageHash(msgHash common.Hash)
	ExecuteTransfer(ctx context.Context, msgHash common.Hash, batch *batchProcessor.ArgListsBatch, batchId uint64, quorum int) (string, error)
	GetTransactionsStatuses(ctx context.Context, batchId uint64) ([]byte, error)
	GetQuorumSize(ctx context.Context) (*big.Int, error)
	IsQuorumReached(ctx context.Context, msgHash common.Hash) (bool, error)
	GetBatchSCMetadata(ctx context.Context, nonce uint64) ([]*contract.SCExecProxyERC20SCDeposit, error)
	CheckClientAvailability(ctx context.Context) error
	CheckRequiredBalance(ctx context.Context, erc20Address common.Address, value *big.Int) error
	TotalBalances(ctx context.Context, token common.Address) (*big.Int, error)
	MintBalances(ctx context.Context, token common.Address) (*big.Int, error)
	BurnBalances(ctx context.Context, token common.Address) (*big.Int, error)
	MintBurnTokens(ctx context.Context, token common.Address) (bool, error)
	NativeTokens(ctx context.Context, token common.Address) (bool, error)
	WhitelistedTokens(ctx context.Context, token common.Address) (bool, error)
	IsInterfaceNil() bool
}

// TopologyProvider is able to manage the current relayers topology
type TopologyProvider interface {
	MyTurnAsLeader() bool
	IsInterfaceNil() bool
}

// SignaturesHolder defines the operations for a component that can hold and manage signatures
type SignaturesHolder interface {
	Signatures(messageHash []byte) [][]byte
	ClearStoredSignatures()
	IsInterfaceNil() bool
}
