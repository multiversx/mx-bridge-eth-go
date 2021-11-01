package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/cmd/node/factory"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/common/logging"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"
	_ "github.com/urfave/cli"
)

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2

	filePathPlaceholder = "[path]"

	defaultLogsPath = "logs"
	logFilePrefix   = "elrond-eth-bridge"
)

var log = logger.GetOrCreate("main")

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//            go build -i -v -ldflags="-X main.appVersion=$(git describe --tags --long --dirty)"
// windows:
//            for /f %i in ('git describe --tags --long --dirty') do set VERS=%i
//            go build -i -v -ldflags="-X main.appVersion=%VERS%"
var appVersion = common.UnVersionedAppString

func main() {
	app := cli.NewApp()
	app.Name = "Relay CLI app"
	app.Usage = "This is the entry point for the bridge relay"
	app.Flags = getFlags()
	machineID := core.GetAnonymizedMachineID(app.Name)
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

	config, err := loadConfig(flagsConfig.ConfigurationFile)
	if err != nil {
		return err
	}

	if !check.IfNil(fileLogging) {
		err := fileLogging.ChangeFileLifeSpan(time.Second * time.Duration(config.Logs.LogFileLifeSpanInSec))
		if err != nil {
			return err
		}
	}

	proxy := blockchain.NewElrondProxy(config.Elrond.NetworkAddress, nil)

	ethClient, err := ethclient.Dial(config.Eth.NetworkAddress)
	if err != nil {
		return err
	}

	ethInstance, err := contract.NewBridge(ethCommon.HexToAddress(config.Eth.BridgeAddress), ethClient)
	if err != nil {
		return err
	}

	args := relay.ArgsRelayer{
		Config:      *config,
		FlagsConfig: *flagsConfig,
		Name:        "EthToElrRelay",
		Proxy:       proxy,
		EthClient:   ethClient,
		EthInstance: ethInstance,
	}
	ethToElrRelay, err := relay.NewRelay(args)
	if err != nil {
		return err
	}

	mainLoop(ethToElrRelay)

	return nil
}

func mainLoop(r *relay.Relay) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Starting relay")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		signal.Stop(sigs)
		cancel()
	}()

	go func() {
		select {
		case <-sigs:
			cancel()
		case <-ctx.Done():
		}
		<-sigs
		os.Exit(exitCodeInterrupt)
	}()

	if err := r.Start(ctx); err != nil {
		log.Error(err.Error())
		os.Exit(exitCodeErr)
	}
}

func loadConfig(filepath string) (*relay.Config, error) {
	cfg := &relay.Config{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func attachFileLogger(log logger.Logger, flagsConfig *relay.ContextFlagsConfig) (factory.FileLoggingHandler, error) {
	var fileLogging factory.FileLoggingHandler
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
