package factory

import "errors"

var (
	errNilProxy                = errors.New("nil proxy")
	errNilEthClient            = errors.New("nil eth client")
	errNilMessenger            = errors.New("nil network messenger")
	errNilStatusStorer         = errors.New("nil status storer")
	errNilErc20ContractsHolder = errors.New("nil ERC20 contracts holder")
	errMissingConfig           = errors.New("missing config")
	errInvalidValue            = errors.New("invalid value")
	errNilMetricsHolder        = errors.New("nil metrics holder")
	errNilStatusHandler        = errors.New("nil status handler")
)
