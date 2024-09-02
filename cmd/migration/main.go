package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	ethereumClient "github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx/mappers"
	"github.com/multiversx/mx-bridge-eth-go/cmd/migration/disabled"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/executors/ethereum"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/urfave/cli"
)

const (
	filePathPlaceholder = "[path]"
	generateMode        = "generate"
	signMode            = "sign"
	executeMode         = "execute"
)

var log = logger.GetOrCreate("main")

func main() {
	app := cli.NewApp()
	app.Name = "Funds migration CLI tool"
	app.Usage = "This is the entry point for the migration CLI tool"
	app.Flags = getFlags()
	app.Authors = []cli.Author{
		{
			Name:  "The MultiversX Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		return execute(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func execute(ctx *cli.Context) error {
	flagsConfig := getFlagsConfig(ctx)

	err := logger.SetLogLevel(flagsConfig.LogLevel)
	if err != nil {
		return err
	}

	cfg, err := loadConfig(flagsConfig.ConfigurationFile)
	if err != nil {
		return err
	}

	log.Info("starting migration help tool", "pid", os.Getpid())

	operationMode := strings.ToLower(ctx.GlobalString(mode.Name))
	switch operationMode {
	case generateMode:
		return generate(cfg)
	case signMode:
		//TODO: implement
	case executeMode:
		//TODO: implement
	}

	return fmt.Errorf("unknown execution mode: %s", operationMode)
}

func generate(cfg config.MigrationToolConfig) error {
	argsProxy := blockchain.ArgsProxy{
		ProxyURL:            cfg.MultiversX.NetworkAddress,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       cfg.MultiversX.Proxy.FinalityCheck,
		AllowedDeltaToFinal: cfg.MultiversX.Proxy.MaxNoncesDelta,
		CacheExpirationTime: time.Second * time.Duration(cfg.MultiversX.Proxy.CacherExpirationSeconds),
		EntityType:          sdkCore.RestAPIEntityType(cfg.MultiversX.Proxy.RestAPIEntityType),
	}
	proxy, err := blockchain.NewProxy(argsProxy)
	if err != nil {
		return err
	}

	emptyAddress := data.NewAddressFromBytes(make([]byte, 0))
	safeAddress, err := data.NewAddressFromBech32String(cfg.MultiversX.SafeContractAddress)
	if err != nil {
		return err
	}

	argsMXClientDataGetter := multiversx.ArgsMXClientDataGetter{
		MultisigContractAddress: emptyAddress,
		SafeContractAddress:     safeAddress,
		RelayerAddress:          emptyAddress,
		Proxy:                   proxy,
		Log:                     log,
	}
	mxDataGetter, err := multiversx.NewMXClientDataGetter(argsMXClientDataGetter)
	if err != nil {
		return err
	}

	tokensWrapper, err := mappers.NewMultiversXToErc20Mapper(mxDataGetter)
	if err != nil {
		return err
	}

	ethClient, err := ethclient.Dial(cfg.Eth.NetworkAddress)
	if err != nil {
		return err
	}

	argsContractsHolder := ethereumClient.ArgsErc20SafeContractsHolder{
		EthClient:              ethClient,
		EthClientStatusHandler: &disabled.StatusHandler{},
	}
	erc20ContractsHolder, err := ethereumClient.NewErc20SafeContractsHolder(argsContractsHolder)
	if err != nil {
		return err
	}

	argsCreator := ethereum.ArgsMigrationBatchCreator{
		TokensList:           cfg.WhitelistedTokens.List,
		TokensMapper:         tokensWrapper,
		Erc20ContractsHolder: erc20ContractsHolder,
		SafeContractAddress:  common.Address{},
		SafeContractWrapper:  nil,
	}

	creator, err := ethereum.NewMigrationBatchCreator(argsCreator)
	if err != nil {
		return err
	}

	batchInfo, err := creator.CreateBatchInfo(context.Background())
	if err != nil {
		return err
	}

	//TODO: save in a file
	val, err := json.MarshalIndent(batchInfo, "", "  ")
	if err != nil {
		return err
	}

	log.Info(string(val))

	return nil
}

func loadConfig(filepath string) (config.MigrationToolConfig, error) {
	cfg := config.MigrationToolConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.MigrationToolConfig{}, err
	}

	return cfg, nil
}
