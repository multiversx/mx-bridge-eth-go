package wrappers

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// ArgsEthereumChainWrapper is the DTO used to construct a ethereumChainWrapper instance
type ArgsEthereumChainWrapper struct {
	StatusHandler    core.StatusHandler
	MultiSigContract multiSigContract
	SafeContract     safeContract
	BlockchainClient blockchainClient
}

type ethereumChainWrapper struct {
	core.StatusHandler
	multiSigContract multiSigContract
	safeContract     safeContract
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
		safeContract:     args.SafeContract,
		blockchainClient: args.BlockchainClient,
	}, nil
}

func checkArgs(args ArgsEthereumChainWrapper) error {
	if check.IfNil(args.StatusHandler) {
		return clients.ErrNilStatusHandler
	}
	if check.IfNilReflect(args.MultiSigContract) {
		return errNilMultiSigContract
	}
	if check.IfNilReflect(args.SafeContract) {
		return errNilSafeContract
	}
	if check.IfNilReflect(args.BlockchainClient) {
		return errNilBlockchainClient
	}

	return nil
}

// GetBatch returns the batch of transactions by providing the batch nonce
func (wrapper *ethereumChainWrapper) GetBatch(ctx context.Context, batchNonce *big.Int) (contract.Batch, bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.GetBatch(&bind.CallOpts{Context: ctx}, batchNonce)
}

// GetBatchDeposits returns the transactions of a batch by providing the batch nonce
func (wrapper *ethereumChainWrapper) GetBatchDeposits(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.GetBatchDeposits(&bind.CallOpts{Context: ctx}, batchNonce)
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

// FilterLogs executes a query and returns matching logs and events
func (wrapper *ethereumChainWrapper) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.blockchainClient.FilterLogs(ctx, q)
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
func (wrapper *ethereumChainWrapper) ExecuteTransfer(opts *bind.TransactOpts, mvxTransactions []contract.MvxTransaction, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientTransactions, 1)
	return wrapper.multiSigContract.ExecuteTransfer(opts, mvxTransactions, batchNonce, signatures)
}

// Quorum returns the current set quorum value
func (wrapper *ethereumChainWrapper) Quorum(ctx context.Context) (*big.Int, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.Quorum(&bind.CallOpts{Context: ctx})
}

// GetStatusesAfterExecution returns the statuses of the last executed transfer
func (wrapper *ethereumChainWrapper) GetStatusesAfterExecution(ctx context.Context, batchID *big.Int) ([]byte, bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.multiSigContract.GetStatusesAfterExecution(&bind.CallOpts{Context: ctx}, batchID)
}

// BalanceAt returns the wei balance of the given account.
// The block number can be nil, in which case the balance is taken from the latest known block.
func (wrapper *ethereumChainWrapper) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.blockchainClient.BalanceAt(ctx, account, blockNumber)
}

// TotalBalances returns the total balance of the given token
func (wrapper *ethereumChainWrapper) TotalBalances(ctx context.Context, token common.Address) (*big.Int, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.safeContract.TotalBalances(&bind.CallOpts{Context: ctx}, token)
}

// MintBalances returns the mint balance of the given token
func (wrapper *ethereumChainWrapper) MintBalances(ctx context.Context, token common.Address) (*big.Int, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.safeContract.MintBalances(&bind.CallOpts{Context: ctx}, token)
}

// BurnBalances returns the burn balance of the given token
func (wrapper *ethereumChainWrapper) BurnBalances(ctx context.Context, token common.Address) (*big.Int, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.safeContract.BurnBalances(&bind.CallOpts{Context: ctx}, token)
}

// MintBurnTokens returns true if the token is a mintBurn token
func (wrapper *ethereumChainWrapper) MintBurnTokens(ctx context.Context, token common.Address) (bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.safeContract.MintBurnTokens(&bind.CallOpts{Context: ctx}, token)
}

// NativeTokens returns true if the token is a native token
func (wrapper *ethereumChainWrapper) NativeTokens(ctx context.Context, token common.Address) (bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.safeContract.NativeTokens(&bind.CallOpts{Context: ctx}, token)
}

// WhitelistedTokens returns true if the token is a native token
func (wrapper *ethereumChainWrapper) WhitelistedTokens(ctx context.Context, token common.Address) (bool, error) {
	wrapper.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.safeContract.WhitelistedTokens(&bind.CallOpts{Context: ctx}, token)
}

// IsPaused returns true if the multisig contract is paused
func (wrapper *ethereumChainWrapper) IsPaused(ctx context.Context) (bool, error) {
	return wrapper.multiSigContract.Paused(&bind.CallOpts{Context: ctx})
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrapper *ethereumChainWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}
