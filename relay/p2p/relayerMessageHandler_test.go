package p2p

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	cryptoMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/crypto"
	p2pMocks "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/p2p"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/stretchr/testify/assert"
)

func TestRelayerMessageHandler_preProcess(t *testing.T) {
	t.Parallel()

	t.Run("preProcess errors if unmarshal fails", func(t *testing.T) {
		rmh := &relayerMessageHandler{
			marshalizer:  &testsCommon.MarshalizerMock{},
			singleSigner: &cryptoMocks.SingleSignerStub{},
		}
		p2pmsg := &p2pMocks.P2PMessageMock{
			DataField: []byte("gibberish"),
		}

		msg, err := rmh.preProcessMessage(p2pmsg)
		assert.Nil(t, msg)
		assert.NotNil(t, err)
	})
	t.Run("preProcess errors if keygen fails", func(t *testing.T) {
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
		_, buff := createSignedMessageAndMarshaledBytes()

		p2pmsg := &p2pMocks.P2PMessageMock{
			DataField: buff,
		}

		msg, err := rmh.preProcessMessage(p2pmsg)
		assert.Nil(t, msg)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("preProcess errors if verify fails", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		rmh := &relayerMessageHandler{
			marshalizer: &testsCommon.MarshalizerMock{},
			singleSigner: &cryptoMocks.SingleSignerStub{
				VerifyCalled: func(public crypto.PublicKey, msg []byte, sig []byte) error {
					return expectedErr
				},
			},
			keyGen: &cryptoMocks.KeyGenStub{},
		}
		_, buff := createSignedMessageAndMarshaledBytes()

		p2pmsg := &p2pMocks.P2PMessageMock{
			DataField: buff,
		}

		msg, err := rmh.preProcessMessage(p2pmsg)
		assert.Nil(t, msg)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("preProcess should work", func(t *testing.T) {
		originalMsg, buff := createSignedMessageAndMarshaledBytes()
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

		msg, err := rmh.preProcessMessage(p2pmsg)
		assert.Equal(t, originalMsg, msg)
		assert.Nil(t, err)
		assert.True(t, verifyCalled)
	})
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
		expectedMsg := &SignedMessage{
			Payload:        payload,
			PublicKeyBytes: rmh.publicKeyBytes,
			Signature:      sig,
			Nonce:          counter,
		}

		assert.Equal(t, expectedMsg, msg)
		assert.Nil(t, err)

		counter++
		msg, err = rmh.createMessage(payload)
		expectedMsg = &SignedMessage{
			Payload:        payload,
			PublicKeyBytes: rmh.publicKeyBytes,
			Signature:      sig,
			Nonce:          counter,
		}

		assert.Equal(t, expectedMsg, msg)
		assert.Nil(t, err)
	})
}

func createSignedMessageAndMarshaledBytes() (*SignedMessage, []byte) {
	return createSignedMessageAndMarshaledBytesWithValues([]byte("payload"), []byte("pk"))
}

func createSignedMessageAndMarshaledBytesWithValues(payload []byte, pk []byte) (*SignedMessage, []byte) {
	msg := &SignedMessage{
		Payload:        payload,
		PublicKeyBytes: pk,
		Signature:      []byte("sig"),
		Nonce:          34,
	}

	marshalizer := &testsCommon.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(msg)

	return msg, buff
}
