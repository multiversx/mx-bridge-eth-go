package relay

import (
	"context"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ethereum/go-ethereum/common"
)

// TopologyProvider defines the topology provider functions
type TopologyProvider interface {
	AmITheLeader() bool
	Clean()
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

// ElrondRoleProvider defines the operations for the Elrond role provider
type ElrondRoleProvider interface {
	Execute(ctx context.Context) error
	IsWhitelisted(address erdgoCore.AddressHandler) bool
	IsInterfaceNil() bool
}

// EthereumRoleProvider defines the operations for the Ethereum role provider
type EthereumRoleProvider interface {
	Execute(ctx context.Context) error
	VerifyEthSignature(signature []byte, messageHash []byte) error
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
