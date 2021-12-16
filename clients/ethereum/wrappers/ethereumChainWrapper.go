package wrappers

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ArgsEthereumChainWrapper is the DTO used to construct a ethereumChainWrapper instance
type ArgsEthereumChainWrapper struct {
	StatusHandler    core.StatusHandler
	MultiSigContract multiSigContract
	BlockchainClient blockchainClient
}

type ethereumChainWrapper struct {
	core.StatusHandler
	multiSigContract multiSigContract
	blockchainClient blockchainClient
}

// NewEthereumChainWrapper creates a new instance of type ethereumChainWrapper
func NewEthereumChainWrapper(args ArgsEthereumChainWrapper) (*ethereumChainWrapper, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &ethereumChainWrapper{
		StatusHandler:    args.StatusHandler,
		multiSigContract: args.MultiSigContract,
		blockchainClient: args.BlockchainClient,
	}, nil
}

func checkArgs(args ArgsEthereumChainWrapper) error {
	if check.IfNil(args.StatusHandler) {
		return errNilStatusHandler
	}
	if check.IfNilReflect(args.MultiSigContract) {
		return errNilMultiSigContract
	}
	if check.IfNilReflect(args.BlockchainClient) {
		return errNilBlockchainClient
	}

	return nil
}

// GetBatch returns the batch of transactions by providing the batch nonce
func (wrapper *ethereumChainWrapper) GetBatch(ctx context.Context, batchNonce *big.Int) (contract.Batch, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.GetBatch(&bind.CallOpts{Context: ctx}, batchNonce)
}

// GetRelayers returns all whitelisted ethereum addresses
func (wrapper *ethereumChainWrapper) GetRelayers(ctx context.Context) ([]common.Address, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.GetRelayers(&bind.CallOpts{Context: ctx})
}

// WasBatchExecuted returns true if the batch was executed
func (wrapper *ethereumChainWrapper) WasBatchExecuted(ctx context.Context, batchNonce *big.Int) (bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.WasBatchExecuted(&bind.CallOpts{Context: ctx}, batchNonce)
}

// ChainID returns the chain ID
func (wrapper *ethereumChainWrapper) ChainID(ctx context.Context) (*big.Int, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.blockchainClient.ChainID(ctx)
}

// BlockNumber returns the current ethereum block number
func (wrapper *ethereumChainWrapper) BlockNumber(ctx context.Context) (uint64, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	val, err := wrapper.blockchainClient.BlockNumber(ctx)
	if err != nil {
		return 0, err
	}

	wrapper.SetIntMetric(core.MetricLastQueriedEthereumBlockNumber, int(val))

	return val, nil
}

// NonceAt returns the account's nonce at the specified block number
func (wrapper *ethereumChainWrapper) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.blockchainClient.NonceAt(ctx, account, blockNumber)
}

// ExecuteTransfer will send an execute-transfer transaction on the ethereum chain
func (wrapper *ethereumChainWrapper) ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientTransactions, 1)
	_ = nonces // TODO: decide if we need to pass the nonces as well
	return wrapper.multiSigContract.ExecuteTransfer(opts, tokens, recipients, amounts, batchNonce, signatures)
}

// Quorum returns the current set quorum value
func (wrapper *ethereumChainWrapper) Quorum(ctx context.Context) (*big.Int, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.Quorum(&bind.CallOpts{Context: ctx})
}

// GetStatusesAfterExecution returns the statuses of the last executed transfer
func (wrapper *ethereumChainWrapper) GetStatusesAfterExecution(ctx context.Context, batchID *big.Int) ([]byte, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.GetStatusesAfterExecution(&bind.CallOpts{Context: ctx}, batchID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrapper *ethereumChainWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
