package p2p

import (
	"encoding/binary"
	"fmt"
	"sync/atomic"

	"github.com/multiversx/mx-bridge-eth-go/core"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/factory"
)

const absolutMaxSliceSize = 1024

type relayerMessageHandler struct {
	marshalizer         marshal.Marshalizer
	keyGen              crypto.KeyGenerator
	singleSigner        crypto.SingleSigner
	counter             uint64
	publicKeyBytes      []byte
	privateKey          crypto.PrivateKey
	antifloodComponents *factory.AntiFloodComponents
}

// canProcessMessage will check if a specific message can be processed
func (rmh *relayerMessageHandler) canProcessMessage(message p2p.MessageP2P, fromConnectedPeer chainCore.PeerID) error {
	if check.IfNil(message) {
		return ErrNilMessage
	}
	err := rmh.antifloodComponents.AntiFloodHandler.CanProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return fmt.Errorf("%w on resolver topic %s", err, message.Topic())
	}
	err = rmh.antifloodComponents.AntiFloodHandler.CanProcessMessagesOnTopic(fromConnectedPeer, message.Topic(), 1, uint64(len(message.Data())), message.SeqNo())
	if err != nil {
		return fmt.Errorf("%w on resolver topic %s", err, message.Topic())
	}
	return nil
}

// preProcessMessage is able to preprocess the received p2p message
func (rmh *relayerMessageHandler) preProcessMessage(message p2p.MessageP2P, fromConnectedPeer chainCore.PeerID) (*core.SignedMessage, error) {
	msg := &core.SignedMessage{}
	err := rmh.marshalizer.Unmarshal(msg, message.Data())
	if err != nil {
		reason := "unmarshalable data got on request topic " + message.Topic()
		rmh.antifloodComponents.AntiFloodHandler.BlacklistPeer(message.Peer(), reason, common.InvalidMessageBlacklistDuration)
		rmh.antifloodComponents.AntiFloodHandler.BlacklistPeer(fromConnectedPeer, reason, common.InvalidMessageBlacklistDuration)
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
		reason := "unverifiable signature on request topic " + message.Topic()
		rmh.antifloodComponents.AntiFloodHandler.BlacklistPeer(message.Peer(), reason, common.InvalidMessageBlacklistDuration)
		rmh.antifloodComponents.AntiFloodHandler.BlacklistPeer(fromConnectedPeer, reason, common.InvalidMessageBlacklistDuration)
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
