package ethereum

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

// ClientWrapper represents the Ethereum client wrapper that the ethereum client can rely on
type ClientWrapper interface {
	core.StatusHandler
	GetBatch(ctx context.Context, batchNonce *big.Int) (contract.Batch, error)
	GetRelayers(ctx context.Context) ([]common.Address, error)
	WasBatchExecuted(ctx context.Context, batchNonce *big.Int) (bool, error)
	GetStatusesAfterExecution(ctx context.Context, batchNonce *big.Int) ([]uint8, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address,
		recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int,
		signatures [][]byte) (*types.Transaction, error)
	Quorum(ctx context.Context) (*big.Int, error)
}

// ArgsBalanceOf is the argument DTO for Erc20ContractHolder.BalanceOf function
type ArgsBalanceOf struct {
	Context      context.Context
	Address      common.Address
	ERC20Address common.Address
}

// Erc20ContractsHolder defines the Ethereum ERC20 contract operations
type Erc20ContractsHolder interface {
	BalanceOf(args ArgsBalanceOf) (*big.Int, error)
	IsInterfaceNil() bool
}

// Broadcaster defines the operations for a component used for communication with other peers
type Broadcaster interface {
	BroadcastSignature(signature []byte, messageHash []byte)
	IsInterfaceNil() bool
}

// TokenMapper can convert a token bytes from one chain to another
type TokenMapper interface {
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
	IsInterfaceNil() bool
}
