package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/wrappers"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/factory"
	"github.com/ElrondNetwork/elrond-eth-bridge/p2p"
	"github.com/ElrondNetwork/elrond-eth-bridge/status"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go-core/marshal/factory"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	elrondFactory "github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	elrondCommon "github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/logging"
	elrondConfig "github.com/ElrondNetwork/elrond-go/config"
	elrondP2P "github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
	"github.com/ElrondNetwork/elrond-go/update/disabled"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"
	_ "github.com/urfave/cli"
)

const (
	filePathPlaceholder      = "[path]"
	defaultLogsPath          = "logs"
	logFilePrefix            = "elrond-eth-bridge"
	p2pPeerNetworkDiscoverer = "optimized"
	nilListSharderType       = "NilListSharder"
	dbPath                   = "db"
	timeForBootstrap         = time.Second * 20
	timeBeforeRepeatJoin     = time.Minute * 5
)

var log = logger.GetOrCreate("main")

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//            go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
// windows:
//            for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//            go build -i -v -ldflags="-X main.appVersion=%VERS%"
var appVersion = elrondCommon.UnVersionedAppString

func main() {
	app := cli.NewApp()
	app.Name = "Relay CLI app"
	app.Usage = "This is the entry point for the bridge relay"
	app.Flags = getFlags()
	machineID := elrondCore.GetAnonymizedMachineID(app.Name)
	app.Version = fmt.Sprintf("%s/%s/%s-%s/%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, machineID)
	app.Authors = []cli.Author{
		{
			Name:  "The Agile Freaks team",
			Email: "office@agilefreaks.com",
		},
		{
			Name:  "The Elrond Team",
			Email: "contact@elrond.com",
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
		err = fileLogging.ChangeFileLifeSpan(time.Second * time.Duration(cfg.Logs.LogFileLifeSpanInSec))
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

	if len(cfg.Elrond.NetworkAddress) == 0 {
		return fmt.Errorf("empty Elrond.NetworkAddress in config file")
	}

	proxy := blockchain.NewElrondProxy(cfg.Elrond.NetworkAddress, nil)
	proxyWithCacher, err := blockchain.NewElrondProxyWithCache(proxy, time.Second*time.Duration(cfg.Elrond.ProxyCacherExpirationSeconds))
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

	argsContractsHolder := ethereum.ArgsErc20SafeContractsHolder{
		EthClient:              ethClient,
		EthClientStatusHandler: ethClientStatusHandler,
	}
	erc20ContractsHolder, err := ethereum.NewErc20SafeContractsHolder(argsContractsHolder)
	if err != nil {
		return err
	}

	marshalizer, err := factoryMarshalizer.NewMarshalizer(cfg.Relayer.Marshalizer.Type)
	if err != nil {
		return err
	}

	messenger, err := buildNetMessenger(cfg, marshalizer)
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
		BlockchainClient: ethClient,
	}

	clientWrapper, err := wrappers.NewEthereumChainWrapper(argsClientWrapper)
	if err != nil {
		return err
	}

	args := factory.ArgsEthereumToElrondBridge{
		Configs:              configs,
		Messenger:            messenger,
		StatusStorer:         statusStorer,
		Proxy:                proxyWithCacher,
		Erc20ContractsHolder: erc20ContractsHolder,
		ClientWrapper:        clientWrapper,
		TimeForBootstrap:     timeForBootstrap,
		TimeBeforeRepeatJoin: timeBeforeRepeatJoin,
		MetricsHolder:        metricsHolder,
	}
	ethToElrondComponents, err := factory.NewEthElrondBridgeComponents(args)
	if err != nil {
		return err
	}

	webServer, err := factory.StartWebServer(configs, metricsHolder)
	if err != nil {
		return err
	}

	log.Info("Starting relay")

	err = ethToElrondComponents.Start()
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("application closing, calling Close on all subcomponents...")

	var lastErr error
	err = ethToElrondComponents.Close()
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
	err := elrondCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

// LoadApiConfig returns a ApiRoutesConfig by reading the config file provided
func loadApiConfig(filepath string) (config.ApiRoutesConfig, error) {
	cfg := config.ApiRoutesConfig{}
	err := elrondCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.ApiRoutesConfig{}, err
	}

	return cfg, nil
}

func attachFileLogger(log logger.Logger, flagsConfig config.ContextFlagsConfig) (elrondFactory.FileLoggingHandler, error) {
	var fileLogging elrondFactory.FileLoggingHandler
	var err error
	if flagsConfig.SaveLogFile {
		fileLogging, err = logging.NewFileLogging(flagsConfig.WorkingDir, defaultLogsPath, logFilePrefix)
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

func buildNetMessenger(cfg config.Config, marshalizer marshal.Marshalizer) (p2p.NetMessenger, error) {
	nodeConfig := elrondConfig.NodeConfig{
		Port:                       cfg.P2P.Port,
		Seed:                       cfg.P2P.Seed,
		MaximumExpectedPeerCount:   0,
		ThresholdMinConnectedPeers: 0,
	}
	peerDiscoveryConfig := elrondConfig.KadDhtPeerDiscoveryConfig{
		Enabled:                          true,
		RefreshIntervalInSec:             5,
		ProtocolID:                       cfg.P2P.ProtocolID,
		InitialPeerList:                  cfg.P2P.InitialPeerList,
		BucketSize:                       0,
		RoutingTableRefreshIntervalInSec: 300,
		Type:                             p2pPeerNetworkDiscoverer,
	}

	p2pConfig := elrondConfig.P2PConfig{
		Node:                nodeConfig,
		KadDhtPeerDiscovery: peerDiscoveryConfig,
		Sharding: elrondConfig.ShardingConfig{
			TargetPeerCount:         0,
			MaxIntraShardValidators: 0,
			MaxCrossShardValidators: 0,
			MaxIntraShardObservers:  0,
			MaxCrossShardObservers:  0,
			Type:                    nilListSharderType,
		},
	}

	args := libp2p.ArgsNetworkMessenger{
		Marshalizer:          marshalizer,
		ListenAddress:        libp2p.ListenAddrWithIp4AndTcp,
		P2pConfig:            p2pConfig,
		SyncTimer:            &libp2p.LocalSyncTimer{},
		PreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		NodeOperationMode:    elrondP2P.NormalOperation,
	}

	messenger, err := libp2p.NewNetworkMessenger(args)
	if err != nil {
		panic(err)
	}

	return messenger, nil
}
