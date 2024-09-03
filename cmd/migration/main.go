package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	ethereumClient "github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
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
	signMode            = "sign"
	executeMode         = "execute"
	configPath          = "config"
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

	log.Info("process finished successfully")
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
	case signMode:
		return generateAndSign(ctx, cfg)
	case executeMode:
		//TODO: implement
	}

	return fmt.Errorf("unknown execution mode: %s", operationMode)
}

func generateAndSign(ctx *cli.Context, cfg config.MigrationToolConfig) error {
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

	dummyAddress := data.NewAddressFromBytes(bytes.Repeat([]byte{0x1}, 32))
	multisigAddress, err := data.NewAddressFromBech32String(cfg.MultiversX.MultisigContractAddress)
	if err != nil {
		return err
	}

	safeAddress, err := data.NewAddressFromBech32String(cfg.MultiversX.SafeContractAddress)
	if err != nil {
		return err
	}

	argsMXClientDataGetter := multiversx.ArgsMXClientDataGetter{
		MultisigContractAddress: multisigAddress,
		SafeContractAddress:     safeAddress,
		RelayerAddress:          dummyAddress,
		Proxy:                   proxy,
		Log:                     log,
	}
	mxDataGetter, err := multiversx.NewMXClientDataGetter(argsMXClientDataGetter)
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

	safeEthAddress := common.HexToAddress(cfg.Eth.SafeContractAddress)
	safeInstance, err := contract.NewERC20Safe(safeEthAddress, ethClient)
	if err != nil {
		return err
	}

	argsCreator := ethereum.ArgsMigrationBatchCreator{
		MvxDataGetter:        mxDataGetter,
		Erc20ContractsHolder: erc20ContractsHolder,
		SafeContractAddress:  safeEthAddress,
		SafeContractWrapper:  safeInstance,
		Logger:               log,
	}

	creator, err := ethereum.NewMigrationBatchCreator(argsCreator)
	if err != nil {
		return err
	}

	newSafeAddressString := ctx.GlobalString(newSafeAddress.Name)
	if len(newSafeAddressString) == 0 {
		return fmt.Errorf("invalid new safe address for Ethereum")
	}
	newSafeAddressValue := common.HexToAddress(ctx.GlobalString(newSafeAddress.Name))

	batchInfo, err := creator.CreateBatchInfo(context.Background(), newSafeAddressValue)
	if err != nil {
		return err
	}

	val, err := json.MarshalIndent(batchInfo, "", "  ")
	if err != nil {
		return err
	}

	cryptoHandler, err := ethereumClient.NewCryptoHandler(cfg.Eth.PrivateKeyFile)
	if err != nil {
		return err
	}

	log.Info("signing batch", "message hash", batchInfo.MessageHash.String(),
		"public key", cryptoHandler.GetAddress().String())

	signature, err := cryptoHandler.Sign(batchInfo.MessageHash)
	if err != nil {
		return err
	}

	log.Info("Migration .json file contents: \n" + string(val))

	jsonFilename := ctx.GlobalString(migrationJsonFile.Name)
	err = os.WriteFile(jsonFilename, val, os.ModePerm)
	if err != nil {
		return err
	}

	sigInfo := &ethereum.SignatureInfo{
		PublicKey:   cryptoHandler.GetAddress().String(),
		MessageHash: batchInfo.MessageHash.String(),
		Signature:   hex.EncodeToString(signature),
	}

	sigFilename := path.Join(configPath, fmt.Sprintf("%s.json", sigInfo.PublicKey))
	val, err = json.MarshalIndent(sigInfo, "", "  ")
	if err != nil {
		return err
	}

	log.Info("Signature .json file contents: \n" + string(val))

	return os.WriteFile(sigFilename, val, os.ModePerm)
}

func loadConfig(filepath string) (config.MigrationToolConfig, error) {
	cfg := config.MigrationToolConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.MigrationToolConfig{}, err
	}

	return cfg, nil
}
