package antiflood

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type topicFloodPreventer struct {
	mutTopicFloodPreventer           sync.RWMutex
	topicMaxNumMessages              map[string]uint32
	counterMap                       map[string]map[core.PeerID]uint32
	defaultMaxNumMessegesPerInterval uint32
}

// NewTopicFloodPreventer creates a new flood preventer based on a topic
func NewTopicFloodPreventer(defaultMaxMessagesPerInterval uint32) (*topicFloodPreventer, error) {
	if defaultMaxMessagesPerInterval < topicMinMessages {
		return nil, fmt.Errorf("error %w, defaultMaxMessagesPerInterval: provided %d, minimum %d",
			errInvalidNumberOfMessages, defaultMaxMessagesPerInterval, topicMinMessages)
	}

	return &topicFloodPreventer{
		topicMaxNumMessages:              make(map[string]uint32),
		counterMap:                       make(map[string]map[core.PeerID]uint32),
		defaultMaxNumMessegesPerInterval: defaultMaxMessagesPerInterval,
	}, nil
}

// IncreaseLoad tries to increment the counter values held at "identifier" position for the given topic
// It returns nil if it had succeeded incrementing (existing counter value is lower than provided topicMaxNumMessages)
func (tfp *topicFloodPreventer) IncreaseLoad(pid core.PeerID, topic string, numMessages uint32) error {
	tfp.mutTopicFloodPreventer.Lock()
	defer tfp.mutTopicFloodPreventer.Unlock()

	_, found := tfp.counterMap[topic]
	if !found {
		tfp.counterMap[topic] = make(map[core.PeerID]uint32)
	}

	numMessagesAfterIncrease := numMessages + tfp.counterMap[topic][pid]
	limitExceeded := numMessagesAfterIncrease > tfp.maxMessagesForTopic(topic)
	if limitExceeded {
		return errSystemBusy
	}

	tfp.counterMap[topic][pid] = numMessagesAfterIncrease

	return nil
}

// ResetForTopic clears all map values for a given topic
func (tfp *topicFloodPreventer) ResetForTopic(topic string) {
	tfp.mutTopicFloodPreventer.Lock()
	tfp.counterMap[topic] = make(map[core.PeerID]uint32)
	tfp.mutTopicFloodPreventer.Unlock()
}

// SetMaxMessagesForTopic will update the maximum number of messages that can be received from a peer in a topic
func (tfp *topicFloodPreventer) SetMaxMessagesForTopic(topic string, maxNum uint32) {
	tfp.mutTopicFloodPreventer.Lock()
	tfp.topicMaxNumMessages[topic] = maxNum
	tfp.mutTopicFloodPreventer.Unlock()
}

func (tfp *topicFloodPreventer) maxMessagesForTopic(topic string) uint32 {
	_, ok := tfp.topicMaxNumMessages[topic]
	if !ok {
		tfp.topicMaxNumMessages[topic] = tfp.defaultMaxNumMessegesPerInterval
	}

	return tfp.topicMaxNumMessages[topic]
}

// IsInterfaceNil returns true if there is no value under the interface
func (tfp *topicFloodPreventer) IsInterfaceNil() bool {
	return tfp == nil
}
