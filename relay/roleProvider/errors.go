package roleProvider

import "errors"

// ErrNilElrondChainInteractor signals that a nil Elrond chain interactor was provided
var ErrNilElrondChainInteractor = errors.New("nil Elrond chain interactor")

// ErrNilEthereumChainInteractor signals that a nil Ethereum chain interactor was provided
var ErrNilEthereumChainInteractor = errors.New("nil Ethereum chain interactor")

// ErrNilLogger signals that a nil logger was provided
var ErrNilLogger = errors.New("nil logger")

// ErrAddressIsNotWhitelisted signals that the address is not whitelisted
var ErrAddressIsNotWhitelisted = errors.New("address is not whitelisted")

// ErrInvalidSignature signals that an invalid signature has been provided
var ErrInvalidSignature = errors.New("invalid signature")
