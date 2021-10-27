package p2p

import (
	"encoding/binary"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type relayerMessageHandler struct {
	marshalizer    marshal.Marshalizer
	keyGen         crypto.KeyGenerator
	singleSigner   crypto.SingleSigner
	counter        uint64
	publicKeyBytes []byte
	privateKey     crypto.PrivateKey
}

// preProcessMessage is able to preprocess the received p2p message
func (rmh *relayerMessageHandler) preProcessMessage(message p2p.MessageP2P) (*SignedMessage, error) {
	msg := &SignedMessage{}
	err := rmh.marshalizer.Unmarshal(msg, message.Data())
	if err != nil {
		return nil, err
	}

	pk, err := rmh.keyGen.PublicKeyFromByteArray(msg.PublicKeyBytes)
	if err != nil {
		return nil, err
	}

	buffNonce := make([]byte, 8)
	binary.BigEndian.PutUint64(buffNonce, msg.Nonce)
	msgWithNonce := append(msg.Payload, buffNonce...)

	err = rmh.singleSigner.Verify(pk, msgWithNonce, msg.Signature)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// createMessage will create a new message ready to be broadcast
func (rmh *relayerMessageHandler) createMessage(payload []byte) (*SignedMessage, error) {
	nonce := atomic.AddUint64(&rmh.counter, 1)

	buffNonce := make([]byte, 8)
	binary.BigEndian.PutUint64(buffNonce, nonce)
	msgWithNonce := append(payload, buffNonce...)

	sig, err := rmh.singleSigner.Sign(rmh.privateKey, msgWithNonce)
	if err != nil {
		return nil, err
	}

	return &SignedMessage{
		Payload:        payload,
		PublicKeyBytes: rmh.publicKeyBytes,
		Signature:      sig,
		Nonce:          nonce,
	}, nil
}
