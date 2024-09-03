package ethereum

import "errors"

var (
	errEmptyTokensList           = errors.New("empty tokens list")
	errNilMvxDataGetter          = errors.New("nil MultiversX data getter")
	errNilErc20ContractsHolder   = errors.New("nil ERC20 contracts holder")
	errNilSafeContractWrapper    = errors.New("nil safe contract wrapper")
	errWrongERC20AddressResponse = errors.New("wrong ERC20 address response")
)
