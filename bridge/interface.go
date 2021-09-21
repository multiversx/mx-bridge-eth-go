package bridge

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// Broadcaster defines the operations for a component used for communication with other peers
type Broadcaster interface {
	Signatures() [][]byte
	SendSignature(signature []byte)
}

// Mapper defines the mapping operations
type Mapper interface {
	GetTokenId(string) string
	GetErc20Address(string) string
}

// RoleProvider defines the operations for a role provider
type RoleProvider interface {
	IsWhitelisted(string) bool
}

// WalletAddressProvider defines the operations for a wallet address provider
type WalletAddressProvider interface {
	GetHexWalletAddress() string
}

// Bridge defines the operations available for a validator operating on a bridge between 2 chains
type Bridge interface {
	GetPending(context.Context) *Batch
	ProposeSetStatus(context.Context, *Batch)
	ProposeTransfer(context.Context, *Batch) (string, error)
	WasProposedTransfer(context.Context, *Batch) bool
	GetActionIdForProposeTransfer(context.Context, *Batch) ActionId
	WasProposedSetStatus(context.Context, *Batch) bool
	GetActionIdForSetStatusOnPendingTransfer(context.Context, *Batch) ActionId
	WasExecuted(context.Context, ActionId, BatchId) bool
	Sign(context.Context, ActionId) (string, error)
	Execute(context.Context, ActionId, *Batch) (string, error)
	SignersCount(context.Context, ActionId) uint
}

// ElrondProxy defines the behavior of a proxy able to serve Elrond blockchain requests
type ElrondProxy interface {
	GetNetworkConfig() (*data.NetworkConfig, error)
	SendTransaction(*data.Transaction) (string, error)
	GetTransactionInfoWithResults(hash string) (*data.TransactionInfo, error)
	ExecuteVMQuery(vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
	GetAccount(address core.AddressHandler) (*data.Account, error)
	IsInterfaceNil() bool
}
