package bridge

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// Broadcaster defines the operations for a component used for communication with other peers
type Broadcaster interface {
	BroadcastSignature(signature []byte, messageHash []byte)
	IsInterfaceNil() bool
}

// Mapper defines the mapping operations
type Mapper interface {
	GetTokenId(string) string
	GetErc20Address(string) string
	IsInterfaceNil() bool
}

// QuorumProvider defines the operations for a quorum provider
type QuorumProvider interface {
	GetQuorum(ctx context.Context) (uint, error)
	IsInterfaceNil() bool
}

// Bridge defines the operations available for a validator operating on a bridge between 2 chains
type Bridge interface {
	GetPending(context.Context) *Batch
	ProposeSetStatus(context.Context, *Batch)
	ProposeTransfer(context.Context, *Batch) (string, error)
	WasProposedTransfer(context.Context, *Batch) bool
	GetActionIdForProposeTransfer(context.Context, *Batch) ActionID
	WasProposedSetStatus(context.Context, *Batch) bool
	GetActionIdForSetStatusOnPendingTransfer(context.Context, *Batch) ActionID
	WasExecuted(context.Context, ActionID, BatchID) bool
	Sign(context.Context, ActionID, *Batch) (string, error)
	Execute(context.Context, ActionID, *Batch, SignaturesHolder) (string, error)
	SignersCount(*Batch, ActionID, SignaturesHolder) uint
	GetTransactionsStatuses(ctx context.Context, batchID BatchID) ([]uint8, error)
	IsInterfaceNil() bool
}

// ElrondProxy defines the behavior of a proxy able to serve Elrond blockchain requests
type ElrondProxy interface {
	GetNetworkConfig() (*data.NetworkConfig, error)
	SendTransaction(*data.Transaction) (string, error)
	SendTransactions(txs []*data.Transaction) ([]string, error)
	ExecuteVMQuery(vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
	GetAccount(address core.AddressHandler) (*data.Account, error)
	IsInterfaceNil() bool
}

// GasHandler defines the component able to fetch the current gas price
type GasHandler interface {
	GetCurrentGasPrice() (*big.Int, error)
	IsInterfaceNil() bool
}

// SignaturesHolder defines the operations for a component that can hold and manage signatures
type SignaturesHolder interface {
	Signatures(messageHash []byte) [][]byte
	IsInterfaceNil() bool
}
