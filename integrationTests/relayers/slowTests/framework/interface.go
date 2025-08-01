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
	PeerChainRelayerAddress() common.Address
	Start() error
	Close() error
}

// ChainSimulatorWrapper defines the wrapper over the chain simulator
type ChainSimulatorWrapper interface {
	Proxy() multiversx.Proxy
	GetNetworkAddress() string
	DeploySC(ctx context.Context, path string, ownerSK []byte, gasLimit uint64, extraParams []string) (*MvxAddress, string, *data.TransactionOnNetwork)
	ScCall(ctx context.Context, senderSK []byte, contract *MvxAddress, value string, gasLimit uint64, function string, parameters []string) (string, *data.TransactionOnNetwork)
	ScCallWithoutGenerateBlocks(ctx context.Context, senderSK []byte, contract *MvxAddress, value string, gasLimit uint64, function string, parameters []string) string
	SendTx(ctx context.Context, senderSK []byte, receiver *MvxAddress, value string, gasLimit uint64, dataField []byte) (string, *data.TransactionOnNetwork)
	SendTxWithoutGenerateBlocks(ctx context.Context, senderSK []byte, receiver *MvxAddress, value string, gasLimit uint64, dataField []byte) string
	FundWallets(ctx context.Context, wallets []string)
	GenerateBlocksUntilEpochReached(ctx context.Context, epoch uint32)
	GenerateBlocks(ctx context.Context, numBlocks int)
	GetESDTBalance(ctx context.Context, address *MvxAddress, token string) string
	GetBlockchainTimeStamp(ctx context.Context) uint64
	GetTransactionResult(ctx context.Context, hash string) *data.TransactionOnNetwork
	ExecuteVMQuery(ctx context.Context, scAddress *MvxAddress, function string, hexParams []string) [][]byte
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

// MoveContract defines the operations of a Move contract
type MoveContract interface {
	BalanceOf(ctx context.Context, account []byte) (*big.Int, error)
	Mint(ctx context.Context, recipientAddress []byte, amount *big.Int) (string, error)
}

// TokensRegistry defines the registry used for the tokens in tests
type TokensRegistry interface {
	AddToken(params IssueTokenParams)
	RegisterPeerChainAddressAndContract(
		abstractTokenIdentifier string,
		peerChainAddress interface{}, // Can be common.Address for Ethereum or []byte for Sui
		peerChainContract interface{}, // Can be ERC20Contract for Ethereum or MoveContract for Sui
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

// BlockchainHandler defines the common interface for blockchain handlers
type BlockchainHandler interface {
	DeployContracts(ctx context.Context)

	GetBalance(receiver interface{}, abstractTokenIdentifier string) *big.Int
	IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams)
	Mint(ctx context.Context, params TestTokenParams, valueToMint *big.Int)

	CreateBatchOnPeerChain(ctx context.Context, params CreateBatchParams)
	SendFromPeerChainToMultiversX(ctx context.Context, params TestTransferParams)

	PauseContractsForTokenChanges(ctx context.Context)
	UnPauseContractsAfterTokenChanges(ctx context.Context)

	Close() error

	GetChainType() string
}

type SuiBlockchainClient interface {
	GetBalance(ctx context.Context, address []byte, coinType string) (*big.Int, error)
	ExecuteTransaction(ctx context.Context, txBytes []byte) (string, error)
	GetTransactionBlock(ctx context.Context, digest string) (interface{}, error)
}
