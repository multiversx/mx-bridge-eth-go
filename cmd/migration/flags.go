package main

import (
	"path"

	"github.com/multiversx/mx-bridge-eth-go/config"
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
		Value: "*:" + logger.LogInfo.String(),
	}
	configurationFile = cli.StringFlag{
		Name: "config",
		Usage: "The `" + filePathPlaceholder + "` for the main configuration file. This TOML file contain the main " +
			"configurations such as storage setups, epoch duration and so on.",
		Value: "config/config.toml",
	}
	mode = cli.StringFlag{
		Name:  "mode",
		Usage: "This flag specifies the operation mode. Usage: query, sign or execute",
		Value: queryMode,
	}
	migrationJsonFile = cli.StringFlag{
		Name:  "migration-file",
		Usage: "The output .json file containing the migration data",
		Value: path.Join(configPath, "migration-"+timestampPlaceholder+".json"),
	}
	signatureJsonFile = cli.StringFlag{
		Name:  "signature-file",
		Usage: "The output .json file containing the signature data",
		Value: path.Join(configPath, publicKeyPlaceholder+"-"+timestampPlaceholder+".json"),
	}
	newSafeAddress = cli.StringFlag{
		Name:  "new-safe-address",
		Usage: "The new safe address on Ethereum",
		Value: "",
	}
	partialMigration = cli.StringFlag{
		Name: "partial-migration",
		Usage: "If a partial migration is wanted, this option can be very handy. We can specify an unlimited tuples in a single string, like this: " +
			"`-partial-migration token1:amount1,token2:amount2,token1:amount3` and so on. You can see that the same token can be specified multiple times, " +
			"the amounts will be added. The amount should be specified as a denominated value (does not contain all decimals, the conversion will be done " +
			"automatically by the tool). Real example: `-partial-migration token1:amount1,token2:amount2,token1:amount3`",
	}
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		logLevel,
		configurationFile,
		mode,
		migrationJsonFile,
		signatureJsonFile,
		newSafeAddress,
		partialMigration,
	}
}
func getFlagsConfig(ctx *cli.Context) config.ContextFlagsConfig {
	flagsConfig := config.ContextFlagsConfig{}

	flagsConfig.LogLevel = ctx.GlobalString(logLevel.Name)
	flagsConfig.ConfigurationFile = ctx.GlobalString(configurationFile.Name)

	return flagsConfig
}
