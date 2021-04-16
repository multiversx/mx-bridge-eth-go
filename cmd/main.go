package main

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	_ "github.com/urfave/cli"
)

var log = logger.GetOrCreate("eth-relay")

func main() {
	// TODO: set on a default config file
	ethNetworkAddress := "http://127.0.0.1:8545"
	ethSafeAddress := "0x6224Dde04296e2528eF5C5705Db49bfCbF043721"
	elrondNetworkAddress := "http://localhost:7950"
	elrondSafeAddress := "erd1qqqqqqqqqqqqqpgqfzydqmdw7m2vazsp6u5p95yxz76t2p9rd8ss0zp9ts"
	elrondPrivateKeyPath := "../mytestnet/testnet/wallets/users/alice.pem"

	log.Debug("Starting relay")
	ethToElrRelay, err := relay.NewRelay(ethNetworkAddress, ethSafeAddress, elrondNetworkAddress, elrondSafeAddress, elrondPrivateKeyPath)

	if err != nil {
		panic(err)
	}

	fmt.Println("Relay started")

	ethToElrRelay.Start(context.Background())
	defer ethToElrRelay.Stop()
}
