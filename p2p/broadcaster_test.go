package p2p

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	cryptoMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/crypto"
	p2pMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/p2p"
	roleProvidersMock "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/roleProviders"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsBroadcaster() ArgsBroadcaster {
	return ArgsBroadcaster{
		Messenger:          &p2pMocks.MessengerStub{},
		Log:                logger.GetOrCreate("test"),
		ElrondRoleProvider: &roleProvidersMock.ElrondRoleProviderStub{},
		KeyGen:             &cryptoMocks.KeyGenStub{},
		SingleSigner:       &cryptoMocks.SingleSignerStub{},
		PrivateKey:         &cryptoMocks.PrivateKeyStub{},
		SignatureProcessor: &testsCommon.SignatureProcessorStub{},
		Name:               "test",
		AntifloodConfig: config.TopicsAntifloodConfig{
			DefaultMaxMessagesPerInterval: 15000,
			IntervalDuration:              time.Second,
			MaxMessages: []config.TopicMaxMessagesConfig{
				{
					Topic: "test" + signTopicSuffix,
				},
				{
					Topic: "test" + joinTopicSuffix,
				},
			},
		},
	}
}

func TestNewBroadcaster(t *testing.T) {
	t.Parallel()

	t.Run("empty name should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.Name = ""

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrEmptyName, err)
	})
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
	t.Run("nil Elrond role provider should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.ElrondRoleProvider = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilElrondRoleProvider, err)
	})
	t.Run("nil messenger should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.Messenger = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilMessenger, err)
	})
	t.Run("nil signature processor should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.SignatureProcessor = nil

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.Equal(t, ErrNilSignatureProcessor, err)
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
	t.Run("invalid default value on antiflood handler should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		args.AntifloodConfig.DefaultMaxMessagesPerInterval = 0

		b, err := NewBroadcaster(args)
		assert.True(t, check.IfNil(b))
		assert.True(t, strings.Contains(err.Error(), "invalid number of messages"))
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
		topics := []string{args.Name + joinTopicSuffix, args.Name + signTopicSuffix}
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
		msg, buff := createSignedMessageAndMarshaledBytes(0)

		args.ElrondRoleProvider = &roleProvidersMock.ElrondRoleProviderStub{
			IsWhitelistedCalled: func(address erdgoCore.AddressHandler) bool {
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
	t.Run("invalid nonce should error", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		msg, buff := createSignedMessageAndMarshaledBytes(0)

		args.ElrondRoleProvider = &roleProvidersMock.ElrondRoleProviderStub{}

		b, _ := NewBroadcaster(args)
		b.nonces[string(msg.PublicKeyBytes)] = msg.Nonce + 1
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField: buff,
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.Equal(t, ErrNonceTooLowInReceivedMessage, err)

		b.nonces[string(msg.PublicKeyBytes)] = msg.Nonce
		err = b.ProcessReceivedMessage(p2pMsg, "")
		assert.Equal(t, ErrNonceTooLowInReceivedMessage, err)
	})
	t.Run("joined topic should send stored messages from clients", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		msg1, buff1 := createSignedMessageForEthSig(0)

		client := &testsCommon.BroadcastClientStub{
			AllStoredSignaturesCalled: func() []*core.SignedMessage {
				return []*core.SignedMessage{msg1}
			},
		}

		sendWasCalled := false
		pid := elrondCore.PeerID("pid1")
		args.Messenger = &p2pMocks.MessengerStub{
			SendToConnectedPeerCalled: func(topic string, buff []byte, peerID elrondCore.PeerID) error {
				assert.Equal(t, args.Name+signTopicSuffix, topic)
				assert.Equal(t, pid, peerID)
				assert.Equal(t, buff1, buff) // test that the original, stored message is sent
				sendWasCalled = true

				return nil
			},
		}

		b, _ := NewBroadcaster(args)
		err := b.AddBroadcastClient(client)
		require.Nil(t, err)
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField:  buff1,
			TopicField: args.Name + signTopicSuffix,
			PeerField:  pid,
		}
		_ = b.ProcessReceivedMessage(p2pMsg, "")

		msg2, buff2 := createSignedMessageAndMarshaledBytes(1)
		p2pMsg = &p2pMocks.P2PMessageMock{
			DataField:  buff2,
			TopicField: args.Name + joinTopicSuffix,
			PeerField:  pid,
		}

		err = b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)

		assert.Equal(t, [][]byte{msg1.PublicKeyBytes, msg2.PublicKeyBytes}, b.SortedPublicKeys())
	})
	t.Run("not a valid signature as payload (unmarshalled failed) should add the message's nonce", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		_, buff1 := createSignedMessageAndMarshaledBytes(0)
		_, buff2 := createSignedMessageAndMarshaledBytes(1)
		args.Messenger = &p2pMocks.MessengerStub{}
		args.SignatureProcessor = &testsCommon.SignatureProcessorStub{}

		b, _ := NewBroadcaster(args)
		_ = b.AddBroadcastClient(&testsCommon.BroadcastClientStub{
			ProcessNewMessageCalled: func(msg *core.SignedMessage, ethMsg *core.EthereumSignature) {
				require.Fail(t, "should have not called process")
			},
		})
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField:  buff2,
			TopicField: args.Name + signTopicSuffix,
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)

		p2pMsg = &p2pMocks.P2PMessageMock{
			DataField:  buff1,
			TopicField: args.Name + signTopicSuffix,
		}

		err = b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)

		assert.Equal(t, 2, len(b.SortedPublicKeys()))
	})
	t.Run("not a valid signature as payload (verify failed) should add the message's nonce", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		_, buff1 := createSignedMessageForEthSig(0)
		args.Messenger = &p2pMocks.MessengerStub{}
		args.SignatureProcessor = &testsCommon.SignatureProcessorStub{
			VerifyEthSignatureCalled: func(signature []byte, messageHash []byte) error {
				return errors.New("invalid signature as payload")
			},
		}

		b, _ := NewBroadcaster(args)
		_ = b.AddBroadcastClient(&testsCommon.BroadcastClientStub{
			ProcessNewMessageCalled: func(msg *core.SignedMessage, ethMsg *core.EthereumSignature) {
				require.Fail(t, "should have not called process")
			},
		})
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField:  buff1,
			TopicField: args.Name + signTopicSuffix,
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)

		assert.Equal(t, 1, len(b.SortedPublicKeys()))
	})
	t.Run("system busy on a topic", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		_, buff1 := createSignedMessageForEthSig(0)
		_, buff2 := createSignedMessageForEthSig(1)
		args.Messenger = &p2pMocks.MessengerStub{}
		args.AntifloodConfig.MaxMessages = []config.TopicMaxMessagesConfig{
			{
				Topic:                  args.Name + signTopicSuffix,
				NumMessagesPerInterval: uint32(len(buff2)),
			},
		}

		processedMessages := make([]*core.SignedMessage, 0)
		b, _ := NewBroadcaster(args)
		_ = b.AddBroadcastClient(&testsCommon.BroadcastClientStub{
			ProcessNewMessageCalled: func(msg *core.SignedMessage, ethMsg *core.EthereumSignature) {
				processedMessages = append(processedMessages, msg)
			},
		})
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField:  buff2,
			TopicField: args.Name + signTopicSuffix,
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)

		p2pMsg = &p2pMocks.P2PMessageMock{
			DataField:  buff1,
			TopicField: args.Name + signTopicSuffix,
		}

		err = b.ProcessReceivedMessage(p2pMsg, "")
		assert.True(t, strings.Contains(err.Error(), "system busy"))
	})
	t.Run("sign should store message", func(t *testing.T) {
		args := createMockArgsBroadcaster()
		msg1, buff1 := createSignedMessageForEthSig(0)
		msg2, buff2 := createSignedMessageForEthSig(1)
		args.Messenger = &p2pMocks.MessengerStub{}

		processedMessages := make([]*core.SignedMessage, 0)
		b, _ := NewBroadcaster(args)
		_ = b.AddBroadcastClient(&testsCommon.BroadcastClientStub{
			ProcessNewMessageCalled: func(msg *core.SignedMessage, ethMsg *core.EthereumSignature) {
				processedMessages = append(processedMessages, msg)
			},
		})
		p2pMsg := &p2pMocks.P2PMessageMock{
			DataField:  buff2,
			TopicField: args.Name + signTopicSuffix,
		}

		err := b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)

		p2pMsg = &p2pMocks.P2PMessageMock{
			DataField:  buff1,
			TopicField: args.Name + signTopicSuffix,
		}

		err = b.ProcessReceivedMessage(p2pMsg, "")
		assert.Nil(t, err)

		assert.Equal(t, [][]byte{msg1.PublicKeyBytes, msg2.PublicKeyBytes}, b.SortedPublicKeys())
		assert.Equal(t, []*core.SignedMessage{msg2, msg1}, processedMessages)
	})
}

func TestBroadcaster_BroadcastJoinTopic(t *testing.T) {
	t.Parallel()

	broadcastCalled := false
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
			assert.Equal(t, args.Name+joinTopicSuffix, topic)

			msg := &core.SignedMessage{}
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
	sig := []byte("signature")
	ethSig := []byte("eth signature")
	ethMsg := []byte("eth message")
	args := createMockArgsBroadcaster()
	args.SingleSigner = &cryptoMocks.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return sig, nil
		},
	}
	args.Messenger = &p2pMocks.MessengerStub{
		BroadcastCalled: func(topic string, buff []byte) {
			broadcastCalled = true
			assert.Equal(t, args.Name+signTopicSuffix, topic)

			msg := &core.SignedMessage{}
			err := marshalizer.Unmarshal(msg, buff)
			require.Nil(t, err)
			assert.Equal(t, sig, msg.Signature)

			ethMsgInstance := &core.EthereumSignature{}
			err = marshalizer.Unmarshal(ethMsgInstance, msg.Payload)
			require.Nil(t, err)

			assert.Equal(t, ethSig, ethMsgInstance.Signature)
			assert.Equal(t, ethMsg, ethMsgInstance.MessageHash)
		},
	}
	b, _ := NewBroadcaster(args)

	b.BroadcastSignature(ethSig, ethMsg)
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

func TestBroadcaster_AddBroadcastClientNilClient(t *testing.T) {
	t.Parallel()

	args := createMockArgsBroadcaster()
	b, _ := NewBroadcaster(args)

	err := b.AddBroadcastClient(nil)
	assert.Equal(t, ErrNilBroadcastClient, err)
}

func TestBroadcaster_ShouldFilterIdenticalMessages(t *testing.T) {
	t.Parallel()

	msg1, _ := createSignedMessageAndMarshaledBytes(1)
	msg2, _ := createSignedMessageAndMarshaledBytes(2)
	msg3, _ := createSignedMessageAndMarshaledBytes(3)

	client1 := &testsCommon.BroadcastClientStub{
		AllStoredSignaturesCalled: func() []*core.SignedMessage {
			return []*core.SignedMessage{msg1, msg2}
		},
	}
	client2 := &testsCommon.BroadcastClientStub{
		AllStoredSignaturesCalled: func() []*core.SignedMessage {
			return []*core.SignedMessage{msg2, msg3}
		},
	}

	args := createMockArgsBroadcaster()
	b, _ := NewBroadcaster(args)

	_ = b.AddBroadcastClient(client1)
	_ = b.AddBroadcastClient(client2)

	uniqueMessages := b.retrieveUniqueMessages()
	testSliceInMap(t, []*core.SignedMessage{msg1, msg2, msg3}, uniqueMessages)
}

func testSliceInMap(t *testing.T, slice []*core.SignedMessage, m map[string]*core.SignedMessage) {
	assert.Equal(t, len(slice), len(m))
	for _, msgSlice := range slice {
		found := false
		for _, msgMap := range m {
			if reflect.DeepEqual(msgSlice, msgMap) {
				found = true
				break
			}
		}

		require.True(t, found)
	}
}
