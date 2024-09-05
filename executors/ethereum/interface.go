package ethereum

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
