package wrappers

import "errors"

var (
	errNilErc20Contract       = errors.New("nil ERC20 contract")
	errNilBlockchainClient    = errors.New("nil blockchain client")
	errNilMultiSigContract    = errors.New("nil multi sig contract")
	errNilSCExecProxyContract = errors.New("nil sc exec proxy contract")
	errNilSafeContract        = errors.New("nil safe contract")
)
