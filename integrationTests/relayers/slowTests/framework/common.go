package framework

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	log                       = logger.GetOrCreate("integrationtests/slowtests")
	addressPubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	zeroValueBigInt           = big.NewInt(0)
)
