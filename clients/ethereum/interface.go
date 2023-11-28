package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
)

// ClientWrapper represents the Ethereum client wrapper that the ethereum client can rely on
type ClientWrapper interface {
	core.StatusHandler
	GetBatch(ctx context.Context, batchNonce *big.Int) (contract.Batch, error)
	GetBatchDeposits(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, error)
	GetRelayers(ctx context.Context) ([]common.Address, error)
	WasBatchExecuted(ctx context.Context, batchNonce *big.Int) (bool, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address,
		recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int,
		signatures [][]byte) (*types.Transaction, error)
	Quorum(ctx context.Context) (*big.Int, error)
	GetStatusesAfterExecution(ctx context.Context, batchID *big.Int) ([]byte, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	TokenMintedBalances(ctx context.Context, token common.Address) (*big.Int, error)
	WhitelistedTokensMintBurn(ctx context.Context, token common.Address) (bool, error)
	IsPaused(ctx context.Context) (bool, error)
}

// Erc20ContractsHolder defines the Ethereum ERC20 contract operations
type Erc20ContractsHolder interface {
	BalanceOf(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error)
	IsInterfaceNil() bool
}

// Broadcaster defines the operations for a component used for communication with other peers
type Broadcaster interface {
	BroadcastSignature(signature []byte, messageHash []byte)
	IsInterfaceNil() bool
}

// TokensMapper can convert a token bytes from one chain to another
type TokensMapper interface {
	ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error)
	IsInterfaceNil() bool
}

// GasHandler defines the component able to fetch the current gas price
type GasHandler interface {
	GetCurrentGasPrice() (*big.Int, error)
	IsInterfaceNil() bool
}

// SignaturesHolder defines the operations for a component that can hold and manage signatures
type SignaturesHolder interface {
	Signatures(messageHash []byte) [][]byte
	ClearStoredSignatures()
	IsInterfaceNil() bool
}

type erc20ContractWrapper interface {
	BalanceOf(ctx context.Context, account common.Address) (*big.Int, error)
	IsInterfaceNil() bool
}
