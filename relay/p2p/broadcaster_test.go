package p2p

import (
	"errors"
	"testing"

	cryptoMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/crypto"
	p2pMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/p2p"
	roleProvidersMock "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/roleProviders"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	ergoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsBroadcaster() ArgsBroadcaster {
	return ArgsBroadcaster{
		Messenger:    &p2pMocks.MessengerStub{},
		Log:          logger.GetOrCreate("test"),
		RoleProvider: &roleProvidersMock.ElrondRoleProviderStub{},
		KeyGen:       &cryptoMocks.KeyGenStub{},
		SingleSigner: &cryptoMocks.SingleSignerStub{},
		PrivateKey:   &cryptoMocks.PrivateKeyStub{},
	}
}

func TestNewBroadcaster(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.Log = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil key gen should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.KeyGen = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilKeyGenerator, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.PrivateKey = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilPrivateKey, err)
	})
	t.Run("nil single signer should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.SingleSigner = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilSingleSigner, err)
	})
	t.Run("nil role provider should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.RoleProvider = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilRoleProvider, err)
	})
	t.Run("nil messenger should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.Messenger = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilMessenger, err)
	})
	t.Run("public key conversion fails", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		expectedErr := errors.New("expected error")
		args.PrivateKey = &cryptoMocks.PrivateKeyStub{
			GeneratePublicCalled: func() crypto.PublicKey {
				return &cryptoMocks.PublicKeyStub{
					ToByteArrayCalled: func() ([]byte, error) {
						return nil, expectedErr
					},
				}
			},
		}

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsBroadcaster()

		b, err := NewBroadcaster(args)
		assert.False(t, check.IfNil(b))
		assert.Nil(t, err)
	})
}

func TestBroadcaster_RegisterOnTopics(t *testing.T) {
	t.Parallel()

	t.Run("create topic errors should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		expectedErr := errors.New("expected error")
		args.Messenger = &p2pMocks.MessengerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				return expectedErr
			},
		}

		b, _ := NewBroadcaster(args)
		err := b.RegisterOnTopics()

		require.Equal(t, expectedErr, err)
	})
	t.Run("register errors should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		expectedErr := errors.New("expected error")
		args.Messenger = &p2pMocks.MessengerStub{
			RegisterMessageProcessorCalled: func(topic string, identifier string, processor p2p.MessageProcessor) error {
				return expectedErr
			},
		}

		b, _ := NewBroadcaster(args)
		err := b.RegisterOnTopics()

		require.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		createTopics := make(map[string]int)
		register := make(map[string]int)
		args.Messenger = &p2pMocks.MessengerStub{
			CreateTopicCalled: func(name string, createChannelForTopic bool) error {
				createTopics[name]++
				return nil
			},
			RegisterMessageProcessorCalled: func(topic string, identifier string, processor p2p.MessageProcessor) error {
				register[topic]++
				return nil
			},
		}

		b, _ := NewBroadcaster(args)
		err := b.RegisterOnTopics()

		require.Nil(t, err)
		topics := []string{joinTopicName, signTopicName}
		for _, topic := range topics {
			assert.Equal(t, 1, createTopics[topic])
			assert.Equal(t, 1, register[topic])
		}
	})
}

func TestBroadcaster_ProcessReceivedMessage(t *testing.T) {
	t.Parallel()

	t.Run("pre process fails", func(t *testing.T) {
		args := createMockArgsBroadcaster()

		b, _ := NewBroadcaster(args)
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField: []byte("gibberish"),
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.NotNil(t, err)
	})
	t.Run("public key not whitelisted", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		isWhiteListedCalled := false
		msg, buff := createSignedMessageAndMarshaledBytes()

		args.RoleProvider = &roleProvidersMock.ElrondRoleProviderStub{
			IsWhitelistedCalled: func(address ergoCore.AddressHandler) bool {
				assert.Equal(t, msg.PublicKeyBytes, address.AddressBytes())
				isWhiteListedCalled = true
				return false
			},
		}

		b, _ := NewBroadcaster(args)
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField: buff,
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.True(t, errors.Is(err, ErrPeerNotWhitelisted))
		assert.True(t, isWhiteListedCalled)
	})
	t.Run("joined topic should save message and send stored messages", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		msg1, buff1 := createSignedMessageAndMarshaledBytesWithValues([]byte("payload1"), []byte("pk1"))
		sendWasCalled := false
		pid := core.PeerID("pid1")
		args.Messenger = &p2pMocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID core.PeerID) error {
				assert.Equal(t, signTopicName, topic)
				assert.Equal(t, pid, peerID)
				assert.Equal(t, buff1, buff) // test that the original, stored message is sent
				sendWasCalled = true

				return nil
			},
		}

		b, _ := NewBroadcaster(args)
		b.addSignedMessage(msg1)

		msg2, buff2 := createSignedMessageAndMarshaledBytesWithValues([]byte("payload2"), []byte("pk2"))
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField:  buff2,
			TopicField: joinTopicName,
			PeerField:  pid,
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)

		assert.Equal(t, [][]byte{msg1.PublicKeyBytes, msg2.PublicKeyBytes}, b.SortedPublicKeys())
	})
	t.Run("sign should store message", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		msg1, buff1 := createSignedMessageAndMarshaledBytesWithValues([]byte("payload1"), []byte("pk1"))
		msg2, buff2 := createSignedMessageAndMarshaledBytesWithValues([]byte("payload2"), []byte("pk2"))
		args.Messenger = &p2pMocks.MessengerStub{}

		b, _ := NewBroadcaster(args)
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField:  buff2,
			TopicField: signTopicName,
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)

		p2pMsg = &p2pMocks.P2PMessageMock{
			DataField:  buff1,
			TopicField: signTopicName,
		}

		err = b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)

		assert.Equal(t, [][]byte{msg1.PublicKeyBytes, msg2.PublicKeyBytes}, b.SortedPublicKeys())
	})
}

func TestBroadcaster_BroadcastJoinTopic(t *testing.T) {
	t.Parallel()

	broadcastCalled := false
	marshalizer := &marshal.JsonMarshalizer{}
	sig := []byte("signature")
	args := createMockArgsBroadcaster()
	args.SingleSigner = &cryptoMocks.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return sig, nil
		},
	}
	args.Messenger = &p2pMocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			broadcastCalled = true
			assert.Equal(t, joinTopicName, topic)

			msg := &SignedMessage{}
			err := marshalizer.Unmarshal(msg, buff)
			require.Nil(t, err)
			assert.Equal(t, sig, msg.Signature)
			assert.Equal(t, []byte(joinTopicMessage), msg.Payload)
		},
	}
	b, _ := NewBroadcaster(args)

	b.BroadcastJoinTopic()
	assert.True(t, broadcastCalled)
}

func TestBroadcaster_BroadcastSignature(t *testing.T) {
	t.Parallel()

	broadcastCalled := false
	marshalizer := &marshal.JsonMarshalizer{}
	sig := []byte("signature")
	externalSignature := []byte("external signature")
	args := createMockArgsBroadcaster()
	args.SingleSigner = &cryptoMocks.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return sig, nil
		},
	}
	args.Messenger = &p2pMocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			broadcastCalled = true
			assert.Equal(t, signTopicName, topic)

			msg := &SignedMessage{}
			err := marshalizer.Unmarshal(msg, buff)
			require.Nil(t, err)
			assert.Equal(t, sig, msg.Signature)
			assert.Equal(t, externalSignature, msg.Payload)
		},
	}
	b, _ := NewBroadcaster(args)

	b.BroadcastSignature(externalSignature)
	assert.True(t, broadcastCalled)
}

func TestBroadcaster_Close(t *testing.T) {
	t.Parallel()

	closeWasCalled := true
	args := createMockArgsBroadcaster()
	args.Messenger = &p2pMocks.MessengerStub{
		CloseCalled: func() error {
			closeWasCalled = true
			return nil
		},
	}
	b, _ := NewBroadcaster(args)
	err := b.Close()

	assert.Nil(t, err)
	assert.True(t, closeWasCalled)
}
