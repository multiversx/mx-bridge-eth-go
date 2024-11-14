package ethereum

import "errors"

var (
	errEmptyTokensList               = errors.New("empty tokens list")
	errNilMvxDataGetter              = errors.New("nil MultiversX data getter")
	errNilErc20ContractsHolder       = errors.New("nil ERC20 contracts holder")
	errWrongERC20AddressResponse     = errors.New("wrong ERC20 address response")
	errNilLogger                     = errors.New("nil logger")
	errNilCryptoHandler              = errors.New("nil crypto handler")
	errNilEthereumChainWrapper       = errors.New("nil Ethereum chain wrapper")
	errQuorumNotReached              = errors.New("quorum not reached")
	errInvalidSignature              = errors.New("invalid signature")
	errMultisigContractPaused        = errors.New("multisig contract paused")
	errNilGasHandler                 = errors.New("nil gas handler")
	errInvalidPartialMigrationString = errors.New("invalid partial migration string")
)
