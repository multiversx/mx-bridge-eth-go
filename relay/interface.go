package relay

import (
	"context"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ethereum/go-ethereum/common"
)

// Startable defines an entity that is able to Start or Stop
type Startable interface {
	Start(context.Context) error
	Stop() error
}

// TopologyProvider defines the topology provider functions
type TopologyProvider interface {
	AmITheLeader() bool
	Clean()
}

// Broadcaster defines a component able to communicate with other such instances and manage signatures and other state related data
type Broadcaster interface {
	BroadcastSignature(signature []byte)
	BroadcastJoinTopic()
	ClearSignatures()
	Signatures() [][]byte
	SortedPublicKeys() [][]byte
	RegisterOnTopics() error
	Close() error
	IsInterfaceNil() bool
}

// ElrondRoleProvider defines the operations for the Elrond role provider
type ElrondRoleProvider interface {
	Execute(ctx context.Context) error
	IsWhitelisted(address core.AddressHandler) bool
	IsInterfaceNil() bool
}

// ElrondChainInteractor defines an Elrond client able to respond to VM queries
type ElrondChainInteractor interface {
	ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error)
	IsInterfaceNil() bool
}

// EthereumChainInteractor defines an Ethereum client able to respond to requests
type EthereumChainInteractor interface {
	GetRelayers(ctx context.Context) ([]common.Address, error)
	IsInterfaceNil() bool
}
