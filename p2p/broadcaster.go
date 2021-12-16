package p2p

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

const (
	joinTopicSuffix        = "_join"
	signTopicSuffix        = "_sign"
	defaultTopicIdentifier = "default"
	joinTopicMessage       = "join topic"
)

// ArgsBroadcaster is the DTO used in the broadcaster constructor
type ArgsBroadcaster struct {
	Messenger          NetMessenger
	Log                logger.Logger
	ElrondRoleProvider ElrondRoleProvider
	SignatureProcessor SignatureProcessor
	KeyGen             crypto.KeyGenerator
	SingleSigner       crypto.SingleSigner
	PrivateKey         crypto.PrivateKey
	Name               string
}

type broadcaster struct {
	*relayerMessageHandler
	*noncesOfPublicKeys
	messenger          NetMessenger
	log                logger.Logger
	elrondRoleProvider ElrondRoleProvider
	signatureProcessor SignatureProcessor
	name               string
	mutClients         sync.RWMutex
	clients            []core.BroadcastClient
	joinTopicName      string
	signTopicName      string
}

// NewBroadcaster will create a new broadcaster able to pass messages and signatures
func NewBroadcaster(args ArgsBroadcaster) (*broadcaster, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	b := &broadcaster{
		name:               args.Name,
		messenger:          args.Messenger,
		noncesOfPublicKeys: newNoncesOfPublicKeys(),
		log:                args.Log,
		elrondRoleProvider: args.ElrondRoleProvider,
		signatureProcessor: args.SignatureProcessor,
		relayerMessageHandler: &relayerMessageHandler{
			marshalizer:  &marshal.JsonMarshalizer{},
			keyGen:       args.KeyGen,
			singleSigner: args.SingleSigner,
			counter:      uint64(time.Now().UnixNano()),
			privateKey:   args.PrivateKey,
		},
		clients:       make([]core.BroadcastClient, 0),
		joinTopicName: args.Name + joinTopicSuffix,
		signTopicName: args.Name + signTopicSuffix,
	}

	pk := b.privateKey.GeneratePublic()
	b.publicKeyBytes, err = pk.ToByteArray()
	if err != nil {
		return nil, err
	}

	return b, err
}

func checkArgs(args ArgsBroadcaster) error {
	if len(args.Name) == 0 {
		return ErrEmptyName
	}
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}
	if check.IfNil(args.KeyGen) {
		return ErrNilKeyGenerator
	}
	if check.IfNil(args.PrivateKey) {
		return ErrNilPrivateKey
	}
	if check.IfNil(args.SingleSigner) {
		return ErrNilSingleSigner
	}
	if check.IfNil(args.ElrondRoleProvider) {
		return ErrNilElrondRoleProvider
	}
	if check.IfNil(args.Messenger) {
		return ErrNilMessenger
	}
	if check.IfNil(args.SignatureProcessor) {
		return ErrNilSignatureProcessor
	}

	return nil
}

// RegisterOnTopics will register the messenger on all required topics
func (b *broadcaster) RegisterOnTopics() error {
	topics := []string{b.joinTopicName, b.signTopicName}
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
func (b *broadcaster) ProcessReceivedMessage(message p2p.MessageP2P, _ elrondCore.PeerID) error {
	msg, err := b.preProcessMessage(message)
	if err != nil {
		b.log.Debug("got message", "topic", message.Topic(), "error", err)
		return err
	}

	addr := data.NewAddressFromBytes(msg.PublicKeyBytes)
	hexPkBytes := hex.EncodeToString(msg.PublicKeyBytes)
	if !b.elrondRoleProvider.IsWhitelisted(addr) {
		return fmt.Errorf("%w for peer: %s", ErrPeerNotWhitelisted, hexPkBytes)
	}

	b.log.Debug("got message", "topic", message.Topic(),
		"msg.Payload", msg.Payload, "msg.Nonce", msg.Nonce, "msg.PublicKey", addr.AddressAsBech32String())

	err = b.processNonce(msg)
	if err != nil {
		// someone might try to send old, already seen by the network, messages
		// drop the message and do not resend-it to other relayers
		return err
	}

	switch message.Topic() {
	case b.joinTopicName:
		b.processJoinMessage(message)
	case b.signTopicName:
		b.processSignMessage(msg)
	}

	return nil
}

func (b *broadcaster) processJoinMessage(message p2p.MessageP2P) {
	err := b.broadcastCurrentSignatures(message.Peer())
	if err != nil {
		b.log.Error(err.Error())
	}
}

func (b *broadcaster) getEthereumSignature(msg *core.SignedMessage) (*core.EthereumSignature, error) {
	ethSignature := &core.EthereumSignature{}
	err := b.marshalizer.Unmarshal(ethSignature, msg.Payload)
	if err != nil {
		return nil, err
	}

	err = b.signatureProcessor.VerifyEthSignature(ethSignature.Signature, ethSignature.MessageHash)
	if err != nil {
		return nil, err
	}

	return ethSignature, nil
}

func (b *broadcaster) processSignMessage(msg *core.SignedMessage) {
	ethSignature, err := b.getEthereumSignature(msg)
	if err != nil {
		b.log.Debug("received message does not contain a valid signature", "error", err)
		return
	}

	b.notifyClients(msg, ethSignature)
}

func (b *broadcaster) notifyClients(msg *core.SignedMessage, ethMsg *core.EthereumSignature) {
	b.mutClients.RLock()
	defer b.mutClients.RUnlock()

	for _, client := range b.clients {
		client.ProcessNewMessage(msg, ethMsg)
	}
}

func (b *broadcaster) broadcastCurrentSignatures(peerId elrondCore.PeerID) error {
	allMessages := b.retrieveUniqueMessages()

	for _, msg := range allMessages {
		err := b.sendSignedMessageToPeer(msg, peerId)
		if err != nil {
			b.log.Debug("error sending current stored signatures",
				"error", err.Error(), "peer", peerId.Pretty())
		}
	}

	return nil
}

func (b *broadcaster) retrieveUniqueMessages() map[string]*core.SignedMessage {
	allMessages := make(map[string]*core.SignedMessage)
	for _, client := range b.clients {
		messages := client.AllStoredSignatures()
		for _, msg := range messages {
			allMessages[msg.UniqueID()] = msg
		}
	}

	return allMessages
}

func (b *broadcaster) sendSignedMessageToPeer(msg *core.SignedMessage, peerId elrondCore.PeerID) error {
	buff, err := b.marshalizer.Marshal(msg)
	if err != nil {
		return err
	}

	return b.messenger.SendToConnectedPeer(b.signTopicName, buff, peerId)
}

// BroadcastSignature will send the provided signature as payload in a wrapped signed message to the other peers.
// It will broadcast the message to all available peers
func (b *broadcaster) BroadcastSignature(signature []byte, messageHash []byte) {
	ethSig := &core.EthereumSignature{
		Signature:   signature,
		MessageHash: messageHash,
	}

	payload, err := b.marshalizer.Marshal(ethSig)
	if err != nil {
		b.log.Error("error creating signature payload", "error", err)
	}

	err = b.broadcastMessage(payload, b.signTopicName)
	if err != nil {
		b.log.Error("error sending signature", "error", err)
	}
}

// BroadcastJoinTopic will send the provided signature as payload in a wrapped signed message to the other peers.
// It will broadcast the message to all available peers
func (b *broadcaster) BroadcastJoinTopic() {
	err := b.broadcastMessage([]byte(joinTopicMessage), b.joinTopicName)
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

// AddBroadcastClient will add a client to the list so it can be notified of the newly received
// messages
func (b *broadcaster) AddBroadcastClient(client core.BroadcastClient) error {
	if check.IfNil(client) {
		return ErrNilBroadcastClient
	}

	b.mutClients.Lock()
	b.clients = append(b.clients, client)
	b.mutClients.Unlock()

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
