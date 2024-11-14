package factory

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/disabled"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps/ethToMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps/multiversxToEth"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/topology"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	balanceValidatorManagement "github.com/multiversx/mx-bridge-eth-go/clients/balanceValidator"
	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement/factory"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx/mappers"
	"github.com/multiversx/mx-bridge-eth-go/clients/roleProviders"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-bridge-eth-go/core/timer"
	"github.com/multiversx/mx-bridge-eth-go/p2p"
	"github.com/multiversx/mx-bridge-eth-go/stateMachine"
	"github.com/multiversx/mx-bridge-eth-go/status"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	chainConfig "github.com/multiversx/mx-chain-go/config"
	antifloodFactory "github.com/multiversx/mx-chain-go/process/throttle/antiflood/factory"
	logger "github.com/multiversx/mx-chain-logger-go"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/core/polling"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/interactors"
)

const (
	minTimeForBootstrap     = time.Millisecond * 100
	minTimeBeforeRepeatJoin = time.Second * 30
	pollingDurationOnError  = time.Second * 5
)

var suite = ed25519.NewEd25519()
var keyGen = signing.NewKeyGenerator(suite)
var singleSigner = &singlesig.Ed25519Signer{}

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

type ethMultiversXBridgeComponents struct {
	baseLogger                        logger.Logger
	messenger                         p2p.NetMessenger
	statusStorer                      core.Storer
	multiversXClient                  ethmultiversx.MultiversXClient
	ethClient                         ethmultiversx.EthereumClient
	evmCompatibleChain                chain.Chain
	multiversXMultisigContractAddress sdkCore.AddressHandler
	multiversXSafeContractAddress     sdkCore.AddressHandler
	multiversXRelayerPrivateKey       crypto.PrivateKey
	multiversXRelayerAddress          sdkCore.AddressHandler
	ethereumRelayerAddress            common.Address
	mxDataGetter                      dataGetter
	proxy                             multiversx.Proxy
	multiversXRoleProvider            MultiversXRoleProvider
	ethereumRoleProvider              EthereumRoleProvider
	broadcaster                       Broadcaster
	timer                             core.Timer
	timeForBootstrap                  time.Duration
	metricsHolder                     core.MetricsHolder
	addressConverter                  core.AddressConverter

	ethToMultiversXMachineStates    core.MachineStates
	ethToMultiversXStepDuration     time.Duration
	ethToMultiversXStatusHandler    core.StatusHandler
	ethToMultiversXStateMachine     StateMachine
	ethToMultiversXSignaturesHolder ethmultiversx.SignaturesHolder

	multiversXToEthMachineStates core.MachineStates
	multiversXToEthStepDuration  time.Duration
	multiversXToEthStatusHandler core.StatusHandler
	multiversXToEthStateMachine  StateMachine

	mutClosableHandlers sync.RWMutex
	closableHandlers    []io.Closer

	pollingHandlers []PollingHandler

	timeBeforeRepeatJoin time.Duration
	cancelFunc           func()
	appStatusHandler     chainCore.AppStatusHandler
}

// NewEthMultiversXBridgeComponents creates a new eth-multiversx bridge components holder
func NewEthMultiversXBridgeComponents(args ArgsEthereumToMultiversXBridge) (*ethMultiversXBridgeComponents, error) {
	err := checkArgsEthereumToMultiversXBridge(args)
	if err != nil {
		return nil, err
	}
	evmCompatibleChain := args.Configs.GeneralConfig.Eth.Chain
	ethToMultiversXName := evmCompatibleChain.EvmCompatibleChainToMultiversXName()
	baseLogId := evmCompatibleChain.BaseLogId()
	components := &ethMultiversXBridgeComponents{
		baseLogger:           core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToMultiversXName), baseLogId),
		evmCompatibleChain:   evmCompatibleChain,
		messenger:            args.Messenger,
		statusStorer:         args.StatusStorer,
		closableHandlers:     make([]io.Closer, 0),
		proxy:                args.Proxy,
		timer:                timer.NewNTPTimer(),
		timeForBootstrap:     args.TimeForBootstrap,
		timeBeforeRepeatJoin: args.TimeBeforeRepeatJoin,
		metricsHolder:        args.MetricsHolder,
		appStatusHandler:     args.AppStatusHandler,
	}

	addressConverter, err := converters.NewAddressConverter()
	if err != nil {
		return nil, clients.ErrNilAddressConverter
	}
	components.addressConverter = addressConverter

	components.addClosableComponent(components.timer)

	err = components.createMultiversXKeysAndAddresses(args.Configs.GeneralConfig.MultiversX)
	if err != nil {
		return nil, err
	}

	err = components.createDataGetter()
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXRoleProvider(args)
	if err != nil {
		return nil, err
	}

	err = components.createMultiversXClient(args)
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

func (components *ethMultiversXBridgeComponents) addClosableComponent(closable io.Closer) {
	components.mutClosableHandlers.Lock()
	components.closableHandlers = append(components.closableHandlers, closable)
	components.mutClosableHandlers.Unlock()
}

func checkArgsEthereumToMultiversXBridge(args ArgsEthereumToMultiversXBridge) error {
	if check.IfNil(args.Proxy) {
		return errNilProxy
	}
	if check.IfNil(args.Messenger) {
		return errNilMessenger
	}
	if check.IfNil(args.ClientWrapper) {
		return errNilEthClient
	}
	if check.IfNil(args.StatusStorer) {
		return errNilStatusStorer
	}
	if check.IfNil(args.Erc20ContractsHolder) {
		return errNilErc20ContractsHolder
	}
	if args.TimeForBootstrap < minTimeForBootstrap {
		return fmt.Errorf("%w for TimeForBootstrap, received: %v, minimum: %v", errInvalidValue, args.TimeForBootstrap, minTimeForBootstrap)
	}
	if args.TimeBeforeRepeatJoin < minTimeBeforeRepeatJoin {
		return fmt.Errorf("%w for TimeBeforeRepeatJoin, received: %v, minimum: %v", errInvalidValue, args.TimeBeforeRepeatJoin, minTimeBeforeRepeatJoin)
	}
	if check.IfNil(args.MetricsHolder) {
		return errNilMetricsHolder
	}
	if check.IfNil(args.AppStatusHandler) {
		return errNilStatusHandler
	}

	return nil
}

func (components *ethMultiversXBridgeComponents) createMultiversXKeysAndAddresses(chainConfigs config.MultiversXConfig) error {
	wallet := interactors.NewWallet()
	multiversXPrivateKeyBytes, err := wallet.LoadPrivateKeyFromPemFile(chainConfigs.PrivateKeyFile)
	if err != nil {
		return err
	}

	components.multiversXRelayerPrivateKey, err = keyGen.PrivateKeyFromByteArray(multiversXPrivateKeyBytes)
	if err != nil {
		return err
	}

	components.multiversXRelayerAddress, err = wallet.GetAddressFromPrivateKey(multiversXPrivateKeyBytes)
	if err != nil {
		return err
	}

	components.multiversXMultisigContractAddress, err = data.NewAddressFromBech32String(chainConfigs.MultisigContractAddress)
	if err != nil {
		return fmt.Errorf("%w for chainConfigs.MultisigContractAddress", err)
	}

	components.multiversXSafeContractAddress, err = data.NewAddressFromBech32String(chainConfigs.SafeContractAddress)
	if err != nil {
		return fmt.Errorf("%w for chainConfigs.SafeContractAddress", err)
	}

	return nil
}

func (components *ethMultiversXBridgeComponents) createDataGetter() error {
	multiversXDataGetterLogId := components.evmCompatibleChain.MultiversXDataGetterLogId()
	argsMXClientDataGetter := multiversx.ArgsMXClientDataGetter{
		MultisigContractAddress: components.multiversXMultisigContractAddress,
		SafeContractAddress:     components.multiversXSafeContractAddress,
		RelayerAddress:          components.multiversXRelayerAddress,
		Proxy:                   components.proxy,
		Log:                     core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXDataGetterLogId), multiversXDataGetterLogId),
	}

	var err error
	components.mxDataGetter, err = multiversx.NewMXClientDataGetter(argsMXClientDataGetter)

	return err
}

func (components *ethMultiversXBridgeComponents) createMultiversXClient(args ArgsEthereumToMultiversXBridge) error {
	chainConfigs := args.Configs.GeneralConfig.MultiversX
	tokensMapper, err := mappers.NewMultiversXToErc20Mapper(components.mxDataGetter)
	if err != nil {
		return err
	}
	multiversXClientLogId := components.evmCompatibleChain.MultiversXClientLogId()

	clientArgs := multiversx.ClientArgs{
		GasMapConfig:                 chainConfigs.GasMap,
		Proxy:                        args.Proxy,
		Log:                          core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXClientLogId), multiversXClientLogId),
		RelayerPrivateKey:            components.multiversXRelayerPrivateKey,
		MultisigContractAddress:      components.multiversXMultisigContractAddress,
		SafeContractAddress:          components.multiversXSafeContractAddress,
		IntervalToResendTxsInSeconds: chainConfigs.IntervalToResendTxsInSeconds,
		TokensMapper:                 tokensMapper,
		RoleProvider:                 components.multiversXRoleProvider,
		StatusHandler:                args.MultiversXClientStatusHandler,
		ClientAvailabilityAllowDelta: chainConfigs.ClientAvailabilityAllowDelta,
	}

	components.multiversXClient, err = multiversx.NewClient(clientArgs)
	components.addClosableComponent(components.multiversXClient)

	return err
}

func (components *ethMultiversXBridgeComponents) createEthereumClient(args ArgsEthereumToMultiversXBridge) error {
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
	ethToMultiversXName := components.evmCompatibleChain.EvmCompatibleChainToMultiversXName()
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

	tokensMapper, err := mappers.NewErc20ToMultiversXMapper(components.mxDataGetter)
	if err != nil {
		return err
	}

	signaturesHolder := ethmultiversx.NewSignatureHolder()
	components.ethToMultiversXSignaturesHolder = signaturesHolder
	err = components.broadcaster.AddBroadcastClient(signaturesHolder)
	if err != nil {
		return err
	}

	safeContractAddress := common.HexToAddress(ethereumConfigs.SafeContractAddress)

	ethClientLogId := components.evmCompatibleChain.EvmCompatibleChainClientLogId()
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

func (components *ethMultiversXBridgeComponents) createMultiversXRoleProvider(args ArgsEthereumToMultiversXBridge) error {
	configs := args.Configs.GeneralConfig
	multiversXRoleProviderLogId := components.evmCompatibleChain.MultiversXRoleProviderLogId()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXRoleProviderLogId), multiversXRoleProviderLogId)

	argsRoleProvider := roleproviders.ArgsMultiversXRoleProvider{
		DataGetter: components.mxDataGetter,
		Log:        log,
	}

	var err error
	components.multiversXRoleProvider, err = roleproviders.NewMultiversXRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             "MultiversX role provider",
		PollingInterval:  time.Duration(configs.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.multiversXRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethMultiversXBridgeComponents) createEthereumRoleProvider(args ArgsEthereumToMultiversXBridge) error {
	configs := args.Configs.GeneralConfig
	ethRoleProviderLogId := components.evmCompatibleChain.EvmCompatibleChainRoleProviderLogId()
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

func (components *ethMultiversXBridgeComponents) createEthereumToMultiversXBridge(args ArgsEthereumToMultiversXBridge) error {
	ethToMultiversXName := components.evmCompatibleChain.EvmCompatibleChainToMultiversXName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToMultiversXName), ethToMultiversXName)

	configs, found := args.Configs.GeneralConfig.StateMachine[ethToMultiversXName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, ethToMultiversXName)
	}

	components.ethToMultiversXStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond

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

	components.ethToMultiversXStatusHandler, err = status.NewStatusHandler(ethToMultiversXName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.ethToMultiversXStatusHandler)
	if err != nil {
		return err
	}

	timeForTransferExecution := time.Second * time.Duration(args.Configs.GeneralConfig.Eth.IntervalToWaitForTransferInSeconds)

	balanceValidator, err := components.createBalanceValidator()
	if err != nil {
		return err
	}

	argsBridgeExecutor := ethmultiversx.ArgsBridgeExecutor{
		Log:                          log,
		TopologyProvider:             topologyHandler,
		MultiversXClient:             components.multiversXClient,
		EthereumClient:               components.ethClient,
		StatusHandler:                components.ethToMultiversXStatusHandler,
		TimeForWaitOnEthereum:        timeForTransferExecution,
		SignaturesHolder:             disabled.NewDisabledSignaturesHolder(),
		BalanceValidator:             balanceValidator,
		MaxQuorumRetriesOnEthereum:   args.Configs.GeneralConfig.Eth.MaxRetriesOnQuorumReached,
		MaxQuorumRetriesOnMultiversX: args.Configs.GeneralConfig.MultiversX.MaxRetriesOnQuorumReached,
		MaxRestriesOnWasProposed:     args.Configs.GeneralConfig.MultiversX.MaxRetriesOnWasTransferProposed,
	}

	bridge, err := ethmultiversx.NewBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.ethToMultiversXMachineStates, err = ethtomultiversx.CreateSteps(bridge)
	if err != nil {
		return err
	}

	return nil
}

func (components *ethMultiversXBridgeComponents) createMultiversXToEthereumBridge(args ArgsEthereumToMultiversXBridge) error {
	multiversXToEthName := components.evmCompatibleChain.MultiversXToEvmCompatibleChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXToEthName), multiversXToEthName)

	configs, found := args.Configs.GeneralConfig.StateMachine[multiversXToEthName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, multiversXToEthName)
	}

	components.multiversXToEthStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond
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

	components.multiversXToEthStatusHandler, err = status.NewStatusHandler(multiversXToEthName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.multiversXToEthStatusHandler)
	if err != nil {
		return err
	}

	timeForWaitOnEthereum := time.Second * time.Duration(args.Configs.GeneralConfig.Eth.IntervalToWaitForTransferInSeconds)

	balanceValidator, err := components.createBalanceValidator()
	if err != nil {
		return err
	}

	argsBridgeExecutor := ethmultiversx.ArgsBridgeExecutor{
		Log:                          log,
		TopologyProvider:             topologyHandler,
		MultiversXClient:             components.multiversXClient,
		EthereumClient:               components.ethClient,
		StatusHandler:                components.multiversXToEthStatusHandler,
		TimeForWaitOnEthereum:        timeForWaitOnEthereum,
		SignaturesHolder:             components.ethToMultiversXSignaturesHolder,
		BalanceValidator:             balanceValidator,
		MaxQuorumRetriesOnEthereum:   args.Configs.GeneralConfig.Eth.MaxRetriesOnQuorumReached,
		MaxQuorumRetriesOnMultiversX: args.Configs.GeneralConfig.MultiversX.MaxRetriesOnQuorumReached,
		MaxRestriesOnWasProposed:     args.Configs.GeneralConfig.MultiversX.MaxRetriesOnWasTransferProposed,
	}

	bridge, err := ethmultiversx.NewBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.multiversXToEthMachineStates, err = multiversxtoeth.CreateSteps(bridge)
	if err != nil {
		return err
	}

	return nil
}

func (components *ethMultiversXBridgeComponents) startPollingHandlers() error {
	for _, pollingHandler := range components.pollingHandlers {
		err := pollingHandler.StartProcessingLoop()
		if err != nil {
			return err
		}
	}

	return nil
}

// Start will start the bridge
func (components *ethMultiversXBridgeComponents) Start() error {
	err := components.messenger.Bootstrap()
	if err != nil {
		return err
	}

	components.baseLogger.Info("waiting for p2p bootstrap", "time", components.timeForBootstrap)
	time.Sleep(components.timeForBootstrap)

	err = components.broadcaster.RegisterOnTopics()
	if err != nil {
		return err
	}

	components.broadcaster.BroadcastJoinTopic()

	err = components.startPollingHandlers()
	if err != nil {
		return err
	}

	var ctx context.Context
	ctx, components.cancelFunc = context.WithCancel(context.Background())
	go components.startBroadcastJoinRetriesLoop(ctx)

	return nil
}

func (components *ethMultiversXBridgeComponents) createBalanceValidator() (ethmultiversx.BalanceValidator, error) {
	argsBalanceValidator := balanceValidatorManagement.ArgsBalanceValidator{
		Log:              components.baseLogger,
		MultiversXClient: components.multiversXClient,
		EthereumClient:   components.ethClient,
	}

	return balanceValidatorManagement.NewBalanceValidator(argsBalanceValidator)
}

func (components *ethMultiversXBridgeComponents) createEthereumToMultiversXStateMachine() error {
	ethToMultiversXName := components.evmCompatibleChain.EvmCompatibleChainToMultiversXName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToMultiversXName), ethToMultiversXName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     ethToMultiversXName,
		Steps:                components.ethToMultiversXMachineStates,
		StartStateIdentifier: ethtomultiversx.GettingPendingBatchFromEthereum,
		Log:                  log,
		StatusHandler:        components.ethToMultiversXStatusHandler,
	}

	var err error
	components.ethToMultiversXStateMachine, err = stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             ethToMultiversXName + " State machine",
		PollingInterval:  components.ethToMultiversXStepDuration,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.ethToMultiversXStateMachine,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethMultiversXBridgeComponents) createMultiversXToEthereumStateMachine() error {
	multiversXToEthName := components.evmCompatibleChain.MultiversXToEvmCompatibleChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXToEthName), multiversXToEthName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     multiversXToEthName,
		Steps:                components.multiversXToEthMachineStates,
		StartStateIdentifier: multiversxtoeth.GettingPendingBatchFromMultiversX,
		Log:                  log,
		StatusHandler:        components.multiversXToEthStatusHandler,
	}

	var err error
	components.multiversXToEthStateMachine, err = stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             multiversXToEthName + " State machine",
		PollingInterval:  components.multiversXToEthStepDuration,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.multiversXToEthStateMachine,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethMultiversXBridgeComponents) createAntifloodComponents(antifloodConfig chainConfig.AntifloodConfig) (*antifloodFactory.AntiFloodComponents, error) {
	var err error
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelFunc()
		}
	}()

	cfg := chainConfig.Config{
		Antiflood: antifloodConfig,
	}
	antiFloodComponents, err := antifloodFactory.NewP2PAntiFloodComponents(ctx, cfg, components.appStatusHandler, components.messenger.ID())
	if err != nil {
		return nil, err
	}
	return antiFloodComponents, nil
}

func (components *ethMultiversXBridgeComponents) startBroadcastJoinRetriesLoop(ctx context.Context) {
	broadcastTimer := time.NewTimer(components.timeBeforeRepeatJoin)
	defer broadcastTimer.Stop()

	for {
		broadcastTimer.Reset(components.timeBeforeRepeatJoin)

		select {
		case <-broadcastTimer.C:
			components.baseLogger.Info("broadcast again join topic")
			components.broadcaster.BroadcastJoinTopic()
		case <-ctx.Done():
			components.baseLogger.Info("closing broadcast join topic loop")
			return

		}
	}
}

// Close will close any sub-components started
func (components *ethMultiversXBridgeComponents) Close() error {
	components.mutClosableHandlers.RLock()
	defer components.mutClosableHandlers.RUnlock()

	if components.cancelFunc != nil {
		components.cancelFunc()
	}

	var lastError error
	for _, closable := range components.closableHandlers {
		if closable == nil {
			components.baseLogger.Warn("programming error, nil closable component")
			continue
		}

		err := closable.Close()
		if err != nil {
			lastError = err

			components.baseLogger.Error("error closing component", "error", err)
		}
	}

	return lastError
}

// MultiversXRelayerAddress returns the MultiversX's address associated to this relayer
func (components *ethMultiversXBridgeComponents) MultiversXRelayerAddress() sdkCore.AddressHandler {
	return components.multiversXRelayerAddress
}

// EthereumRelayerAddress returns the Ethereum's address associated to this relayer
func (components *ethMultiversXBridgeComponents) EthereumRelayerAddress() common.Address {
	return components.ethereumRelayerAddress
}
