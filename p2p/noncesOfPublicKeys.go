package p2p

import (
	"bytes"
	"sort"
	"sync"

	"github.com/multiversx/mx-bridge-eth-go/core"
)

type noncesOfPublicKeys struct {
	mut    sync.RWMutex
	nonces map[string]uint64
}

func newNoncesOfPublicKeys() *noncesOfPublicKeys {
	return &noncesOfPublicKeys{
		nonces: make(map[string]uint64),
	}
}

func (holder *noncesOfPublicKeys) processNonce(msg *core.SignedMessage) error {
	holder.mut.Lock()
	defer holder.mut.Unlock()

	oldNonce := holder.nonces[string(msg.PublicKeyBytes)]
	if oldNonce >= msg.Nonce {
		// only accept newer signatures in order to prevent replay attacks from a malicious relayer that stored old
		// signature messages
		return ErrNonceTooLowInReceivedMessage
	}

	holder.nonces[string(msg.PublicKeyBytes)] = msg.Nonce

	return nil
}

// SortedPublicKeys will return all the sorted public keys contained
func (holder *noncesOfPublicKeys) SortedPublicKeys() [][]byte {
	holder.mut.RLock()
	defer holder.mut.RUnlock()

	publicKeys := make([][]byte, 0, len(holder.nonces))
	for pk := range holder.nonces {
		publicKeys = append(publicKeys, []byte(pk))
	}

	sort.Slice(publicKeys, func(i, j int) bool {
		return bytes.Compare(publicKeys[i], publicKeys[j]) < 0
	})

	return publicKeys
}
