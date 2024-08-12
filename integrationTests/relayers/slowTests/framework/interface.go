package framework

import (
	"context"
	"math/big"

	goEthereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

type httpClientWrapper interface {
	GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error)
	PostHTTP(ctx context.Context, endpoint string, data []byte) ([]byte, int, error)
	IsInterfaceNil() bool
}

// Relayer defines the behavior a bridge relayer must implement
type Relayer interface {
	MultiversXRelayerAddress() sdkCore.AddressHandler
	EthereumRelayerAddress() common.Address
	Start() error
	Close() error
}

// ChainSimulatorWrapper defines the wrapper over the chain simulator
type ChainSimulatorWrapper interface {
	Proxy() multiversx.Proxy
	GetNetworkAddress() string
	DeploySC(ctx context.Context, path string, ownerSK []byte, gasLimit uint64, extraParams []string) (*MvxAddress, string, *data.TransactionOnNetwork)
	ScCall(ctx context.Context, senderSK []byte, contract *MvxAddress, value string, gasLimit uint64, function string, parameters []string) (string, *data.TransactionOnNetwork)
	SendTx(ctx context.Context, senderSK []byte, receiver *MvxAddress, value string, gasLimit uint64, dataField []byte) (string, *data.TransactionOnNetwork)
	FundWallets(ctx context.Context, wallets []string)
	GenerateBlocksUntilEpochReached(ctx context.Context, epoch uint32)
	GenerateBlocks(ctx context.Context, numBlocks int)
	GetESDTBalance(ctx context.Context, address *MvxAddress, token string) string
	GetBlockchainTimeStamp(ctx context.Context) uint64
}

// EthereumBlockchainClient defines the operations supported by the Ethereum client
type EthereumBlockchainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	FilterLogs(ctx context.Context, q goEthereum.FilterQuery) ([]types.Log, error)
}

// ERC20Contract defines the operations of an ERC20 contract
type ERC20Contract interface {
	BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error)
	Mint(opts *bind.TransactOpts, recipientAddress common.Address, amount *big.Int) (*types.Transaction, error)
	Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error)
}

// TokensRegistry defines the registry used for the tokens in tests
type TokensRegistry interface {
	AddToken(params IssueTokenParams)
	RegisterEthAddressAndContract(
		abstractTokenIdentifier string,
		ethErc20Address common.Address,
		ethErc20Contract ERC20Contract,
	)
	GetTokenData(abstractTokenIdentifier string) *TokenData
	RegisterUniversalToken(abstractTokenIdentifier string, mvxUniversalToken string)
	RegisterChainSpecificToken(abstractTokenIdentifier string, mvxChainSpecificToken string)
}

// SCCallerModule defines the operation for the module able to execute smart contract calls
type SCCallerModule interface {
	GetNumSentTransaction() uint32
	Close() error
}
