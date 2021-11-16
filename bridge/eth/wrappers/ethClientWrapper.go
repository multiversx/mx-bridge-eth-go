package wrappers

import (
	"context"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ArgsEthClientWrapper is the DTO used to construct an ethClientWrapper instance
type ArgsEthClientWrapper struct {
	BridgeContract   BridgeContract
	BlockchainClient BlockchainClient
	StatusHandler    core.StatusHandler
}

type ethClientWrapper struct {
	core.StatusHandler
	bridgeContract   BridgeContract
	blockchainClient BlockchainClient
}

// NewEthClientWrapper creates a new instance of type ethClientWrapper
func NewEthClientWrapper(args ArgsEthClientWrapper) (*ethClientWrapper, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &ethClientWrapper{
		StatusHandler:    args.StatusHandler,
		bridgeContract:   args.BridgeContract,
		blockchainClient: args.BlockchainClient,
	}, nil
}

func checkArgs(args ArgsEthClientWrapper) error {
	if check.IfNil(args.StatusHandler) {
		return ErrNilStatusHandler
	}
	if check.IfNilReflect(args.BridgeContract) {
		return ErrNilBrdgeContract
	}
	if check.IfNilReflect(args.BlockchainClient) {
		return ErrNilBlockchainClient
	}

	return nil
}

// GetNextPendingBatch returns the next pending batch of transactions
func (wrapper *ethClientWrapper) GetNextPendingBatch(ctx context.Context) (contract.Batch, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.bridgeContract.GetNextPendingBatch(&bind.CallOpts{Context: ctx})
}

// GetRelayers returns all whitelisted ethereum addresses
func (wrapper *ethClientWrapper) GetRelayers(ctx context.Context) ([]common.Address, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.bridgeContract.GetRelayers(&bind.CallOpts{Context: ctx})
}

// WasBatchExecuted returns true if the batch was executed
func (wrapper *ethClientWrapper) WasBatchExecuted(ctx context.Context, batchNonce int64) (bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.bridgeContract.WasBatchExecuted(&bind.CallOpts{Context: ctx}, big.NewInt(batchNonce))
}

// WasBatchFinished returns true if the batch was finished
func (wrapper *ethClientWrapper) WasBatchFinished(ctx context.Context, batchNonce int64) (bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.bridgeContract.WasBatchFinished(&bind.CallOpts{Context: ctx}, big.NewInt(batchNonce))
}

// GetStatusesAfterExecution returns the transaction statuses after execution
func (wrapper *ethClientWrapper) GetStatusesAfterExecution(ctx context.Context, batchNonceElrondETH int64) ([]uint8, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.bridgeContract.GetStatusesAfterExecution(&bind.CallOpts{Context: ctx}, big.NewInt(batchNonceElrondETH))
}

// ChainID returns the chain ID
func (wrapper *ethClientWrapper) ChainID(ctx context.Context) (*big.Int, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.blockchainClient.ChainID(ctx)
}

// BlockNumber returns the current ethereum block number
func (wrapper *ethClientWrapper) BlockNumber(ctx context.Context) (uint64, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	val, err := wrapper.blockchainClient.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}

	wrapper.SetIntMetric(core.MetricLastQueriedEthereumBlockNumber, int(val))

	return val, nil
}

// NonceAt returns the account's nonce at the specified block number
func (wrapper *ethClientWrapper) NonceAt(ctx context.Context, account common.Address, blockNumber uint64) (uint64, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.blockchainClient.NonceAt(ctx, account, big.NewInt(0).SetUint64(blockNumber))
}

// FinishCurrentPendingBatch will send a set-status transaction on the ethereum chain
func (wrapper *ethClientWrapper) FinishCurrentPendingBatch(opts *bind.TransactOpts, batchNonce int64, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientTransactions, 1)
	return wrapper.bridgeContract.FinishCurrentPendingBatch(opts, big.NewInt(batchNonce), newDepositStatuses, signatures)
}

// ExecuteTransfer will send an execute-transfer transaction on the ethereum chain
func (wrapper *ethClientWrapper) ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce int64, signatures [][]byte) (*types.Transaction, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientTransactions, 1)
	return wrapper.bridgeContract.ExecuteTransfer(opts, tokens, recipients, amounts, big.NewInt(batchNonce), signatures)
}

// Quorum returns the current set quorum value
func (wrapper *ethClientWrapper) Quorum(ctx context.Context) (uint64, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	value, err := wrapper.bridgeContract.Quorum(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, err
	}

	if value.Cmp(big.NewInt(math.MaxUint32)) > 0 {
		return 0, ErrInvalidQuorumValue
	}

	return value.Uint64(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrapper *ethClientWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
