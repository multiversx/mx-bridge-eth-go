package sui

import (
	"context"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
)

// Proxy defines the operations for a component that can act as a proxy to interact with the Sui blockchain
type Proxy interface {
	SuiGetLatestCheckpointSequenceNumber(ctx context.Context) (uint64, error)
	SuiXGetBalance(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error)
	SuiDevInspectTransactionBlock(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
	SuiGetObject(ctx context.Context, req models.SuiGetObjectRequest) (models.SuiObjectResponse, error)
	MoveCall(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error)
	SignAndExecuteTransactionBlock(ctx context.Context, req models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error)
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
	Sign(message []byte) (*signer.SignedMessageSerializedSig, error)
}
