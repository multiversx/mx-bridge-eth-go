package eth

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ClientWrapper represents the Ethereum client wrapper that the ethereum client can rely on
type ClientWrapper interface {
	core.StatusHandler
	GetNextPendingBatch(ctx context.Context) (contract.Batch, error)
	GetRelayers(ctx context.Context) ([]common.Address, error)
	WasBatchExecuted(ctx context.Context, batchNonce int64) (bool, error)
	WasBatchFinished(ctx context.Context, batchNonce int64) (bool, error)
	GetStatusesAfterExecution(ctx context.Context, batchNonceElrondETH int64) ([]uint8, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber uint64) (uint64, error)
	FinishCurrentPendingBatch(opts *bind.TransactOpts, batchNonce int64, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error)
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce int64, signatures [][]byte) (*types.Transaction, error)
	Quorum(ctx context.Context) (uint64, error)
}

// Erc20Contract defines the Ethereum ERC20 contract operations
type Erc20Contract interface {
	BalanceOf(ctx context.Context, account common.Address) (*big.Int, error)
	IsInterfaceNil() bool
}
