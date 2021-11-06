package p2p

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessengerMock -
type MessengerMock struct {
	mut                         sync.RWMutex
	peerID                      core.PeerID
	registeredMessageProcessors map[string]p2p.MessageProcessor
	createdTopics               map[string]struct{}
	bootstrapWasCalled          bool
	lastSendTopicName           string
	lastSendData                []byte
	lastSendPeerID              core.PeerID
}

// ID -
func (mock *MessengerMock) ID() core.PeerID {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	return mock.peerID
}

// SetID -
func (mock *MessengerMock) SetID(id core.PeerID) {
	mock.mut.Lock()
	mock.peerID = id
	mock.mut.Unlock()
}

// Bootstrap -
func (mock *MessengerMock) Bootstrap() error {
	mock.mut.Lock()
	mock.bootstrapWasCalled = true
	mock.mut.Unlock()

	return nil
}

// GetBootstrapWasCalled -
func (mock *MessengerMock) GetBootstrapWasCalled() bool {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	return mock.bootstrapWasCalled
}

// RegisterMessageProcessor -
func (mock *MessengerMock) RegisterMessageProcessor(topic string, _ string, handler p2p.MessageProcessor) error {
	mock.mut.Lock()
	defer mock.mut.Unlock()

	if mock.registeredMessageProcessors == nil {
		mock.registeredMessageProcessors = make(map[string]p2p.MessageProcessor)
	}

	mock.registeredMessageProcessors[topic] = handler
	return nil
}

// GetRegisterMessageProcessors -
func (mock *MessengerMock) GetRegisterMessageProcessors() map[string]p2p.MessageProcessor {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	processors := make(map[string]p2p.MessageProcessor)
	for topic, proc := range mock.registeredMessageProcessors {
		processors[topic] = proc
	}

	return processors
}

// HasTopic -
func (mock *MessengerMock) HasTopic(name string) bool {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	_, found := mock.createdTopics[name]

	return found
}

// CreateTopic -
func (mock *MessengerMock) CreateTopic(name string, _ bool) error {
	mock.mut.Lock()
	defer mock.mut.Unlock()

	mock.createdTopics[name] = struct{}{}

	return nil
}

// Addresses -
func (mock *MessengerMock) Addresses() []string {
	return nil
}

// Broadcast -
func (mock *MessengerMock) Broadcast(topic string, data []byte) {
	mock.mut.Lock()
	defer mock.mut.Unlock()

	mock.lastSendTopicName = topic
	mock.lastSendData = data
}

// SendToConnectedPeer -
func (mock *MessengerMock) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	mock.mut.Lock()
	defer mock.mut.Unlock()

	mock.lastSendTopicName = topic
	mock.lastSendData = buff
	mock.lastSendPeerID = peerID

	return nil
}

// LastSendTopicName -
func (mock *MessengerMock) LastSendTopicName() string {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	return mock.lastSendTopicName
}

// LastSendData -
func (mock *MessengerMock) LastSendData() []byte {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	return mock.lastSendData
}

// LastSendPeerID -
func (mock *MessengerMock) LastSendPeerID() core.PeerID {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	return mock.lastSendPeerID
}

// ConnectedAddresses -
func (mock *MessengerMock) ConnectedAddresses() []string {
	return make([]string, 0)
}

// Close -
func (mock *MessengerMock) Close() error {
	return nil
}

// IsInterfaceNil -
func (mock *MessengerMock) IsInterfaceNil() bool {
	return mock == nil
}
