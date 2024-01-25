package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/wrappers"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/factory"
	"github.com/multiversx/mx-bridge-eth-go/p2p"
	"github.com/multiversx/mx-bridge-eth-go/status"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/marshal"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1"
	"github.com/multiversx/mx-chain-crypto-go/signing/secp256k1/singlesig"
	chainFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	chainCommon "github.com/multiversx/mx-chain-go/common"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	p2pFactory "github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/statusHandler/persister"
	"github.com/multiversx/mx-chain-go/storage/cache"
	"github.com/multiversx/mx-chain-go/update/disabled"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/urfave/cli"
)

const (
	filePathPlaceholder      = "[path]"
	defaultLogsPath          = "logs"
	logFilePrefix            = "multiversx-eth-bridge"
	p2pPeerNetworkDiscoverer = "optimized"
	nilListSharderType       = "NilListSharder"
	disabledWatcher          = "disabled"
	dbPath                   = "db"
	timeForBootstrap         = time.Second * 20
	timeBeforeRepeatJoin     = time.Minute * 5
)

var log = logger.GetOrCreate("main")

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//
//	go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
//
// windows:
//
//	for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//	go build -i -v -ldflags="-X main.appVersion=%VERS%"
var appVersion = chainCommon.UnVersionedAppString

func main() {
	app := cli.NewApp()
	app.Name = "Relay CLI app"
	app.Usage = "This is the entry point for the bridge relay"
	app.Flags = getFlags()
	machineID := chainCore.GetAnonymizedMachineID(app.Name)
	app.Version = fmt.Sprintf("%s/%s/%s-%s/%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, machineID)
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return startRelay(c, app.Version)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startRelay(ctx *cli.Context, version string) error {
	flagsConfig := getFlagsConfig(ctx)

	fileLogging, errLogger := attachFileLogger(log, flagsConfig)
	if errLogger != nil {
		return errLogger
	}

	log.Info("starting bridge node", "version", version, "pid", os.Getpid())

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	cfg, err := loadConfig(flagsConfig.ConfigurationFile)
	if err != nil {
		return err
	}

	apiRoutesConfig, err := loadApiConfig(flagsConfig.ConfigurationApiFile)
	if err != nil {
		return err
	}
	log.Debug("config", "file", flagsConfig.ConfigurationApiFile)

	if !check.IfNil(fileLogging) {
		timeLogLifeSpan := time.Second * time.Duration(cfg.Logs.LogFileLifeSpanInSec)
		sizeLogLifeSpanInMB := uint64(cfg.Logs.LogFileLifeSpanInMB)
		err = fileLogging.ChangeFileLifeSpan(timeLogLifeSpan, sizeLogLifeSpanInMB)
		if err != nil {
			return err
		}
	}

	dbFullPath := path.Join(flagsConfig.WorkingDir, dbPath)
	statusStorer, err := factory.CreateUnitStorer(cfg.Relayer.StatusMetricsStorage, dbFullPath)
	if err != nil {
		return err
	}

	metricsHolder := status.NewMetricsHolder()
	ethClientStatusHandler, err := status.NewStatusHandler(core.EthClientStatusHandlerName, statusStorer)
	if err != nil {
		return err
	}
	err = metricsHolder.AddStatusHandler(ethClientStatusHandler)
	if err != nil {
		return err
	}

	multiversXClientStatusHandler, err := status.NewStatusHandler(core.MultiversXClientStatusHandlerName, statusStorer)
	if err != nil {
		return err
	}
	err = metricsHolder.AddStatusHandler(multiversXClientStatusHandler)
	if err != nil {
		return err
	}

	if len(cfg.MultiversX.NetworkAddress) == 0 {
		return fmt.Errorf("empty MultiversX.NetworkAddress in config file")
	}

	argsProxy := blockchain.ArgsProxy{
		ProxyURL:            cfg.MultiversX.NetworkAddress,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       cfg.MultiversX.ProxyFinalityCheck,
		AllowedDeltaToFinal: cfg.MultiversX.ProxyMaxNoncesDelta,
		CacheExpirationTime: time.Second * time.Duration(cfg.MultiversX.ProxyCacherExpirationSeconds),
		EntityType:          sdkCore.RestAPIEntityType(cfg.MultiversX.ProxyRestAPIEntityType),
	}
	proxy, err := blockchain.NewProxy(argsProxy)
	if err != nil {
		return err
	}

	ethClient, err := ethclient.Dial(cfg.Eth.NetworkAddress)
	if err != nil {
		return err
	}

	bridgeEthAddress := ethCommon.HexToAddress(cfg.Eth.MultisigContractAddress)
	multiSigInstance, err := contract.NewBridge(bridgeEthAddress, ethClient)
	if err != nil {
		return err
	}

	safeEthAddress := ethCommon.HexToAddress(cfg.Eth.SafeContractAddress)
	safeInstance, err := contract.NewContract(safeEthAddress, ethClient)
	if err != nil {
		return err
	}

	argsContractsHolder := ethereum.ArgsErc20SafeContractsHolder{
		EthClient:              ethClient,
		EthClientStatusHandler: ethClientStatusHandler,
	}
	erc20ContractsHolder, err := ethereum.NewErc20SafeContractsHolder(argsContractsHolder)
	if err != nil {
		return err
	}

	marshaller, err := factoryMarshaller.NewMarshalizer(cfg.Relayer.Marshalizer.Type)
	if err != nil {
		return err
	}

	messenger, err := buildNetMessenger(cfg, marshaller, log)
	if err != nil {
		return err
	}

	configs := config.Configs{
		GeneralConfig:   cfg,
		ApiRoutesConfig: apiRoutesConfig,
		FlagsConfig:     flagsConfig,
	}

	argsClientWrapper := wrappers.ArgsEthereumChainWrapper{
		StatusHandler:    ethClientStatusHandler,
		MultiSigContract: multiSigInstance,
		SafeContract:     safeInstance,
		BlockchainClient: ethClient,
	}

	clientWrapper, err := wrappers.NewEthereumChainWrapper(argsClientWrapper)
	if err != nil {
		return err
	}

	var appStatusHandlers []chainCore.AppStatusHandler
	statusMetrics := statusHandler.NewStatusMetrics()
	appStatusHandlers = append(appStatusHandlers, statusMetrics)

	persistentHandler, err := persister.NewPersistentStatusHandler(marshaller, uint64ByteSlice.NewBigEndianConverter())
	if err != nil {
		return err
	}
	appStatusHandlers = append(appStatusHandlers, persistentHandler)
	appStatusHandler, err := statusHandler.NewAppStatusFacadeWithHandlers(appStatusHandlers...)
	if err != nil {
		return err
	}

	args := factory.ArgsEthereumToMultiversXBridge{
		Configs:                       configs,
		Messenger:                     messenger,
		StatusStorer:                  statusStorer,
		Proxy:                         proxy,
		Erc20ContractsHolder:          erc20ContractsHolder,
		ClientWrapper:                 clientWrapper,
		TimeForBootstrap:              timeForBootstrap,
		TimeBeforeRepeatJoin:          timeBeforeRepeatJoin,
		MetricsHolder:                 metricsHolder,
		AppStatusHandler:              appStatusHandler,
		MultiversXClientStatusHandler: multiversXClientStatusHandler,
	}

	ethToMultiversXComponents, err := factory.NewEthMultiversXBridgeComponents(args)
	if err != nil {
		return err
	}

	webServer, err := factory.StartWebServer(configs, metricsHolder)
	if err != nil {
		return err
	}

	log.Info("Starting relay")

	err = ethToMultiversXComponents.Start()
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("application closing, calling Close on all subcomponents...")

	var lastErr error
	err = ethToMultiversXComponents.Close()
	if err != nil {
		lastErr = err
	}

	err = webServer.Close()
	if err != nil {
		lastErr = err
	}

	return lastErr
}

func loadConfig(filepath string) (config.Config, error) {
	cfg := config.Config{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

// LoadApiConfig returns a ApiRoutesConfig by reading the config file provided
func loadApiConfig(filepath string) (config.ApiRoutesConfig, error) {
	cfg := config.ApiRoutesConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.ApiRoutesConfig{}, err
	}

	return cfg, nil
}

func attachFileLogger(log logger.Logger, flagsConfig config.ContextFlagsConfig) (chainFactory.FileLoggingHandler, error) {
	var fileLogging chainFactory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		argsFileLogging := file.ArgsFileLogging{
			WorkingDir:      flagsConfig.WorkingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		}
		fileLogging, err = file.NewFileLogging(argsFileLogging)
		if err != nil {
			return nil, fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)
	logger.ToggleLoggerName(flagsConfig.EnableLogName)
	logLevelFlagValue := flagsConfig.LogLevel
	err = logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return nil, err
	}

	if flagsConfig.DisableAnsiColor {
		err = logger.RemoveLogObserver(os.Stdout)
		if err != nil {
			return nil, err
		}

		err = logger.AddLogObserver(os.Stdout, &logger.PlainFormatter{})
		if err != nil {
			return nil, err
		}
	}
	log.Trace("logger updated", "level", logLevelFlagValue, "disable ANSI color", flagsConfig.DisableAnsiColor)

	return fileLogging, nil
}

func buildNetMessenger(cfg config.Config, marshalizer marshal.Marshalizer, log logger.Logger) (p2p.NetMessenger, error) {
	nodeConfig := p2pConfig.NodeConfig{
		Port:                       cfg.P2P.Port,
		MaximumExpectedPeerCount:   0,
		ThresholdMinConnectedPeers: 0,
		Transports:                 cfg.P2P.Transports,
		ResourceLimiter:            cfg.P2P.ResourceLimiter,
	}
	peerDiscoveryConfig := p2pConfig.KadDhtPeerDiscoveryConfig{
		Enabled:                          true,
		RefreshIntervalInSec:             5,
		ProtocolID:                       cfg.P2P.ProtocolID,
		InitialPeerList:                  cfg.P2P.InitialPeerList,
		BucketSize:                       0,
		RoutingTableRefreshIntervalInSec: 300,
		Type:                             p2pPeerNetworkDiscoverer,
	}

	p2pCfg := p2pConfig.P2PConfig{
		Node:                nodeConfig,
		KadDhtPeerDiscovery: peerDiscoveryConfig,
		Sharding: p2pConfig.ShardingConfig{
			TargetPeerCount:         0,
			MaxIntraShardValidators: 0,
			MaxCrossShardValidators: 0,
			MaxIntraShardObservers:  0,
			MaxCrossShardObservers:  0,
			Type:                    nilListSharderType,
		},
	}

	topRatedCache, err := cache.NewLRUCache(cfg.PeersRatingConfig.TopRatedCacheCapacity)
	if err != nil {
		return nil, err
	}
	badRatedCache, err := cache.NewLRUCache(cfg.PeersRatingConfig.BadRatedCacheCapacity)
	if err != nil {
		return nil, err
	}
	argsPeersRatingHandler := p2pFactory.ArgPeersRatingHandler{
		TopRatedCache: topRatedCache,
		BadRatedCache: badRatedCache,
		Logger:        log,
	}
	peersRatingHandler, err := p2pFactory.NewPeersRatingHandler(argsPeersRatingHandler)
	if err != nil {
		return nil, err
	}

	p2pSingleSigner := &singlesig.Secp256k1Signer{}
	p2pKeyGen := signing.NewKeyGenerator(secp256k1.NewSecp256k1())
	p2pPrivKey, _ := p2pKeyGen.GeneratePair()

	args := libp2p.ArgsNetworkMessenger{
		Marshaller:            marshalizer,
		P2pConfig:             p2pCfg,
		SyncTimer:             &libp2p.LocalSyncTimer{},
		PreferredPeersHolder:  disabled.NewPreferredPeersHolder(),
		PeersRatingHandler:    peersRatingHandler,
		ConnectionWatcherType: disabledWatcher,
		P2pPrivateKey:         p2pPrivKey,
		P2pSingleSigner:       p2pSingleSigner,
		P2pKeyGenerator:       p2pKeyGen,
		Logger:                log,
	}

	return libp2p.NewNetworkMessenger(args)
}
