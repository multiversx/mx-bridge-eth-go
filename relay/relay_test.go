package relay

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	p2pMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/p2p"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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
			PrivateKeyFile:               "testdata/grace.sk",
			IntervalToResendTxsInSeconds: 0,
			GasLimit:                     0,
			GasStation: bridge.GasStationConfig{
				Enabled:                  true,
				URL:                      "",
				PollingIntervalInSeconds: 1,
				RequestTimeInSeconds:     1,
				MaximumAllowedGasPrice:   1000,
				GasPriceSelector:         "fast",
			},
		},
		Elrond: bridge.ElrondConfig{
			IntervalToResendTxsInSeconds: 60,
			PrivateKeyFile:               "testdata/grace.pem",
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
			RoleProvider: RoleProviderConfig{
				PollingIntervalInMillis: 1000,
			},
		},
	}
	flagsConfig := ContextFlagsConfig{}
	ethClient, err := ethclient.Dial(cfg.Eth.NetworkAddress)
	require.Nil(t, err)

	ethInstance, err := contract.NewBridge(ethCommon.HexToAddress(cfg.Eth.BridgeAddress), ethClient)
	require.Nil(t, err)

	args := ArgsRelayer{
		Config:      cfg,
		FlagsConfig: flagsConfig,
		Name:        "name",
		Proxy:       blockchain.NewElrondProxy(cfg.Elrond.NetworkAddress, nil),
		EthClient:   ethClient,
		EthInstance: ethInstance,
	}
	r, err := NewRelay(args)
	require.Nil(t, err)
	require.False(t, check.IfNil(r))

	r.Clean()
}

func TestInit(t *testing.T) {
	messenger := &p2pMocks.MessengerMock{}
	timer := testsCommon.NewTimerStub()
	broadcastJoinTopicCalled := false
	relay := Relay{
		messenger: messenger,
		timer:     timer,
		log:       log,

		elrondBridge: &testsCommon.BridgeStub{},
		ethBridge:    &testsCommon.BridgeStub{},

		broadcaster: &testsCommon.BroadcasterStub{
			BroadcastJoinTopicCalled: func() {
				broadcastJoinTopicCalled = true
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_ = relay.Start(ctx)

	assert.True(t, messenger.BootstrapWasCalled)
	assert.Equal(t, 1, timer.GetFunctionCounter("Start"))
	assert.True(t, broadcastJoinTopicCalled)
}

func TestClean(t *testing.T) {
	t.Run("it will clean signatures", func(t *testing.T) {
		clearCalled := false
		relay := Relay{
			broadcaster: &testsCommon.BroadcasterStub{
				ClearSignaturesCalled: func() {
					clearCalled = true
				},
			},
		}

		relay.Clean()

		assert.True(t, clearCalled)
	})
}

func TestAmILeader(t *testing.T) {
	t.Run("will return true when time matches current index", func(t *testing.T) {
		relay := Relay{
			broadcaster: &testsCommon.BroadcasterStub{
				SortedPublicKeysCalled: func() [][]byte {
					return [][]byte{[]byte("self")}
				},
			},
			address:      data.NewAddressFromBytes([]byte("self")),
			messenger:    &p2pMocks.MessengerMock{PeerID: "self"},
			timer:        testsCommon.NewTimerStub(),
			stepDuration: time.Second,
		}

		assert.True(t, relay.AmITheLeader())
	})
	t.Run("will return false when time does not match", func(t *testing.T) {
		stepDuration := time.Second * 6
		timer := testsCommon.NewTimerStub()
		timer.NowUnixCalled = func() int64 {
			return int64(stepDuration.Seconds()) + 1
		}
		relay := Relay{
			stepDuration: stepDuration,
			broadcaster: &testsCommon.BroadcasterStub{
				SortedPublicKeysCalled: func() [][]byte {
					return [][]byte{[]byte("self"), []byte("other")}
				},
			},
			address: data.NewAddressFromBytes([]byte("self")),
			timer:   timer,
		}

		assert.False(t, relay.AmITheLeader())
	})
}

func TestRelay_CreateAndStartBridge(t *testing.T) {
	t.Parallel()
	t.Run("nil bridge should error", func(t *testing.T) {
		relay := &Relay{
			quorumProvider:     &testsCommon.QuorumProviderStub{},
			timer:              &testsCommon.TimerMock{},
			log:                logger.GetOrCreate("test"),
			stateMachineConfig: createMapMockDurationsMapConfig(),
		}

		stateMachine, err := relay.createAndStartBridge(nil, &testsCommon.BridgeStub{}, "EthToElrond")
		require.True(t, check.IfNilReflect(stateMachine))
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "source bridge"))
		require.True(t, strings.Contains(err.Error(), "nil bridge"))
	})
	t.Run("invalid step time duration", func(t *testing.T) {
		t.Run("for first half", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &testsCommon.QuorumProviderStub{},
				timer:              &testsCommon.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[0]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.StepDurationInMillis = 999
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&testsCommon.BridgeStub{}, &testsCommon.BridgeStub{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrInvalidDurationConfig))
		})
		t.Run("for second half", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &testsCommon.QuorumProviderStub{},
				timer:              &testsCommon.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[1]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.StepDurationInMillis = 999
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&testsCommon.BridgeStub{}, &testsCommon.BridgeStub{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrInvalidDurationConfig))
		})
		t.Run("for both parts", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &testsCommon.QuorumProviderStub{},
				timer:              &testsCommon.TimerMock{},
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

			stateMachine, err := relay.createAndStartBridge(&testsCommon.BridgeStub{}, &testsCommon.BridgeStub{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrInvalidDurationConfig))
		})
	})
	t.Run("missing duration for step", func(t *testing.T) {
		t.Run("for first half", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &testsCommon.QuorumProviderStub{},
				timer:              &testsCommon.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[0]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.Steps = halfBridge.Steps[1:]
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&testsCommon.BridgeStub{}, &testsCommon.BridgeStub{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrMissingDurationConfig))
		})
		t.Run("for second half", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &testsCommon.QuorumProviderStub{},
				timer:              &testsCommon.TimerMock{},
				stateMachineConfig: createMapMockDurationsMapConfig(),
				log:                logger.GetOrCreate("test"),
			}
			halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)
			key := halfBridgeKeys[1]
			halfBridge := relay.stateMachineConfig[key]
			halfBridge.Steps = halfBridge.Steps[1:]
			relay.stateMachineConfig[key] = halfBridge

			stateMachine, err := relay.createAndStartBridge(&testsCommon.BridgeStub{}, &testsCommon.BridgeStub{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrMissingDurationConfig))
		})
		t.Run("for both parts", func(t *testing.T) {
			relay := &Relay{
				quorumProvider:     &testsCommon.QuorumProviderStub{},
				timer:              &testsCommon.TimerMock{},
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

			stateMachine, err := relay.createAndStartBridge(&testsCommon.BridgeStub{}, &testsCommon.BridgeStub{}, key)
			require.True(t, check.IfNilReflect(stateMachine))
			require.True(t, errors.Is(err, ErrMissingDurationConfig))
		})
	})
	t.Run("should work", func(t *testing.T) {
		relay := &Relay{
			quorumProvider:     &testsCommon.QuorumProviderStub{},
			timer:              &testsCommon.TimerMock{},
			stateMachineConfig: createMapMockDurationsMapConfig(),
			log:                logger.GetOrCreate("test"),
		}

		halfBridgeKeys := getMapMockDurationsMapConfigKeys(relay.stateMachineConfig)

		stateMachine, err := relay.createAndStartBridge(&testsCommon.BridgeStub{}, &testsCommon.BridgeStub{}, halfBridgeKeys[0])
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
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
