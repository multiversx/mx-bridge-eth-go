package p2p

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// NetMessenger is the definition of an entity able to receive and send messages
type NetMessenger interface {
	ID() core.PeerID
	Bootstrap() error
	Addresses() []string
	RegisterMessageProcessor(topic string, identifier string, processor p2p.MessageProcessor) error
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	Broadcast(topic string, buff []byte)
	SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error
	Close() error
	IsInterfaceNil() bool
}

// RoleProvider defines the operations for a role provider
type RoleProvider interface {
	IsWhitelisted(string) bool
	IsInterfaceNil() bool
}
