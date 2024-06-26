package balanceValidator

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// MultiversXClient defines the behavior of the MultiversX client able to communicate with the MultiversX chain
type MultiversXClient interface {
	IsMintBurnToken(ctx context.Context, token []byte) (bool, error)
	IsNativeToken(ctx context.Context, token []byte) (bool, error)
	TotalBalances(ctx context.Context, token []byte) (*big.Int, error)
	MintBalances(ctx context.Context, token []byte) (*big.Int, error)
	BurnBalances(ctx context.Context, token []byte) (*big.Int, error)
	CheckRequiredBalance(ctx context.Context, token []byte, value *big.Int) error
	IsInterfaceNil() bool
}

// EthereumClient defines the behavior of the Ethereum client able to communicate with the Ethereum chain
type EthereumClient interface {
	TotalBalances(ctx context.Context, token common.Address) (*big.Int, error)
	MintBalances(ctx context.Context, token common.Address) (*big.Int, error)
	BurnBalances(ctx context.Context, token common.Address) (*big.Int, error)
	MintBurnTokens(ctx context.Context, token common.Address) (bool, error)
	NativeTokens(ctx context.Context, token common.Address) (bool, error)
	CheckRequiredBalance(ctx context.Context, erc20Address common.Address, value *big.Int) error
	IsInterfaceNil() bool
}
