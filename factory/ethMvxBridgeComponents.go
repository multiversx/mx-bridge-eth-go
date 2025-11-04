package factory

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/bridges"
	"github.com/multiversx/mx-bridge-eth-go/bridges/disabled"
	"github.com/multiversx/mx-bridge-eth-go/bridges/steps/fromMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/bridges/steps/toMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/bridges/topology"
	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement/factory"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx/mappers/eth"
	roleproviders "github.com/multiversx/mx-bridge-eth-go/clients/roleProviders"
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

// ArgsEthereumToMultiversXBridge is the arguments DTO used for creating an Ethereum to MultiversX bridge
type ArgsEthereumToMultiversXBridge struct {
	Configs                       config.Configs
	Messenger                     p2p.NetMessenger
	StatusStorer                  core.Storer
	Proxy                         multiversx.Proxy
	MultiversXClientStatusHandler core.StatusHandler
	Erc20ContractsHolder          ethereum.Erc20ContractsHolder
	ClientWrapper                 ethereum.ClientWrapper
	TimeForBootstrap              time.Duration
	TimeBeforeRepeatJoin          time.Duration
	MetricsHolder                 core.MetricsHolder
	AppStatusHandler              chainCore.AppStatusHandler
}

type ethMvxBridgeComponents struct {
	*baseBridgeComponents
	evmCompatibleChain     chain.Chain
	ethClient              bridges.PeerChainClient
	ethereumRelayerAddress common.Address
	ethereumRoleProvider   PeerChainRoleProvider
}

func NewEthMvxBridgeComponents(args ArgsEthereumToMultiversXBridge) (*ethMvxBridgeComponents, error) {
	err := checkArgsEthereum(args)
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

	components := &ethMvxBridgeComponents{
		baseBridgeComponents: baseComponents,
		evmCompatibleChain:   args.Configs.GeneralConfig.Eth.Chain,
	}

	err = components.initBaseComponents(commonBridgeArgs)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumRoleProvider(args)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumClient(args)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumToMultiversXBridge(args)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumToMultiversXStateMachine()
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXToEthereumBridge(args)
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXToEthereumStateMachine()
	if err != nil {
		return nil, err
	}

	return components, nil
}

func checkArgsEthereum(args ArgsEthereumToMultiversXBridge) error {
	if check.IfNil(args.ClientWrapper) {
		return errNilEthClient
	}
	if check.IfNil(args.Erc20ContractsHolder) {
		return errNilErc20ContractsHolder
	}
	return nil
}

func (components *ethMvxBridgeComponents) initBaseComponents(args ArgsBridgeCommon) error {
	ethToMultiversXName := components.evmCompatibleChain.PeerChainToMultiversXName()
	baseLogId := components.evmCompatibleChain.BaseLogId()
	components.baseLogger = core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToMultiversXName), baseLogId)

	dataGetterLogId := components.evmCompatibleChain.MultiversXDataGetterLogId()
	err := components.createDataGetter(dataGetterLogId)
	if err != nil {
		return err
	}

	clientLogId := components.evmCompatibleChain.MultiversXClientLogId()
	err = components.createMultiversXRoleProvider(args, clientLogId)
	if err != nil {
		return err
	}

	roleProviderLogId := components.evmCompatibleChain.MultiversXRoleProviderLogId()
	tokensMapper, err := eth.NewMultiversXToErc20Mapper(components.mxDataGetter)
	if err != nil {
		return err
	}
	err = components.createMultiversXClient(args, roleProviderLogId, tokensMapper)
	if err != nil {
		return err
	}

	return nil
}

func (components *ethMvxBridgeComponents) createEthereumRoleProvider(args ArgsEthereumToMultiversXBridge) error {
	configs := args.Configs.GeneralConfig
	ethRoleProviderLogId := components.evmCompatibleChain.PeerChainRoleProviderLogId()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethRoleProviderLogId), ethRoleProviderLogId)
	argsRoleProvider := roleproviders.ArgsEthereumRoleProvider{
		EthereumChainInteractor: args.ClientWrapper,
		Log:                     log,
	}

	var err error
	components.ethereumRoleProvider, err = roleproviders.NewEthereumRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             string(components.evmCompatibleChain) + " role provider",
		PollingInterval:  time.Duration(configs.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.ethereumRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethMvxBridgeComponents) createEthereumClient(args ArgsEthereumToMultiversXBridge) error {
	ethereumConfigs := args.Configs.GeneralConfig.Eth

	gasStationConfig := ethereumConfigs.GasStation
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

	broadcasterLogId := components.evmCompatibleChain.BroadcasterLogId()
	ethToMultiversXName := components.evmCompatibleChain.PeerChainToMultiversXName()
	argsBroadcaster := p2p.ArgsBroadcaster{
		Messenger:              args.Messenger,
		Log:                    core.NewLoggerWithIdentifier(logger.GetOrCreate(broadcasterLogId), broadcasterLogId),
		MultiversXRoleProvider: components.multiversXRoleProvider,
		SignatureProcessor:     components.ethereumRoleProvider,
		KeyGen:                 keyGen,
		SingleSigner:           singleSigner,
		PrivateKey:             components.multiversXRelayerPrivateKey,
		Name:                   ethToMultiversXName,
		AntifloodComponents:    antifloodComponents,
	}

	components.broadcaster, err = p2p.NewBroadcaster(argsBroadcaster)
	if err != nil {
		return err
	}

	cryptoHandler, err := ethereum.NewCryptoHandler(ethereumConfigs.PrivateKeyFile)
	if err != nil {
		return err
	}

	components.ethereumRelayerAddress = cryptoHandler.GetAddress()

	tokensMapper, err := eth.NewErc20ToMultiversXMapper(components.mxDataGetter)
	if err != nil {
		return err
	}

	signaturesHolder := bridges.NewSignatureHolder()
	components.toMultiversXSignaturesHolder = signaturesHolder
	err = components.broadcaster.AddBroadcastClient(signaturesHolder)
	if err != nil {
		return err
	}

	safeContractAddress := common.HexToAddress(ethereumConfigs.SafeContractAddress)

	ethClientLogId := components.evmCompatibleChain.PeerChainClientLogId()
	argsEthClient := ethereum.ArgsEthereumClient{
		ClientWrapper:                args.ClientWrapper,
		Erc20ContractsHandler:        args.Erc20ContractsHolder,
		Log:                          core.NewLoggerWithIdentifier(logger.GetOrCreate(ethClientLogId), ethClientLogId),
		AddressConverter:             components.addressConverter,
		Broadcaster:                  components.broadcaster,
		CryptoHandler:                cryptoHandler,
		TokensMapper:                 tokensMapper,
		SignatureHolder:              signaturesHolder,
		SafeContractAddress:          safeContractAddress,
		GasHandler:                   gs,
		TransferGasLimitBase:         ethereumConfigs.GasLimitBase,
		TransferGasLimitForEach:      ethereumConfigs.GasLimitForEach,
		ClientAvailabilityAllowDelta: ethereumConfigs.ClientAvailabilityAllowDelta,
		EventsBlockRangeFrom:         ethereumConfigs.EventsBlockRangeFrom,
		EventsBlockRangeTo:           ethereumConfigs.EventsBlockRangeTo,
	}

	components.ethClient, err = ethereum.NewEthereumClient(argsEthClient)

	return err
}

func (components *ethMvxBridgeComponents) createEthereumToMultiversXBridge(args ArgsEthereumToMultiversXBridge) error {
	ethToMultiversXName := components.evmCompatibleChain.PeerChainToMultiversXName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToMultiversXName), ethToMultiversXName)

	configs, found := args.Configs.GeneralConfig.StateMachine[ethToMultiversXName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, ethToMultiversXName)
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

	components.toMultiversXStatusHandler, err = status.NewStatusHandler(ethToMultiversXName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.toMultiversXStatusHandler)
	if err != nil {
		return err
	}

	timeForTransferExecution := time.Second * time.Duration(args.Configs.GeneralConfig.Eth.IntervalToWaitForTransferInSeconds)

	balanceValidator, err := components.createBalanceValidator(components.baseLogger, components.ethClient)
	if err != nil {
		return err
	}

	argsBridgeExecutor := bridges.ArgsBridgeExecutor{
		Log:                          log,
		TopologyProvider:             topologyHandler,
		MultiversXClient:             components.multiversXClient,
		PeerChainClient:              components.ethClient,
		StatusHandler:                components.toMultiversXStatusHandler,
		TimeForWaitOnPeerClient:      timeForTransferExecution,
		SignaturesHolder:             disabled.NewDisabledSignaturesHolder(),
		BalanceValidator:             balanceValidator,
		MaxQuorumRetriesOnPeerClient: args.Configs.GeneralConfig.Eth.MaxRetriesOnQuorumReached,
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

func (components *ethMvxBridgeComponents) createEthereumToMultiversXStateMachine() error {
	ethToMultiversXName := components.evmCompatibleChain.PeerChainToMultiversXName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToMultiversXName), ethToMultiversXName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     ethToMultiversXName,
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
		Name:             ethToMultiversXName + " State machine",
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

func (components *ethMvxBridgeComponents) createMultiversXToEthereumBridge(args ArgsEthereumToMultiversXBridge) error {
	multiversXToEthName := components.evmCompatibleChain.MultiversXToPeerChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXToEthName), multiversXToEthName)

	configs, found := args.Configs.GeneralConfig.StateMachine[multiversXToEthName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, multiversXToEthName)
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

	components.fromMultiversXStatusHandler, err = status.NewStatusHandler(multiversXToEthName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.fromMultiversXStatusHandler)
	if err != nil {
		return err
	}

	timeForWaitOnEthereum := time.Second * time.Duration(args.Configs.GeneralConfig.Eth.IntervalToWaitForTransferInSeconds)

	balanceValidator, err := components.createBalanceValidator(components.baseLogger, components.ethClient)
	if err != nil {
		return err
	}

	argsBridgeExecutor := bridges.ArgsBridgeExecutor{
		Log:                          log,
		TopologyProvider:             topologyHandler,
		MultiversXClient:             components.multiversXClient,
		PeerChainClient:              components.ethClient,
		StatusHandler:                components.fromMultiversXStatusHandler,
		TimeForWaitOnPeerClient:      timeForWaitOnEthereum,
		SignaturesHolder:             components.toMultiversXSignaturesHolder,
		BalanceValidator:             balanceValidator,
		MaxQuorumRetriesOnPeerClient: args.Configs.GeneralConfig.Eth.MaxRetriesOnQuorumReached,
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

func (components *ethMvxBridgeComponents) createMultiversXToEthereumStateMachine() error {
	multiversXToEthName := components.evmCompatibleChain.MultiversXToPeerChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXToEthName), multiversXToEthName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     multiversXToEthName,
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
		Name:             multiversXToEthName + " State machine",
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
func (components *ethMvxBridgeComponents) Start() error {
	return components.start()
}

// Close will close the bridge
func (components *ethMvxBridgeComponents) Close() error {
	return components.close()
}

// PeerChainRelayerAddress returns the Ethereum's address associated to this relayer
func (components *ethMvxBridgeComponents) PeerChainRelayerAddress() string {
	return components.ethereumRelayerAddress.String()
}
