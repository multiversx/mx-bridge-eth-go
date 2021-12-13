package factory

import (
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement/factory"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/elrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/elrond/mappers"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/roleProviders"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/core/polling"
	"github.com/ElrondNetwork/elrond-eth-bridge/core/timer"
	v2 "github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond/steps"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/topology"
	"github.com/ElrondNetwork/elrond-eth-bridge/p2p"
	"github.com/ElrondNetwork/elrond-eth-bridge/stateMachine"
	"github.com/ElrondNetwork/elrond-eth-bridge/status"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	pollingDurationOnError  = time.Second * 5
	ethToElrondName         = "EthToElrond"
	baseLogId               = "EthElrond-Base"
	elrondClientLogId       = "EthElrond-ElrondClient"
	ethClientLogId          = "EthElrond-EthClient"
	elrondRoleProviderLogId = "EthElrond-ElrondRoleProvider"
	ethRoleProviderLogId    = "EthElrond-EthRoleProvider"
)

var suite = ed25519.NewEd25519()
var keyGen = signing.NewKeyGenerator(suite)
var singleSigner = &singlesig.Ed25519Signer{}

// ArgsEthereumToElrondBridge is the arguments DTO used for creating an Ethereum to Elrond bridge
type ArgsEthereumToElrondBridge struct {
	Configs              config.Configs
	Messenger            p2p.NetMessenger
	StatusStorer         core.Storer
	Proxy                elrond.ElrondProxy
	Erc20ContractsHolder ethereum.Erc20ContractsHolder
	ClientWrapper        ethereum.ClientWrapper
}

type ethElrondBridgeComponents struct {
	configs                       config.Configs
	baseLogger                    logger.Logger
	messenger                     p2p.NetMessenger
	statusStorer                  core.Storer
	elrondClient                  v2.ElrondClient
	ethClient                     v2.EthereumClient
	elrondMultisigContractAddress erdgoCore.AddressHandler
	elrondRelayerPrivateKey       crypto.PrivateKey
	elrondRelayerAddress          erdgoCore.AddressHandler
	dataGetter                    dataGetter
	proxy                         elrond.ElrondProxy
	elrondRoleProvider            ElrondRoleProvider
	ethereumRoleProvider          EthereumRoleProvider
	broadcaster                   Broadcaster
	timer                         core.Timer

	ethToElrondBridge        ethToElrond.EthToElrondBridge
	ethToElrondMachineStates core.MachineStates
	ethToElrondStepDuration  time.Duration

	mutClosableHandlers sync.RWMutex
	closableHandlers    []io.Closer
}

// NewEthElrondBridgeComponents creates a new eth-elrond bridge components holder
func NewEthElrondBridgeComponents(args ArgsEthereumToElrondBridge) (*ethElrondBridgeComponents, error) {
	err := checkArgsEthereumToElrondBridge(args)
	if err != nil {
		return nil, err
	}

	components := &ethElrondBridgeComponents{
		baseLogger:       core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToElrondName), baseLogId),
		messenger:        args.Messenger,
		statusStorer:     args.StatusStorer,
		configs:          args.Configs,
		closableHandlers: make([]io.Closer, 0),
		proxy:            args.Proxy,
		timer:            timer.NewNTPTimer(),
	}
	components.addClosableComponent(components.timer)

	err = components.createElrondKeysAndAddresses(args.Configs.GeneralConfig.Elrond)
	if err != nil {
		return nil, err
	}

	err = components.createDataGetter()
	if err != nil {
		return nil, err
	}

	err = components.createElrondClient(args)
	if err != nil {
		return nil, err
	}

	err = components.createElrondRoleProvider(args)
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
	args := elrond.ArgsDataGetter{
		MultisigContractAddress: components.elrondMultisigContractAddress,
		RelayerAddress:          components.elrondRelayerAddress,
		Proxy:                   components.proxy,
	}

	var err error
	components.dataGetter, err = elrond.NewDataGetter(args)

	return err
}

func (components *ethElrondBridgeComponents) createElrondClient(args ArgsEthereumToElrondBridge) error {
	elrondConfigs := args.Configs.GeneralConfig.Elrond
	tokensMapper, err := mappers.NewElrondToErc20Mapper(components.dataGetter)
	if err != nil {
		return err
	}

	clientArgs := elrond.ClientArgs{
		GasMapConfig:                 elrondConfigs.GasMap,
		Proxy:                        args.Proxy,
		Log:                          core.NewLoggerWithIdentifier(logger.GetOrCreate(elrondClientLogId), elrondClientLogId),
		RelayerPrivateKey:            components.elrondRelayerPrivateKey,
		MultisigContractAddress:      components.elrondMultisigContractAddress,
		IntervalToResendTxsInSeconds: elrondConfigs.IntervalToResendTxsInSeconds,
		TokensMapper:                 tokensMapper,
		MaxRetriesOnQuorumReached:    elrondConfigs.MaxRetriesOnQuorumReached,
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
		RequestTime:            time.Duration(gasStationConfig.RequestTimeInSeconds) * time.Second,
		MaximumGasPrice:        gasStationConfig.MaximumAllowedGasPrice,
		GasPriceSelector:       core.EthGasPriceSelector(gasStationConfig.GasPriceSelector),
	}

	gs, err := factory.CreateGasStation(argsGasStation, gasStationConfig.Enabled)
	if err != nil {
		return err
	}

	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethClientLogId), ethClientLogId)

	argsBroadcaster := p2p.ArgsBroadcaster{
		Messenger:          args.Messenger,
		Log:                log,
		ElrondRoleProvider: components.elrondRoleProvider,
		SignatureProcessor: components.ethereumRoleProvider,
		KeyGen:             keyGen,
		SingleSigner:       singleSigner,
		PrivateKey:         components.elrondRelayerPrivateKey,
		Name:               ethToElrondName,
	}

	components.broadcaster, err = p2p.NewBroadcaster(argsBroadcaster)
	if err != nil {
		return err
	}

	privateKeyBytes, err := ioutil.ReadFile(ethereumConfigs.PrivateKeyFile)
	if err != nil {
		return err
	}
	privateKeyString := core.TrimWhiteSpaceCharacters(string(privateKeyBytes))
	privateKey, err := ethCrypto.HexToECDSA(privateKeyString)
	if err != nil {
		return err
	}

	tokensMapper, err := mappers.NewErc20ToElrondMapper(components.dataGetter)
	if err != nil {
		return err
	}

	safeContractAddress := common.HexToAddress(ethereumConfigs.SafeContractAddress)
	argsEthClient := ethereum.ArgsEthereumClient{
		ClientWrapper:             args.ClientWrapper,
		Erc20ContractsHandler:     args.Erc20ContractsHolder,
		Log:                       log,
		AddressConverter:          core.NewAddressConverter(),
		Broadcaster:               components.broadcaster,
		PrivateKey:                privateKey,
		TokensMapper:              tokensMapper,
		SignatureHolder:           &disabledSignatureHolder{}, //TODO replace this with the real component
		SafeContractAddress:       safeContractAddress,
		GasHandler:                gs,
		TransferGasLimit:          ethereumConfigs.GasLimit,
		MaxRetriesOnQuorumReached: ethereumConfigs.MaxRetriesOnQuorumReached,
	}

	components.ethClient, err = ethereum.NewEthereumClient(argsEthClient)

	return err
}

func (components *ethElrondBridgeComponents) createElrondRoleProvider(args ArgsEthereumToElrondBridge) error {
	configs := args.Configs.GeneralConfig
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

	return pollingHandler.StartProcessingLoop()
}

func (components *ethElrondBridgeComponents) createEthereumRoleProvider(args ArgsEthereumToElrondBridge) error {
	configs := args.Configs.GeneralConfig

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
		Name:             "Ethereum role provider",
		PollingInterval:  time.Duration(configs.Relayer.RoleProvider.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: pollingDurationOnError,
		Executor:         components.ethereumRoleProvider,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	components.addClosableComponent(pollingHandler)

	return pollingHandler.StartProcessingLoop()
}

func (components *ethElrondBridgeComponents) createEthereumToElrondBridge(args ArgsEthereumToElrondBridge) error {
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToElrondName), ethToElrondName)

	configs, found := args.Configs.GeneralConfig.StateMachine[ethToElrondName]
	if !found {
		return fmt.Errorf("%w for %q", errMissingConfig, ethToElrondName)
	}

	components.ethToElrondStepDuration = time.Duration(configs.StepDurationInMillis) * time.Millisecond

	argsTopologyHandler := topology.ArgsTopologyHandler{
		PublicKeysProvider: components.broadcaster,
		Timer:              components.timer,
		StepDuration:       components.ethToElrondStepDuration,
		AddressBytes:       components.elrondRelayerAddress.AddressBytes(),
	}

	topologyHandler, err := topology.NewTopologyHandler(argsTopologyHandler)
	if err != nil {
		return err
	}

	argsBridgeExecutor := v2.ArgsEthToElrondBridgeExecutor{
		Log:              log,
		TopologyProvider: topologyHandler,
		ElrondClient:     components.elrondClient,
		EthereumClient:   components.ethClient,
	}

	components.ethToElrondBridge, err = v2.NewEthToElrondBridgeExecutor(argsBridgeExecutor)
	if err != nil {
		return err
	}

	components.ethToElrondMachineStates, err = steps.CreateSteps(components.ethToElrondBridge)
	if err != nil {
		return err
	}

	return nil
}

// Start will start the bridge
func (components *ethElrondBridgeComponents) Start() error {
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(ethToElrondName), ethToElrondName)

	//TODO replace this with the real status handler
	ethToElrondStatusHandler, _ := status.NewStatusHandler("dummy", testsCommon.NewStorerMock())

	argsStateMachine := stateMachine.ArgsStateMachine{
		StateMachineName:     ethToElrondName,
		Steps:                components.ethToElrondMachineStates,
		StartStateIdentifier: ethToElrond.GettingPendingBatchFromEthereum,
		DurationBetweenSteps: components.ethToElrondStepDuration,
		Log:                  log,
		StatusHandler:        ethToElrondStatusHandler,
	}

	sm, err := stateMachine.NewStateMachine(argsStateMachine)
	if err != nil {
		return err
	}

	components.addClosableComponent(sm)

	return nil
}

// Close will close any sub-components started
func (components *ethElrondBridgeComponents) Close() error {
	components.mutClosableHandlers.RLock()
	defer components.mutClosableHandlers.RUnlock()

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
