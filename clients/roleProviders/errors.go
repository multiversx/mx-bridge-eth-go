package roleproviders

import "errors"

var (
	// ErrNilEthereumChainInteractor signals that a nil Ethereum chain interactor was provided
	ErrNilEthereumChainInteractor = errors.New("nil Ethereum chain interactor")

	// ErrAddressIsNotWhitelisted signals that the address is not whitelisted
	ErrAddressIsNotWhitelisted = errors.New("address is not whitelisted")

	// ErrInvalidSignature signals that an invalid signature has been provided
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidAddressBytes signals that an invalid address bytes were provided
	ErrInvalidAddressBytes = errors.New("invalid address bytes")

	// ErrInvalidSuiAddress signals that an invalid sui address was provided
	ErrInvalidSuiAddress = errors.New("invalid sui address")

	// ErrInvalidSignaturesArray signals that an invalid array of signatures was provided
	ErrInvalidSignaturesArray = errors.New("invalid signatures array")

	// ErrInvalidMessagesArray signals that an invalid array of messages was provided
	ErrInvalidMessagesArray = errors.New("invalid messages array")

	// ErrInvalidSignaturesCount signals that an invalid number of signatures was provided
	ErrInvalidSignaturesCount = errors.New("invalid signatures count")
)
