package wrappers

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BridgeContract defines the supported Ethereum contract operations
type BridgeContract interface {
	GetNextPendingBatch(opts *bind.CallOpts) (contract.Batch, error)
	FinishCurrentPendingBatch(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error)
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	WasBatchExecuted(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	WasBatchFinished(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	Quorum(opts *bind.CallOpts) (*big.Int, error)
	GetStatusesAfterExecution(opts *bind.CallOpts, batchNonceElrondETH *big.Int) ([]uint8, error)
	GetRelayers(opts *bind.CallOpts) ([]common.Address, error)
}

// GenericErc20Contract defines the Ethereum ERC20 contract operations
type GenericErc20Contract interface {
	BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error)
}

// BlockchainClient defines the RPC operations on the Ethereum node
type BlockchainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
}
