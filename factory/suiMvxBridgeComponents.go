package factory

import (
	"fmt"
	"os"
	"time"

	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/multiversx/mx-bridge-eth-go/bridges"
	"github.com/multiversx/mx-bridge-eth-go/bridges/disabled"
	multiversxtoeth "github.com/multiversx/mx-bridge-eth-go/bridges/steps/fromMultiversX"
	ethtomultiversx "github.com/multiversx/mx-bridge-eth-go/bridges/steps/toMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/bridges/topology"
	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement/factory"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	mapper "github.com/multiversx/mx-bridge-eth-go/clients/multiversx/mappers/sui"
	roleproviders "github.com/multiversx/mx-bridge-eth-go/clients/roleProviders"
	suiClient "github.com/multiversx/mx-bridge-eth-go/clients/sui"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/p2p"
	"github.com/multiversx/mx-bridge-eth-go/stateMachine"
	"github.com/multiversx/mx-bridge-eth-go/status"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/core/polling"
)

// ArgsSuiToMultiversXBridge is the arguments DTO used for creating a Sui to MultiversX bridge
type ArgsSuiToMultiversXBridge struct {
	Configs                       config.Configs
	Messenger                     p2p.NetMessenger
	StatusStorer                  core.Storer
	Proxy                         multiversx.Proxy
	MultiversXClientStatusHandler core.StatusHandler
	SuiProxy                      sui.ISuiAPI
	SuiClientStatusHandler        core.StatusHandler
	TimeForBootstrap              time.Duration
	TimeBeforeRepeatJoin          time.Duration
	MetricsHolder                 core.MetricsHolder
	AppStatusHandler              chainCore.AppStatusHandler
}

type suiMvxBridgeComponents struct {
	*baseBridgeComponents
	chain           chain.Chain
	suiApi          sui.ISuiAPI
	suiClient       bridges.PeerChainClient
	suiDataGetter   suiDataGetter
	suiSigner       *signer.Signer
	suiRoleProvider PeerChainRoleProvider
	suiPackageId    string
}

func NewSuiMvxBridgeComponents(args ArgsSuiToMultiversXBridge) (*suiMvxBridgeComponents, error) {
	err := checkArgsSui(args)
	if err != nil {
		return nil, err
	}

	commonBridgeArgs := ArgsBridgeCommon{
		Configs:                       args.Configs,
		Messenger:                     args.Messenger,
		StatusStorer:                  args.StatusStorer,
		Proxy:                         args.Proxy,
		MultiversXClientStatusHandler: args.MultiversXClientStatusHandler,
		TimeForBootstrap:              args.TimeForBootstrap,
		TimeBeforeRepeatJoin:          args.TimeBeforeRepeatJoin,
		MetricsHolder:                 args.MetricsHolder,
		AppStatusHandler:              args.AppStatusHandler,
	}
	baseComponents, err := NewBaseComponents(commonBridgeArgs)
	if err != nil {
		return nil, err
	}

	components := &suiMvxBridgeComponents{
		baseBridgeComponents: baseComponents,
		chain:                args.Configs.GeneralConfig.Sui.Chain,
		suiApi:               args.SuiProxy,
	}

	err = components.initBaseComponents(commonBridgeArgs)
	if err != nil {
		return nil, err
	}

	err = components.createSuiKeysAndAddresses(args.Configs.GeneralConfig.Sui)
	if err != nil {
		return nil, err
	}

	err = components.createSuiDataGetter(args)
	if err != nil {
		return nil, err
	}

	err = components.createSuiRoleProvider(args)
	if err != nil {
		return nil, err
	}

	err = components.createSuiClient(args)
	if err != nil {
		return nil, err
	}

	err = components.createSuiToMultiversXBridge(args)
	if err != nil {
		return nil, err
	}

	err = components.createSuiToMultiversXStateMachine()
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXToSuiBridge(args)
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXToSuiStateMachine()
	if err != nil {
		return nil, err
	}

	return components, nil
}

func checkArgsSui(args ArgsSuiToMultiversXBridge) error {
	if args.SuiProxy == nil {
		return fmt.Errorf("%w for Sui proxy", errNilProxy)
	}
	if check.IfNil(args.SuiClientStatusHandler) {
		return errNilStatusHandler
	}

	return nil
}

func (components *suiMvxBridgeComponents) initBaseComponents(args ArgsBridgeCommon) error {
	suiToMultiversXName := components.chain.PeerChainToMultiversXName()
	baseLogId := components.chain.BaseLogId()
	components.baseLogger = core.NewLoggerWithIdentifier(logger.GetOrCreate(suiToMultiversXName), baseLogId)

	dataGetterLogId := components.chain.MultiversXDataGetterLogId()
	err := components.createDataGetter(dataGetterLogId)
	if err != nil {
		return err
	}

	clientLogId := components.chain.MultiversXClientLogId()
	err = components.createMultiversXRoleProvider(args, clientLogId)
	if err != nil {
		return err
	}

	roleProviderLogId := components.chain.MultiversXRoleProviderLogId()
	tokensMapper, err := mapper.NewMultiversXToSuiMapper(components.mxDataGetter)
	if err != nil {
		return err
	}
	err = components.createMultiversXClient(args, roleProviderLogId, tokensMapper)
	if err != nil {
		return err
	}

	return nil
}

func (components *suiMvxBridgeComponents) createSuiKeysAndAddresses(suiConfigs config.SuiConfig) error {
	privKey, err := loadPrivateKeyFromFile(suiConfigs.PrivateKeyFile)
	if err != nil {
		return err
	}
	seed, err := getSeedFromPrivateKey(privKey)
	if err != nil {
		return err
	}

	components.suiSigner = signer.NewSigner(seed)
	components.suiPackageId = suiConfigs.PackageId

	return nil
}

func (components *suiMvxBridgeComponents) createSuiDataGetter(args ArgsSuiToMultiversXBridge) error {
	suiConfig := args.Configs.GeneralConfig.Sui
	suiDataGetterLogId := components.chain.PeerChainDataGetterLogId()
	argsSuiDataGetter := suiClient.ArgsSuiClientDataGetter{
		PackageId:                  components.suiPackageId,
		SafeObjectId:               suiConfig.SafeObjectId,
		SafeInitialSharedVersion:   suiConfig.SafeObjectInitialSharedVersion,
		BridgeObjectId:             suiConfig.BridgeObjectId,
		BridgeInitialSharedVersion: suiConfig.BridgeObjectInitialSharedVersion,
		RelayerAddress:             components.suiSigner.Address,
		Proxy:                      components.suiApi,
		Log:                        core.NewLoggerWithIdentifier(logger.GetOrCreate(suiDataGetterLogId), suiDataGetterLogId),
	}

	var err error
	components.suiDataGetter, err = suiClient.NewSuiClientDataGetter(argsSuiDataGetter)

	return err
}

func (components *suiMvxBridgeComponents) createSuiRoleProvider(args ArgsSuiToMultiversXBridge) error {
	configs := args.Configs.GeneralConfig
	suiRoleProviderLogId := components.chain.PeerChainRoleProviderLogId()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(suiRoleProviderLogId), suiRoleProviderLogId)
	argsRoleProvider := roleproviders.ArgsSuiRoleProvider{
		DataGetter: components.suiDataGetter,
		Log:        log,
	}

	var err error
	components.suiRoleProvider, err = roleproviders.NewSuiRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             string(components.chain) + " role provider",
		PollingInterval:  time.Duration(configs.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.suiRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *suiMvxBridgeComponents) createSuiClient(args ArgsSuiToMultiversXBridge) error {
	suiConfig := args.Configs.GeneralConfig.Sui

	gasStationConfig := suiConfig.GasStation
	argsGasStation := gasManagement.ArgsGasStation{
		RequestURL:             gasStationConfig.URL,
		RequestPollingInterval: time.Duration(gasStationConfig.PollingIntervalInSeconds) * time.Second,
		RequestRetryDelay:      time.Duration(gasStationConfig.RequestRetryDelayInSeconds) * time.Second,
		MaximumFetchRetries:    gasStationConfig.MaxFetchRetries,
		RequestTime:            time.Duration(gasStationConfig.RequestTimeInSeconds) * time.Second,
		MaximumGasPrice:        gasStationConfig.MaximumAllowedGasPrice,
		GasPriceSelector:       core.EthGasPriceSelector(gasStationConfig.GasPriceSelector),
		GasPriceMultiplier:     gasStationConfig.GasPriceMultiplier,
	}

	gs, err := factory.CreateGasStation(argsGasStation, gasStationConfig.Enabled)
	if err != nil {
		return err
	}

	components.addClosableComponent(gs)

	antifloodComponents, err := components.createAntifloodComponents(args.Configs.GeneralConfig.P2P.AntifloodConfig)
	if err != nil {
		return err
	}

	peerDenialEvaluator, err := p2p.NewPeerDenialEvaluator(antifloodComponents.BlacklistHandler, antifloodComponents.PubKeysCacher)
	if err != nil {
		return err
	}
	err = args.Messenger.SetPeerDenialEvaluator(peerDenialEvaluator)
	if err != nil {
		return err
	}

	broadcasterLogId := components.chain.BroadcasterLogId()
	suiToMultiversXName := components.chain.PeerChainToMultiversXName()
	argsBroadcaster := p2p.ArgsBroadcaster{
		Messenger:              args.Messenger,
		Log:                    core.NewLoggerWithIdentifier(logger.GetOrCreate(broadcasterLogId), broadcasterLogId),
		MultiversXRoleProvider: components.multiversXRoleProvider,
		SignatureProcessor:     components.suiRoleProvider, // TODO
		KeyGen:                 keyGen,
		SingleSigner:           singleSigner,
		PrivateKey:             components.multiversXRelayerPrivateKey,
		Name:                   suiToMultiversXName,
		AntifloodComponents:    antifloodComponents,
	}

	components.broadcaster, err = p2p.NewBroadcaster(argsBroadcaster)
	if err != nil {
		return err
	}

	tokensMapper, err := mapper.NewSuiToMultiversXMapper(components.mxDataGetter) // TODO
	if err != nil {
		return err
	}

	signaturesHolder := bridges.NewSignatureHolder()
	components.toMultiversXSignaturesHolder = signaturesHolder
	err = components.broadcaster.AddBroadcastClient(signaturesHolder)
	if err != nil {
		return err
	}

	suiClientLogId := components.chain.PeerChainClientLogId()
	argsSuiClient := suiClient.ArgsSuiClient{
		Proxy:                        components.suiApi,
		Log:                          core.NewLoggerWithIdentifier(logger.GetOrCreate(suiClientLogId), suiClientLogId),
		Signer:                       components.suiSigner,
		PackageId:                    components.suiPackageId,
		SafeObjectId:                 suiConfig.SafeObjectId,
		SafeInitialSharedVersion:     suiConfig.SafeObjectInitialSharedVersion,
		BridgeObjectId:               suiConfig.BridgeObjectId,
		BridgeInitialSharedVersion:   suiConfig.BridgeObjectInitialSharedVersion,
		Broadcaster:                  components.broadcaster,
		TokensMapper:                 tokensMapper,
		SignatureHolder:              signaturesHolder,
		StatusHandler:                args.SuiClientStatusHandler,
		ClientAvailabilityAllowDelta: suiConfig.ClientAvailabilityAllowDelta,
	}

	components.suiClient, err = suiClient.NewSuiClient(argsSuiClient)

	return err
}

func (components *suiMvxBridgeComponents) createSuiToMultiversXBridge(args ArgsSuiToMultiversXBridge) error {
	suiToMultiversXName := components.chain.PeerChainToMultiversXName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(suiToMultiversXName), suiToMultiversXName)

	configs, found := args.Configs.GeneralConfig.StateMachine[suiToMultiversXName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, suiToMultiversXName)
	}

	components.toMultiversXStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond

	argsTopologyHandler := topology.ArgsTopologyHandler{
		PublicKeysProvider: components.multiversXRoleProvider,
		Timer:              components.timer,
		IntervalForLeader:  time.Second * time.Duration(configs.IntervalForLeaderInSeconds),
		AddressBytes:       components.multiversXRelayerAddress.AddressBytes(),
		Log:                log,
		AddressConverter:   components.addressConverter,
	}

	topologyHandler, err := topology.NewTopologyHandler(argsTopologyHandler)
	if err != nil {
		return err
	}

	components.toMultiversXStatusHandler, err = status.NewStatusHandler(suiToMultiversXName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.toMultiversXStatusHandler)
	if err != nil {
		return err
	}

	timeForTransferExecution := time.Second * time.Duration(args.Configs.GeneralConfig.Sui.IntervalToWaitForTransferInSeconds)

	balanceValidator, err := components.createBalanceValidator(components.baseLogger, components.suiClient)
	if err != nil {
		return err
	}

	argsBridgeExecutor := bridges.ArgsBridgeExecutor{
		Log:                          log,
		TopologyProvider:             topologyHandler,
		MultiversXClient:             components.multiversXClient,
		PeerChainClient:              components.suiClient,
		StatusHandler:                components.toMultiversXStatusHandler,
		TimeForWaitOnPeerClient:      timeForTransferExecution,
		SignaturesHolder:             disabled.NewDisabledSignaturesHolder(),
		BalanceValidator:             balanceValidator,
		MaxQuorumRetriesOnPeerClient: args.Configs.GeneralConfig.Sui.MaxRetriesOnQuorumReached,
		MaxQuorumRetriesOnMultiversX: args.Configs.GeneralConfig.MultiversX.MaxRetriesOnQuorumReached,
		MaxRestriesOnWasProposed:     args.Configs.GeneralConfig.MultiversX.MaxRetriesOnWasTransferProposed,
	}

	bridge, err := bridges.NewBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.toMultiversXMachineStates, err = ethtomultiversx.CreateSteps(bridge)
	if err != nil {
		return err
	}

	return nil
}

func (components *suiMvxBridgeComponents) createSuiToMultiversXStateMachine() error {
	suiToMultiversXName := components.chain.PeerChainToMultiversXName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(suiToMultiversXName), suiToMultiversXName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     suiToMultiversXName,
		Steps:                components.toMultiversXMachineStates,
		StartStateIdentifier: ethtomultiversx.GettingPendingBatchFromPeerChain,
		Log:                  log,
		StatusHandler:        components.toMultiversXStatusHandler,
	}

	var err error
	components.toMultiversXStateMachine, err = stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             suiToMultiversXName + " State machine",
		PollingInterval:  components.toMultiversXStepDuration,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.toMultiversXStateMachine,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *suiMvxBridgeComponents) createMultiversXToSuiBridge(args ArgsSuiToMultiversXBridge) error {
	multiversXToSuiName := components.chain.MultiversXToPeerChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXToSuiName), multiversXToSuiName)

	configs, found := args.Configs.GeneralConfig.StateMachine[multiversXToSuiName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, multiversXToSuiName)
	}

	components.fromMultiversXStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond
	argsTopologyHandler := topology.ArgsTopologyHandler{
		PublicKeysProvider: components.multiversXRoleProvider,
		Timer:              components.timer,
		IntervalForLeader:  time.Second * time.Duration(configs.IntervalForLeaderInSeconds),
		AddressBytes:       components.multiversXRelayerAddress.AddressBytes(),
		Log:                log,
		AddressConverter:   components.addressConverter,
	}

	topologyHandler, err := topology.NewTopologyHandler(argsTopologyHandler)
	if err != nil {
		return err
	}

	components.fromMultiversXStatusHandler, err = status.NewStatusHandler(multiversXToSuiName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.fromMultiversXStatusHandler)
	if err != nil {
		return err
	}

	timeForWaitOnSui := time.Second * time.Duration(args.Configs.GeneralConfig.Sui.IntervalToWaitForTransferInSeconds)

	balanceValidator, err := components.createBalanceValidator(components.baseLogger, components.suiClient)
	if err != nil {
		return err
	}

	argsBridgeExecutor := bridges.ArgsBridgeExecutor{
		Log:                          log,
		TopologyProvider:             topologyHandler,
		MultiversXClient:             components.multiversXClient,
		PeerChainClient:              components.suiClient,
		StatusHandler:                components.fromMultiversXStatusHandler,
		TimeForWaitOnPeerClient:      timeForWaitOnSui,
		SignaturesHolder:             components.toMultiversXSignaturesHolder,
		BalanceValidator:             balanceValidator,
		MaxQuorumRetriesOnPeerClient: args.Configs.GeneralConfig.Sui.MaxRetriesOnQuorumReached,
		MaxQuorumRetriesOnMultiversX: args.Configs.GeneralConfig.MultiversX.MaxRetriesOnQuorumReached,
		MaxRestriesOnWasProposed:     args.Configs.GeneralConfig.MultiversX.MaxRetriesOnWasTransferProposed,
	}

	bridge, err := bridges.NewBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.fromMultiversXMachineStates, err = multiversxtoeth.CreateSteps(bridge)
	if err != nil {
		return err
	}

	return nil
}

func (components *suiMvxBridgeComponents) createMultiversXToSuiStateMachine() error {
	multiversXToSuiName := components.chain.MultiversXToPeerChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXToSuiName), multiversXToSuiName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     multiversXToSuiName,
		Steps:                components.fromMultiversXMachineStates,
		StartStateIdentifier: multiversxtoeth.GettingPendingBatchFromMultiversX,
		Log:                  log,
		StatusHandler:        components.fromMultiversXStatusHandler,
	}

	var err error
	components.fromMultiversXStateMachine, err = stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             multiversXToSuiName + " State machine",
		PollingInterval:  components.fromMultiversXStepDuration,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.fromMultiversXStateMachine,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

// Start will start the bridge
func (components *suiMvxBridgeComponents) Start() error {
	return components.start()
}

// Close will close the bridge
func (components *suiMvxBridgeComponents) Close() error {
	return components.close()
}

// PeerChainRelayerAddress returns the Sui address associated to this relayer
func (components *suiMvxBridgeComponents) PeerChainRelayerAddress() string {
	return components.suiSigner.Address
}

func loadPrivateKeyFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getSeedFromPrivateKey(privKey string) ([]byte, error) {
	_, data, err := bech32.Decode(privKey)
	if err != nil {
		return nil, err
	}
	decoded, err := bech32.ConvertBits(data, 5, 8, false)
	if err != nil {
		return nil, err
	}
	if len(decoded) < 33 {
		return nil, err
	}

	seed := decoded[1:33]
	return seed, nil
}
