package v2

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateSignedMessage(index uint64) *core.SignedMessage {
	return &core.SignedMessage{
		Payload:        []byte(fmt.Sprintf("payload %d", index)),
		Signature:      []byte(fmt.Sprintf("sig %d", index)),
		PublicKeyBytes: []byte(fmt.Sprintf("pk %d", index)),
		Nonce:          index,
	}
}

func generateEthMessage(index uint64) *core.EthereumSignature {
	return &core.EthereumSignature{
		Signature:   []byte(fmt.Sprintf("sig %d", index)),
		MessageHash: []byte("message hash"),
	}
}

func TestSignatureHolder_ProcessNewMessage(t *testing.T) {
	t.Parallel()

	t.Run("nil messages", func(t *testing.T) {
		t.Parallel()

		msg := generateSignedMessage(0)
		ethMsg := generateEthMessage(0)

		sh := NewSignatureHolder()
		sh.ProcessNewMessage(nil, ethMsg)
		assert.Equal(t, 0, len(sh.signedMessages))
		assert.Equal(t, 0, len(sh.ethMessages))

		sh.ProcessNewMessage(msg, nil)
		assert.Equal(t, 0, len(sh.signedMessages))
		assert.Equal(t, 0, len(sh.ethMessages))
	})
	t.Run("first message should add", func(t *testing.T) {
		t.Parallel()

		msg := generateSignedMessage(0)
		ethMsg := generateEthMessage(0)

		sh := NewSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		assert.Equal(t, []*core.SignedMessage{msg}, sh.AllStoredSignatures())
		assert.Equal(t, []*core.EthereumSignature{ethMsg}, sh.ethMessages)
	})
	t.Run("two messages should add", func(t *testing.T) {
		t.Parallel()

		msg := generateSignedMessage(0)
		ethMsg := generateEthMessage(0)

		msg1 := generateSignedMessage(1)
		ethMsg1 := generateEthMessage(1)

		sh := NewSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		sh.ProcessNewMessage(msg1, ethMsg1)
		compareEthSignatureMessageLists(t, []*core.EthereumSignature{ethMsg, ethMsg1}, sh.ethMessages)
		compareSignedMessageLists(t, []*core.SignedMessage{msg, msg1}, sh.AllStoredSignatures())
	})
}

func TestSignatureHolder_Signatures(t *testing.T) {
	t.Parallel()

	t.Run("unique signatures should work", func(t *testing.T) {
		t.Parallel()

		msg := generateSignedMessage(0)
		ethMsg := generateEthMessage(0)

		msg1 := generateSignedMessage(1)
		ethMsg1 := generateEthMessage(1)

		sh := NewSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		sh.ProcessNewMessage(msg1, ethMsg1)

		compareBytesSlicesLists(t, [][]byte{ethMsg.Signature, ethMsg1.Signature}, sh.Signatures(ethMsg.MessageHash))

		sh.ClearStoredSignatures()

		assert.Equal(t, 0, len(sh.Signatures([]byte("eth msg"))))
	})
	t.Run("same signatures should return the unique ones", func(t *testing.T) {
		t.Parallel()

		msg := generateSignedMessage(0)
		ethMsg := generateEthMessage(0)

		msg1 := generateSignedMessage(1)
		ethMsg1 := generateEthMessage(1)

		msg2 := generateSignedMessage(2)
		ethMsg2 := generateEthMessage(2)
		ethMsg2.Signature = ethMsg1.Signature

		sh := NewSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		sh.ProcessNewMessage(msg1, ethMsg1)
		sh.ProcessNewMessage(msg2, ethMsg2)

		compareBytesSlicesLists(t, [][]byte{ethMsg.Signature, ethMsg1.Signature}, sh.Signatures(ethMsg.MessageHash))
	})
	t.Run("same signatures should return filter by message", func(t *testing.T) {
		t.Parallel()

		msg := generateSignedMessage(0)
		ethMsg := generateEthMessage(0)
		ethMsg.MessageHash = []byte("eth msg 1")

		msg1 := generateSignedMessage(1)
		ethMsg1 := generateEthMessage(1)

		msg2 := generateSignedMessage(2)
		ethMsg2 := generateEthMessage(2)

		sh := NewSignatureHolder()
		sh.ProcessNewMessage(msg, ethMsg)
		sh.ProcessNewMessage(msg1, ethMsg1)
		sh.ProcessNewMessage(msg2, ethMsg2)

		compareBytesSlicesLists(t, [][]byte{ethMsg1.Signature, ethMsg2.Signature}, sh.Signatures(ethMsg1.MessageHash))
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
