package balanceValidator

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
)

// MultiversXClient defines the behavior of the MultiversX client able to communicate with the MultiversX chain
type MultiversXClient interface {
	GetPendingBatch(ctx context.Context) (*bridgeCore.TransferBatch, error)
	GetBatch(ctx context.Context, batchID uint64) (*bridgeCore.TransferBatch, error)
	GetLastExecutedEthBatchID(ctx context.Context) (uint64, error)
	GetLastMvxBatchID(ctx context.Context) (uint64, error)
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
	GetBatch(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error)
	TotalBalances(ctx context.Context, token common.Address) (*big.Int, error)
	MintBalances(ctx context.Context, token common.Address) (*big.Int, error)
	BurnBalances(ctx context.Context, token common.Address) (*big.Int, error)
	MintBurnTokens(ctx context.Context, token common.Address) (bool, error)
	NativeTokens(ctx context.Context, token common.Address) (bool, error)
	CheckRequiredBalance(ctx context.Context, erc20Address common.Address, value *big.Int) error
	WasExecuted(ctx context.Context, mvxBatchID uint64) (bool, error)
	IsInterfaceNil() bool
}
