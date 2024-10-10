package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TokensMapper can convert a token bytes from one chain to another
type TokensMapper interface {
	ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error)
	IsInterfaceNil() bool
}

// Erc20ContractsHolder defines the Ethereum ERC20 contract operations
type Erc20ContractsHolder interface {
	BalanceOf(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error)
	IsInterfaceNil() bool
}

// SafeContractWrapper defines the operations for the safe contract
type SafeContractWrapper interface {
	DepositsCount(opts *bind.CallOpts) (uint64, error)
	BatchesCount(opts *bind.CallOpts) (uint64, error)
}

// MvxDataGetter defines the operations for the data getter operating on MultiversX chain
type MvxDataGetter interface {
	GetAllKnownTokens(ctx context.Context) ([][]byte, error)
	GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error)
	IsInterfaceNil() bool
}

// EthereumChainWrapper defines the operations of the Ethereum wrapper
type EthereumChainWrapper interface {
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address,
		recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int,
		signatures [][]byte) (*types.Transaction, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	Quorum(ctx context.Context) (*big.Int, error)
	GetRelayers(ctx context.Context) ([]common.Address, error)
	IsPaused(ctx context.Context) (bool, error)
}

// CryptoHandler defines the operations for a component that expose some crypto primitives
type CryptoHandler interface {
	Sign(msgHash common.Hash) ([]byte, error)
	GetAddress() common.Address
	CreateKeyedTransactor(chainId *big.Int) (*bind.TransactOpts, error)
	IsInterfaceNil() bool
}

// GasHandler defines the component able to fetch the current gas price
type GasHandler interface {
	GetCurrentGasPrice() (*big.Int, error)
	IsInterfaceNil() bool
}
