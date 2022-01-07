package antiflood

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTopicFloodPreventer(t *testing.T) {
	t.Parallel()

	t.Run("invalid max num of messages should error", func(t *testing.T) {
		t.Parallel()

		tfp, err := NewTopicFloodPreventer(0)
		assert.Nil(t, tfp)
		assert.True(t, strings.Contains(err.Error(), errInvalidNumberOfMessages.Error()))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tfp, err := NewTopicFloodPreventer(10)
		assert.Nil(t, err)
		assert.False(t, tfp.IsInterfaceNil())
	})
}

func Test_topicFloodPreventer_IncreaseLoad(t *testing.T) {
	t.Parallel()

	tfp, err := NewTopicFloodPreventer(10)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(providedPid, providedTopic, 5)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(providedPid, providedTopic, 5)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(providedPid, providedTopic, 1)
	assert.Equal(t, errSystemBusy, err)
}

func Test_topicFloodPreventer_ResetForTopic(t *testing.T) {
	t.Parallel()

	tfp, err := NewTopicFloodPreventer(10)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(providedPid, providedTopic, 5)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(providedPid, providedTopic, 5)
	assert.Nil(t, err)
	tfp.ResetForTopic(providedTopic)
	err = tfp.IncreaseLoad(providedPid, providedTopic, 5)
	assert.Nil(t, err)
}

func Test_topicFloodPreventer_SetMaxMessagesForTopic(t *testing.T) {
	t.Parallel()

	defaultMaxMessages := uint32(10)
	tfp, err := NewTopicFloodPreventer(defaultMaxMessages)
	assert.Nil(t, err)

	tfp.SetMaxMessagesForTopic(providedTopic, defaultMaxMessages*2)
	err = tfp.IncreaseLoad(providedPid, providedTopic, defaultMaxMessages)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(providedPid, providedTopic, defaultMaxMessages)
	assert.Nil(t, err)
	err = tfp.IncreaseLoad(providedPid, providedTopic, 1)
	assert.Equal(t, errSystemBusy, err)
}
