package main

import (
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-chain-go/facade"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/urfave/cli"
)

var (
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogDebug.String(),
	}
	// configurationFile defines a flag for the path to the main toml configuration file
	configurationFile = cli.StringFlag{
		Name: "config",
		Usage: "The `" + filePathPlaceholder + "` for the main configuration file. This TOML file contain the main " +
			"configurations such as monitored SC, gateway URL, timings and so on",
		Value: "config/config.toml",
	}
	// logFile is used when the log output needs to be logged in a file
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
	// profileMode defines a flag for profiling the binary
	// If enabled, it will open the pprof routes over the default gin rest webserver.
	// There are several routes that will be available for profiling (profiling can be analyzed with: go tool pprof):
	//  /debug/pprof/ (can be accessed in the browser, will list the available options)
	//  /debug/pprof/goroutine
	//  /debug/pprof/heap
	//  /debug/pprof/threadcreate
	//  /debug/pprof/block
	//  /debug/pprof/mutex
	//  /debug/pprof/profile (CPU profile)
	//  /debug/pprof/trace?seconds=5 (CPU trace) -> being a trace, can be analyzed with: go tool trace
	// Usage: go tool pprof http(s)://ip.of.the.server/debug/pprof/xxxxx
	profileMode = cli.BoolFlag{
		Name: "profile-mode",
		Usage: "Boolean option for enabling the profiling mode. If set, the /debug/pprof routes will be available " +
			"on the node for profiling the application.",
	}
	// restApiInterface defines a flag for the interface on which the rest API will try to bind with
	restApiInterface = cli.StringFlag{
		Name: "rest-api-interface",
		Usage: "The interface `address and port` to which the REST API will attempt to bind. " +
			"To bind to all available interfaces, set this flag to :8080",
		Value: facade.DefaultRestInterface,
	}
	// workingDirectory defines a flag for the path for the working directory.
	workingDirectory = cli.StringFlag{
		Name:  "working-directory",
		Usage: "This flag specifies the `directory` where the node will store databases and logs.",
		Value: "",
	}
	// disableAnsiColor defines if the logger subsystem should prevent displaying ANSI colors
	disableAnsiColor = cli.BoolFlag{
		Name:  "disable-ansi-color",
		Usage: "Boolean option for disabling ANSI colors in the logging system.",
	}
	// logWithLoggerName is used to enable log correlation elements
	logWithLoggerName = cli.BoolFlag{
		Name:  "log-logger-name",
		Usage: "Boolean option for logger name in the logs.",
	}
	// networkAddress is used to specify the network address used
	networkAddress = cli.StringFlag{
		Name:  "network-address",
		Usage: "The network address (gateway) to be used. Example: 'https://testnet-explorer.multiversx.com'",
	}
	// privateKeyFile is the MultiversX private key file used to issue transaction for the SC calls
	privateKeyFile = cli.StringFlag{
		Name:  "private-key-file",
		Usage: "The MultiversX private key file used to issue transaction for the SC calls",
	}
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		workingDirectory,
		logLevel,
		disableAnsiColor,
		configurationFile,
		logSaveFile,
		logWithLoggerName,
		profileMode,
		restApiInterface,
		networkAddress,
		privateKeyFile,
	}
}
func getFlagsConfig(ctx *cli.Context) config.ContextFlagsConfig {
	flagsConfig := config.ContextFlagsConfig{}

	flagsConfig.WorkingDir = ctx.GlobalString(workingDirectory.Name)
	flagsConfig.LogLevel = ctx.GlobalString(logLevel.Name)
	flagsConfig.DisableAnsiColor = ctx.GlobalBool(disableAnsiColor.Name)
	flagsConfig.ConfigurationFile = ctx.GlobalString(configurationFile.Name)
	flagsConfig.SaveLogFile = ctx.GlobalBool(logSaveFile.Name)
	flagsConfig.EnableLogName = ctx.GlobalBool(logWithLoggerName.Name)
	flagsConfig.EnablePprof = ctx.GlobalBool(profileMode.Name)
	flagsConfig.RestApiInterface = ctx.GlobalString(restApiInterface.Name)

	return flagsConfig
}
