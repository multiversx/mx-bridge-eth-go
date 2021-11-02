package p2p

import (
	"bytes"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type noncesHolder struct {
	mut    sync.RWMutex
	nonces map[string]uint64
}

func newNoncesHolder() *noncesHolder {
	return &noncesHolder{
		nonces: make(map[string]uint64),
	}
}

func (nh *noncesHolder) processNonce(msg *core.SignedMessage) error {
	nh.mut.Lock()
	defer nh.mut.Unlock()

	oldNonce := nh.nonces[string(msg.PublicKeyBytes)]
	if oldNonce >= msg.Nonce {
		// only accept newer signatures in order to prevent replay attacks from a malicious relayer that stored old
		// signature messages
		return ErrNonceTooLowInReceivedMessage
	}

	nh.nonces[string(msg.PublicKeyBytes)] = msg.Nonce

	return nil
}

// SortedPublicKeys will return all the sorted public keys contained
func (nh *noncesHolder) SortedPublicKeys() [][]byte {
	nh.mut.RLock()
	defer nh.mut.RUnlock()

	publicKeys := make([][]byte, 0, len(nh.nonces))
	for pk := range nh.nonces {
		publicKeys = append(publicKeys, []byte(pk))
	}

	sort.Slice(publicKeys, func(i, j int) bool {
		return bytes.Compare(publicKeys[i], publicKeys[j]) < 0
	})

	return publicKeys
}
