package bridgeV2Wrappers

import "errors"

var (
	errNilBlockchainClient = errors.New("nil blockchain client")
	errNilMultiSigContract = errors.New("nil multi sig contract")
)
