package relay

import (
	"bytes"
	"context"
	"errors"
	"io"

	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/api"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/elrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement/factory"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	coreBridge "github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/core/polling"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/bridgeExecutors"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/facade"
	relayp2p "github.com/ElrondNetwork/elrond-eth-bridge/relay/p2p"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/roleProvider"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondConfig "github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/ntp"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ethereum/go-ethereum/common"
)

const (
	minimumDurationForStep = time.Second
	pollingDurationOnError = time.Second * 5
)

type defaultTimer struct {
	ntpSyncTimer ntp.SyncTimer
}

func NewDefaultTimer() *defaultTimer {
	return &defaultTimer{
		ntpSyncTimer: ntp.NewSyncTime(elrondConfig.NTPConfig{SyncPeriodSeconds: 3600}, nil),
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

func (s *defaultTimer) IsInterfaceNil() bool {
	return s == nil
}

type Relay struct {
	messenger relayp2p.NetMessenger
	timer     coreBridge.Timer
	log       logger.Logger

	ethBridge    bridge.Bridge
	elrondBridge bridge.Bridge

	elrondRoleProvider   ElrondRoleProvider
	ethereumRoleProvider EthereumRoleProvider
	quorumProvider       bridge.QuorumProvider
	stepDuration         time.Duration
	configs              config.Configs
	broadcaster          Broadcaster
	pollingHandlers      []io.Closer
	elrondAddress        erdgoCore.AddressHandler
	ethereumAddress      common.Address
}

// ArgsRelayer is the DTO used in the relayer constructor
type ArgsRelayer struct {
	Config           Config
	Name             string
	Proxy            bridge.ElrondProxy
	EthClient        eth.BlockchainClient
	EthInstance      eth.BridgeContract
	Messenger        relayp2p.NetMessenger
	Erc20Contracts   map[common.Address]eth.GenericErc20Contract
	BridgeEthAddress common.Address
}

// NewRelay creates a new relayer node able to work on 2-half bridges
// TODO refactor even further this struct
func NewRelay(args ArgsRelayer) (*Relay, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	relay := &Relay{
		messenger: args.Messenger,
		log:       logger.GetOrCreate(args.Name),
		configs:   args.Configs,
	}

	cfgs := args.Configs.GeneralConfig
	wallet := interactors.NewWallet()
	privateKey, err := wallet.LoadPrivateKeyFromPemFile(cfgs.Elrond.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	relay.elrondAddress, err = wallet.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	suite := ed25519.NewEd25519()
	keyGen := signing.NewKeyGenerator(suite)
	txSignPrivKey, err := keyGen.PrivateKeyFromByteArray(privateKey)
	if err != nil {
		return nil, err
	}

	clientArgs := elrond.ClientArgs{
		Config:     cfgs.Elrond,
		Proxy:      args.Proxy,
		PrivateKey: txSignPrivKey,
		Address:    relay.elrondAddress,
	}
	elrondBridge, err := elrond.NewClient(clientArgs)
	if err != nil {
		return nil, err
	}
	relay.elrondBridge = elrondBridge

	gasStationConfig := cfgs.Eth.GasStation
	argsGasStation := gasManagement.ArgsGasStation{
		RequestURL:             gasStationConfig.URL,
		RequestPollingInterval: time.Duration(gasStationConfig.PollingIntervalInSeconds) * time.Second,
		RequestTime:            time.Duration(gasStationConfig.RequestTimeInSeconds) * time.Second,
		MaximumGasPrice:        gasStationConfig.MaximumAllowedGasPrice,
		GasPriceSelector:       coreBridge.EthGasPriceSelector(gasStationConfig.GasPriceSelector),
	}

	gs, err := factory.CreateGasStation(argsGasStation, gasStationConfig.Enabled)
	if err != nil {
		return nil, err
	}

	argsClient := eth.ArgsClient{
		Config:         cfgs.Eth,
		Broadcaster:    relay,
		Mapper:         elrondBridge,
		GasHandler:     gs,
		EthClient:      args.EthClient,
		EthInstance:    args.EthInstance,
		Erc20Contracts: args.Erc20Contracts,
		BridgeAddress:  args.BridgeEthAddress,
	}
	ethBridge, err := eth.NewClient(argsClient)
	if err != nil {
		return nil, err
	}
	relay.ethBridge = ethBridge
	relay.quorumProvider = ethBridge
	relay.ethereumAddress = ethBridge.Address()

	err = relay.createRoleProviders(*cfgs)
	if err != nil {
		return nil, err
	}

	argsBroadcaster := relayp2p.ArgsBroadcaster{
		Messenger:          relay.messenger,
		Log:                relay.log,
		ElrondRoleProvider: relay.elrondRoleProvider,
		KeyGen:             keyGen,
		SingleSigner:       &singlesig.Ed25519Signer{},
		PrivateKey:         txSignPrivKey,
		SignatureProcessor: relay.ethereumRoleProvider,
		Name:               "eth-elrond",
	}
	relay.broadcaster, err = relayp2p.NewBroadcaster(argsBroadcaster)
	if err != nil {
		return nil, err
	}

	relay.timer = NewDefaultTimer()

	relay.log.Debug("creating API services")
	_, err = relay.createHttpServer()
	if err != nil {
		return nil, err
	}

	return relay, nil
}

func checkArgs(args ArgsRelayer) error {
	if check.IfNilReflect(args.Configs) {
		return ErrMissingConfig
	}
	if check.IfNilReflect(args.Configs.GeneralConfig) {
		return ErrMissingGeneralConfig
	}
	if check.IfNilReflect(args.Configs.ApiRoutesConfig) {
		return ErrMissingApiRoutesConfig
	}
	if check.IfNilReflect(args.Configs.FlagsConfig) {
		return ErrMissingFlagsConfig
	}
	if check.IfNil(args.Proxy) {
		return ErrNilElrondProxy
	}
	if check.IfNilReflect(args.EthClient) {
		return ErrNilEthClient
	}
	if check.IfNilReflect(args.EthInstance) {
		return ErrNilEthInstance
	}
	if check.IfNil(args.Messenger) {
		return ErrNilMessenger
	}
	if args.Erc20Contracts == nil {
		return ErrNilErc20Contracts
	}
	return nil
}

func (r *Relay) createRoleProviders(config config.Config) error {
	err := r.createElrondRoleProvider(config)
	if err != nil {
		return err
	}

	err = r.createEthereumRoleProvider(config)
	if err != nil {
		return err
	}

	return nil
}

func (r *Relay) createElrondRoleProvider(config config.Config) error {
	chainInteractor, ok := r.elrondBridge.(ElrondChainInteractor)
	if !ok {
		return errors.New("programming error: r.elrondBridge is not of type ElrondChainInteractor")
	}

	argsRoleProvider := roleProvider.ArgsElrondRoleProvider{
		ElrondChainInteractor: chainInteractor,
		Log:                   r.log,
	}

	erp, err := roleProvider.NewElrondRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}
	r.elrondRoleProvider = erp

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              r.log,
		Name:             "Elrond role provider",
		PollingInterval:  time.Duration(config.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         r.elrondRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	r.pollingHandlers = append(r.pollingHandlers, pollingHandler)

	return pollingHandler.StartProcessingLoop()
}

func (r *Relay) createEthereumRoleProvider(config config.Config) error {
	chainInteractor, ok := r.ethBridge.(EthereumChainInteractor)
	if !ok {
		return errors.New("programming error: r.ethBridge is not of type EthereumChainInteractor")
	}

	argsRoleProvider := roleProvider.ArgsEthereumRoleProvider{
		EthereumChainInteractor: chainInteractor,
		Log:                     r.log,
	}

	erp, err := roleProvider.NewEthereumRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}
	r.ethereumRoleProvider = erp

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              r.log,
		Name:             "Ethereum role provider",
		PollingInterval:  time.Duration(config.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         r.ethereumRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	r.pollingHandlers = append(r.pollingHandlers, pollingHandler)

	return pollingHandler.StartProcessingLoop()
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

	err = r.broadcaster.AddBroadcastClient(bridgeExecutor)
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
	cfg, exists := r.configs.GeneralConfig.StateMachine[name]
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
	for _, closer := range r.pollingHandlers {
		err := closer.Close()
		r.log.LogIfError(err)
	}

	err := r.timer.Close()
	if err != nil {
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

		return bytes.Equal(publicKeys[index], r.elrondAddress.AddressBytes())
	}
}

// BroadcastSignature will broadcast the signature to other peers
func (r *Relay) BroadcastSignature(sig []byte, messageHash []byte) {
	r.broadcaster.BroadcastSignature(sig, messageHash)
}

// ElrondAddress returns the Elrond's address associated to this relayer
func (r *Relay) ElrondAddress() erdgoCore.AddressHandler {
	return r.elrondAddress
}

// EthereumAddress returns the Ethereum's address associated to this relayer
func (r *Relay) EthereumAddress() common.Address {
	return r.ethereumAddress
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

func (r *Relay) createHttpServer() (api.UpgradeableHttpServerHandler, error) {
	httpServerArgs := api.ArgsNewWebServer{
		Facade:    facade.NewRelayerFacade(r.configs.FlagsConfig.RestApiInterface, r.configs.FlagsConfig.EnablePprof),
		ApiConfig: *r.configs.ApiRoutesConfig,
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
