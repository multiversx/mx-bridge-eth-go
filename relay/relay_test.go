package relay

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	relayMock "github.com/ElrondNetwork/elrond-eth-bridge/relay/mock"
	"github.com/ElrondNetwork/elrond-eth-bridge/testHelpers"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// implements interface
var (
	_ = Startable(&Relay{})
	_ = bridge.Broadcaster(&Relay{})
)

var log = logger.GetOrCreate("main")

func TestNewRelay(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Eth: bridge.EthereumConfig{
			NetworkAddress:               "http://127.0.0.1:8545",
			BridgeAddress:                "5DdDe022a65F8063eE9adaC54F359CBF46166068",
			PrivateKey:                   "9bb971db41e3815a669a71c3f1bcb24e0b81f21e04bf11faa7a34b9b40e7cfb1",
			IntervalToResendTxsInSeconds: 0,
			GasLimit:                     0,
			GasStation: bridge.GasStationConfig{
				URL:                      "",
				PollingIntervalInSeconds: 1,
				RequestTimeInSeconds:     1,
				MaximumAllowedGasPrice:   1000,
				GasPriceSelector:         "fast",
			},
		},
		Elrond: bridge.ElrondConfig{
			IntervalToResendTxsInSeconds: 60,
			PrivateKey:                   "testdata/grace.pem",
			NetworkAddress:               "http://127.0.0.1:8079",
			BridgeAddress:                "erd1qqqqqqqqqqqqqpgqgftcwj09u0nhmskrw7xxqcqh8qmzwyexd8ss7ftcxx",
		},
		P2P: ConfigP2P{
			Port:            "0",
			Seed:            "",
			InitialPeerList: nil,
			ProtocolID:      "erd/1.1.0",
		},
		Relayer: ConfigRelayer{
			Marshalizer: config.MarshalizerConfig{
				Type:           "gogo protobuf",
				SizeCheckDelta: 10,
			},
		},
	}
	flagsConfig := ContextFlagsConfig{}
	r, err := NewRelay(cfg, flagsConfig, "name")
	require.Nil(t, err)
	require.False(t, check.IfNil(r))

	r.Clean()
}

func TestInit(t *testing.T) {
	testHelpers.SetTestLogLevel()

	messenger := &netMessengerStub{}
	timer := testHelpers.TimerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &timer,
		log:       log,

		elrondBridge: &bridgeMock{},
		ethBridge:    &bridgeMock{},

		elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	assert.True(t, messenger.bootstrapWasCalled)
	assert.Contains(t, messenger.createdTopics, privateTopicName)
	assert.Contains(t, messenger.createdTopics, joinTopicName)
	assert.Contains(t, messenger.createdTopics, signTopicName)
	assert.Contains(t, messenger.registeredMessageProcessors, privateTopicName)
	assert.Contains(t, messenger.registeredMessageProcessors, joinTopicName)
	assert.Contains(t, messenger.registeredMessageProcessors, signTopicName)
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

		elrondBridge: &bridgeMock{},
		ethBridge:    &bridgeMock{},

		elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	privateMessageProcessor := messenger.registeredMessageProcessors[privateTopicName]
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

			elrondBridge: &bridgeMock{},
			ethBridge:    &bridgeMock{},

			peers:      Peers{"first", "second"},
			signatures: Signatures{"first": []byte("first signature")},

			roleProvider:                &roleProviderStub{isWhitelisted: true},
			elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		joinMessageProcessor := messenger.registeredMessageProcessors[joinTopicName]
		_ = joinMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("other"), "peer_near_me")

		dec := gob.NewDecoder(bytes.NewReader(messenger.lastSendData))
		var got Topology
		if err := dec.Decode(&got); err != nil {
			t.Fatal(err)
		}

		expectedPeers := Peers{"first", "other", "second"}
		expectedSignatures := Signatures{"first": []byte("first signature")}

		assert.Equal(t, expectedPeers, got.Peers)
		assert.Equal(t, expectedSignatures, got.Signatures)
	})
	t.Run("on joined action when there are more peers then self and the peer is not whitelisted it will broadcast to private", func(t *testing.T) {
		messenger := &netMessengerStub{}
		relay := Relay{
			messenger: messenger,
			timer:     &testHelpers.TimerStub{},
			log:       log,

			elrondBridge: &bridgeMock{},
			ethBridge:    &bridgeMock{},

			peers: Peers{"first", "second"},

			roleProvider:                &roleProviderStub{isWhitelisted: false},
			elrondWalletAddressProvider: &walletAddressProviderStub{address: ""},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		joinMessageProcessor := messenger.registeredMessageProcessors[joinTopicName]
		_ = joinMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("other"), "peer_near_me")

		assert.Empty(t, messenger.lastSendData)
	})
	t.Run("when self joined will not broadcast to private", func(t *testing.T) {
		messenger := &netMessengerStub{peerID: "self"}
		relay := Relay{
			messenger: messenger,
			timer:     &testHelpers.TimerStub{},
			log:       log,

			elrondBridge: &bridgeMock{},
			ethBridge:    &bridgeMock{},

			roleProvider:                &roleProviderStub{isWhitelisted: true},
			elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		joinMessageProcessor := messenger.registeredMessageProcessors[joinTopicName]
		_ = joinMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("self"), "peer_near_me")

		assert.NotEqual(t, privateTopicName, messenger.lastSendTopicName)
	})
	t.Run("on joined will not add to the peer list if already present", func(t *testing.T) {
		messenger := &netMessengerStub{peerID: "self"}
		expected := Peers{"first", "second"}
		relay := Relay{
			messenger: messenger,
			timer:     &testHelpers.TimerStub{},
			log:       log,

			elrondBridge: &bridgeMock{},
			ethBridge:    &bridgeMock{},

			peers: Peers{"first", "second"},

			roleProvider:                &roleProviderStub{isWhitelisted: true},
			elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		_ = relay.Start(ctx)

		joinMessageProcessor := messenger.registeredMessageProcessors[joinTopicName]
		_ = joinMessageProcessor.ProcessReceivedMessage(buildJoinedMessage("first"), "peer_near_me")

		assert.Equal(t, expected, relay.peers)
	})
}

func TestJoin(t *testing.T) {
	testHelpers.SetTestLogLevel()

	messenger := &netMessengerStub{}
	relay := Relay{
		messenger: messenger,
		timer:     &testHelpers.TimerStub{},
		log:       log,

		elrondBridge: &bridgeMock{},
		ethBridge:    &bridgeMock{},

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

		elrondBridge: &bridgeMock{},
		ethBridge:    &bridgeMock{},
	}

	_, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	expectedData := []byte("signature")
	relay.SendSignature(expectedData)

	assert.Equal(t, signTopicName, messenger.lastSendTopicName)
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

		elrondBridge: &bridgeMock{},
		ethBridge:    &bridgeMock{},

		peers: Peers{"first", "second"},

		elrondWalletAddressProvider: &walletAddressProviderStub{address: "address1"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	signMessageProcessor := messenger.registeredMessageProcessors[signTopicName]
	expected := []byte("signature")
	_ = signMessageProcessor.ProcessReceivedMessage(buildSignMessage("second", expected), "peer_near_me")

	assert.Equal(t, expected, relay.Signatures()[0])
}

func TestAmILeader(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("will return true when time matches current index", func(t *testing.T) {
		relay := Relay{
			peers:        Peers{"self"},
			messenger:    &netMessengerStub{peerID: "self"},
			timer:        &testHelpers.TimerStub{TimeNowUnix: 0},
			stepDuration: time.Second,
		}

		assert.True(t, relay.AmITheLeader())
	})
	t.Run("will return false when time does not match", func(t *testing.T) {
		stepDuration := time.Second * 6
		relay := Relay{
			stepDuration: stepDuration,
			peers:        Peers{"self", "other"},
			messenger:    &netMessengerStub{peerID: "self"},
			timer:        &testHelpers.TimerStub{TimeNowUnix: int64(stepDuration.Seconds()) + 1},
		}

		assert.False(t, relay.AmITheLeader())
	})
}

func TestRelay_CreateAndStartBridge(t *testing.T) {
	t.Parallel()
	t.Run("nil bridge should error", func(t *testing.T) {
		relay := &Relay{
			quorumProvider:     &relayMock.QuorumProviderStub{},
			timer:              &relayMock.TimerMock{},
			log:                logger.GetOrCreate("test"),
			stateMachineConfig: createMapMockDurationsMapConfig(),
		}

		stateMachine, err := relay.createAndStartBridge(nil, &bridgeMock{}, "EthToElrond")
		require.True(t, check.IfNilReflect(stateMachine))
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "source bridge"))
		require.True(t, strings.Contains(err.Error(), "nil bridge"))
	})
	t.Run("invalid step time duration", func(t *testing.T) {
		t.Run("for first half", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &relayMock.QuorumProviderStub{},
				timer:              &relayMock.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[0]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.StepDurationInMillis = 999
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&bridgeMock{}, &bridgeMock{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrInvalidDurationConfig))
		})
		t.Run("for second half", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &relayMock.QuorumProviderStub{},
				timer:              &relayMock.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[1]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.StepDurationInMillis = 999
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&bridgeMock{}, &bridgeMock{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrInvalidDurationConfig))
		})
		t.Run("for both parts", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &relayMock.QuorumProviderStub{},
				timer:              &relayMock.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[0]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.StepDurationInMillis = 999
			relay.stateMachineConfig[key] = halfBridge

			key = halfBridgeKeys[1]
			halfBridge = relay.stateMachineConfig[key]
			halfBridge.StepDurationInMillis = 999
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&bridgeMock{}, &bridgeMock{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrInvalidDurationConfig))
		})
	})
	t.Run("missing duration for step", func(t *testing.T) {
		t.Run("for first half", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &relayMock.QuorumProviderStub{},
				timer:              &relayMock.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[0]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.Steps = halfBridge.Steps[1:]
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&bridgeMock{}, &bridgeMock{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrMissingDurationConfig))
		})
		t.Run("for second half", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &relayMock.QuorumProviderStub{},
				timer:              &relayMock.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[1]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.Steps = halfBridge.Steps[1:]
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&bridgeMock{}, &bridgeMock{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrMissingDurationConfig))
		})
		t.Run("for both parts", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &relayMock.QuorumProviderStub{},
				timer:              &relayMock.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[0]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.Steps = halfBridge.Steps[1:]
			relay.stateMachineConfig[key] = halfBridge

			key = halfBridgeKeys[1]
			halfBridge = relay.stateMachineConfig[key]
			halfBridge.Steps = halfBridge.Steps[1:]
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&bridgeMock{}, &bridgeMock{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrMissingDurationConfig))
		})
	})
	t.Run("should work", func(t *testing.T) {
		relay := &Relay{
			quorumProvider:     &relayMock.QuorumProviderStub{},
			timer:              &relayMock.TimerMock{},
			stateMachineConfig: createMapMockDurationsMapConfig(),
			log:                logger.GetOrCreate("test"),
		}

		halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)

		stateMachine, err := relay.createAndStartBridge(&bridgeMock{}, &bridgeMock{}, halfBridgeKeys[0])
		require.Nil(t, err)
		require.NotNil(t, stateMachine)

		_ = stateMachine.Close()
	})
}

func createMockDurationsMapConfig() ConfigStateMachine {
	return ConfigStateMachine{
		StepDurationInMillis: 1000,
		Steps: []StepConfig{
			{
				Name:             ethToElrond.GettingPending,
				DurationInMillis: 1,
			},
			{
				Name:             ethToElrond.ProposingTransfer,
				DurationInMillis: 1,
			},
			{
				Name:             ethToElrond.WaitingSignaturesForProposeTransfer,
				DurationInMillis: 1,
			},
			{
				Name:             ethToElrond.ExecutingTransfer,
				DurationInMillis: 1,
			},
			{
				Name:             ethToElrond.ProposingSetStatus,
				DurationInMillis: 1,
			},
			{
				Name:             ethToElrond.WaitingSignaturesForProposeSetStatus,
				DurationInMillis: 1,
			},
			{
				Name:             ethToElrond.ExecutingSetStatus,
				DurationInMillis: 1,
			},
		},
	}
}

func createMapMockDurationsMapConfig() map[string]ConfigStateMachine {
	m := make(map[string]ConfigStateMachine)
	m["ElrondToEth"] = createMockDurationsMapConfig()
	m["EthToElrond"] = createMockDurationsMapConfig()
	return m
}

func getMapMockDurationsMapConfigKeys(m map[string]ConfigStateMachine) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func buildPrivateMessage(peerID core.PeerID, peers Peers) p2p.MessageP2P {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	err := enc.Encode(Topology{Peers: peers})
	if err != nil {
		panic(err)
	}

	return &mock.P2PMessageMock{
		TopicField: privateTopicName,
		PeerField:  peerID,
		DataField:  data.Bytes(),
	}
}

func buildJoinedMessage(peerID core.PeerID) p2p.MessageP2P {
	return &mock.P2PMessageMock{
		TopicField: joinTopicName,
		PeerField:  peerID,
		DataField:  []byte("address"),
	}
}

func buildSignMessage(peerID core.PeerID, signature []byte) p2p.MessageP2P {
	return &mock.P2PMessageMock{
		TopicField: signTopicName,
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

func (p *netMessengerStub) RegisterMessageProcessor(topic string, _ string, handler p2p.MessageProcessor) error {
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
	if topic == joinTopicName && string(data) == "address" {
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
