package testsCommon

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-bridge-eth-go/core"
)

// SignaturesHolderMock -
type SignaturesHolderMock struct {
	mut            sync.RWMutex
	signedMessages map[string]*core.SignedMessage
	ethMessages    []*core.EthereumSignature
}

// NewSignaturesHolderMock -
func NewSignaturesHolderMock() *SignaturesHolderMock {
	return &SignaturesHolderMock{
		signedMessages: make(map[string]*core.SignedMessage),
		ethMessages:    make([]*core.EthereumSignature, 0),
	}
}

// ProcessNewMessage will store the new messages
func (mock *SignaturesHolderMock) ProcessNewMessage(msg *core.SignedMessage, ethMsg *core.EthereumSignature) {
	mock.mut.Lock()
	defer mock.mut.Unlock()

	mock.signedMessages[msg.UniqueID()] = msg
	mock.ethMessages = append(mock.ethMessages, ethMsg)
}

// AllStoredSignatures will return the stored signatures
func (mock *SignaturesHolderMock) AllStoredSignatures() []*core.SignedMessage {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	result := make([]*core.SignedMessage, 0, len(mock.signedMessages))
	for _, msg := range mock.signedMessages {
		result = append(result, msg)
	}

	return result
}

// Signatures will provide all gathered signatures
func (mock *SignaturesHolderMock) Signatures(msgHash []byte) [][]byte {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	uniqueEthSigs := make(map[string]struct{})
	for _, ethMsg := range mock.ethMessages {
		if bytes.Equal(ethMsg.MessageHash, msgHash) {
			uniqueEthSigs[string(ethMsg.Signature)] = struct{}{}
		}
	}

	result := make([][]byte, 0, len(mock.signedMessages))
	for sig := range uniqueEthSigs {
		result = append(result, []byte(sig))
	}

	return result
}

// ClearStoredSignatures -
func (mock *SignaturesHolderMock) ClearStoredSignatures() {
	mock.mut.Lock()
	defer mock.mut.Unlock()

	mock.signedMessages = make(map[string]*core.SignedMessage)
	mock.ethMessages = make([]*core.EthereumSignature, 0)
}

// IsInterfaceNil -
func (mock *SignaturesHolderMock) IsInterfaceNil() bool {
	return mock == nil
}
