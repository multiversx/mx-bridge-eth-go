package p2p

import "errors"

// ErrPeerNotWhitelisted signals that a peer is not whitelisted
var ErrPeerNotWhitelisted = errors.New("current peer is not whitelisted")
