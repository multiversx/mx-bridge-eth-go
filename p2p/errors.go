package p2p

import "errors"

// ErrPeerNotWhitelisted signals that a peer is not whitelisted
var ErrPeerNotWhitelisted = errors.New("current peer is not whitelisted")

// ErrNilLogger signals that a nil logger was provided
var ErrNilLogger = errors.New("nil logger")

// ErrNilKeyGenerator signals that a nil key generator was provided
var ErrNilKeyGenerator = errors.New("nil key generator")

// ErrNilPrivateKey signals that a nil private key was provided
var ErrNilPrivateKey = errors.New("nil private key")

// ErrNilSingleSigner signals that a nil single signer was provided
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrNilElrondRoleProvider signals that a nil Elrond role provider was given
var ErrNilElrondRoleProvider = errors.New("nil Elrond role provider")

// ErrNilMessenger signals that a nil network messenger was provided
var ErrNilMessenger = errors.New("nil network messenger")

// ErrInvalidSize signals that a slice has an invalid size
var ErrInvalidSize = errors.New("invalid size")

// ErrNilSignatureProcessor signals that a nil signature processor was provided
var ErrNilSignatureProcessor = errors.New("nil signature processor")

// ErrNonceTooLowInReceivedMessage signals that a too low nonce was provided in the message
var ErrNonceTooLowInReceivedMessage = errors.New("nonce too low in received message")

// ErrEmptyName signals that an empty name is not allowed
var ErrEmptyName = errors.New("empty name")

// ErrNilBroadcastClient signals that a nil broadcast client was provided
var ErrNilBroadcastClient = errors.New("nil broadcast client")

// ErrNilStatusHandler signals that a nil status handler was provided
var ErrNilStatusHandler = errors.New("nil status handler")
