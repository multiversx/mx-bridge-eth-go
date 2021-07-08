package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/elrond"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/urfave/cli"
	_ "github.com/urfave/cli"
)

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2

	filePathPlaceholder = "[path]"
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

	configurationFile = cli.StringFlag{
		Name: "config",
		Usage: "The `" + filePathPlaceholder + "` for the main configuration file. This TOML file contain the main " +
			"configurations such as the marshalizer type",
		Value: "./config.toml",
	}
)

var log = logger.GetOrCreate("main")

func main() {
	app := cli.NewApp()
	app.Name = "Relay CLI app"
	app.Usage = "This is the entry point for the bridge relay"
	app.Flags = []cli.Flag{
		logLevel,
		configurationFile,
	}
	app.Version = "v0.0.1"
	app.Authors = []cli.Author{
		{
			Name:  "The Agile Freaks team",
			Email: "office@agilefreaks.com",
		},
	}

	app.Action = func(c *cli.Context) error {
		//return startRelay(c)
		return playgroundElrond(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func playgroundElrond(ctx *cli.Context) error {
	log.Info("Playground")

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	config, err := loadConfig(configurationFileName)
	if err != nil {
		return err
	}

	client, err := elrond.NewClient(config.Elrond)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("ERC20 address is %s", client.GetErc20Address("574554482d626631623064")))
	log.Info(fmt.Sprintf("TokenId is %s", client.GetTokenId("6DF7EFEA5d25B76AEb6A53537390c634faeD9AeD")))
	//log.Info(fmt.Sprintf("Signers count %d", client.SignersCount(context.TODO(), bridge.ActionId(7))))

	//batch := client.GetPending(context.TODO())
	//log.Info(fmt.Sprintf("%+v", batch))
	//log.Info(fmt.Sprintf("%+v", batch.Transactions[0]))

	//_ = client.GetActionIdForProposeTransfer(context.TODO(), bridge.NewBatchId(4))

	//ethClient, err := eth.NewClient(config.Eth, &broadcasterStub{}, client)
	//if err != nil {
	//	return err
	//}
	//
	//_, _ = ethClient.ProposeTransfer(context.Background(), tx)
	//
	//hash, err := ethClient.Execute(context.Background(), bridge.NewActionId(0), bridge.NewNonce(0))
	//if err != nil {
	//	return err
	//}
	//log.Info(fmt.Sprintf("Executed with hash %s", hash))
	//

	//for _, tx := range batch.Transactions {
	//	tx.Status = bridge.Executed
	//}
	//client.ProposeSetStatus(context.TODO(), batch)
	//time.Sleep(30 * time.Second)
	//
	//actionId := client.GetActionIdForSetStatusOnPendingTransfer(context.TODO())
	//log.Info(fmt.Sprintf("%v", actionId))
	//
	//wasProposed := client.WasProposedSetStatus(context.TODO(), batch)
	//log.Info(fmt.Sprintf("was proposed: %v", wasProposed))

	//hash, err := client.Sign(context.TODO(), actionId)
	//if err != nil {
	//	log.Error(err.Error())
	//}
	//log.Info(fmt.Sprintf("Sign %s", hash))
	//time.Sleep(30 * time.Second)
	//
	//hash, err = client.Execute(context.TODO(), actionId, batch.Id)
	//if err != nil {
	//	log.Error(err.Error())
	//}
	//log.Info(fmt.Sprintf("Execute %s", hash))

	//batchId := bridge.NewBatchId(1)
	//tx1 := &bridge.DepositTransaction{
	//	To:           "erd1rve9evhhfhuw26ctgctzxmevptj43yv800g9603l8vmua2ew7lcq4tp2an",
	//	From:         "0x765D98061aca14648c7b40EEEdcB0aacBBE33235",
	//	TokenAddress: "0x6DF7EFEA5d25B76AEb6A53537390c634faeD9AeD",
	//	Amount:       big.NewInt(2),
	//}
	//tx2 := &bridge.DepositTransaction{
	//	To:           "erd1rve9evhhfhuw26ctgctzxmevptj43yv800g9603l8vmua2ew7lcq4tp2an",
	//	From:         "0x765D98061aca14648c7b40EEEdcB0aacBBE33235",
	//	TokenAddress: "0x6DF7EFEA5d25B76AEb6A53537390c634faeD9AeD",
	//	Amount:       big.NewInt(3),
	//}
	//batch := &bridge.Batch{
	//	Id:           batchId,
	//	Transactions: []*bridge.DepositTransaction{tx1, tx2},
	//}
	//
	//transfer, err := client.ProposeTransfer(context.TODO(), batch)
	//if err != nil {
	//	log.Error(err.Error())
	//}
	//log.Info(transfer)
	//
	//time.Sleep(30 * time.Second)
	//result := client.WasProposedTransfer(context.TODO(), batch)
	//log.Info(fmt.Sprint(result))
	//
	//time.Sleep(30 * time.Second)
	//actionId := client.GetActionIdForProposeTransfer(context.TODO(), batch)
	//log.Info(fmt.Sprintf("ActionId: %v", actionId))
	//
	//hash, err := client.Sign(context.TODO(), actionId)
	//if err != nil {
	//	return err
	//}
	//log.Info(fmt.Sprintf("Sign hash %q", hash))
	//
	//time.Sleep(30 * time.Second)
	//hash, err = client.Execute(context.TODO(), actionId, nil)
	//if err != nil {
	//	return err
	//}
	//log.Info(fmt.Sprintf("Perform hash %q", hash))
	//
	//time.Sleep(30 * time.Second)
	//log.Info(fmt.Sprintf("%v", client.WasExecuted(context.TODO(), actionId, nil)))

	// deploy
	// issue tokens (snippets)
	// deployCC
	// setLocalRoles
	// issue token -> esdtSafeAddTokenToWhitelist

	//client, err := eth.NewClient(config.Eth)
	//if err != nil {
	//	return err
	//}
	//
	//client.GetPendingDepositTransaction(context.Background())

	return nil
}

//func playgroundEth(ctx *cli.Context) error {
//	log.Info("Playground Eth")
//
//	configurationFileName := ctx.GlobalString(configurationFile.Name)
//	config, err := loadConfig(configurationFileName)
//	if err != nil {
//		return err
//	}
//
//	elrdClient, err := elrond.NewClient(config.Elrond)
//	if err != nil {
//		return err
//	}
//
//	client, err := eth.NewClient(config.Eth, &broadcasterStub{}, elrdClient)
//	if err != nil {
//		return err
//	}
//
//	batchId := bridge.NewBatchId(3)
//	tx1 := &bridge.DepositTransaction{
//		To:           "0x264EeFfE37AA569BEc16A951c51bA25a98e07Dab",
//		From:         "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
//		TokenAddress: "0x5745474c442d666639323737",
//		Amount:       big.NewInt(2),
//	}
//	tx2 := &bridge.DepositTransaction{
//		To:           "0x264EeFfE37AA569BEc16A951c51bA25a98e07Dab",
//		From:         "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
//		TokenAddress: "0x5745474c442d666639323737",
//		Amount:       big.NewInt(3),
//	}
//	batch := &bridge.Batch{
//		Id:           batchId,
//		Transactions: []*bridge.DepositTransaction{tx1, tx2},
//	}
//
//	_, _ = client.ProposeTransfer(context.Background(), batch)

//batch := client.GetPending(context.Background())
//log.Info(fmt.Sprintf("%+v", batch))
//log.Info(fmt.Sprintf("Nonce %v", batch.Id))
//log.Info(fmt.Sprintf("Transactions %+v", batch.Transactions))
//
//batch.SetStatusOnAllTransactions(bridge.Executed, nil)
//
//client.ProposeSetStatus(context.Background(), batch)

//	hash, err := client.Execute(context.Background(), bridge.NewActionId(0), batch.Id)
//	if err != nil {
//		return err
//	}
//	log.Info(fmt.Sprintf("Executed with hash %q", hash))
//
//	return nil
//}

//type broadcasterStub struct {
//	lastBroadcastSignature []byte
//}
//
//func (b *broadcasterStub) SendSignature(signature []byte) {
//	b.lastBroadcastSignature = signature
//}
//
//func (b *broadcasterStub) Signatures() [][]byte {
//	return [][]byte{b.lastBroadcastSignature}
//}

func startRelay(ctx *cli.Context) error {
	logLevelFlagValue := ctx.GlobalString(logLevel.Name)
	err := logger.SetLogLevel(logLevelFlagValue)
	if err != nil {
		return err
	}

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	config, err := loadConfig(configurationFileName)
	if err != nil {
		return err
	}

	ethToElrRelay, err := relay.NewRelay(config, "EthToErlRelay")
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
