package p2p

import (
	"time"

	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

// NetMessenger is the definition of an entity able to receive and send messages
type NetMessenger interface {
	ID() chainCore.PeerID
	Bootstrap() error
	Addresses() []string
	RegisterMessageProcessor(topic string, identifier string, processor p2p.MessageProcessor) error
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	Broadcast(topic string, buff []byte)
	SendToConnectedPeer(topic string, buff []byte, peerID chainCore.PeerID) error
	SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error
	ConnectedAddresses() []string
	Close() error
	IsInterfaceNil() bool
}

// MultiversXRoleProvider defines the operations for an MultiversX role provider
type MultiversXRoleProvider interface {
	IsWhitelisted(address sdkCore.AddressHandler) bool
	IsInterfaceNil() bool
}

// SignatureProcessor defines the operations needed to process signatures
type SignatureProcessor interface {
	VerifyEthSignature(signature []byte, messageHash []byte) error
	IsInterfaceNil() bool
}

// PeerDenialEvaluator defines the behavior of a component that is able to decide if a peer ID is black listed or not
type PeerDenialEvaluator interface {
	IsDenied(pid chainCore.PeerID) bool
	UpsertPeerID(pid chainCore.PeerID, duration time.Duration) error
	IsInterfaceNil() bool
}
