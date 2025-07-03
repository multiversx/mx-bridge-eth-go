package sui

import (
	"context"
	"github.com/block-vision/sui-go-sdk/models"
)

// SuiClient defines the operations for a component that can interact with the Sui blockchain
type SuiClient interface {
	SuiGetLatestCheckpointSequenceNumber(ctx context.Context) (uint64, error)
	SuiGetChainIdentifier(ctx context.Context) (string, error)
	SuiXGetBalance(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error)
	MoveCall(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error)
	SignAndExecuteTransactionBlock(ctx context.Context, req models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	SuiDryRunTransactionBlock(ctx context.Context, req models.SuiDryRunTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	SuiXGetCoins(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error)
	IsInterfaceNil() bool
}

// TokensMapper can convert a token bytes from one chain to another
type TokensMapper interface {
	ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error)
	IsInterfaceNil() bool
}

// SignaturesHolder defines the operations for a component that can hold and manage signatures
type SignaturesHolder interface {
	Signatures(messageHash []byte) [][]byte
	ClearStoredSignatures()
	IsInterfaceNil() bool
}

// Broadcaster defines the operations for a component used for communication with other peers
type Broadcaster interface {
	BroadcastSignature(signature []byte, messageHash []byte)
	IsInterfaceNil() bool
}

type txHandler interface {
	SendTransactionReturnHash(ctx context.Context, moveCallRequest models.MoveCallRequest) (string, error)
	Sign(message []byte) ([]byte, error)
}
