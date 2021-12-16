package p2p

import (
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
)

// NetMessenger is the definition of an entity able to receive and send messages
type NetMessenger interface {
	ID() elrondCore.PeerID
	Bootstrap() error
	Addresses() []string
	RegisterMessageProcessor(topic string, identifier string, processor p2p.MessageProcessor) error
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	Broadcast(topic string, buff []byte)
	SendToConnectedPeer(topic string, buff []byte, peerID elrondCore.PeerID) error
	ConnectedAddresses() []string
	Close() error
	IsInterfaceNil() bool
}

// ElrondRoleProvider defines the operations for an Elrond role provider
type ElrondRoleProvider interface {
	IsWhitelisted(address erdgoCore.AddressHandler) bool
	IsInterfaceNil() bool
}

// SignatureProcessor defines the operations needed to process signatures
type SignatureProcessor interface {
	VerifyEthSignature(signature []byte, messageHash []byte) error
	IsInterfaceNil() bool
}

// SignaturesHolder defines the operations for a component that can hold and manage signatures
type SignaturesHolder interface {
	Signatures(messageHash []byte) [][]byte
	IsInterfaceNil() bool
}
