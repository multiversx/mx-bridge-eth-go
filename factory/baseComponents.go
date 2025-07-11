package factory

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	ethmultiversx "github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	balanceValidatorManagement "github.com/multiversx/mx-bridge-eth-go/clients/balanceValidator"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	roleproviders "github.com/multiversx/mx-bridge-eth-go/clients/roleProviders"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-bridge-eth-go/core/timer"
	"github.com/multiversx/mx-bridge-eth-go/p2p"
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

type ArgsBridgeCommon struct {
	Configs                       config.Configs
	Messenger                     p2p.NetMessenger
	StatusStorer                  core.Storer
	Proxy                         multiversx.Proxy
	MultiversXClientStatusHandler core.StatusHandler
	TimeForBootstrap              time.Duration
	TimeBeforeRepeatJoin          time.Duration
	MetricsHolder                 core.MetricsHolder
	AppStatusHandler              chainCore.AppStatusHandler
}

type baseBridgeComponents struct {
	baseLogger                        logger.Logger
	messenger                         p2p.NetMessenger
	statusStorer                      core.Storer
	multiversXClient                  ethmultiversx.MultiversXClient
	multiversXMultisigContractAddress sdkCore.AddressHandler
	multiversXSafeContractAddress     sdkCore.AddressHandler
	multiversXRelayerPrivateKey       crypto.PrivateKey
	multiversXRelayerAddress          sdkCore.AddressHandler
	mxDataGetter                      dataGetter
	proxy                             multiversx.Proxy
	multiversXRoleProvider            MultiversXRoleProvider
	broadcaster                       Broadcaster
	timer                             core.Timer
	timeForBootstrap                  time.Duration
	metricsHolder                     core.MetricsHolder
	addressConverter                  core.AddressConverter

	toMultiversXMachineStates    core.MachineStates
	toMultiversXStepDuration     time.Duration
	toMultiversXStatusHandler    core.StatusHandler
	toMultiversXStateMachine     StateMachine
	toMultiversXSignaturesHolder ethmultiversx.SignaturesHolder

	fromMultiversXMachineStates core.MachineStates
	fromMultiversXStepDuration  time.Duration
	fromMultiversXStatusHandler core.StatusHandler
	fromMultiversXStateMachine  StateMachine

	mutClosableHandlers sync.RWMutex
	closableHandlers    []io.Closer

	pollingHandlers []PollingHandler

	timeBeforeRepeatJoin time.Duration
	cancelFunc           func()
	appStatusHandler     chainCore.AppStatusHandler
}

// NewBaseComponents creates a base component holder
func NewBaseComponents(args ArgsBridgeCommon) (*baseBridgeComponents, error) {
	err := checkCommonArgs(args)
	if err != nil {
		return nil, err
	}

	components := &baseBridgeComponents{
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

	return components, nil
}

func checkCommonArgs(args ArgsBridgeCommon) error {
	if check.IfNil(args.Proxy) {
		return errNilProxy
	}
	if check.IfNil(args.Messenger) {
		return errNilMessenger
	}
	if check.IfNil(args.StatusStorer) {
		return errNilStatusStorer
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

func (components *baseBridgeComponents) addClosableComponent(closable io.Closer) {
	components.mutClosableHandlers.Lock()
	components.closableHandlers = append(components.closableHandlers, closable)
	components.mutClosableHandlers.Unlock()
}

func (components *baseBridgeComponents) createMultiversXKeysAndAddresses(chainConfigs config.MultiversXConfig) error {
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

func (components *baseBridgeComponents) createDataGetter(logId string) error {
	argsMXClientDataGetter := multiversx.ArgsMXClientDataGetter{
		MultisigContractAddress: components.multiversXMultisigContractAddress,
		SafeContractAddress:     components.multiversXSafeContractAddress,
		RelayerAddress:          components.multiversXRelayerAddress,
		Proxy:                   components.proxy,
		Log:                     core.NewLoggerWithIdentifier(logger.GetOrCreate(logId), logId),
	}

	var err error
	components.mxDataGetter, err = multiversx.NewMXClientDataGetter(argsMXClientDataGetter)

	return err
}

func (components *baseBridgeComponents) createMultiversXClient(args ArgsBridgeCommon, logId string, tokensMapper multiversx.TokensMapper) error {
	chainConfigs := args.Configs.GeneralConfig.MultiversX

	clientArgs := multiversx.ClientArgs{
		GasMapConfig:                 chainConfigs.GasMap,
		Proxy:                        args.Proxy,
		Log:                          core.NewLoggerWithIdentifier(logger.GetOrCreate(logId), logId),
		RelayerPrivateKey:            components.multiversXRelayerPrivateKey,
		MultisigContractAddress:      components.multiversXMultisigContractAddress,
		SafeContractAddress:          components.multiversXSafeContractAddress,
		IntervalToResendTxsInSeconds: chainConfigs.IntervalToResendTxsInSeconds,
		TokensMapper:                 tokensMapper,
		RoleProvider:                 components.multiversXRoleProvider,
		StatusHandler:                args.MultiversXClientStatusHandler,
		ClientAvailabilityAllowDelta: chainConfigs.ClientAvailabilityAllowDelta,
	}

	var err error
	components.multiversXClient, err = multiversx.NewClient(clientArgs)
	components.addClosableComponent(components.multiversXClient)

	return err
}

func (components *baseBridgeComponents) createMultiversXRoleProvider(args ArgsBridgeCommon, logId string) error {
	configs := args.Configs.GeneralConfig
	log := core.NewLoggerWithIdentifier(logger.GetOrCreate(logId), logId)

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

func (components *baseBridgeComponents) createAntifloodComponents(antifloodConfig chainConfig.AntifloodConfig) (*antifloodFactory.AntiFloodComponents, error) {
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

func (components *baseBridgeComponents) createBalanceValidator(logger logger.Logger, client ethmultiversx.PeerChainClient) (ethmultiversx.BalanceValidator, error) {
	argsBalanceValidator := balanceValidatorManagement.ArgsBalanceValidator{
		Log:              logger,
		MultiversXClient: components.multiversXClient,
		PeerChainClient:  client,
	}

	return balanceValidatorManagement.NewBalanceValidator(argsBalanceValidator)
}

func (components *baseBridgeComponents) startPollingHandlers() error {
	for _, pollingHandler := range components.pollingHandlers {
		err := pollingHandler.StartProcessingLoop()
		if err != nil {
			return err
		}
	}

	return nil
}

func (components *baseBridgeComponents) startBroadcastJoinRetriesLoop(ctx context.Context) {
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

// Start will start the bridge
func (components *baseBridgeComponents) Start() error {
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

// Close will close any sub-components started
func (components *baseBridgeComponents) Close() error {
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
func (components *baseBridgeComponents) MultiversXRelayerAddress() sdkCore.AddressHandler {
	return components.multiversXRelayerAddress
}
