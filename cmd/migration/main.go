package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	ethereumClient "github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement"
	"github.com/multiversx/mx-bridge-eth-go/clients/gasManagement/factory"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	"github.com/multiversx/mx-bridge-eth-go/cmd/migration/disabled"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/executors/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/executors/ethereum/bridgeV2Wrappers"
	"github.com/multiversx/mx-bridge-eth-go/executors/ethereum/bridgeV2Wrappers/contract"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/urfave/cli"
)

const (
	filePathPlaceholder  = "[path]"
	queryMode            = "query"
	signMode             = "sign"
	executeMode          = "execute"
	configPath           = "config"
	timestampPlaceholder = "[timestamp]"
	publicKeyPlaceholder = "[public-key]"
)

var log = logger.GetOrCreate("main")

type internalComponents struct {
	creator              BatchCreator
	batch                *ethereum.BatchInfo
	cryptoHandler        ethereumClient.CryptoHandler
	ethClient            *ethclient.Client
	ethereumChainWrapper ethereum.EthereumChainWrapper
}

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
	case queryMode:
		return executeQuery(cfg)
	case signMode:
		_, err = generateAndSign(ctx, cfg)
		return err
	case executeMode:
		return executeTransfer(ctx, cfg)
	}

	return fmt.Errorf("unknown execution mode: %s", operationMode)
}

func executeQuery(cfg config.MigrationToolConfig) error {
	components, err := createInternalComponentsWithBatchCreator(cfg)
	if err != nil {
		return err
	}

	dummyEthAddress := common.Address{}
	info, err := components.creator.CreateBatchInfo(context.Background(), dummyEthAddress, nil)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Token balances for ERC20 safe address %s\n%s",
		cfg.Eth.SafeContractAddress,
		ethereum.TokensBalancesDisplayString(info),
	))

	return nil
}

func createInternalComponentsWithBatchCreator(cfg config.MigrationToolConfig) (*internalComponents, error) {
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
		return nil, err
	}

	dummyAddress := data.NewAddressFromBytes(bytes.Repeat([]byte{0x1}, 32))
	multisigAddress, err := data.NewAddressFromBech32String(cfg.MultiversX.MultisigContractAddress)
	if err != nil {
		return nil, err
	}

	safeAddress, err := data.NewAddressFromBech32String(cfg.MultiversX.SafeContractAddress)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	ethClient, err := ethclient.Dial(cfg.Eth.NetworkAddress)
	if err != nil {
		return nil, err
	}

	argsContractsHolder := ethereumClient.ArgsErc20SafeContractsHolder{
		EthClient:              ethClient,
		EthClientStatusHandler: &disabled.StatusHandler{},
	}
	erc20ContractsHolder, err := ethereumClient.NewErc20SafeContractsHolder(argsContractsHolder)
	if err != nil {
		return nil, err
	}

	safeEthAddress := common.HexToAddress(cfg.Eth.SafeContractAddress)

	bridgeEthAddress := common.HexToAddress(cfg.Eth.MultisigContractAddress)
	multiSigInstance, err := contract.NewBridge(bridgeEthAddress, ethClient)
	if err != nil {
		return nil, err
	}

	argsClientWrapper := bridgeV2Wrappers.ArgsEthereumChainWrapper{
		StatusHandler:    &disabled.StatusHandler{},
		MultiSigContract: multiSigInstance,
		BlockchainClient: ethClient,
	}
	ethereumChainWrapper, err := bridgeV2Wrappers.NewEthereumChainWrapper(argsClientWrapper)
	if err != nil {
		return nil, err
	}

	argsCreator := ethereum.ArgsMigrationBatchCreator{
		MvxDataGetter:        mxDataGetter,
		Erc20ContractsHolder: erc20ContractsHolder,
		SafeContractAddress:  safeEthAddress,
		EthereumChainWrapper: ethereumChainWrapper,
		Logger:               log,
	}

	creator, err := ethereum.NewMigrationBatchCreator(argsCreator)
	if err != nil {
		return nil, err
	}

	return &internalComponents{
		creator:              creator,
		ethClient:            ethClient,
		ethereumChainWrapper: ethereumChainWrapper,
	}, nil
}

func generateAndSign(ctx *cli.Context, cfg config.MigrationToolConfig) (*internalComponents, error) {
	components, err := createInternalComponentsWithBatchCreator(cfg)
	if err != nil {
		return nil, err
	}

	newSafeAddressString := ctx.GlobalString(newSafeAddress.Name)
	if len(newSafeAddressString) == 0 {
		return nil, fmt.Errorf("invalid new safe address for Ethereum")
	}
	newSafeAddressValue := common.HexToAddress(ctx.GlobalString(newSafeAddress.Name))

	partialMigration, err := ethereum.ConvertPartialMigrationStringToMap(ctx.GlobalString(partialMigration.Name))
	if err != nil {
		return nil, err
	}

	components.batch, err = components.creator.CreateBatchInfo(context.Background(), newSafeAddressValue, partialMigration)
	if err != nil {
		return nil, err
	}

	val, err := json.MarshalIndent(components.batch, "", "  ")
	if err != nil {
		return nil, err
	}

	components.cryptoHandler, err = ethereumClient.NewCryptoHandler(cfg.Eth.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	log.Info("signing batch", "message hash", components.batch.MessageHash.String(),
		"public key", components.cryptoHandler.GetAddress().String())

	signature, err := components.cryptoHandler.Sign(components.batch.MessageHash)
	if err != nil {
		return nil, err
	}

	log.Info("Migration .json file contents: \n" + string(val))

	jsonFilename := ctx.GlobalString(migrationJsonFile.Name)
	jsonFilename = applyTimestamp(jsonFilename)
	err = os.WriteFile(jsonFilename, val, os.ModePerm)
	if err != nil {
		return nil, err
	}

	sigInfo := &ethereum.SignatureInfo{
		Address:     components.cryptoHandler.GetAddress().String(),
		MessageHash: components.batch.MessageHash.String(),
		Signature:   hex.EncodeToString(signature),
	}

	sigFilename := ctx.GlobalString(signatureJsonFile.Name)
	sigFilename = applyTimestamp(sigFilename)
	sigFilename = applyPublicKey(sigFilename, sigInfo.Address)
	val, err = json.MarshalIndent(sigInfo, "", "  ")
	if err != nil {
		return nil, err
	}

	log.Info("Signature .json file contents: \n" + string(val))

	err = os.WriteFile(sigFilename, val, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return components, nil
}

func executeTransfer(ctx *cli.Context, cfg config.MigrationToolConfig) error {
	components, err := generateAndSign(ctx, cfg)
	if err != nil {
		return err
	}

	gasStationConfig := cfg.Eth.GasStation
	argsGasStation := gasManagement.ArgsGasStation{
		RequestURL:             gasStationConfig.URL,
		RequestPollingInterval: time.Duration(gasStationConfig.PollingIntervalInSeconds) * time.Second,
		RequestRetryDelay:      time.Duration(gasStationConfig.RequestRetryDelayInSeconds) * time.Second,
		MaximumFetchRetries:    gasStationConfig.MaxFetchRetries,
		RequestTime:            time.Duration(gasStationConfig.RequestTimeInSeconds) * time.Second,
		MaximumGasPrice:        gasStationConfig.MaximumAllowedGasPrice,
		GasPriceSelector:       core.EthGasPriceSelector(gasStationConfig.GasPriceSelector),
		GasPriceMultiplier:     gasStationConfig.GasPriceMultiplier,
	}
	gs, err := factory.CreateGasStation(argsGasStation, gasStationConfig.Enabled)
	if err != nil {
		return err
	}

	err = waitForGasPrice(gs)
	if err != nil {
		return err
	}

	args := ethereum.ArgsMigrationBatchExecutor{
		EthereumChainWrapper:    components.ethereumChainWrapper,
		CryptoHandler:           components.cryptoHandler,
		Batch:                   *components.batch,
		Signatures:              ethereum.LoadAllSignatures(log, configPath),
		Logger:                  log,
		GasHandler:              gs,
		TransferGasLimitBase:    cfg.Eth.GasLimitBase,
		TransferGasLimitForEach: cfg.Eth.GasLimitForEach,
	}

	executor, err := ethereum.NewMigrationBatchExecutor(args)
	if err != nil {
		return err
	}

	return executor.ExecuteTransfer(context.Background())
}

func waitForGasPrice(gs clients.GasHandler) error {
	log.Info("Fetching a gas price value. Please wait...")
	numRetries := 5
	timeBetweenChecks := time.Second

	var err error
	var gasPrice *big.Int
	for i := 0; i < numRetries; i++ {
		time.Sleep(timeBetweenChecks)
		gasPrice, err = gs.GetCurrentGasPrice()
		if err != nil {
			log.Debug("waitForGasPrice", "error", err)
			continue
		}

		log.Info("Fetched the gas price", "value", gasPrice.String())
		return nil
	}

	return err
}

func loadConfig(filepath string) (config.MigrationToolConfig, error) {
	cfg := config.MigrationToolConfig{}
	err := chainCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.MigrationToolConfig{}, err
	}

	return cfg, nil
}

func applyTimestamp(input string) string {
	actualTimestamp := time.Now().Format("2006-01-02T15-04-05")
	actualTimestamp = strings.Replace(actualTimestamp, "T", "-", 1)

	return strings.Replace(input, timestampPlaceholder, actualTimestamp, 1)
}

func applyPublicKey(input string, publickey string) string {
	return strings.Replace(input, publicKeyPlaceholder, publickey, 1)
}
