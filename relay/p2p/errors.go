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

// ErrNilRoleProvider signals that a nil role provider was given
var ErrNilRoleProvider = errors.New("nil role provider")

// ErrNilMessenger signals that a nil network messenger was provided
var ErrNilMessenger = errors.New("nil network messenger")
