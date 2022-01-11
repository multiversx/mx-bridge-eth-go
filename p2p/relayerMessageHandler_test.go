package p2p

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	cryptoMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/crypto"
	p2pMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/p2p"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var marshalizer = &testsCommon.MarshalizerMock{}

const fromPeer = "from-peer"
const pid = elrondCore.PeerID("pid1")

func TestRelayerMessageHandler_preProcess(t *testing.T) {
	t.Parallel()

	t.Run("preProcess errors if unmarshal fails", preProcessUnmarshal)
	t.Run("preProcess errors if fields lengths exceeds the limit", preProcessLimits)
	t.Run("preProcess errors if keygen fails", preProcessKeygenFails)
	t.Run("preProcess errors if verify fails", preProcessVerifyFails)
	t.Run("preProcess should work", preProcessShouldWork)
}

func preProcessUnmarshal(t *testing.T) {
	blackList := make(map[elrondCore.PeerID]string)
	rmh := &relayerMessageHandler{
		marshalizer:  &testsCommon.MarshalizerMock{},
		singleSigner: &cryptoMocks.SingleSignerStub{},
		antifloodComponents: &factory.AntiFloodComponents{
			AntiFloodHandler: &mock.P2PAntifloodHandlerStub{
				BlacklistPeerCalled: func(peer elrondCore.PeerID, reason string, duration time.Duration) {
					blackList[peer] = reason
				},
			},
		},
	}
	p2pmsg := &p2pMocks.P2PMessageMock{
		PeerField: pid,
		DataField: []byte("gibberish"),
	}

	msg, err := rmh.preProcessMessage(p2pmsg, fromPeer)
	assert.Nil(t, msg)
	assert.NotNil(t, err)

	assert.Equal(t, 2, len(blackList))
	reason, ok := blackList[pid]
	assert.True(t, ok)
	assert.True(t, strings.Contains(reason, "unmarshalable data"))
	reason, ok = blackList[fromPeer]
	assert.True(t, ok)
	assert.True(t, strings.Contains(reason, "unmarshalable data"))
}

func preProcessLimits(t *testing.T) {
	rmh := &relayerMessageHandler{
		marshalizer:  &testsCommon.MarshalizerMock{},
		singleSigner: &cryptoMocks.SingleSignerStub{},
		keyGen:       &cryptoMocks.KeyGenStub{},
	}

	largeBuff := bytes.Repeat([]byte{1}, absolutMaxSliceSize+1)
	err := preProcessMessageInvalidLimits(t, rmh, []byte("payload"), largeBuff, []byte("sig"))
	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "PublicKeyBytes"))

	err = preProcessMessageInvalidLimits(t, rmh, largeBuff, []byte("pk"), []byte("sig"))
	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Payload"))

	err = preProcessMessageInvalidLimits(t, rmh, []byte("payload"), []byte("pk"), largeBuff)
	require.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "Signature"))
}

func preProcessMessageInvalidLimits(
	t *testing.T,
	rmh *relayerMessageHandler,
	payload []byte,
	pubKey []byte,
	sig []byte,
) error {
	msg := &core.SignedMessage{
		Payload:        payload,
		PublicKeyBytes: pubKey,
		Signature:      sig,
		Nonce:          34,
	}
	buff, _ := marshalizer.Marshal(msg)

	p2pmsg := &p2pMocks.P2PMessageMock{
		DataField: buff,
	}

	msg, err := rmh.preProcessMessage(p2pmsg, fromPeer)
	require.Nil(t, msg)
	assert.True(t, errors.Is(err, ErrInvalidSize))

	return err
}

func preProcessKeygenFails(t *testing.T) {
	expectedErr := errors.New("expected error")
	rmh := &relayerMessageHandler{
		marshalizer:  &testsCommon.MarshalizerMock{},
		singleSigner: &cryptoMocks.SingleSignerStub{},
		keyGen: &cryptoMocks.KeyGenStub{
			PublicKeyFromByteArrayStub: func(b []byte) (crypto.PublicKey, error) {
				return nil, expectedErr
			},
		},
	}
	_, buff := createSignedMessageAndMarshaledBytes(0)

	p2pmsg := &p2pMocks.P2PMessageMock{
		DataField: buff,
	}

	msg, err := rmh.preProcessMessage(p2pmsg, fromPeer)
	assert.Nil(t, msg)
	assert.Equal(t, expectedErr, err)
}

func preProcessVerifyFails(t *testing.T) {
	blackList := make(map[elrondCore.PeerID]string)
	expectedErr := errors.New("expected error")
	rmh := &relayerMessageHandler{
		marshalizer: &testsCommon.MarshalizerMock{},
		singleSigner: &cryptoMocks.SingleSignerStub{
			VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				return expectedErr
			},
		},
		keyGen: &cryptoMocks.KeyGenStub{},
		antifloodComponents: &factory.AntiFloodComponents{
			AntiFloodHandler: &mock.P2PAntifloodHandlerStub{
				BlacklistPeerCalled: func(peer elrondCore.PeerID, reason string, duration time.Duration) {
					blackList[peer] = reason
				},
			},
		},
	}
	_, buff := createSignedMessageAndMarshaledBytes(0)

	p2pmsg := &p2pMocks.P2PMessageMock{
		PeerField: pid,
		DataField: buff,
	}

	msg, err := rmh.preProcessMessage(p2pmsg, fromPeer)
	assert.Nil(t, msg)
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, 2, len(blackList))
	reason, ok := blackList[pid]
	assert.True(t, ok)
	assert.True(t, strings.Contains(reason, "unverifiable signature"))
	reason, ok = blackList[fromPeer]
	assert.True(t, ok)
	assert.True(t, strings.Contains(reason, "unverifiable signature"))
}

func preProcessShouldWork(t *testing.T) {
	originalMsg, buff := createSignedMessageAndMarshaledBytes(0)
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, originalMsg.Nonce)
	signedMessage := append(originalMsg.Payload, nonceBytes...)

	verifyCalled := false
	rmh := &relayerMessageHandler{
		marshalizer: &testsCommon.MarshalizerMock{},
		singleSigner: &cryptoMocks.SingleSignerStub{
			VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
				assert.Equal(t, msg, signedMessage)
				assert.Equal(t, originalMsg.Signature, sig)
				verifyCalled = true

				return nil
			},
		},
		keyGen: &cryptoMocks.KeyGenStub{},
	}

	p2pmsg := &p2pMocks.P2PMessageMock{
		DataField: buff,
	}

	msg, err := rmh.preProcessMessage(p2pmsg, fromPeer)
	assert.Equal(t, originalMsg, msg)
	assert.Nil(t, err)
	assert.True(t, verifyCalled)
}

func TestRelayerMessageHandler_createMessage(t *testing.T) {
	t.Parallel()

	t.Run("createMessage errors if sign fails", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		rmh := &relayerMessageHandler{
			marshalizer: &testsCommon.MarshalizerMock{},
			singleSigner: &cryptoMocks.SingleSignerStub{
				SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
					return nil, expectedErr
				},
			},
		}

		msg, err := rmh.createMessage([]byte("payload"))
		assert.Nil(t, msg)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("createMessage should work", func(t *testing.T) {
		payload := []byte("payload")
		sig := []byte("sig")
		counter := uint64(22322)
		numSignCalled := 0
		rmh := &relayerMessageHandler{
			counter:     counter,
			marshalizer: &testsCommon.MarshalizerMock{},
			singleSigner: &cryptoMocks.SingleSignerStub{
				SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
					nonceBytes := make([]byte, 8)
					binary.BigEndian.PutUint64(nonceBytes, counter)
					signedMessage := append(payload, nonceBytes...)
					assert.Equal(t, signedMessage, msg)

					numSignCalled++
					return sig, nil
				},
			},
			publicKeyBytes: []byte("pk"),
		}
		counter++

		msg, err := rmh.createMessage(payload)
		expectedMsg := &core.SignedMessage{
			Payload:        payload,
			PublicKeyBytes: rmh.publicKeyBytes,
			Signature:      sig,
			Nonce:          counter,
		}

		assert.Equal(t, expectedMsg, msg)
		assert.Nil(t, err)

		counter++
		msg, err = rmh.createMessage(payload)
		expectedMsg = &core.SignedMessage{
			Payload:        payload,
			PublicKeyBytes: rmh.publicKeyBytes,
			Signature:      sig,
			Nonce:          counter,
		}

		assert.Equal(t, expectedMsg, msg)
		assert.Nil(t, err)
	})
}

func createSignedMessageAndMarshaledBytes(index int) (*core.SignedMessage, []byte) {
	msg := &core.SignedMessage{
		Payload:        []byte(fmt.Sprintf("payload %d", index)),
		PublicKeyBytes: []byte(fmt.Sprintf("pk %d", index)),
		Signature:      []byte(fmt.Sprintf("sig %d", index)),
		Nonce:          34,
	}

	buff, _ := marshalizer.Marshal(msg)

	return msg, buff
}

func createSignedMessageForEthSig(index int) (*core.SignedMessage, []byte) {
	e := &core.EthereumSignature{
		Signature:   []byte(fmt.Sprintf("eth sig %d", index)),
		MessageHash: []byte("eth msg hash"),
	}
	payload, _ := marshalizer.Marshal(e)

	msg := &core.SignedMessage{
		Payload:        payload,
		PublicKeyBytes: []byte(fmt.Sprintf("pk %d", index)),
		Signature:      []byte(fmt.Sprintf("sig %d", index)),
		Nonce:          34,
	}
	buff, _ := marshalizer.Marshal(msg)

	return msg, buff
}
