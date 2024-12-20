package framework

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// HalfBridgeIdentifier is the type that holds the half-bridge identifier (counter)
type HalfBridgeIdentifier string

// TokenBalanceType represents the token type that should be checked for balance
type TokenBalanceType string

const (
	// FirstHalfBridge represents the first half bridge in the tests
	FirstHalfBridge HalfBridgeIdentifier = "first half bridge"
	// SecondHalfBridge represents the second half bridge in the tests
	SecondHalfBridge HalfBridgeIdentifier = "second half bridge"

	// UniversalToken is the universal token identifier
	UniversalToken TokenBalanceType = "universal"
	// ChainSpecificToken is the chain-specific token identifier
	ChainSpecificToken TokenBalanceType = "chain-specific"
)

var (
	log                       = logger.GetOrCreate("integrationtests/slowtests")
	addressPubkeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	zeroValueBigInt           = big.NewInt(0)
)
