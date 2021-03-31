package elrond

import "github.com/ElrondNetwork/elrond-eth-bridge/safe"

var (
	_ = safe.Safe(&Client{})
)
