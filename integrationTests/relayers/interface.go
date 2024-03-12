package relayers

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

type bridgeComponents interface {
	MultiversXRelayerAddress() sdkCore.AddressHandler
	EthereumRelayerAddress() common.Address
	Start() error
	Close() error
}

type blockchainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}
