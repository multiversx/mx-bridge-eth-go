package p2p

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerMock -
type MessengerMock struct {
	PeerID                      core.PeerID
	RegisteredMessageProcessors map[string]p2p.MessageProcessor
	CreatedTopics               []string
	BootstrapWasCalled          bool
	LastSendTopicName           string
	LastSendData                []byte
	LastSendPeerID              core.PeerID
}

// ID -
func (mock *MessengerMock) ID() core.PeerID {
	return mock.PeerID
}

// Bootstrap -
func (mock *MessengerMock) Bootstrap() error {
	mock.BootstrapWasCalled = true
	return nil
}

// RegisterMessageProcessor -
func (mock *MessengerMock) RegisterMessageProcessor(topic string, _ string, handler p2p.MessageProcessor) error {
	if mock.RegisteredMessageProcessors == nil {
		mock.RegisteredMessageProcessors = make(map[string]p2p.MessageProcessor)
	}

	mock.RegisteredMessageProcessors[topic] = handler
	return nil
}

// HasTopic -
func (mock *MessengerMock) HasTopic(name string) bool {
	for _, topic := range mock.CreatedTopics {
		if topic == name {
			return true
		}
	}
	return false
}

// CreateTopic -
func (mock *MessengerMock) CreateTopic(name string, _ bool) error {
	mock.CreatedTopics = append(mock.CreatedTopics, name)
	return nil
}

// Addresses -
func (mock *MessengerMock) Addresses() []string {
	return nil
}

// Broadcast -
func (mock *MessengerMock) Broadcast(topic string, data []byte) {
	mock.LastSendTopicName = topic
	mock.LastSendData = data
}

// SendToConnectedPeer -
func (mock *MessengerMock) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	mock.LastSendTopicName = topic
	mock.LastSendData = buff
	mock.LastSendPeerID = peerID

	return nil
}

// Close -
func (mock *MessengerMock) Close() error {
	return nil
}

// IsInterfaceNil -
func (mock *MessengerMock) IsInterfaceNil() bool {
	return mock == nil
}
