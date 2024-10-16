package bridgeV2Wrappers

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/executors/ethereum/bridgeV2Wrappers/contract"
)

type multiSigContract interface {
	GetBatch(opts *bind.CallOpts, batchNonce *big.Int) (contract.Batch, error)
	GetBatchDeposits(opts *bind.CallOpts, batchNonce *big.Int) ([]contract.Deposit, error)
	GetRelayers(opts *bind.CallOpts) ([]common.Address, error)
	WasBatchExecuted(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, depositNonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	Quorum(opts *bind.CallOpts) (*big.Int, error)
	GetStatusesAfterExecution(opts *bind.CallOpts, batchID *big.Int) ([]byte, error)
	Paused(opts *bind.CallOpts) (bool, error)
}

type blockchainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
}
