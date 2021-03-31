package main

import (
	"context"
	"fmt"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	_ "github.com/urfave/cli"
)

var log = logger.GetOrCreate("eth-bridge")

func main() {
	ethNetworkAddress := "http://127.0.0.1:8545"
	ethSafeAddress := "0x6224Dde04296e2528eF5C5705Db49bfCbF043721"

	log.Debug("Starting bridge")
	ethToElrBridge, err := bridge.NewBridge(ethNetworkAddress, ethSafeAddress)

	if err != nil {
		panic(err)
	}

	fmt.Println("Bridge started")

	ethToElrBridge.Start(context.Background())
	defer ethToElrBridge.Stop()
}
