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
			Nonce:          0,
		}

		sh := newSignatureHolder()
		sh.addSignedMessage(msg)
		assert.Equal(t, []*SignedMessage{msg}, sh.storedSignedMessages())
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
		sh.addSignedMessage(msg)
		sh.addSignedMessage(msg2)

		compareSignedMessageLists(t, []*SignedMessage{msg, msg2}, sh.storedSignedMessages())
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
		sh.addSignedMessage(msg)
		assert.Equal(t, []*SignedMessage{msg}, sh.storedSignedMessages())

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          0,
		}

		sh.addSignedMessage(msg2)
		assert.Equal(t, []*SignedMessage{msg2}, sh.storedSignedMessages())
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
		sh.addSignedMessage(msg)

		msg2 := &SignedMessage{
			Payload:        []byte("payload2"),
			Signature:      []byte("sig2"),
			PublicKeyBytes: []byte("pk"),
			Nonce:          3,
		}

		sh.addSignedMessage(msg2)
		assert.Equal(t, []*SignedMessage{msg}, sh.storedSignedMessages())
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
	sh.addSignedMessage(msg)
	sh.addSignedMessage(msg2)

	compareBytesSlicesLists(t, [][]byte{[]byte("payload"), []byte("payload2")}, sh.Signatures())
	assert.Equal(t, [][]byte{[]byte("pk"), []byte("pk2")}, sh.SortedPublicKeys())

	sh.ClearSignatures()

	assert.Equal(t, 0, len(sh.Signatures()))
	assert.Equal(t, [][]byte{[]byte("pk"), []byte("pk2")}, sh.SortedPublicKeys())
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
