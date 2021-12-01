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

// TokensMapper can convert a token bytes from one chain to another
type TokensMapper interface {
	ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error)
	IsInterfaceNil() bool
}
