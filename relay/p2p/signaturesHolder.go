package p2p

import (
	"bytes"
	"sort"
	"sync"
)

type signaturesHolder struct {
	mut            sync.RWMutex
	signedMessages map[string]*SignedMessage
	nonces         map[string]uint64
}

func newSignatureHolder() *signaturesHolder {
	return &signaturesHolder{
		signedMessages: make(map[string]*SignedMessage),
		nonces:         make(map[string]uint64),
	}
}

func (sh *signaturesHolder) addSignedMessage(msg *SignedMessage) {
	sh.mut.Lock()
	defer sh.mut.Unlock()

	oldNonce := sh.nonces[string(msg.PublicKeyBytes)]
	if oldNonce > msg.Nonce {
		// only accept newer signatures in order to prevent replay attacks from a malicious relayer that stored old
		// signature messages
		return
	}

	sh.nonces[string(msg.PublicKeyBytes)] = msg.Nonce
	sh.signedMessages[string(msg.PublicKeyBytes)] = msg
}

func (sh *signaturesHolder) addJoinedMessage(msg *SignedMessage) {
	sh.mut.Lock()
	defer sh.mut.Unlock()

	oldNonce := sh.nonces[string(msg.PublicKeyBytes)]
	if oldNonce > msg.Nonce {
		// only accept newer signatures in order to prevent replay attacks from a malicious relayer that stored old
		// signature messages
		return
	}

	sh.nonces[string(msg.PublicKeyBytes)] = msg.Nonce
}

// ClearSignatures will clear any stored signatures
func (sh *signaturesHolder) ClearSignatures() {
	sh.mut.Lock()
	defer sh.mut.Unlock()

	sh.signedMessages = make(map[string]*SignedMessage)
}

// Signatures will provide all gathered signatures
func (sh *signaturesHolder) Signatures() [][]byte {
	sh.mut.RLock()
	defer sh.mut.RUnlock()

	result := make([][]byte, 0, len(sh.signedMessages))

	// the ethereum signatures are stored in the Payload field. The Signature field is the ed25519 sig applied
	// over the Payload and Nonce
	for _, msg := range sh.signedMessages {
		result = append(result, msg.Payload)
	}

	return result
}

func (sh *signaturesHolder) storedSignedMessages() []*SignedMessage {
	sh.mut.RLock()
	defer sh.mut.RUnlock()

	result := make([]*SignedMessage, 0, len(sh.signedMessages))

	// the ethereum signatures are stored in the Payload field. The Signature field is the ed25519 sig applied
	// over the Payload and Nonce
	for _, msg := range sh.signedMessages {
		result = append(result, msg)
	}

	return result
}

// SortedPublicKeys will return all the sorted public keys contained
func (sh *signaturesHolder) SortedPublicKeys() [][]byte {
	sh.mut.RLock()
	defer sh.mut.RUnlock()

	publicKeys := make([][]byte, 0, len(sh.nonces))
	for pk := range sh.nonces {
		publicKeys = append(publicKeys, []byte(pk))
	}

	sort.Slice(publicKeys, func(i, j int) bool {
		return bytes.Compare(publicKeys[i], publicKeys[j]) < 0
	})

	return publicKeys
}
