package ethereum

import "errors"

var (
	errEmptyTokensList         = errors.New("empty tokens list")
	errNilTokensMapper         = errors.New("nil MultiversX to Ethereum tokens mapper")
	errNilErc20ContractsHolder = errors.New("nil ERC20 contracts holder")
	errNilSafeContractWrapper  = errors.New("nil safe contract wrapper")
	errPublicKeyCast           = errors.New("error casting public key to ECDSA")
)
