package ethereum

import "errors"

var (
	errInsufficientErc20Balance = errors.New("insufficient ERC20 balance")
	errInsufficientBalance      = errors.New("insufficient balance")
	errPublicKeyCast            = errors.New("error casting public key to ECDSA")
	errNilERC20ContractsHandler = errors.New("nil ERC20 contracts handler")
	errNilGasHandler            = errors.New("nil gas handler")
	errNilEthClient             = errors.New("nil eth client")
)
