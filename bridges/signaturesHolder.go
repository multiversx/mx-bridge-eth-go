package bridges

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-bridge-eth-go/core"
)

type signaturesHolder struct {
	mut            sync.RWMutex
	signedMessages map[string]*core.SignedMessage
	peerMessages   []*core.PeerChainSignature
}

// NewSignatureHolder creates a new signatureHolder
func NewSignatureHolder() *signaturesHolder {
	return &signaturesHolder{
		signedMessages: make(map[string]*core.SignedMessage),
		peerMessages:   make([]*core.PeerChainSignature, 0),
	}
}

// ProcessNewMessage will store the new messages
func (sh *signaturesHolder) ProcessNewMessage(msg *core.SignedMessage, peerMsg *core.PeerChainSignature) {
	if msg == nil || peerMsg == nil {
		return
	}

	sh.mut.Lock()
	defer sh.mut.Unlock()

	sh.signedMessages[msg.UniqueID()] = msg
	sh.peerMessages = append(sh.peerMessages, peerMsg)
}

// AllStoredSignatures will return the stored signatures
func (sh *signaturesHolder) AllStoredSignatures() []*core.SignedMessage {
	sh.mut.RLock()
	defer sh.mut.RUnlock()

	result := make([]*core.SignedMessage, 0, len(sh.signedMessages))
	for _, msg := range sh.signedMessages {
		result = append(result, msg)
	}

	return result
}

// Signatures will provide all gathered signatures for a given message hash
func (sh *signaturesHolder) Signatures(msgHash []byte) [][]byte {
	sh.mut.RLock()
	defer sh.mut.RUnlock()

	uniquePeerSigs := make(map[string]struct{})
	for _, peerMsg := range sh.peerMessages {
		if bytes.Equal(peerMsg.MessageHash, msgHash) {
			uniquePeerSigs[string(peerMsg.Signature)] = struct{}{}
		}
	}

	result := make([][]byte, 0, len(sh.signedMessages))
	for sig := range uniquePeerSigs {
		result = append(result, []byte(sig))
	}

	return result
}

// ClearStoredSignatures will clear any stored signatures
func (sh *signaturesHolder) ClearStoredSignatures() {
	sh.mut.Lock()
	defer sh.mut.Unlock()

	sh.signedMessages = make(map[string]*core.SignedMessage)
	sh.peerMessages = make([]*core.PeerChainSignature, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sh *signaturesHolder) IsInterfaceNil() bool {
	return sh == nil
}
