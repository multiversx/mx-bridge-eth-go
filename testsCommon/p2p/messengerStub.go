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
	SetPeerDenialEvaluatorCalled   func(handler p2p.PeerDenialEvaluator) error
	ConnectedAddressesCalled       func() []string
	CloseCalled                    func() error
}

// SetPeerDenialEvaluator -
func (stub *MessengerStub) SetPeerDenialEvaluator(handler p2p.PeerDenialEvaluator) error {
	if stub.SetPeerDenialEvaluatorCalled != nil {
		return stub.SetPeerDenialEvaluatorCalled(handler)
	}

	return nil
}

// ID -
func (stub *MessengerStub) ID() core.PeerID {
	if stub.IDCalled != nil {
		return stub.IDCalled()
	}

	return ""
}

// Bootstrap -
func (stub *MessengerStub) Bootstrap() error {
	if stub.BootstrapCalled != nil {
		return stub.BootstrapCalled()
	}

	return nil
}

// Addresses -
func (stub *MessengerStub) Addresses() []string {
	if stub.AddressesCalled != nil {
		return stub.AddressesCalled()
	}

	return make([]string, 0)
}

// RegisterMessageProcessor -
func (stub *MessengerStub) RegisterMessageProcessor(topic string, identifier string, processor p2p.MessageProcessor) error {
	if stub.RegisterMessageProcessorCalled != nil {
		return stub.RegisterMessageProcessorCalled(topic, identifier, processor)
	}

	return nil
}

// HasTopic -
func (stub *MessengerStub) HasTopic(name string) bool {
	if stub.HasTopicCalled != nil {
		return stub.HasTopicCalled(name)
	}

	return false
}

// CreateTopic -
func (stub *MessengerStub) CreateTopic(name string, createChannelForTopic bool) error {
	if stub.CreateTopicCalled != nil {
		return stub.CreateTopicCalled(name, createChannelForTopic)
	}

	return nil
}

// Broadcast -
func (stub *MessengerStub) Broadcast(topic string, buff []byte) {
	if stub.BroadcastCalled != nil {
		stub.BroadcastCalled(topic, buff)
	}
}

// SendToConnectedPeer -
func (stub *MessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	if stub.SendToConnectedPeerCalled != nil {
		return stub.SendToConnectedPeerCalled(topic, buff, peerID)
	}

	return nil
}

// ConnectedAddresses -
func (stub *MessengerStub) ConnectedAddresses() []string {
	if stub.ConnectedAddressesCalled != nil {
		return stub.ConnectedAddressesCalled()
	}

	return make([]string, 0)
}

// Close -
func (stub *MessengerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *MessengerStub) IsInterfaceNil() bool {
	return stub == nil
}
