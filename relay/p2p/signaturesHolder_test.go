package p2p

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureHolder_addSignedMessage(t *testing.T) {
	t.Parallel()

	t.Run("first message should add", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.addSignedMessage(msg, ethMsg)
		assert.Equal(t, []*SignedMessage{msg}, sh.storedSignedMessages())
		assert.Equal(t, []*EthereumSignature{ethMsg}, sh.ethMessages)
		assert.Equal(t, [][]byte{[]byte("pk")}, sh.SortedPublicKeys())
	})
	t.Run("two messages should add", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          2,
		}
		ethMsg2 := &EthereumSignature{
			Signature:   []byte("eth sig2"),
			MessageHash: []byte("eth msg2"),
		}

		sh := newSignatureHolder()
		sh.addSignedMessage(msg, ethMsg)
		sh.addSignedMessage(msg2, ethMsg2)
		compareEthSignatureMessageLists(t, []*EthereumSignature{ethMsg, ethMsg2}, sh.ethMessages)
		compareSignedMessageLists(t, []*SignedMessage{msg, msg2}, sh.storedSignedMessages())
		assert.Equal(t, [][]byte{[]byte("pk"), []byte("pk2")}, sh.SortedPublicKeys())
	})
	t.Run("same nonce should not rewrite", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.addSignedMessage(msg, ethMsg)
		assert.Equal(t, []*SignedMessage{msg}, sh.storedSignedMessages())
		assert.Equal(t, []*EthereumSignature{ethMsg}, sh.ethMessages)

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg2 := &EthereumSignature{
			Signature:   []byte("eth sig2"),
			MessageHash: []byte("eth msg2"),
		}

		sh.addSignedMessage(msg2, ethMsg2)
		assert.Equal(t, []*SignedMessage{msg}, sh.storedSignedMessages())
		assert.Equal(t, []*EthereumSignature{ethMsg}, sh.ethMessages)
		assert.Equal(t, [][]byte{[]byte("pk")}, sh.SortedPublicKeys())
	})
	t.Run("lower nonce should not write", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          4,
		}
		ethMsg := &EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.addSignedMessage(msg, ethMsg)

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          3,
		}
		ethMsg2 := &EthereumSignature{
			Signature:   []byte("eth sig2"),
			MessageHash: []byte("eth msg2"),
		}

		sh.addSignedMessage(msg2, ethMsg2)
		assert.Equal(t, []*SignedMessage{msg}, sh.storedSignedMessages())
		assert.Equal(t, []*EthereumSignature{ethMsg}, sh.ethMessages)
		assert.Equal(t, [][]byte{[]byte("pk")}, sh.SortedPublicKeys())
	})
}

func TestSignatureHolder_addJoinedMessage(t *testing.T) {
	t.Parallel()

	t.Run("first message should add", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          0,
		}

		sh := newSignatureHolder()
		sh.addJoinedMessage(msg)
		assert.Equal(t, 0, len(sh.storedSignedMessages()))
		assert.Equal(t, [][]byte{[]byte("pk")}, sh.SortedPublicKeys())
	})
	t.Run("two messages should add", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          0,
		}

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          0,
		}

		sh := newSignatureHolder()
		sh.addJoinedMessage(msg)
		sh.addJoinedMessage(msg2)

		assert.Equal(t, 0, len(sh.storedSignedMessages()))
		assert.Equal(t, [][]byte{[]byte("pk"), []byte("pk2")}, sh.SortedPublicKeys())
	})
	t.Run("same nonce should rewrite", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          0,
		}

		sh := newSignatureHolder()
		sh.addJoinedMessage(msg)
		assert.Equal(t, 0, len(sh.storedSignedMessages()))

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          0,
		}

		sh.addJoinedMessage(msg2)
		assert.Equal(t, 0, len(sh.storedSignedMessages()))
		assert.Equal(t, [][]byte{[]byte("pk")}, sh.SortedPublicKeys())
	})
	t.Run("lower nonce should not write", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          4,
		}

		sh := newSignatureHolder()
		sh.addJoinedMessage(msg)

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          3,
		}

		sh.addJoinedMessage(msg2)
		assert.Equal(t, 0, len(sh.storedSignedMessages()))
		assert.Equal(t, [][]byte{[]byte("pk")}, sh.SortedPublicKeys())
	})
}

func TestSignatureHolder_Signatures(t *testing.T) {
	t.Parallel()

	t.Run("unique signatures should work", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          1,
		}
		ethMsg2 := &EthereumSignature{
			Signature:   []byte("eth sig 2"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.addSignedMessage(msg, ethMsg)
		sh.addSignedMessage(msg2, ethMsg2)

		compareBytesSlicesLists(t, [][]byte{[]byte("eth sig"), []byte("eth sig 2")}, sh.Signatures([]byte("eth msg")))
		assert.Equal(t, [][]byte{[]byte("pk"), []byte("pk2")}, sh.SortedPublicKeys())

		sh.ClearSignatures()

		assert.Equal(t, 0, len(sh.Signatures([]byte("eth msg"))))
		assert.Equal(t, [][]byte{[]byte("pk"), []byte("pk2")}, sh.SortedPublicKeys())
	})
	t.Run("same signatures should return the unique ones", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg"),
		}

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          1,
		}
		ethMsg2 := &EthereumSignature{
			Signature:   []byte("eth sig 2"),
			MessageHash: []byte("eth msg"),
		}

		msg3 := &SignedMessage{
			Payload:        []byte("payload3"),
			Signature:      []byte("sig3"),
			PublicKeyBytes: []byte("pk3"),
			Nonce:          1,
		}
		ethMsg3 := &EthereumSignature{
			Signature:   []byte("eth sig 2"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.addSignedMessage(msg, ethMsg)
		sh.addSignedMessage(msg2, ethMsg2)
		sh.addSignedMessage(msg3, ethMsg3)

		compareBytesSlicesLists(t, [][]byte{[]byte("eth sig"), []byte("eth sig 2")}, sh.Signatures([]byte("eth msg")))
		assert.Equal(t, [][]byte{[]byte("pk"), []byte("pk2"), []byte("pk3")}, sh.SortedPublicKeys())
	})
	t.Run("same signatures should return filter by message", func(t *testing.T) {
		msg := &SignedMessage{
			Payload:        []byte("payload"),
			Signature:      []byte("sig"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          1,
		}
		ethMsg := &EthereumSignature{
			Signature:   []byte("eth sig"),
			MessageHash: []byte("eth msg 1"),
		}

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk2"),
			Nonce:          1,
		}
		ethMsg2 := &EthereumSignature{
			Signature:   []byte("eth sig 2"),
			MessageHash: []byte("eth msg"),
		}

		msg3 := &SignedMessage{
			Payload:        []byte("payload3"),
			Signature:      []byte("sig3"),
			PublicKeyBytes: []byte("pk3"),
			Nonce:          1,
		}
		ethMsg3 := &EthereumSignature{
			Signature:   []byte("eth sig 3"),
			MessageHash: []byte("eth msg"),
		}

		sh := newSignatureHolder()
		sh.addSignedMessage(msg, ethMsg)
		sh.addSignedMessage(msg2, ethMsg2)
		sh.addSignedMessage(msg3, ethMsg3)

		compareBytesSlicesLists(t, [][]byte{[]byte("eth sig 2"), []byte("eth sig 3")}, sh.Signatures([]byte("eth msg")))
		assert.Equal(t, [][]byte{[]byte("pk"), []byte("pk2"), []byte("pk3")}, sh.SortedPublicKeys())
	})
}

func compareSignedMessageLists(t *testing.T, list1 []*SignedMessage, list2 []*SignedMessage) {
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

func compareEthSignatureMessageLists(t *testing.T, list1 []*EthereumSignature, list2 []*EthereumSignature) {
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
