package sui

import "errors"

var (
	errInsufficientCoinBalance     = errors.New("insufficient coin balance")
	errNilObjectId                 = errors.New("nil object id")
	errNilPackageId                = errors.New("nil package id")
	errNilAddress                  = errors.New("nil address")
	errNilProxy                    = errors.New("nil proxy")
	errInvalidCoinType             = errors.New("invalid coin type format")
	errInvalidInitialSharedVersion = errors.New("invalid initial shared version")
)
