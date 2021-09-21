package relay

import (
	"bytes"
	"context"
	"encoding/gob"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"

	"github.com/ElrondNetwork/elrond-eth-bridge/testHelpers"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

// implements interface
var (
	_ = Startable(&Relay{})
	_ = bridge.Broadcaster(&Relay{})
)

var log = logger.GetOrCreate("main")

func TestInit(t *testing.T) {
	testHelpers.SetTestLogLevel()

	messenger := &netMessengerStub{}
	timer := testHelpers.TimerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &timer,
		log:       log,

		elrondBridge: &bridgeStub{},
		ethBridge:    &bridgeStub{},

		elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	assert.True(t, messenger.bootstrapWasCalled)
	assert.Contains(t, messenger.createdTopics, PrivateTopicName)
	assert.Contains(t, messenger.createdTopics, JoinTopicName)
	assert.Contains(t, messenger.createdTopics, SignTopicName)
	assert.Contains(t, messenger.registeredMessageProcessors, PrivateTopicName)
	assert.Contains(t, messenger.registeredMessageProcessors, JoinTopicName)
	assert.Contains(t, messenger.registeredMessageProcessors, SignTopicName)
	assert.True(t, timer.WasStarted)
}

func TestClean(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("it will clean signatures", func(t *testing.T) {
		relay := Relay{
			signatures: Signatures{"peer": []byte("some signature")},
		}

		relay.Clean()

		assert.Empty(t, relay.signatures)
	})
}

func TestPrivateTopicProcessor(t *testing.T) {
	testHelpers.SetTestLogLevel()

	messenger := &netMessengerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &testHelpers.TimerStub{},
		log:       log,

		elrondBridge: &bridgeStub{},
		ethBridge:    &bridgeStub{},

		elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
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

func TestJoinTopicProcessor(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("on joined action when there are more peers then self will broadcast to private", func(t *testing.T) {
		messenger := &netMessengerStub{}
		relay := Relay{
			messenger: messenger,
			timer:     &testHelpers.TimerStub{},
			log:       log,

			elrondBridge: &bridgeStub{},
			ethBridge:    &bridgeStub{},

			peers: Peers{"first", "second"},

			roleProvider:                &roleProviderStub{isWhitelisted: true},
			elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		joinMessageProcessor := messenger.registeredMessageProcessors[JoinTopicName]
		_ = joinMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("other"), "peer_near_me")

		dec := gob.NewDecoder(bytes.NewReader(messenger.lastSendData))
		var got Peers
		if err := dec.Decode(&got); err != nil {
			t.Fatal(err)
		}

		expected := Peers{"first", "other", "second"}

		assert.Equal(t, expected, got)
	})
	t.Run("on joined action when there are more peers then self and the peer is not whitelisted it will broadcast to private", func(t *testing.T) {
		messenger := &netMessengerStub{}
		relay := Relay{
			messenger: messenger,
			timer:     &testHelpers.TimerStub{},
			log:       log,

			elrondBridge: &bridgeStub{},
			ethBridge:    &bridgeStub{},

			peers: Peers{"first", "second"},

			roleProvider:                &roleProviderStub{isWhitelisted: false},
			elrondWalletAddressProvider: &walletAddressProviderStub{address: ""},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		joinMessageProcessor := messenger.registeredMessageProcessors[JoinTopicName]
		_ = joinMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("other"), "peer_near_me")

		assert.Empty(t, messenger.lastSendData)
	})
	t.Run("when self joined will not broadcast to private", func(t *testing.T) {
		messenger := &netMessengerStub{peerID: "self"}
		relay := Relay{
			messenger: messenger,
			timer:     &testHelpers.TimerStub{},
			log:       log,

			elrondBridge: &bridgeStub{},
			ethBridge:    &bridgeStub{},

			roleProvider:                &roleProviderStub{isWhitelisted: true},
			elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		joinMessageProcessor := messenger.registeredMessageProcessors[JoinTopicName]
		_ = joinMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("self"), "peer_near_me")

		assert.NotEqual(t, PrivateTopicName, messenger.lastSendTopicName)
	})
}

func TestJoin(t *testing.T) {
	testHelpers.SetTestLogLevel()

	messenger := &netMessengerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &testHelpers.TimerStub{},
		log:       log,

		elrondBridge: &bridgeStub{},
		ethBridge:    &bridgeStub{},

		elrondWalletAddressProvider: &walletAddressProviderStub{address: "address"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	assert.True(t, messenger.joinedWasCalled)
}

func TestSendSignature(t *testing.T) {
	testHelpers.SetTestLogLevel()

	messenger := &netMessengerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &testHelpers.TimerStub{},
		log:       log,

		elrondBridge: &bridgeStub{},
		ethBridge:    &bridgeStub{},
	}

	_, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	expectedData := []byte("signature")
	relay.SendSignature(expectedData)

	assert.Equal(t, SignTopicName, messenger.lastSendTopicName)
	assert.Equal(t, expectedData, messenger.lastSendData)
}

func TestSignTopicProcessor(t *testing.T) {
	testHelpers.SetTestLogLevel()

	messenger := &netMessengerStub{peerID: "first"}
	relay := Relay{
		messenger:  messenger,
		timer:      &testHelpers.TimerStub{},
		log:        log,
		signatures: make(map[core.PeerID][]byte),

		elrondBridge: &bridgeStub{},
		ethBridge:    &bridgeStub{},

		peers: Peers{"first", "second"},

		elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	signMessageProcessor := messenger.registeredMessageProcessors[SignTopicName]
	expected := []byte("signature")
	_ = signMessageProcessor.ProcessReceivedMessage(buildSignMessage("second", expected), "peer_near_me")

	assert.Equal(t, expected, relay.Signatures()[0])
}

func TestAmILeader(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("will return true when time matches current index", func(t *testing.T) {
		relay := Relay{
			peers:     Peers{"self"},
			messenger: &netMessengerStub{peerID: "self"},
			timer:     &testHelpers.TimerStub{TimeNowUnix: 0},
		}

		assert.True(t, relay.AmITheLeader())
	})
	t.Run("will return false when time does not match", func(t *testing.T) {
		relay := Relay{
			peers:     Peers{"self", "other"},
			messenger: &netMessengerStub{peerID: "self"},
			timer:     &testHelpers.TimerStub{TimeNowUnix: int64(Timeout.Seconds()) + 1},
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
		TopicField: JoinTopicName,
		PeerField:  peerID,
		DataField:  []byte("address"),
	}
}

func buildSignMessage(peerID core.PeerID, signature []byte) p2p.MessageP2P {
	return &mock.P2PMessageMock{
		TopicField: SignTopicName,
		PeerField:  peerID,
		DataField:  signature,
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
	if topic == JoinTopicName && string(data) == "address" {
		p.joinedWasCalled = true
	}

	p.lastSendTopicName = topic
	p.lastSendData = data
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

type roleProviderStub struct {
	isWhitelisted bool
}

func (r *roleProviderStub) IsWhitelisted(string) bool {
	return r.isWhitelisted
}

type walletAddressProviderStub struct {
	address string
}

func (r *walletAddressProviderStub) GetHexWalletAddress() string {
	return r.address
}
