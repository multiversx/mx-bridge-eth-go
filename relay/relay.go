package relay

import (
	"bytes"
	"context"
	"io"

	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/api"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/elrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement/factory"
	coreBridge "github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/bridgeExecutors"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/facade"
	relayp2p "github.com/ElrondNetwork/elrond-eth-bridge/relay/p2p"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/roleProvider"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go-core/marshal/factory"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/update/disabled"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

const (
	p2pPeerNetworkDiscoverer = "optimized"
	minimumDurationForStep   = time.Second
)

type Peers []core.PeerID

type Signatures map[core.PeerID][]byte

type Topology struct {
	Peers      Peers
	Signatures Signatures
}

type Timer interface {
	After(d time.Duration) <-chan time.Time
	NowUnix() int64
	Start()
	Close() error
}

type defaultTimer struct {
	ntpSyncTimer ntp.SyncTimer
}

func NewDefaultTimer() *defaultTimer {
	return &defaultTimer{
		ntpSyncTimer: ntp.NewSyncTime(config.NTPConfig{SyncPeriodSeconds: 3600}, nil),
	}
}

func (s *defaultTimer) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (s *defaultTimer) NowUnix() int64 {
	return s.ntpSyncTimer.CurrentTime().Unix()
}

func (s *defaultTimer) Start() {
	s.ntpSyncTimer.StartSyncingTime()
}

func (s *defaultTimer) Close() error {
	return s.ntpSyncTimer.Close()
}

type Relay struct {
	messenger relayp2p.NetMessenger
	timer     Timer
	log       logger.Logger

	ethBridge    bridge.Bridge
	elrondBridge bridge.Bridge

	roleProvider                RoleProvider
	elrondWalletAddressProvider bridge.WalletAddressProvider
	quorumProvider              bridge.QuorumProvider
	stepDuration                time.Duration
	stateMachineConfig          map[string]ConfigStateMachine
	flagsConfig                 ContextFlagsConfig
	broadcaster                 Broadcaster
	address                     erdgoCore.AddressHandler
}

func NewRelay(config Config, flagsConfig ContextFlagsConfig, name string) (*Relay, error) {
	relay := &Relay{
		stateMachineConfig: config.StateMachine,
		log:                logger.GetOrCreate(name),
	}

	wallet := interactors.NewWallet()
	privateKey, err := wallet.LoadPrivateKeyFromPemFile(config.Elrond.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	relay.address, err = wallet.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	suite := ed25519.NewEd25519()
	keyGen := signing.NewKeyGenerator(suite)
	txSignPrivKey, err := keyGen.PrivateKeyFromByteArray(privateKey)
	if err != nil {
		return nil, err
	}

	proxy := blockchain.NewElrondProxy(config.Elrond.NetworkAddress, nil)
	clientArgs := elrond.ClientArgs{
		Config:     config.Elrond,
		Proxy:      proxy,
		PrivateKey: txSignPrivKey,
		Address:    relay.address,
	}
	elrondBridge, err := elrond.NewClient(clientArgs)
	if err != nil {
		return nil, err
	}
	relay.elrondBridge = elrondBridge

	argsRoleProvider := roleProvider.ArgsElrondRoleProvider{
		ChainInteractor: elrondBridge,
		Log:             relay.log,
		PollingInterval: time.Duration(config.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
	}

	erp, err := roleProvider.NewElrondRoleProvider(argsRoleProvider)
	if err != nil {
		return nil, err
	}

	relay.roleProvider = erp
	relay.elrondWalletAddressProvider = elrondBridge

	argsGasStation := gasManagement.ArgsGasStation{
		RequestURL:             config.Eth.GasStation.URL,
		RequestPollingInterval: time.Duration(config.Eth.GasStation.PollingIntervalInSeconds) * time.Second,
		RequestTime:            time.Duration(config.Eth.GasStation.RequestTimeInSeconds) * time.Second,
		MaximumGasPrice:        config.Eth.GasStation.MaximumAllowedGasPrice,
		GasPriceSelector:       coreBridge.EthGasPriceSelector(config.Eth.GasStation.GasPriceSelector),
	}

	gs, err := factory.CreateGasStation(argsGasStation, config.Eth.GasStation.Enabled)
	if err != nil {
		return nil, err
	}

	ethBridge, err := eth.NewClient(config.Eth, relay, elrondBridge, gs)
	if err != nil {
		return nil, err
	}
	relay.ethBridge = ethBridge
	relay.quorumProvider = ethBridge

	marshalizer, err := factoryMarshalizer.NewMarshalizer(config.Relayer.Marshalizer.Type)
	if err != nil {
		return nil, err
	}

	messenger, err := buildNetMessenger(&config, marshalizer)
	if err != nil {
		return nil, err
	}
	relay.messenger = messenger

	argsBroadcaster := relayp2p.ArgsBroadcaster{
		Messenger:    messenger,
		Log:          relay.log,
		RoleProvider: relay.roleProvider,
		KeyGen:       keyGen,
		SingleSigner: &singlesig.Ed25519Signer{},
		PrivateKey:   txSignPrivKey,
	}
	relay.broadcaster, err = relayp2p.NewBroadcaster(argsBroadcaster)
	if err != nil {
		return nil, err
	}

	relay.timer = NewDefaultTimer()
	relay.flagsConfig = flagsConfig

	relay.log.Debug("creating API services")
	_, err = relay.createHttpServer()
	if err != nil {
		return nil, err
	}

	return relay, nil
}

func (r *Relay) Start(ctx context.Context) error {
	err := r.init(ctx)
	if err != nil {
		return nil
	}
	r.broadcaster.BroadcastJoinTopic()

	r.timer.Start()

	smEthToElrond, err := r.createAndStartBridge(r.ethBridge, r.elrondBridge, "EthToElrond")
	if err != nil {
		return err
	}
	smElrondToEth, err := r.createAndStartBridge(r.elrondBridge, r.ethBridge, "ElrondToEth")
	if err != nil {
		return err
	}

	<-ctx.Done()
	err = smEthToElrond.Close()
	r.log.LogIfError(err)

	err = smElrondToEth.Close()
	r.log.LogIfError(err)

	err = r.Stop()
	r.log.LogIfError(err)

	return nil
}

func (r *Relay) createAndStartBridge(
	sourceBridge bridge.Bridge,
	destinationBridge bridge.Bridge,
	name string,
) (io.Closer, error) {
	durationsMap, err := r.processStateMachineConfigDurations(name)
	if err != nil {
		return nil, err
	}

	logExecutor := logger.GetOrCreate(name + "/executor")
	argsExecutor := bridgeExecutors.ArgsEthElrondBridgeExecutor{
		ExecutorName:      name,
		Logger:            logExecutor,
		SourceBridge:      sourceBridge,
		DestinationBridge: destinationBridge,
		TopologyProvider:  r,
		QuorumProvider:    r.quorumProvider,
		Timer:             r.timer,
		DurationsMap:      durationsMap,
	}

	bridgeExecutor, err := bridgeExecutors.NewEthElrondBridgeExecutor(argsExecutor)
	if err != nil {
		return nil, err
	}

	stepsMap, err := steps.CreateSteps(bridgeExecutor)
	if err != nil {
		return nil, err
	}

	err = r.checkDurations(stepsMap, durationsMap)
	if err != nil {
		return nil, err
	}

	logStateMachine := logger.GetOrCreate(name + "/statemachine")
	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     name,
		Steps:                stepsMap,
		StartStateIdentifier: ethToElrond.GettingPending,
		DurationBetweenSteps: r.stepDuration,
		Log:                  logStateMachine,
		Timer:                r.timer,
	}

	return stateMachine.NewStateMachine(argsStateMachine)
}

func (r *Relay) processStateMachineConfigDurations(name string) (map[coreBridge.StepIdentifier]time.Duration, error) {
	cfg, exists := r.stateMachineConfig[name]
	if !exists {
		return nil, fmt.Errorf("%w for %q", ErrMissingConfig, name)
	}
	r.stepDuration = time.Duration(cfg.StepDurationInMillis) * time.Millisecond
	r.log.Debug("loaded state machine StepDuration from configs", "duration", r.stepDuration)

	durationsMap := make(map[coreBridge.StepIdentifier]time.Duration)
	for _, stepCfg := range cfg.Steps {
		d := time.Duration(stepCfg.DurationInMillis) * time.Millisecond
		durationsMap[coreBridge.StepIdentifier(stepCfg.Name)] = d
		r.log.Debug("loaded StepDuration from configs", "step", stepCfg.Name, "duration", d)
	}

	return durationsMap, nil
}

func (r *Relay) checkDurations(
	steps map[coreBridge.StepIdentifier]coreBridge.Step,
	stepsDurations map[coreBridge.StepIdentifier]time.Duration,
) error {
	if r.stepDuration < minimumDurationForStep {
		return fmt.Errorf("%w for config %q", ErrInvalidDurationConfig, "StepDurationInMillis")
	}

	for stepIdentifer := range steps {
		_, found := stepsDurations[stepIdentifer]
		if !found {
			return fmt.Errorf("%w for step %q", ErrMissingDurationConfig, stepIdentifer)
		}
	}

	return nil
}

func (r *Relay) Stop() error {
	if err := r.timer.Close(); err != nil {
		r.log.Error(err.Error())
	}
	return r.broadcaster.Close()
}

// AmITheLeader returns true if the current relayer is the leader in this round
// TODO since now we can have different values for the step duration, move this to the bridge executor
func (r *Relay) AmITheLeader() bool {
	publicKeys := r.broadcaster.SortedPublicKeys()

	if len(publicKeys) == 0 {
		return false
	} else {
		numberOfPeers := int64(len(publicKeys))
		index := (r.timer.NowUnix() / int64(r.stepDuration.Seconds())) % numberOfPeers

		return bytes.Equal(publicKeys[index], r.address.AddressBytes())
	}
}

// Clean will clean any stored signatures
func (r *Relay) Clean() {
	r.broadcaster.ClearSignatures()
}

// Signatures returns any stored signatures
func (r *Relay) Signatures() [][]byte {
	return r.broadcaster.Signatures()
}

// SendSignature will broadcast the signature to other peers
func (r *Relay) SendSignature(sig []byte) {
	r.broadcaster.BroadcastSignature(sig)
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *Relay) IsInterfaceNil() bool {
	return r == nil
}

func (r *Relay) init(ctx context.Context) error {
	err := r.messenger.Bootstrap()
	if err != nil {
		return err
	}

	select {
	case <-r.timer.After(10 * time.Second):
		r.log.Info(fmt.Sprint(r.messenger.Addresses()))

		err = r.broadcaster.RegisterOnTopics()
		if err != nil {
			return nil
		}
	case <-ctx.Done():
		return nil
	}

	return nil
}

func (r *Relay) createHttpServer() (shared.UpgradeableHttpServerHandler, error) {
	httpServerArgs := api.ArgsNewWebServer{
		Facade: facade.NewRelayerFacade(r.flagsConfig.RestApiInterface, r.flagsConfig.EnablePprof),
	}

	httpServerWrapper, err := api.NewWebServerHandler(httpServerArgs)
	if err != nil {
		return nil, err
	}

	err = httpServerWrapper.StartHttpServer()
	if err != nil {
		return nil, err
	}

	return httpServerWrapper, nil
}

func buildNetMessenger(cfg *Config, marshalizer marshal.Marshalizer) (relayp2p.NetMessenger, error) {
	nodeConfig := config.NodeConfig{
		Port:                       cfg.P2P.Port,
		Seed:                       cfg.P2P.Seed,
		MaximumExpectedPeerCount:   0,
		ThresholdMinConnectedPeers: 0,
	}
	peerDiscoveryConfig := config.KadDhtPeerDiscoveryConfig{
		Enabled:                          true,
		RefreshIntervalInSec:             5,
		ProtocolID:                       cfg.P2P.ProtocolID,
		InitialPeerList:                  cfg.P2P.InitialPeerList,
		BucketSize:                       0,
		RoutingTableRefreshIntervalInSec: 300,
		Type:                             p2pPeerNetworkDiscoverer,
	}

	p2pConfig := config.P2PConfig{
		Node:                nodeConfig,
		KadDhtPeerDiscovery: peerDiscoveryConfig,
		Sharding: config.ShardingConfig{
			TargetPeerCount:         0,
			MaxIntraShardValidators: 0,
			MaxCrossShardValidators: 0,
			MaxIntraShardObservers:  0,
			MaxCrossShardObservers:  0,
			Type:                    "NilListSharder",
		},
	}

	args := libp2p.ArgsNetworkMessenger{
		Marshalizer:          marshalizer,
		ListenAddress:        libp2p.ListenAddrWithIp4AndTcp,
		P2pConfig:            p2pConfig,
		SyncTimer:            &libp2p.LocalSyncTimer{},
		PreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		NodeOperationMode:    p2p.NormalOperation,
	}

	messenger, err := libp2p.NewNetworkMessenger(args)
	if err != nil {
		panic(err)
	}

	return messenger, nil
}
