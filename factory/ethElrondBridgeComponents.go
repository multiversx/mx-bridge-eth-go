package factory

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/disabled"
	elrondToEthSteps "github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps/elrondToEth"
	ethToElrondSteps "github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/steps/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond/topology"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	batchValidatorManagement "github.com/ElrondNetwork/elrond-eth-bridge/clients/batchValidator"
	batchManagementFactory "github.com/ElrondNetwork/elrond-eth-bridge/clients/batchValidator/factory"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/chain"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/elrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/elrond/mappers"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/gasManagement"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/gasManagement/factory"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/roleProviders"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/core/converters"
	"github.com/ElrondNetwork/elrond-eth-bridge/core/timer"
	"github.com/ElrondNetwork/elrond-eth-bridge/p2p"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine"
	"github.com/ElrondNetwork/elrond-eth-bridge/status"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondConfig "github.com/ElrondNetwork/elrond-go/config"
	antifloodFactory "github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/factory"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core/polling"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	minTimeForBootstrap     = time.Millisecond * 100
	minTimeBeforeRepeatJoin = time.Second * 30
	pollingDurationOnError  = time.Second * 5
)

var suite = ed25519.NewEd25519()
var keyGen = signing.NewKeyGenerator(suite)
var singleSigner = &singlesig.Ed25519Signer{}

// ArgsEthereumToElrondBridge is the arguments DTO used for creating an Ethereum to Elrond bridge
type ArgsEthereumToElrondBridge struct {
	Configs                   config.Configs
	Messenger                 p2p.NetMessenger
	StatusStorer              core.Storer
	Proxy                     elrond.ElrondProxy
	ElrondClientStatusHandler core.StatusHandler
	Erc20ContractsHolder      ethereum.Erc20ContractsHolder
	ClientWrapper             ethereum.ClientWrapper
	TimeForBootstrap          time.Duration
	TimeBeforeRepeatJoin      time.Duration
	MetricsHolder             core.MetricsHolder
	AppStatusHandler          elrondCore.AppStatusHandler
}

type ethElrondBridgeComponents struct {
	baseLogger                    logger.Logger
	messenger                     p2p.NetMessenger
	statusStorer                  core.Storer
	elrondClient                  ethElrond.ElrondClient
	ethClient                     ethElrond.EthereumClient
	evmCompatibleChain            chain.Chain
	elrondMultisigContractAddress erdgoCore.AddressHandler
	elrondRelayerPrivateKey       crypto.PrivateKey
	elrondRelayerAddress          erdgoCore.AddressHandler
	ethereumRelayerAddress        common.Address
	dataGetter                    dataGetter
	proxy                         elrond.ElrondProxy
	elrondRoleProvider            ElrondRoleProvider
	ethereumRoleProvider          EthereumRoleProvider
	broadcaster                   Broadcaster
	timer                         core.Timer
	timeForBootstrap              time.Duration
	metricsHolder                 core.MetricsHolder
	addressConverter              core.AddressConverter

	ethToElrondMachineStates    core.MachineStates
	ethToElrondStepDuration     time.Duration
	ethToElrondStatusHandler    core.StatusHandler
	ethToElrondStateMachine     StateMachine
	ethToElrondSignaturesHolder ethElrond.SignaturesHolder

	elrondToEthMachineStates core.MachineStates
	elrondToEthStepDuration  time.Duration
	elrondToEthStatusHandler core.StatusHandler
	elrondToEthStateMachine  StateMachine

	mutClosableHandlers sync.RWMutex
	closableHandlers    []io.Closer

	pollingHandlers []PollingHandler

	timeBeforeRepeatJoin time.Duration
	cancelFunc           func()
	appStatusHandler     elrondCore.AppStatusHandler
}

// NewEthElrondBridgeComponents creates a new eth-elrond bridge components holder
func NewEthElrondBridgeComponents(args ArgsEthereumToElrondBridge) (*ethElrondBridgeComponents, error) {
	err := checkArgsEthereumToElrondBridge(args)
	if err != nil {
		return nil, err
	}
	evmCompatibleChain := args.Configs.GeneralConfig.Eth.Chain
	ethToElrondName := evmCompatibleChain.EvmCompatibleChainToElrondName()
	baseLogId := evmCompatibleChain.BaseLogId()
	components := &ethElrondBridgeComponents{
		baseLogger:           core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToElrondName), baseLogId),
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

	err = components.createElrondKeysAndAddresses(args.Configs.GeneralConfig.Elrond)
	if err != nil {
		return nil, err
	}

	err = components.createDataGetter()
	if err != nil {
		return nil, err
	}

	err = components.createElrondRoleProvider(args)
	if err != nil {
		return nil, err
	}

	err = components.createElrondClient(args)
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

	err = components.createEthereumToElrondBridge(args)
	if err != nil {
		return nil, err
	}

	err = components.createEthereumToElrondStateMachine()
	if err != nil {
		return nil, err
	}

	err = components.createElrondToEthereumBridge(args)
	if err != nil {
		return nil, err
	}

	err = components.createElrondToEthereumStateMachine()
	if err != nil {
		return nil, err
	}

	return components, nil
}

func (components *ethElrondBridgeComponents) addClosableComponent(closable io.Closer) {
	components.mutClosableHandlers.Lock()
	components.closableHandlers = append(components.closableHandlers, closable)
	components.mutClosableHandlers.Unlock()
}

func checkArgsEthereumToElrondBridge(args ArgsEthereumToElrondBridge) error {
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

func (components *ethElrondBridgeComponents) createElrondKeysAndAddresses(elrondConfigs config.ElrondConfig) error {
	wallet := interactors.NewWallet()
	elrondPrivateKeyBytes, err := wallet.LoadPrivateKeyFromPemFile(elrondConfigs.PrivateKeyFile)
	if err != nil {
		return err
	}

	components.elrondRelayerPrivateKey, err = keyGen.PrivateKeyFromByteArray(elrondPrivateKeyBytes)
	if err != nil {
		return err
	}

	components.elrondRelayerAddress, err = wallet.GetAddressFromPrivateKey(elrondPrivateKeyBytes)
	if err != nil {
		return err
	}

	components.elrondMultisigContractAddress, err = data.NewAddressFromBech32String(elrondConfigs.MultisigContractAddress)
	if err != nil {
		return fmt.Errorf("%w for elrondConfigs.MultisigContractAddress", err)
	}

	return nil
}

func (components *ethElrondBridgeComponents) createDataGetter() error {
	elrondDataGetterLogId := components.evmCompatibleChain.ElrondDataGetterLogId()
	argsDataGetter := elrond.ArgsDataGetter{
		MultisigContractAddress: components.elrondMultisigContractAddress,
		RelayerAddress:          components.elrondRelayerAddress,
		Proxy:                   components.proxy,
		Log:                     core.NewLoggerWithIdentifier(logger.GetOrCreate(elrondDataGetterLogId), elrondDataGetterLogId),
	}

	var err error
	components.dataGetter, err = elrond.NewDataGetter(argsDataGetter)

	return err
}

func (components *ethElrondBridgeComponents) createElrondClient(args ArgsEthereumToElrondBridge) error {
	elrondConfigs := args.Configs.GeneralConfig.Elrond
	tokensMapper, err := mappers.NewElrondToErc20Mapper(components.dataGetter)
	if err != nil {
		return err
	}
	elrondClientLogId := components.evmCompatibleChain.ElrondClientLogId()

	clientArgs := elrond.ClientArgs{
		GasMapConfig:                 elrondConfigs.GasMap,
		Proxy:                        args.Proxy,
		Log:                          core.NewLoggerWithIdentifier(logger.GetOrCreate(elrondClientLogId), elrondClientLogId),
		RelayerPrivateKey:            components.elrondRelayerPrivateKey,
		MultisigContractAddress:      components.elrondMultisigContractAddress,
		IntervalToResendTxsInSeconds: elrondConfigs.IntervalToResendTxsInSeconds,
		TokensMapper:                 tokensMapper,
		RoleProvider:                 components.elrondRoleProvider,
		StatusHandler:                args.ElrondClientStatusHandler,
		AllowDelta:                   uint64(elrondConfigs.ProxyMaxNoncesDelta),
	}

	components.elrondClient, err = elrond.NewClient(clientArgs)
	components.addClosableComponent(components.elrondClient)

	return err
}

func (components *ethElrondBridgeComponents) createEthereumClient(args ArgsEthereumToElrondBridge) error {
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
	ethToElrondName := components.evmCompatibleChain.EvmCompatibleChainToElrondName()
	argsBroadcaster := p2p.ArgsBroadcaster{
		Messenger:           args.Messenger,
		Log:                 core.NewLoggerWithIdentifier(logger.GetOrCreate(broadcasterLogId), broadcasterLogId),
		ElrondRoleProvider:  components.elrondRoleProvider,
		SignatureProcessor:  components.ethereumRoleProvider,
		KeyGen:              keyGen,
		SingleSigner:        singleSigner,
		PrivateKey:          components.elrondRelayerPrivateKey,
		Name:                ethToElrondName,
		AntifloodComponents: antifloodComponents,
	}

	components.broadcaster, err = p2p.NewBroadcaster(argsBroadcaster)
	if err != nil {
		return err
	}

	privateKeyBytes, err := ioutil.ReadFile(ethereumConfigs.PrivateKeyFile)
	if err != nil {
		return err
	}
	privateKeyString := converters.TrimWhiteSpaceCharacters(string(privateKeyBytes))
	privateKey, err := ethCrypto.HexToECDSA(privateKeyString)
	if err != nil {
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return errPublicKeyCast
	}
	components.ethereumRelayerAddress = ethCrypto.PubkeyToAddress(*publicKeyECDSA)

	tokensMapper, err := mappers.NewErc20ToElrondMapper(components.dataGetter)
	if err != nil {
		return err
	}

	signaturesHolder := ethElrond.NewSignatureHolder()
	components.ethToElrondSignaturesHolder = signaturesHolder
	err = components.broadcaster.AddBroadcastClient(signaturesHolder)
	if err != nil {
		return err
	}

	safeContractAddress := common.HexToAddress(ethereumConfigs.SafeContractAddress)

	ethClientLogId := components.evmCompatibleChain.EvmCompatibleChainClientLogId()
	argsEthClient := ethereum.ArgsEthereumClient{
		ClientWrapper:           args.ClientWrapper,
		Erc20ContractsHandler:   args.Erc20ContractsHolder,
		Log:                     core.NewLoggerWithIdentifier(logger.GetOrCreate(ethClientLogId), ethClientLogId),
		AddressConverter:        components.addressConverter,
		Broadcaster:             components.broadcaster,
		PrivateKey:              privateKey,
		TokensMapper:            tokensMapper,
		SignatureHolder:         signaturesHolder,
		SafeContractAddress:     safeContractAddress,
		GasHandler:              gs,
		TransferGasLimitBase:    ethereumConfigs.GasLimitBase,
		TransferGasLimitForEach: ethereumConfigs.GasLimitForEach,
		AllowDelta:              ethereumConfigs.MaxBlocksDelta,
	}

	components.ethClient, err = ethereum.NewEthereumClient(argsEthClient)

	return err
}

func (components *ethElrondBridgeComponents) createElrondRoleProvider(args ArgsEthereumToElrondBridge) error {
	configs := args.Configs.GeneralConfig
	elrondRoleProviderLogId := components.evmCompatibleChain.ElrondRoleProviderLogId()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(elrondRoleProviderLogId), elrondRoleProviderLogId)

	argsRoleProvider := roleProviders.ArgsElrondRoleProvider{
		DataGetter: components.dataGetter,
		Log:        log,
	}

	var err error
	components.elrondRoleProvider, err = roleProviders.NewElrondRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             "Elrond role provider",
		PollingInterval:  time.Duration(configs.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.elrondRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethElrondBridgeComponents) createEthereumRoleProvider(args ArgsEthereumToElrondBridge) error {
	configs := args.Configs.GeneralConfig
	evmCompatibleChain := args.Configs.GeneralConfig.Eth.Chain
	ethRoleProviderLogId := components.evmCompatibleChain.EvmCompatibleChainRoleProviderLogId()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethRoleProviderLogId), ethRoleProviderLogId)
	argsRoleProvider := roleProviders.ArgsEthereumRoleProvider{
		EthereumChainInteractor: args.ClientWrapper,
		Log:                     log,
	}

	var err error
	components.ethereumRoleProvider, err = roleProviders.NewEthereumRoleProvider(argsRoleProvider)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             string(evmCompatibleChain) + " role provider",
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

func (components *ethElrondBridgeComponents) createEthereumToElrondBridge(args ArgsEthereumToElrondBridge) error {
	ethToElrondName := components.evmCompatibleChain.EvmCompatibleChainToElrondName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToElrondName), ethToElrondName)

	configs, found := args.Configs.GeneralConfig.StateMachine[ethToElrondName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, ethToElrondName)
	}

	components.ethToElrondStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond

	argsTopologyHandler := topology.ArgsTopologyHandler{
		PublicKeysProvider: components.elrondRoleProvider,
		Timer:              components.timer,
		IntervalForLeader:  time.Second * time.Duration(configs.IntervalForLeaderInSeconds),
		AddressBytes:       components.elrondRelayerAddress.AddressBytes(),
		Log:                log,
		AddressConverter:   components.addressConverter,
	}

	topologyHandler, err := topology.NewTopologyHandler(argsTopologyHandler)
	if err != nil {
		return err
	}

	components.ethToElrondStatusHandler, err = status.NewStatusHandler(ethToElrondName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.ethToElrondStatusHandler)
	if err != nil {
		return err
	}

	timeForTransferExecution := time.Second * time.Duration(args.Configs.GeneralConfig.Eth.IntervalToWaitForTransferInSeconds)

	batchValidator, err := components.createBatchValidator(args.Configs.GeneralConfig.Eth.Chain, chain.Elrond, args.Configs.GeneralConfig.BatchValidator)
	if err != nil {
		return err
	}

	argsBridgeExecutor := ethElrond.ArgsBridgeExecutor{
		Log:                        log,
		TopologyProvider:           topologyHandler,
		ElrondClient:               components.elrondClient,
		EthereumClient:             components.ethClient,
		StatusHandler:              components.ethToElrondStatusHandler,
		TimeForWaitOnEthereum:      timeForTransferExecution,
		SignaturesHolder:           disabled.NewDisabledSignaturesHolder(),
		BatchValidator:             batchValidator,
		MaxQuorumRetriesOnEthereum: args.Configs.GeneralConfig.Eth.MaxRetriesOnQuorumReached,
		MaxQuorumRetriesOnElrond:   args.Configs.GeneralConfig.Elrond.MaxRetriesOnQuorumReached,
		MaxRestriesOnWasProposed:   args.Configs.GeneralConfig.Elrond.MaxRetriesOnWasTransferProposed,
	}

	bridge, err := ethElrond.NewBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.ethToElrondMachineStates, err = ethToElrondSteps.CreateSteps(bridge)
	if err != nil {
		return err
	}

	return nil
}

func (components *ethElrondBridgeComponents) createElrondToEthereumBridge(args ArgsEthereumToElrondBridge) error {
	elrondToEthName := components.evmCompatibleChain.ElrondToEvmCompatibleChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(elrondToEthName), elrondToEthName)

	configs, found := args.Configs.GeneralConfig.StateMachine[elrondToEthName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, elrondToEthName)
	}

	components.elrondToEthStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond
	argsTopologyHandler := topology.ArgsTopologyHandler{
		PublicKeysProvider: components.elrondRoleProvider,
		Timer:              components.timer,
		IntervalForLeader:  time.Second * time.Duration(configs.IntervalForLeaderInSeconds),
		AddressBytes:       components.elrondRelayerAddress.AddressBytes(),
		Log:                log,
		AddressConverter:   components.addressConverter,
	}

	topologyHandler, err := topology.NewTopologyHandler(argsTopologyHandler)
	if err != nil {
		return err
	}

	components.elrondToEthStatusHandler, err = status.NewStatusHandler(elrondToEthName, components.statusStorer)
	if err != nil {
		return err
	}

	err = components.metricsHolder.AddStatusHandler(components.elrondToEthStatusHandler)
	if err != nil {
		return err
	}

	timeForWaitOnEthereum := time.Second * time.Duration(args.Configs.GeneralConfig.Eth.IntervalToWaitForTransferInSeconds)

	batchValidator, err := components.createBatchValidator(chain.Elrond, args.Configs.GeneralConfig.Eth.Chain, args.Configs.GeneralConfig.BatchValidator)
	if err != nil {
		return err
	}

	argsBridgeExecutor := ethElrond.ArgsBridgeExecutor{
		Log:                        log,
		TopologyProvider:           topologyHandler,
		ElrondClient:               components.elrondClient,
		EthereumClient:             components.ethClient,
		StatusHandler:              components.elrondToEthStatusHandler,
		TimeForWaitOnEthereum:      timeForWaitOnEthereum,
		SignaturesHolder:           components.ethToElrondSignaturesHolder,
		BatchValidator:             batchValidator,
		MaxQuorumRetriesOnEthereum: args.Configs.GeneralConfig.Eth.MaxRetriesOnQuorumReached,
		MaxQuorumRetriesOnElrond:   args.Configs.GeneralConfig.Elrond.MaxRetriesOnQuorumReached,
		MaxRestriesOnWasProposed:   args.Configs.GeneralConfig.Elrond.MaxRetriesOnWasTransferProposed,
	}

	bridge, err := ethElrond.NewBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.elrondToEthMachineStates, err = elrondToEthSteps.CreateSteps(bridge)
	if err != nil {
		return err
	}

	return nil
}

func (components *ethElrondBridgeComponents) startPollingHandlers() error {
	for _, pollingHandler := range components.pollingHandlers {
		err := pollingHandler.StartProcessingLoop()
		if err != nil {
			return err
		}
	}

	return nil
}

// Start will start the bridge
func (components *ethElrondBridgeComponents) Start() error {
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

	go components.startBroadcastJoinRetriesLoop()

	return nil
}

func (components *ethElrondBridgeComponents) createBatchValidator(sourceChain chain.Chain, destinationChain chain.Chain, args config.BatchValidatorConfig) (clients.BatchValidator, error) {
	argsBatchValidator := batchValidatorManagement.ArgsBatchValidator{
		SourceChain:      sourceChain,
		DestinationChain: destinationChain,
		RequestURL:       args.URL,
		RequestTime:      time.Second * time.Duration(args.RequestTimeInSeconds),
	}

	batchValidator, err := batchManagementFactory.CreateBatchValidator(argsBatchValidator, args.Enabled)
	if err != nil {
		return nil, err
	}
	return batchValidator, err
}

func (components *ethElrondBridgeComponents) createEthereumToElrondStateMachine() error {
	ethToElrondName := components.evmCompatibleChain.EvmCompatibleChainToElrondName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToElrondName), ethToElrondName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     ethToElrondName,
		Steps:                components.ethToElrondMachineStates,
		StartStateIdentifier: ethToElrondSteps.GettingPendingBatchFromEthereum,
		Log:                  log,
		StatusHandler:        components.ethToElrondStatusHandler,
	}

	var err error
	components.ethToElrondStateMachine, err = stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             ethToElrondName + " State machine",
		PollingInterval:  components.ethToElrondStepDuration,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.ethToElrondStateMachine,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethElrondBridgeComponents) createElrondToEthereumStateMachine() error {
	elrondToEthName := components.evmCompatibleChain.ElrondToEvmCompatibleChainName()
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(elrondToEthName), elrondToEthName)

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     elrondToEthName,
		Steps:                components.elrondToEthMachineStates,
		StartStateIdentifier: elrondToEthSteps.GettingPendingBatchFromElrond,
		Log:                  log,
		StatusHandler:        components.elrondToEthStatusHandler,
	}

	var err error
	components.elrondToEthStateMachine, err = stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             elrondToEthName + " State machine",
		PollingInterval:  components.elrondToEthStepDuration,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.elrondToEthStateMachine,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)
	components.pollingHandlers = append(components.pollingHandlers, pollingHandler)

	return nil
}

func (components *ethElrondBridgeComponents) createAntifloodComponents(antifloodConfig elrondConfig.AntifloodConfig) (*antifloodFactory.AntiFloodComponents, error) {
	var err error
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancelFunc()
		}
	}()

	cfg := elrondConfig.Config{
		Antiflood: antifloodConfig,
	}
	antiFloodComponents, err := antifloodFactory.NewP2PAntiFloodComponents(ctx, cfg, components.appStatusHandler, components.messenger.ID())
	if err != nil {
		return nil, err
	}
	return antiFloodComponents, nil
}

func (components *ethElrondBridgeComponents) startBroadcastJoinRetriesLoop() {
	broadcastTimer := time.NewTimer(components.timeBeforeRepeatJoin)
	defer broadcastTimer.Stop()

	var ctx context.Context
	ctx, components.cancelFunc = context.WithCancel(context.Background())
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
func (components *ethElrondBridgeComponents) Close() error {
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

// ElrondRelayerAddress returns the Elrond's address associated to this relayer
func (components *ethElrondBridgeComponents) ElrondRelayerAddress() erdgoCore.AddressHandler {
	return components.elrondRelayerAddress
}

// EthereumRelayerAddress returns the Ethereum's address associated to this relayer
func (components *ethElrondBridgeComponents) EthereumRelayerAddress() common.Address {
	return components.ethereumRelayerAddress
}
