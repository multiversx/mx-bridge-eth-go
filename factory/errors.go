package factory

import "errors"

var (
	errNilProxy                = errors.New("nil ElrondProxy")
	errNilEthClient            = errors.New("nil eth client")
	errNilMessenger            = errors.New("nil network messenger")
	errNilStatusStorer         = errors.New("nil status storer")
	errNilErc20ContractsHolder = errors.New("nil ERC20 contracts holder")
)
