package integrationTests

import (
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/p2p"
)

type messengerWrapper struct {
	p2p.Messenger
}

// ConnectTo will try to initiate a connection to the provided parameter
func (mw *messengerWrapper) ConnectTo(connectable Connectable) error {
	if check.IfNil(connectable) {
		return fmt.Errorf("trying to connect to a nil Connectable parameter")
	}

	return mw.ConnectToPeer(connectable.GetConnectableAddress())
}

// GetConnectableAddress returns a non circuit, non windows default connectable p2p address
func (mw *messengerWrapper) GetConnectableAddress() string {
	if mw == nil {
		return "nil"
	}

	return getConnectableAddress(mw)
}

// GetConnectableAddress returns a non circuit, non windows default connectable address for provided messenger
func getConnectableAddress(mes p2p.Messenger) string {
	for _, addr := range mes.Addresses() {
		if strings.Contains(addr, "circuit") || strings.Contains(addr, "169.254") {
			continue
		}
		return addr
	}
	return ""
}
