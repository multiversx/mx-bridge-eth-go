package v2

import "errors"

var (
	errNilBatch            = errors.New("nil batch")
	errInvalidDepositNonce = errors.New("invalid deposit nonce")
	errNilLogger           = errors.New("nil logger")
	errNilElrondClient     = errors.New("nil Elrond client")
	errNilEthereumClient   = errors.New("nil Ethereum client")
	errNilTopologyProvider = errors.New("nil topology provider")
)
