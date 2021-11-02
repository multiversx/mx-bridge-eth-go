package bridgeExecutors

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureHolder_ProcessNewMessage(t *testing.T) {
	t.Parallel()

	t.Run("nil messages", func(t *testing.T) {
		msg := &core.SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &core.EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.ProcessNewMessage(nil, ethMsg)
		assert.Equal(t, 0, len(sh.signedMessages))
		assert.Equal(t, 0, len(sh.ethMessages))

		sh.ProcessNewMessage(msg, nil)
		assert.Equal(t, 0, len(sh.signedMessages))
		assert.Equal(t, 0, len(sh.ethMessages))
	})
	t.Run("first message should add", func(t *testing.T) {
		msg := &core.SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &core.EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		assert.Equal(t, []*core.SignedMessage{msg}, sh.AllStoredSignatures())
		assert.Equal(t, []*core.EthereumSignature{ethMsg}, sh.ethMessages)
	})
	t.Run("two messages should add", func(t *testing.T) {
		msg := &core.SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &core.EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		msg2 := &core.SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          2,
		}
		ethMsg2 := &core.EthereumSignature{
			Signature:   []byte("eth sig2"),
			MessageHash: []byte("eth msg2"),
		}

		sh := newSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		sh.ProcessNewMessage(msg2, ethMsg2)
		compareEthSignatureMessageLists(t, []*core.EthereumSignature{ethMsg, ethMsg2}, sh.ethMessages)
		compareSignedMessageLists(t, []*core.SignedMessage{msg, msg2}, sh.AllStoredSignatures())
	})
}

func TestSignatureHolder_Signatures(t *testing.T) {
	t.Parallel()

	t.Run("unique signatures should work", func(t *testing.T) {
		msg := &core.SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &core.EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		msg2 := &core.SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          1,
		}
		ethMsg2 := &core.EthereumSignature{
			Signature:   []byte("eth sig 2"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		sh.ProcessNewMessage(msg2, ethMsg2)

		compareBytesSlicesLists(t, [][]byte{[]byte("eth sig"), []byte("eth sig 2")}, sh.Signatures([]byte("eth msg")))

		sh.clearStoredSignatures()

		assert.Equal(t, 0, len(sh.Signatures([]byte("eth msg"))))
	})
	t.Run("same signatures should return the unique ones", func(t *testing.T) {
		msg := &core.SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &core.EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		msg2 := &core.SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          1,
		}
		ethMsg2 := &core.EthereumSignature{
			Signature:   []byte("eth sig 2"),
			MessageHash: []byte("eth msg"),
		}

		msg3 := &core.SignedMessage{
			Payload:        []byte("payload3"),
			Signature:      []byte("sig3"),
			PublicKeyBytes: []byte("pk3"),
			Nonce:          1,
		}
		ethMsg3 := &core.EthereumSignature{
			Signature:   []byte("eth sig 2"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		sh.ProcessNewMessage(msg2, ethMsg2)
		sh.ProcessNewMessage(msg3, ethMsg3)

		compareBytesSlicesLists(t, [][]byte{[]byte("eth sig"), []byte("eth sig 2")}, sh.Signatures([]byte("eth msg")))
	})
	t.Run("same signatures should return filter by message", func(t *testing.T) {
		msg := &core.SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &core.EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg 1"),
		}

		msg2 := &core.SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          1,
		}
		ethMsg2 := &core.EthereumSignature{
			Signature:   []byte("eth sig 2"),
			MessageHash: []byte("eth msg"),
		}

		msg3 := &core.SignedMessage{
			Payload:        []byte("payload3"),
			Signature:      []byte("sig3"),
			PublicKeyBytes: []byte("pk3"),
			Nonce:          1,
		}
		ethMsg3 := &core.EthereumSignature{
			Signature:   []byte("eth sig 3"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		sh.ProcessNewMessage(msg2, ethMsg2)
		sh.ProcessNewMessage(msg3, ethMsg3)

		compareBytesSlicesLists(t, [][]byte{[]byte("eth sig 2"), []byte("eth sig 3")}, sh.Signatures([]byte("eth msg")))
	})
}

func compareSignedMessageLists(t *testing.T, list1 []*core.SignedMessage, list2 []*core.SignedMessage) {
	require.Equal(t, len(list1), len(list2))
	for _, obj1 := range list1 {
		found := false
		for _, obj2 := range list2 {
			if reflect.DeepEqual(obj1, obj2) {
				found = true
			}
		}
		require.True(t, found)
	}

	for _, obj2 := range list2 {
		found := false
		for _, obj1 := range list1 {
			if reflect.DeepEqual(obj1, obj2) {
				found = true
			}
		}
		require.True(t, found)
	}
}

func compareEthSignatureMessageLists(t *testing.T, list1 []*core.EthereumSignature, list2 []*core.EthereumSignature) {
	require.Equal(t, len(list1), len(list2))
	for _, obj1 := range list1 {
		found := false
		for _, obj2 := range list2 {
			if reflect.DeepEqual(obj1, obj2) {
				found = true
			}
		}
		require.True(t, found)
	}

	for _, obj2 := range list2 {
		found := false
		for _, obj1 := range list1 {
			if reflect.DeepEqual(obj1, obj2) {
				found = true
			}
		}
		require.True(t, found)
	}
}

func compareBytesSlicesLists(t *testing.T, list1 [][]byte, list2 [][]byte) {
	require.Equal(t, len(list1), len(list2))
	for _, slice1 := range list1 {
		found := false
		for _, slice2 := range list2 {
			if bytes.Equal(slice1, slice2) {
				found = true
			}
		}
		require.True(t, found)
	}

	for _, slice2 := range list2 {
		found := false
		for _, slice1 := range list1 {
			if bytes.Equal(slice1, slice2) {
				found = true
			}
		}
		require.True(t, found)
	}
}
