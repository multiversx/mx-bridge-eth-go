package antiflood

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/p2p/antiflood"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/assert"
)

var (
	providedPid         = core.PeerID("provided pid")
	providedTopic       = "provided topic"
	providedNumMessages = uint32(5)
)

func TestNewAntifloodHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil topic flood preventer should error", func(t *testing.T) {
		t.Parallel()

		ah, err := NewAntifloodHandler(nil)
		assert.Nil(t, ah)
		assert.Equal(t, errNilTopicFloodPreventer, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ah, err := NewAntifloodHandler(&antiflood.TopicFloodPreventerStub{})
		assert.Nil(t, err)
		assert.False(t, ah.IsInterfaceNil())
	})
}

func Test_antifloodHandler_CanProcessMessagesOnTopic(t *testing.T) {
	t.Parallel()

	t.Run("system busy should error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		expectedError := errors.New("expected error")
		tpfs := &antiflood.TopicFloodPreventerStub{
			IncreaseLoadCalled: func(pid core.PeerID, topic string, numMessages uint32) error {
				wasCalled = true
				return expectedError
			},
		}
		ah, err := NewAntifloodHandler(tpfs)
		assert.Nil(t, err)
		assert.False(t, ah.IsInterfaceNil())

		err = ah.CanProcessMessagesOnTopic(providedPid, providedTopic, providedNumMessages)
		assert.Equal(t, expectedError, err)
		assert.True(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		tpfs := &antiflood.TopicFloodPreventerStub{
			IncreaseLoadCalled: func(pid core.PeerID, topic string, numMessages uint32) error {
				wasCalled = true
				assert.Equal(t, providedPid, pid)
				assert.Equal(t, providedTopic, topic)
				assert.Equal(t, providedNumMessages, numMessages)
				return nil
			},
		}
		ah, err := NewAntifloodHandler(tpfs)
		assert.Nil(t, err)
		assert.False(t, ah.IsInterfaceNil())

		err = ah.CanProcessMessagesOnTopic(providedPid, providedTopic, providedNumMessages)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}

func Test_antifloodHandler_ResetForTopic(t *testing.T) {
	t.Parallel()

	wasCalled := false
	tpfs := &antiflood.TopicFloodPreventerStub{
		ResetForTopicCalled: func(topic string) {
			wasCalled = true
			assert.Equal(t, providedTopic, topic)
		},
	}

	ah, err := NewAntifloodHandler(tpfs)
	assert.Nil(t, err)
	assert.False(t, ah.IsInterfaceNil())

	ah.ResetForTopic(providedTopic)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func Test_antifloodHandler_SetMaxMessagesForTopic(t *testing.T) {
	t.Parallel()

	t.Run("invalid max num of messages", func(t *testing.T) {
		t.Parallel()

		tpfs := &antiflood.TopicFloodPreventerStub{}
		ah, err := NewAntifloodHandler(tpfs)
		assert.Nil(t, err)
		assert.False(t, ah.IsInterfaceNil())

		err = ah.SetMaxMessagesForTopic("topic", 0)
		assert.Equal(t, errInvalidNumberOfMessages, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		tpfs := &antiflood.TopicFloodPreventerStub{
			SetMaxMessagesForTopicCalled: func(topic string, maxNum uint32) {
				wasCalled = true
				assert.Equal(t, providedTopic, topic)
				assert.Equal(t, providedNumMessages, maxNum)
			},
		}
		ah, err := NewAntifloodHandler(tpfs)
		assert.Nil(t, err)
		assert.False(t, ah.IsInterfaceNil())

		err = ah.SetMaxMessagesForTopic(providedTopic, providedNumMessages)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})
}
