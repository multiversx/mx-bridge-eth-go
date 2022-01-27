package factory

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
)

type dataGetter interface {
	GetTokenIdForErc20Address(ctx context.Context, erc20Address []byte) ([][]byte, error)
	GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error)
	GetAllStakedRelayers(ctx context.Context) ([][]byte, error)
	IsInterfaceNil() bool
}

// ElrondRoleProvider defines the operations for the Elrond role provider
type ElrondRoleProvider interface {
	Execute(ctx context.Context) error
	IsWhitelisted(address erdgoCore.AddressHandler) bool
	SortedPublicKeys() [][]byte
	IsInterfaceNil() bool
}

// EthereumRoleProvider defines the operations for the Ethereum role provider
type EthereumRoleProvider interface {
	Execute(ctx context.Context) error
	VerifyEthSignature(signature []byte, messageHash []byte) error
	IsInterfaceNil() bool
}

// Broadcaster defines a component able to communicate with other such instances and manage signatures and other state related data
type Broadcaster interface {
	BroadcastSignature(signature []byte, messageHash []byte)
	BroadcastJoinTopic()
	SortedPublicKeys() [][]byte
	RegisterOnTopics() error
	AddBroadcastClient(client core.BroadcastClient) error
	Close() error
	IsInterfaceNil() bool
}

// StateMachine defines a state machine component
type StateMachine interface {
	Execute(ctx context.Context) error
	IsInterfaceNil() bool
}

// PollingHandler defines a polling handler component
type PollingHandler interface {
	StartProcessingLoop() error
	IsInterfaceNil() bool
}
