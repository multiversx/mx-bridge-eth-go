package ethereum

import "errors"

var (
	errQuorumNotReached         = errors.New("quorum not reached")
	errInsufficientErc20Balance = errors.New("insufficient ERC20 balance")
	errPublicKeyCast            = errors.New("error casting public key to ECDSA")
	errNilClientWrapper         = errors.New("nil client wrapper")
	errNilERC20ContractsHandler = errors.New("nil ERC20 contracts handler")
	errNilTokensMapper          = errors.New("nil tokens mapper")
	errNilLogger                = errors.New("nil logger")
	errNilAddressConverter      = errors.New("nil address converter")
	errNilBroadcaster           = errors.New("nil broadcaster")
	errNilPrivateKey            = errors.New("nil private key")
	errNilBatch                 = errors.New("nil batch")
	errNilSignaturesHolder      = errors.New("nil signatures holder")
	errNilGasHandler            = errors.New("nil gas handler")
	errInvalidGasLimit          = errors.New("invalid gas limit")
	errNilStatusHandler         = errors.New("nil status handler")
	errNilEthClient             = errors.New("nil eth client")
)
