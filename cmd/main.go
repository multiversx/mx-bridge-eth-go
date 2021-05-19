package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

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
		return startRelay(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

//func playgroundElrond(ctx *cli.Context) error {
//	log.Info("Playground")
//
//	configurationFileName := ctx.GlobalString(configurationFile.Name)
//	config, err := loadConfig(configurationFileName)
//	if err != nil {
//		return err
//	}
//
//	client, err := elrond.NewClient(config.Elrond)
//	if err != nil {
//		return err
//	}

//log.Info(fmt.Sprintf("ERC20 address is %s", client.getERC20Address("574554482d386538333666")))
//log.Info(fmt.Sprintf("TokenId is %s", client.getTokenId("90d2bd2d7d7EE1b46FE4193cB18B02Cb67d7A130")))
//log.Info(fmt.Sprintf("Signers count %d", client.SignersCount(context.TODO(), bridge.ActionId(7))))

//tx := client.GetPendingDepositTransaction(context.TODO())
//log.Info(fmt.Sprintf("%+v", tx))

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
//client.ProposeSetStatus(context.TODO(), bridge.Executed, tx.DepositNonce)
//time.Sleep(30 * time.Second)
//
//actionId := client.GetActionIdForSetStatusOnPendingTransfer(context.TODO())
//log.Info(fmt.Sprintf("%v", actionId))
//
//wasProposed := client.WasProposedSetStatusOnPendingTransfer(context.TODO(), bridge.Executed)
//log.Info(fmt.Sprintf("was proposed: %v", wasProposed))
//hash, err = client.Sign(context.TODO(), actionId)
//if err != nil {
//	log.Error(err.Error())
//}
//log.Info(fmt.Sprintf("Sign %s", hash))
//time.Sleep(30 * time.Second)
//
//hash, err = client.Execute(context.TODO(), actionId, tx.DepositNonce)
//if err != nil {
//	log.Error(err.Error())
//}
//log.Info(fmt.Sprintf("Execute %s", hash))

//nonce := bridge.Nonce(45)
//transfer, err := client.ProposeTransfer(context.TODO(), &bridge.DepositTransaction{
//	To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
//	From:         "0x132A150926691F08a693721503a38affeD18d524",
//	TokenAddress: "574554482d393761323662",
//	Amount:       big.NewInt(3),
//	DepositNonce: nonce,
//})
//if err != nil {
//	return err
//}
//log.Info(transfer)
//
//result := client.WasProposedTransfer(context.TODO(), nonce)
//log.Info(fmt.Sprint(result))
//
//time.Sleep(10 * time.Second)
//actionId := client.GetActionIdForProposeTransfer(context.TODO(), nonce)
//log.Info(fmt.Sprintf("ActionId: %d", actionId))
//
//hash, err := client.Sign(context.TODO(), actionId)
//if err != nil {
//	return err
//}
//log.Info(fmt.Sprintf("Sign hash %q", hash))
//
//time.Sleep(10 * time.Second)
//hash, err = client.Execute(context.TODO(), actionId)
//if err != nil {
//	return err
//}
//log.Info(fmt.Sprintf("Perform hash %q", hash))
//
//time.Sleep(10 * time.Second)
//log.Info(fmt.Sprintf("%v", client.WasExecuted(context.TODO(), actionId)))

// deploy
// deployCC
// stake
// MultiTransferEsdt_WrappedEthIssue
// MultiTransferEsdt_TransferEsdt

//client, err := eth.NewClient(config.Eth)
//if err != nil {
//	return err
//}
//
//client.GetPendingDepositTransaction(context.Background())
//
//	return nil
//}

//
//func playgroundEth(ctx *cli.Context) error {
//	log.Info("Playground Eth")
//
//	configurationFileName := ctx.GlobalString(configurationFile.Name)
//	config, err := loadConfig(configurationFileName)
//	if err != nil {
//		return err
//	}
//
//	client, err := eth.NewClient(config.Eth, &broadcasterStub{})
//	if err != nil {
//		return err
//	}
//
//	tx := client.GetPendingDepositTransaction(context.Background())
//	log.Info(fmt.Sprintf("%+v", tx))
//	log.Info(fmt.Sprintf("Nonce %v", tx.DepositNonce))
//
//	client.ProposeSetStatus(context.Background(), tx.DepositNonce)
//	hash, err := client.Execute(context.Background(), bridge.NewActionId(0), tx.DepositNonce)
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
