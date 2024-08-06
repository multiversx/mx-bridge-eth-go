package balanceValidator

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCommon "github.com/multiversx/mx-bridge-eth-go/common"
)

// MultiversXClient defines the behavior of the MultiversX client able to communicate with the MultiversX chain
type MultiversXClient interface {
	GetPendingBatch(ctx context.Context) (*bridgeCommon.TransferBatch, error)
	GetBatch(ctx context.Context, batchID uint64) (*bridgeCommon.TransferBatch, error)
	GetLastExecutedEthBatchID(ctx context.Context) (uint64, error)
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
	GetBatch(ctx context.Context, nonce uint64) (*bridgeCommon.TransferBatch, bool, error)
	TotalBalances(ctx context.Context, token common.Address) (*big.Int, error)
	MintBalances(ctx context.Context, token common.Address) (*big.Int, error)
	BurnBalances(ctx context.Context, token common.Address) (*big.Int, error)
	MintBurnTokens(ctx context.Context, token common.Address) (bool, error)
	NativeTokens(ctx context.Context, token common.Address) (bool, error)
	CheckRequiredBalance(ctx context.Context, erc20Address common.Address, value *big.Int) error
	GetTransactionsStatuses(ctx context.Context, batchId uint64) ([]byte, error)
	IsInterfaceNil() bool
}
