package p2p

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerStub -
type MessengerStub struct {
	IDCalled                       func() core.PeerID
	BootstrapCalled                func() error
	AddressesCalled                func() []string
	RegisterMessageProcessorCalled func(topic string, identifier string, processor p2p.MessageProcessor) error
	HasTopicCalled                 func(name string) bool
	CreateTopicCalled              func(name string, createChannelForTopic bool) error
	BroadcastCalled                func(topic string, buff []byte)
	SendToConnectedPeerCalled      func(topic string, buff []byte, peerID core.PeerID) error
	CloseCalled                    func() error
}

// ID -
func (ms *MessengerStub) ID() core.PeerID {
	if ms.IDCalled != nil {
		return ms.IDCalled()
	}

	return ""
}

// Bootstrap -
func (ms *MessengerStub) Bootstrap() error {
	if ms.BootstrapCalled != nil {
		return ms.BootstrapCalled()
	}

	return nil
}

// Addresses -
func (ms *MessengerStub) Addresses() []string {
	if ms.AddressesCalled != nil {
		return ms.AddressesCalled()
	}

	return make([]string, 0)
}

// RegisterMessageProcessor -
func (ms *MessengerStub) RegisterMessageProcessor(topic string, identifier string, processor p2p.MessageProcessor) error {
	if ms.RegisterMessageProcessorCalled != nil {
		return ms.RegisterMessageProcessorCalled(topic, identifier, processor)
	}

	return nil
}

// HasTopic -
func (ms *MessengerStub) HasTopic(name string) bool {
	if ms.HasTopicCalled != nil {
		return ms.HasTopicCalled(name)
	}

	return false
}

// CreateTopic -
func (ms *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	if ms.CreateTopicCalled != nil {
		return ms.CreateTopicCalled(name, createChannelForTopic)
	}

	return nil
}

// Broadcast -
func (ms *MessengerStub) Broadcast(topic string, buff []byte) {
	if ms.BroadcastCalled != nil {
		ms.BroadcastCalled(topic, buff)
	}
}

// SendToConnectedPeer -
func (ms *MessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	if ms.SendToConnectedPeerCalled != nil {
		return ms.SendToConnectedPeerCalled(topic, buff, peerID)
	}

	return nil
}

// Close -
func (ms *MessengerStub) Close() error {
	if ms.CloseCalled != nil {
		return ms.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (ms *MessengerStub) IsInterfaceNil() bool {
	return ms == nil
}
