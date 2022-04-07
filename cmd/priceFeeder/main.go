package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/aggregator"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/aggregator/fetchers"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/aggregator/notifees"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core/polling"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

const configPath = "config/config.toml"

var log = logger.GetOrCreate("priceFeeder/main")

func main() {
	_ = logger.SetLogLevel("*:DEBUG")

	log.Info("Price feeder will fetch the price of a defined pair from a bunch of exchanges, and will" +
		" write to the contract if the price changed")
	log.Info("application started")

	err := runApp()
	if err != nil {
		log.Error(err.Error())
	} else {
		log.Info("application gracefully closed")
	}
}

func runApp() error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if len(cfg.GeneralConfig.NetworkAddress) == 0 {
		return fmt.Errorf("empty NetworkAddress in config file")
	}

	proxy := blockchain.NewElrondProxy(cfg.GeneralConfig.NetworkAddress, nil)
	proxyWithCacher, err := blockchain.NewElrondProxyWithCache(proxy, time.Second*time.Duration(cfg.GeneralConfig.ProxyCacherExpirationSeconds))
	if err != nil {
		return err
	}

	priceFetchers, err := createPriceFetchers()
	if err != nil {
		return err
	}

	argsPriceAggregator := aggregator.ArgsPriceAggregator{
		PriceFetchers: priceFetchers,
		MinResultsNum: cfg.GeneralConfig.MinResultsNum,
	}
	priceAggregator, err := aggregator.NewPriceAggregator(argsPriceAggregator)
	if err != nil {
		return err
	}

	txBuilder, err := builders.NewTxBuilder(blockchain.NewTxSigner())
	if err != nil {
		return err
	}

	txNonceHandler, err := interactors.NewNonceTransactionHandler(proxyWithCacher, time.Second*time.Duration(cfg.GeneralConfig.IntervalToResendTxsInSeconds), true)
	if err != nil {
		return err
	}

	aggregatorAddress, err := data.NewAddressFromBech32String(cfg.GeneralConfig.AggregatorContractAddress)
	if err != nil {
		return err
	}

	var keyGen = signing.NewKeyGenerator(ed25519.NewEd25519())
	wallet := interactors.NewWallet()
	privateKeyBytes, err := wallet.LoadPrivateKeyFromPemFile(cfg.GeneralConfig.PrivateKeyFile)
	if err != nil {
		return err
	}

	privateKey, err := keyGen.PrivateKeyFromByteArray(privateKeyBytes)

	if err != nil {
		return err
	}
	argsElrondNotifee := notifees.ArgsElrondNotifee{
		Proxy:           proxyWithCacher,
		TxBuilder:       txBuilder,
		TxNonceHandler:  txNonceHandler,
		ContractAddress: aggregatorAddress,
		PrivateKey:      privateKey,
		BaseGasLimit:    cfg.GeneralConfig.BaseGasLimit,
		GasLimitForEach: cfg.GeneralConfig.GasLimitForEach,
	}
	elrondNotifee, err := notifees.NewElrondNotifee(argsElrondNotifee)
	if err != nil {
		return err
	}

	argsPriceNotifier := aggregator.ArgsPriceNotifier{
		Pairs:            []*aggregator.ArgsPair{},
		Fetcher:          priceAggregator,
		Notifee:          elrondNotifee,
		AutoSendInterval: time.Second * time.Duration(cfg.GeneralConfig.AutoSendIntervalInSeconds),
	}
	for _, pair := range cfg.Pairs {
		argsPair := &aggregator.ArgsPair{
			Base:                      pair.Base,
			Quote:                     pair.Quote,
			PercentDifferenceToNotify: pair.PercentDifferenceToNotify,
			TrimPrecision:             pair.TrimPrecision,
			DenominationFactor:        pair.DenominationFactor,
		}
		argsPriceNotifier.Pairs = append(argsPriceNotifier.Pairs, argsPair)
	}
	priceNotifier, err := aggregator.NewPriceNotifier(argsPriceNotifier)
	if err != nil {
		return err
	}

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              log,
		Name:             "price notifier polling handler",
		PollingInterval:  time.Second * time.Duration(cfg.GeneralConfig.PollIntervalInSeconds),
		PollingWhenError: time.Second * time.Duration(cfg.GeneralConfig.PollIntervalInSeconds),
		Executor:         priceNotifier,
	}

	pollingHandler, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	log.Info("Starting Elrond Notifee")

	err = pollingHandler.StartProcessingLoop()
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("application closing, closing polling handler...")

	err = pollingHandler.Close()
	return err
}

func loadConfig(filepath string) (config.PriceNotifierConfig, error) {
	cfg := config.PriceNotifierConfig{}
	err := elrondCore.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.PriceNotifierConfig{}, err
	}

	return cfg, nil
}

func createPriceFetchers() ([]aggregator.PriceFetcher, error) {
	exchanges := fetchers.ImplementedFetchers
	priceFetchers := make([]aggregator.PriceFetcher, 0, len(exchanges))
	for _, exchangeName := range exchanges {
		priceFetcher, err := fetchers.NewPriceFetcher(exchangeName, &aggregator.HttpResponseGetter{})
		if err != nil {
			return nil, err
		}

		priceFetchers = append(priceFetchers, priceFetcher)
	}

	return priceFetchers, nil
}
