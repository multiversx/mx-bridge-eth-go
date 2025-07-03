package sui

import "errors"

var (
	errInsufficientCoinBalance = errors.New("insufficient coin balance")
	errNilObjectId             = errors.New("nil object id")
	errNilPackageId            = errors.New("nil package id")
	errNilRelayerSigner        = errors.New("nil relayer signer")
	errNilClient               = errors.New("nil client")
	errGetCoinObjectsFailed    = errors.New("get coin objects failed")
	errNoGasObjects            = errors.New("no gas objects found")
	errInvalidCoinType         = errors.New("invalid coin type format")
	errNilBlockchainClient     = errors.New("nil blockchain client")
)
