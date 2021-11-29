package elrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// ElrondProxy defines the behavior of a proxy able to serve Elrond blockchain requests
type ElrondProxy interface {
	GetNetworkConfig(ctx context.Context) (*data.NetworkConfig, error)
	SendTransaction(ctx context.Context, tx *data.Transaction) (string, error)
	SendTransactions(ctx context.Context, txs []*data.Transaction) ([]string, error)
	ExecuteVMQuery(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
	GetAccount(ctx context.Context, address core.AddressHandler) (*data.Account, error)
	IsInterfaceNil() bool
}

// NonceTransactionsHandler represents the interface able to handle the current nonce and the transactions resend mechanism
type NonceTransactionsHandler interface {
	GetNonce(ctx context.Context, address core.AddressHandler) (uint64, error)
	SendTransaction(ctx context.Context, tx *data.Transaction) (string, error)
	Close() error
}

// elrondClientDataGetter defines the interface able to handle get requests for Elrond blockchain
type DataGetter interface {
	ExecuteQueryReturningBytes(ctx context.Context, request *data.VmValueRequest) ([][]byte, error)
	ExecuteQueryReturningBool(ctx context.Context, request *data.VmValueRequest) (bool, error)
	ExecuteQueryReturningUint64(ctx context.Context, request *data.VmValueRequest) (uint64, error)
	GetCurrentBatchAsDataBytes(ctx context.Context) ([][]byte, error)
	GetTokenIdForErc20Address(ctx context.Context, erc20Address []byte) ([][]byte, error)
	GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error)
	IsInterfaceNil() bool
}
