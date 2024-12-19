package multiversx

import (
	"context"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// Proxy defines the behavior of a proxy able to serve MultiversX blockchain requests
type Proxy interface {
	GetNetworkConfig(ctx context.Context) (*data.NetworkConfig, error)
	SendTransaction(ctx context.Context, tx *transaction.FrontendTransaction) (string, error)
	SendTransactions(ctx context.Context, txs []*transaction.FrontendTransaction) ([]string, error)
	ExecuteVMQuery(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
	GetAccount(ctx context.Context, address core.AddressHandler) (*data.Account, error)
	GetNetworkStatus(ctx context.Context, shardID uint32) (*data.NetworkStatus, error)
	GetShardOfAddress(ctx context.Context, bech32Address string) (uint32, error)
	GetESDTTokenData(ctx context.Context, address core.AddressHandler, tokenIdentifier string, queryOptions api.AccountQueryOptions) (*data.ESDTFungibleTokenData, error)
	GetTransactionInfoWithResults(ctx context.Context, hash string) (*data.TransactionInfo, error)
	ProcessTransactionStatus(ctx context.Context, hexTxHash string) (transaction.TxStatus, error)
	IsInterfaceNil() bool
}

// NonceTransactionsHandler represents the interface able to handle the current nonce and the transactions resend mechanism
type NonceTransactionsHandler interface {
	ApplyNonceAndGasPrice(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error
	SendTransaction(ctx context.Context, tx *transaction.FrontendTransaction) (string, error)
	Close() error
	IsInterfaceNil() bool
}

// ScCallsExecuteFilter defines the operations supported by a filter that allows selective executions of batches
type ScCallsExecuteFilter interface {
	ShouldExecute(callData bridgeCore.ProxySCCompleteCallData) bool
	IsInterfaceNil() bool
}

// Codec defines the operations implemented by a MultiversX codec
type Codec interface {
	DecodeProxySCCompleteCallData(buff []byte) (bridgeCore.ProxySCCompleteCallData, error)
	ExtractGasLimitFromRawCallData(buff []byte) (uint64, error)
	EncodeCallDataStrict(callData bridgeCore.CallData) []byte
	DecodeCallData(buff []byte) (bridgeCore.CallData, error)
	IsInterfaceNil() bool
}
