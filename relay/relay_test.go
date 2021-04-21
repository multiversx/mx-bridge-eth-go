package relay

import (
	"bytes"
	"context"
	"encoding/gob"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p/mock"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

// implements interface
var (
	_ = Startable(&Relay{})
)

var log = logger.GetOrCreate("main")

func TestInit(t *testing.T) {
	setLoggerLevel()

	messenger := &netMessengerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &timerStub{},
		log:       log,

		elrondBridge: &bridgeStub{},
		ethBridge:    &bridgeStub{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	assert.True(t, messenger.bootstrapWasCalled)
	assert.Contains(t, messenger.createdTopics, PrivateTopicName)
	assert.Contains(t, messenger.createdTopics, ActionsTopicName)
	assert.Contains(t, messenger.registeredMessageProcessors, PrivateTopicName)
	assert.Contains(t, messenger.registeredMessageProcessors, ActionsTopicName)
}

func TestPrivateTopicProcessor(t *testing.T) {
	setLoggerLevel()

	messenger := &netMessengerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &timerStub{},
		log:       log,

		elrondBridge: &bridgeStub{},
		ethBridge:    &bridgeStub{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	privateMessageProcessor := messenger.registeredMessageProcessors[PrivateTopicName]
	expected := Peers{"first", "second"}
	message := buildPrivateMessage("other", expected)
	_ = privateMessageProcessor.ProcessReceivedMessage(message, "peer_near_me")

	assert.Equal(t, expected, relay.peers)
}

func TestActionsTopicProcessor(t *testing.T) {
	setLoggerLevel()

	t.Run("on joined action when there are more peers then self will broadcast to private", func(t *testing.T) {
		messenger := &netMessengerStub{}
		relay := Relay{
			messenger: messenger,
			timer:     &timerStub{},
			log:       log,

			elrondBridge: &bridgeStub{},
			ethBridge:    &bridgeStub{},

			peers: Peers{"first", "second"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		actionsMessageProcessor := messenger.registeredMessageProcessors[ActionsTopicName]
		_ = actionsMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("other"), "peer_near_me")

		dec := gob.NewDecoder(bytes.NewReader(messenger.lastSendData))
		var got Peers
		if err := dec.Decode(&got); err != nil {
			t.Fatal(err)
		}

		expected := Peers{"first", "other", "second"}

		assert.Equal(t, expected, got)
	})
	t.Run("when self joined will not broadcast to private", func(t *testing.T) {
		messenger := &netMessengerStub{peerID: "self"}
		relay := Relay{
			messenger: messenger,
			timer:     &timerStub{},
			log:       log,

			elrondBridge: &bridgeStub{},
			ethBridge:    &bridgeStub{},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		actionsMessageProcessor := messenger.registeredMessageProcessors[ActionsTopicName]
		_ = actionsMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("self"), "peer_near_me")

		assert.Nil(t, messenger.lastSendData)
	})
}

func TestJoin(t *testing.T) {
	setLoggerLevel()

	messenger := &netMessengerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &timerStub{},
		log:       log,

		elrondBridge: &bridgeStub{},
		ethBridge:    &bridgeStub{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	assert.True(t, messenger.joinedWasCalled)
}

func TestAmILeader(t *testing.T) {
	setLoggerLevel()

	t.Run("will return true when time matches current index", func(t *testing.T) {
		relay := Relay{
			peers:     Peers{"self"},
			messenger: &netMessengerStub{peerID: "self"},
			timer:     &timerStub{timeNowUnix: 0},
		}

		assert.True(t, relay.AmITheLeader())
	})
	t.Run("will return false when time does not match", func(t *testing.T) {
		relay := Relay{
			peers:     Peers{"self", "other"},
			messenger: &netMessengerStub{peerID: "self"},
			timer:     &timerStub{timeNowUnix: int64(Timeout.Seconds()) + 1},
		}

		assert.False(t, relay.AmITheLeader())
	})
}

func buildPrivateMessage(peerID core.PeerID, peers Peers) p2p.MessageP2P {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(peers)
	if err != nil {
		panic(err)
	}

	return &mock.P2PMessageMock{
		TopicField: PrivateTopicName,
		PeerField:  peerID,
		DataField:  data.Bytes(),
	}
}

func buildJoinedMessage(peerID core.PeerID) p2p.MessageP2P {
	return &mock.P2PMessageMock{
		TopicField: ActionsTopicName,
		PeerField:  peerID,
		DataField:  []byte(JoinedAction),
	}
}

type netMessengerStub struct {
	peerID                      core.PeerID
	registeredMessageProcessors map[string]p2p.MessageProcessor
	createdTopics               []string

	joinedWasCalled    bool
	bootstrapWasCalled bool

	lastSendTopicName string
	lastSendData      []byte
	lastSendPeerID    core.PeerID
}

func (p *netMessengerStub) ID() core.PeerID {
	return p.peerID
}

func (p *netMessengerStub) Bootstrap() error {
	p.bootstrapWasCalled = true
	return nil
}

func (p *netMessengerStub) RegisterMessageProcessor(topic string, handler p2p.MessageProcessor) error {
	if p.registeredMessageProcessors == nil {
		p.registeredMessageProcessors = make(map[string]p2p.MessageProcessor)
	}

	p.registeredMessageProcessors[topic] = handler
	return nil
}

func (p *netMessengerStub) HasTopic(name string) bool {
	for _, topic := range p.createdTopics {
		if topic == name {
			return true
		}
	}
	return false
}

func (p *netMessengerStub) CreateTopic(name string, _ bool) error {
	p.createdTopics = append(p.createdTopics, name)
	return nil
}

func (p *netMessengerStub) Addresses() []string {
	return nil
}

func (p *netMessengerStub) Broadcast(topic string, data []byte) {
	if topic == ActionsTopicName && string(data) == JoinedAction {
		p.joinedWasCalled = true
	}
}

func (p *netMessengerStub) SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error {
	p.lastSendTopicName = topic
	p.lastSendData = buff
	p.lastSendPeerID = peerID

	return nil
}

func (p *netMessengerStub) Close() error {
	return nil
}
