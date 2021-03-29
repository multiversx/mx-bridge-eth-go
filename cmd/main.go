package main

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	_ "github.com/urfave/cli"
)

var log = logger.GetOrCreate("eth-bridge")

func main() {
	//AFCoin deployed to: 0x5abc5e20F56Dc6Ce962C458A3142FC289A757F4E
	//ERC20Safe deployed to: 0x6224Dde04296e2528eF5C5705Db49bfCbF043721

	log.Debug("Some debug message")
}
