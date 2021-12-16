package wrappers

import "errors"

var (
	errNilErc20Contract    = errors.New("nil ERC20 contract")
	errNilStatusHandler    = errors.New("nil status handler")
	errNilBlockchainClient = errors.New("nil blockchain client")
	errNilMultiSigContract = errors.New("nil multi sig contract")
)
