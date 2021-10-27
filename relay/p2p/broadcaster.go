package p2p

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

const (
	joinTopicName          = "join/1"
	signTopicName          = "sign/1"
	defaultTopicIdentifier = "default"
)

// ArgsBroadcaster is the DTO used in the broadcaster constructor
type ArgsBroadcaster struct {
	Messenger    NetMessenger
	Log          logger.Logger
	RoleProvider RoleProvider
	KeyGen       crypto.KeyGenerator
	SingleSigner crypto.SingleSigner
	PrivateKey   crypto.PrivateKey
}

type broadcaster struct {
	*relayerMessageHandler
	*signaturesHolder
	messenger    NetMessenger
	log          logger.Logger
	roleProvider RoleProvider
}

// NewBroadcaster will create a new broadcaster able to pass messages and signatures
func NewBroadcaster(args ArgsBroadcaster) (*broadcaster, error) {
	// TODO add checks

	b := &broadcaster{
		messenger:        args.Messenger,
		signaturesHolder: newSignatureHolder(),
		log:              args.Log,
		roleProvider:     args.RoleProvider,
		relayerMessageHandler: &relayerMessageHandler{
			marshalizer:  &marshal.JsonMarshalizer{},
			keyGen:       args.KeyGen,
			singleSigner: args.SingleSigner,
			counter:      uint64(time.Now().UnixNano()),
			privateKey:   args.PrivateKey,
		},
	}

	var err error
	pk := b.privateKey.GeneratePublic()
	b.publicKeyBytes, err = pk.ToByteArray()
	if err != nil {
		return nil, err
	}

	return b, err
}

// RegisterOnTopics will register the messenger on all required topics
func (b *broadcaster) RegisterOnTopics() error {
	topics := []string{joinTopicName, signTopicName}
	for _, topic := range topics {
		err := b.messenger.CreateTopic(topic, true)
		if err != nil {
			return err
		}

		err = b.messenger.RegisterMessageProcessor(topic, defaultTopicIdentifier, b)
		if err != nil {
			return err
		}

		b.log.Info("registered", "topic", topic)
	}

	return nil
}

// ProcessReceivedMessage will be called by the network messenger whenever a new message is received
func (b *broadcaster) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	msg, err := b.preProcessMessage(message)
	if err != nil {
		b.log.Debug("got message", "topic", message.Topic(), "error", err)
		return err
	}

	hexPkBytes := hex.EncodeToString(msg.PublicKeyBytes)
	if !b.roleProvider.IsWhitelisted(hexPkBytes) {
		return fmt.Errorf("%w for peer: %s", ErrPeerNotWhitelisted, hexPkBytes)
	}

	addr := data.NewAddressFromBytes(msg.PublicKeyBytes)
	b.log.Debug("got message", "topic", message.Topic(),
		"msg.Payload", msg.Payload, "msg.Nonce", msg.Nonce, "msg.PublicKey", addr.AddressAsBech32String())

	switch message.Topic() {
	case joinTopicName:
		b.addJoinedMessage(msg)
		err = b.broadcastCurrentSignatures(message.Peer())
		if err != nil {
			b.log.Error(err.Error())
		}
	case signTopicName:
		b.addSignedMessage(msg)
	}

	return nil
}

func (b *broadcaster) broadcastCurrentSignatures(peerId core.PeerID) error {
	signedMessages := b.storedSignedMessages()
	for _, msg := range signedMessages {
		err := b.sendSignedMessageToPeer(msg, peerId)
		if err != nil {
			b.log.Debug("error sending current stored signatures",
				"error", err.Error(), "peer", peerId.Pretty())
		}
	}

	return nil
}

func (b *broadcaster) sendSignedMessageToPeer(msg *SignedMessage, peerId core.PeerID) error {
	buff, err := b.marshalizer.Marshal(msg)
	if err != nil {
		return err
	}

	return b.messenger.SendToConnectedPeer(signTopicName, buff, peerId)
}

// BroadcastSignature will send the provided signature as payload in a wrapped signed message to the other peers.
// It will broadcast the message to all available peers
func (b *broadcaster) BroadcastSignature(signature []byte) {
	err := b.broadcastMessage(signature, signTopicName)
	if err != nil {
		b.log.Error("error sending signature", "error", err)
	}
}

// BroadcastJoinTopic will send the provided signature as payload in a wrapped signed message to the other peers.
// It will broadcast the message to all available peers
func (b *broadcaster) BroadcastJoinTopic() {
	err := b.broadcastMessage([]byte("dummy"), joinTopicName)
	if err != nil {
		b.log.Error("error sending signature", "error", err)
	}
}

func (b *broadcaster) broadcastMessage(payload []byte, topic string) error {
	msg, err := b.createMessage(payload)
	if err != nil {
		return err
	}

	buff, err := b.marshalizer.Marshal(msg)
	if err != nil {
		return err
	}

	b.messenger.Broadcast(topic, buff)

	return nil
}

// Close will close any containing members and clean any go routines associated
func (b *broadcaster) Close() error {
	return b.messenger.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *broadcaster) IsInterfaceNil() bool {
	return b == nil
}
