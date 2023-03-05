package testsCommon

import "github.com/multiversx/mx-bridge-eth-go/core"

// BroadcastClientStub -
type BroadcastClientStub struct {
	ProcessNewMessageCalled   func(msg *core.SignedMessage, ethMsg *core.EthereumSignature)
	AllStoredSignaturesCalled func() []*core.SignedMessage
}

// ProcessNewMessage -
func (stub *BroadcastClientStub) ProcessNewMessage(msg *core.SignedMessage, ethMsg *core.EthereumSignature) {
	if stub.ProcessNewMessageCalled != nil {
		stub.ProcessNewMessageCalled(msg, ethMsg)
	}
}

// AllStoredSignatures -
func (stub *BroadcastClientStub) AllStoredSignatures() []*core.SignedMessage {
	if stub.AllStoredSignaturesCalled != nil {
		return stub.AllStoredSignaturesCalled()
	}

	return make([]*core.SignedMessage, 0)
}

// IsInterfaceNil -
func (stub *BroadcastClientStub) IsInterfaceNil() bool {
	return stub == nil
}
