package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"

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
		return playgroundEth(c)
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

	// carol: erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8

	//result := client.WasProposedTransfer(context.TODO(), bridge.Nonce(1))
	//log.Info(fmt.Sprint(result))
	//
	//log.Info(fmt.Sprintf("ActionId: %d", client.GetActionIdForProposeTransfer(context.TODO(), bridge.Nonce(1))))
	//
	//hash, err := client.Sign(context.TODO(), bridge.ActionId(2))
	//if err != nil {
	//	log.Error(err.Error())
	//}
	//log.Info(fmt.Sprintf("Sign hash %q", hash))
	//
	//hash, err = client.Execute(context.TODO(), bridge.ActionId(2))
	//if err != nil {
	//	log.Error(err.Error())
	//}
	//log.Info(fmt.Sprintf("Perform hash %q", hash))

	log.Info(fmt.Sprintf("%v", client.WasExecuted(context.TODO(), bridge.ActionId(2))))

	// deploy
	// deployCC
	// stake
	// proposeMultiTransferEsdtSetLocalMintRole
	// issueToken

	//client, err := eth.NewClient(config.Eth)
	//if err != nil {
	//	return err
	//}
	//
	//client.GetPendingDepositTransaction(context.Background())

	return nil
}

func playgroundEth(ctx *cli.Context) error {
	log.Info("Playground Eth")

	configurationFileName := ctx.GlobalString(configurationFile.Name)
	config, err := loadConfig(configurationFileName)
	if err != nil {
		return err
	}

	client, err := eth.NewClient(config.Eth, &broadcasterStub{})
	if err != nil {
		return err
	}

	tx := client.GetPendingDepositTransaction(context.Background())
	log.Info(fmt.Sprintf("%v", tx))

	client.ProposeSetStatusSuccessOnPendingTransfer(context.Background())
	hash, err := client.Execute(context.Background(), bridge.ActionId(0))
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Executed with hash %q", hash))

	return nil
}

type broadcasterStub struct {
	lastSignData           string
	lastBroadcastSignature string
}

func (b *broadcasterStub) SendSignature(signData, signature string) {
	b.lastSignData = signData
	b.lastBroadcastSignature = signature
}

func (b *broadcasterStub) Signatures() [][]byte {
	return [][]byte{[]byte(b.lastBroadcastSignature)}
}

func (b *broadcasterStub) SignData() string {
	return b.lastSignData
}

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
