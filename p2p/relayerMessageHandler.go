package p2p

import (
	"encoding/binary"
	"fmt"
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/p2p/antiflood"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

const absolutMaxSliceSize = 1024

type relayerMessageHandler struct {
	marshalizer      marshal.Marshalizer
	keyGen           crypto.KeyGenerator
	singleSigner     crypto.SingleSigner
	counter          uint64
	publicKeyBytes   []byte
	privateKey       crypto.PrivateKey
	antifloodHandler antiflood.AntifloodHandler
}

// canProcessMessage will check if a specific message can be processed
func (rmh *relayerMessageHandler) canProcessMessage(message p2p.MessageP2P) error {
	return rmh.antifloodHandler.CanProcessMessagesOnTopic(message.Peer(), message.Topic(), uint32(len(message.Data())))
}

// preProcessMessage is able to preprocess the received p2p message
func (rmh *relayerMessageHandler) preProcessMessage(message p2p.MessageP2P) (*core.SignedMessage, error) {
	msg := &core.SignedMessage{}
	err := rmh.marshalizer.Unmarshal(msg, message.Data())
	if err != nil {
		return nil, err
	}

	err = checkLengths(msg)
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

func checkLengths(msg *core.SignedMessage) error {
	if len(msg.PublicKeyBytes) > absolutMaxSliceSize {
		return fmt.Errorf("%w for PublicKeyBytes field", ErrInvalidSize)
	}
	if len(msg.Signature) > absolutMaxSliceSize {
		return fmt.Errorf("%w for Signature field", ErrInvalidSize)
	}
	if len(msg.Payload) > absolutMaxSliceSize {
		return fmt.Errorf("%w for Payload field", ErrInvalidSize)
	}

	return nil
}

// createMessage will create a new message ready to be broadcast
func (rmh *relayerMessageHandler) createMessage(payload []byte) (*core.SignedMessage, error) {
	nonce := atomic.AddUint64(&rmh.counter, 1)

	buffNonce := make([]byte, 8)
	binary.BigEndian.PutUint64(buffNonce, nonce)
	msgWithNonce := append(payload, buffNonce...)

	sig, err := rmh.singleSigner.Sign(rmh.privateKey, msgWithNonce)
	if err != nil {
		return nil, err
	}

	return &core.SignedMessage{
		Payload:        payload,
		PublicKeyBytes: rmh.publicKeyBytes,
		Signature:      sig,
		Nonce:          nonce,
	}, nil
}
