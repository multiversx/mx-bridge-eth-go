package roleProviders

import "errors"

// ErrNilDataGetter signals that a nil data getter was provided
var ErrNilDataGetter = errors.New("nil data getter")

// ErrNilLogger signals that a nil logger was provided
var ErrNilLogger = errors.New("nil logger")

// ErrNilEthereumChainInteractor signals that a nil Ethereum chain interactor was provided
var ErrNilEthereumChainInteractor = errors.New("nil Ethereum chain interactor")

// ErrAddressIsNotWhitelisted signals that the address is not whitelisted
var ErrAddressIsNotWhitelisted = errors.New("address is not whitelisted")

// ErrInvalidSignature signals that an invalid signature has been provided
var ErrInvalidSignature = errors.New("invalid signature")

// ErrInvalidAddressBytes signals that an invalid address bytes were provided
var ErrInvalidAddressBytes = errors.New("invalid address bytes")
