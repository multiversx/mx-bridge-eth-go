package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/executors/multiversx/module"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	chainFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	chainCommon "github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/urfave/cli"
)

const (
	filePathPlaceholder = "[path]"
	defaultLogsPath     = "logs"
	logFilePrefix       = "sc-calls-executor"
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
	app.Name = "SC calls executor CLI app"
	app.Usage = "This is the entry point for the module that periodically tries to execute SC calls"
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
		return startExecutor(c, app.Version)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func startExecutor(ctx *cli.Context, version string) error {
	flagsConfig := getFlagsConfig(ctx)

	fileLogging, errLogger := attachFileLogger(log, flagsConfig)
	if errLogger != nil {
		return errLogger
	}

	log.Info("starting SC calls executor node", "version", version, "pid", os.Getpid())

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	cfg, err := loadConfig(flagsConfig.ConfigurationFile)
	if err != nil {
		return err
	}

	if !check.IfNil(fileLogging) {
		timeLogLifeSpan := time.Second * time.Duration(cfg.Logs.LogFileLifeSpanInSec)
		sizeLogLifeSpanInMB := uint64(cfg.Logs.LogFileLifeSpanInMB)
		err = fileLogging.ChangeFileLifeSpan(timeLogLifeSpan, sizeLogLifeSpanInMB)
		if err != nil {
			return err
		}
	}

	if ctx.IsSet(scProxyBech32Address.Name) {
		cfg.ScProxyBech32Address = ctx.GlobalString(scProxyBech32Address.Name)
		log.Info("using flag-defined SC proxy address", "address", cfg.ScProxyBech32Address)
	}
	if ctx.IsSet(networkAddress.Name) {
		cfg.NetworkAddress = ctx.GlobalString(networkAddress.Name)
		log.Info("using flag-defined network address", "address", cfg.NetworkAddress)
	}
	if ctx.IsSet(privateKeyFile.Name) {
		cfg.PrivateKeyFile = ctx.GlobalString(privateKeyFile.Name)
		log.Info("using flag-defined private key file", "filename", cfg.PrivateKeyFile)
	}

	if len(cfg.NetworkAddress) == 0 {
		return fmt.Errorf("empty NetworkAddress in config file")
	}

	args := config.ScCallsModuleConfig{
		ScProxyBech32Address:         cfg.ScProxyBech32Address,
		ExtraGasToExecute:            cfg.ExtraGasToExecute,
		MaxGasLimitToUse:             cfg.MaxGasLimitToUse,
		NetworkAddress:               cfg.NetworkAddress,
		ProxyMaxNoncesDelta:          cfg.ProxyMaxNoncesDelta,
		ProxyFinalityCheck:           cfg.ProxyFinalityCheck,
		ProxyCacherExpirationSeconds: cfg.ProxyCacherExpirationSeconds,
		ProxyRestAPIEntityType:       cfg.ProxyRestAPIEntityType,
		IntervalToResendTxsInSeconds: cfg.IntervalToResendTxsInSeconds,
		PrivateKeyFile:               cfg.PrivateKeyFile,
		PollingIntervalInMillis:      cfg.PollingIntervalInMillis,
		Filter:                       cfg.Filter,
		Logs:                         cfg.Logs,
		TransactionChecks:            cfg.TransactionChecks,
	}

	chCloseApp := make(chan struct{}, 1)
	scCallsExecutor, err := module.NewScCallsModule(args, log, chCloseApp)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		log.Info("application closing by user error input, calling Close on all subcomponents...")
	case <-chCloseApp:
		log.Info("application closing, requested internally, calling Close on all subcomponents...")
	}

	return scCallsExecutor.Close()
}

func loadConfig(filepath string) (config.ScCallsModuleConfig, error) {
	cfg := config.ScCallsModuleConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.ScCallsModuleConfig{}, err
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
