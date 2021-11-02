package bridgeExecutors

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type signaturesHolder struct {
	mut            sync.RWMutex
	signedMessages map[string]*core.SignedMessage
	ethMessages    []*core.EthereumSignature
}

func newSignatureHolder() *signaturesHolder {
	return &signaturesHolder{
		signedMessages: make(map[string]*core.SignedMessage),
		ethMessages:    make([]*core.EthereumSignature, 0),
	}
}

// ProcessNewMessage will store the new messages
func (sh *signaturesHolder) ProcessNewMessage(msg *core.SignedMessage, ethMsg *core.EthereumSignature) {
	if msg == nil || ethMsg == nil {
		return
	}

	sh.mut.Lock()
	defer sh.mut.Unlock()

	sh.signedMessages[msg.UniqueID()] = msg
	sh.ethMessages = append(sh.ethMessages, ethMsg)
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

// Signatures will provide all gathered signatures
func (sh *signaturesHolder) Signatures(msgHash []byte) [][]byte {
	sh.mut.RLock()
	defer sh.mut.RUnlock()

	uniqueEthSigs := make(map[string]struct{})
	for _, ethMsg := range sh.ethMessages {
		if bytes.Equal(ethMsg.MessageHash, msgHash) {
			uniqueEthSigs[string(ethMsg.Signature)] = struct{}{}
		}
	}

	result := make([][]byte, 0, len(sh.signedMessages))
	for sig := range uniqueEthSigs {
		result = append(result, []byte(sig))
	}

	return result
}

// clearStoredSignatures will clear any stored signatures
func (sh *signaturesHolder) clearStoredSignatures() {
	sh.mut.Lock()
	defer sh.mut.Unlock()

	sh.signedMessages = make(map[string]*core.SignedMessage)
	sh.ethMessages = make([]*core.EthereumSignature, 0)
}
