package relay

import (
	"bytes"
	"context"
	"encoding/gob"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/p2p/mock"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
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

func TestReadPendingTransaction(t *testing.T) {
	t.Run("it will read the next pending transaction", func(t *testing.T) {
		expected := &bridge.DepositTransaction{Hash: "hash"}
		ethBridge := &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expected}}
		relay := Relay{
			messenger: &netMessengerStub{},
			timer:     &timerStub{},
			log:       log,

			elrondBridge: &bridgeStub{},
			ethBridge:    ethBridge,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expected, relay.pendingTransaction)
	})
	t.Run("it will sleep and try again if there is no pending transaction", func(t *testing.T) {
		expected := &bridge.DepositTransaction{Hash: "hash"}
		ethBridge := &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{nil, expected}}
		relay := Relay{
			messenger: &netMessengerStub{},
			timer:     &timerStub{sleepDuration: 1 * time.Millisecond},
			log:       log,

			elrondBridge: &bridgeStub{},
			ethBridge:    ethBridge,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expected, relay.pendingTransaction)
		assert.GreaterOrEqual(t, ethBridge.pendingTransactionCallIndex, 1)
	})
}

func TestPropose(t *testing.T) {
	t.Run("it will propose eth transaction when leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{Hash: "hash"}
		elrondBridge := &bridgeStub{}
		relay := Relay{
			peers:     Peers{"first"},
			messenger: &netMessengerStub{peerID: "first"},
			timer:     &timerStub{timeNowUnix: 0},
			log:       log,

			elrondBridge: elrondBridge,
			ethBridge:    &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expect, elrondBridge.lastProposedTransaction)
	})
	t.Run("it will wait for proposal if not leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{Hash: "hash"}
		elrondBridge := &bridgeStub{}
		relay := Relay{
			peers:     Peers{"first", "second"},
			messenger: &netMessengerStub{peerID: "first"},
			timer:     &timerStub{timeNowUnix: int64(Timeout.Seconds()) + 1},
			log:       log,

			elrondBridge: elrondBridge,
			ethBridge:    &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expect, elrondBridge.lastWasProposedTransaction)
	})
	t.Run("it will sign proposed transaction if not leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{Hash: "hash"}
		elrondBridge := &bridgeStub{wasProposed: true}
		relay := Relay{
			peers:     Peers{"first", "second"},
			messenger: &netMessengerStub{peerID: "first"},
			timer:     &timerStub{timeNowUnix: int64(Timeout.Seconds()) + 1},
			log:       log,

			elrondBridge: elrondBridge,
			ethBridge:    &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expect, elrondBridge.lastSignedTransaction)
	})
	t.Run("it will try to propose again if timeout", func(t *testing.T) {
		expect := &bridge.DepositTransaction{Hash: "hash"}
		elrondBridge := &bridgeStub{wasProposed: false}
		timer := &timerStub{timeNowUnix: int64(Timeout.Seconds()) + 1, afterDuration: 5 * time.Millisecond}
		relay := Relay{
			peers:     Peers{"first", "second"},
			messenger: &netMessengerStub{peerID: "first"},
			timer:     timer,
			log:       log,

			elrondBridge: elrondBridge,
			ethBridge:    &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
		}

		go func() {
			time.Sleep(3 * time.Millisecond)
			timer.timeNowUnix = 0
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expect, elrondBridge.lastProposedTransaction)
	})
}

func TestWaitForSignatures(t *testing.T) {
	t.Run("it will execute when number of signatures is > 67%", func(t *testing.T) {
		expect := &bridge.DepositTransaction{Hash: "hash"}
		elrondBridge := &bridgeStub{signersCount: 3}
		relay := Relay{
			peers:     Peers{"first", "other", "second", "third"},
			messenger: &netMessengerStub{peerID: "first"},
			timer:     &timerStub{timeNowUnix: 0},
			log:       log,

			elrondBridge: elrondBridge,
			ethBridge:    &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expect, elrondBridge.lastExecutedTransaction)
	})
	t.Run("it will sleep and try to wait for signatures again", func(t *testing.T) {
		expect := &bridge.DepositTransaction{Hash: "hash"}
		elrondBridge := &bridgeStub{signersCount: 0}
		relay := Relay{
			peers:     Peers{"first", "other", "second", "third"},
			messenger: &netMessengerStub{peerID: "first"},
			timer:     &timerStub{timeNowUnix: 0, afterDuration: 3 * time.Millisecond},
			log:       log,

			elrondBridge: elrondBridge,
			ethBridge:    &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
		}

		go func() {
			time.Sleep(2 * time.Millisecond)
			elrondBridge.signersCount = 3
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expect, elrondBridge.lastExecutedTransaction)
	})
}

func TestExecute(t *testing.T) {
	t.Run("it will wait for execution when not leader", func(t *testing.T) {
		expect := &bridge.DepositTransaction{Hash: "hash"}
		elrondBridge := &bridgeStub{signersCount: 3, wasExecuted: false, wasProposed: true}
		timer := &timerStub{timeNowUnix: int64(Timeout.Seconds()) + 1, afterDuration: 3 * time.Millisecond}
		relay := Relay{
			peers:     Peers{"first", "other", "second", "third"},
			messenger: &netMessengerStub{peerID: "first"},
			timer:     timer,
			log:       log,

			elrondBridge: elrondBridge,
			ethBridge:    &bridgeStub{pendingTransactions: []*bridge.DepositTransaction{expect}},
		}

		go func() {
			time.Sleep(2 * time.Millisecond)
			timer.timeNowUnix = 0
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		assert.Equal(t, expect, elrondBridge.lastExecutedTransaction)
	})
}

func TestWaitForExecute(t *testing.T) {}

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

type bridgeStub struct {
	pendingTransactionCallIndex int
	pendingTransactions         []*bridge.DepositTransaction
	wasProposed                 bool
	lastProposedTransaction     *bridge.DepositTransaction
	lastWasProposedTransaction  *bridge.DepositTransaction
	lastSignedTransaction       *bridge.DepositTransaction
	signersCount                uint
	lastExecutedTransaction     *bridge.DepositTransaction
	wasExecuted                 bool
}

func (b *bridgeStub) GetPendingDepositTransaction(context.Context) *bridge.DepositTransaction {
	defer func() { b.pendingTransactionCallIndex++ }()

	if b.pendingTransactionCallIndex >= len(b.pendingTransactions) {
		return nil
	} else {
		return b.pendingTransactions[b.pendingTransactionCallIndex]
	}
}

func (b *bridgeStub) Propose(_ context.Context, tx *bridge.DepositTransaction) {
	b.lastProposedTransaction = tx
}

func (b *bridgeStub) WasProposed(_ context.Context, tx *bridge.DepositTransaction) bool {
	b.lastWasProposedTransaction = tx
	return b.wasProposed
}

func (b *bridgeStub) WasExecuted(context.Context, *bridge.DepositTransaction) bool {
	return b.wasExecuted
}

func (b *bridgeStub) Sign(_ context.Context, tx *bridge.DepositTransaction) {
	b.lastSignedTransaction = tx
}

func (b *bridgeStub) Execute(_ context.Context, tx *bridge.DepositTransaction) (string, error) {
	b.lastExecutedTransaction = tx
	return "", nil
}

func (b *bridgeStub) SignersCount(context.Context, *bridge.DepositTransaction) uint {
	return b.signersCount
}

type timerStub struct {
	sleepDuration time.Duration
	afterDuration time.Duration
	timeNowUnix   int64
}

func (s *timerStub) sleep(time.Duration) {
	time.Sleep(s.sleepDuration)
}

func (s *timerStub) after(time.Duration) <-chan time.Time {
	return time.After(s.afterDuration)
}

func (s *timerStub) nowUnix() int64 {
	return s.timeNowUnix
}
