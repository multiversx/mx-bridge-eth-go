package antiflood

import "github.com/ElrondNetwork/elrond-go-core/core"

type TopicFloodPreventerStub struct {
	IncreaseLoadCalled           func(pid core.PeerID, topic string, numMessages uint32) error
	ResetForTopicCalled          func(topic string)
	SetMaxMessagesForTopicCalled func(topic string, maxNum uint32)
	IsInterfaceNilCalled         func() bool
}

// IncreaseLoad -
func (t *TopicFloodPreventerStub) IncreaseLoad(pid core.PeerID, topic string, numMessages uint32) error {
	if t.IncreaseLoadCalled != nil {
		return t.IncreaseLoadCalled(pid, topic, numMessages)
	}
	return nil
}

// ResetForTopic -
func (t *TopicFloodPreventerStub) ResetForTopic(topic string) {
	if t.ResetForTopicCalled != nil {
		t.ResetForTopicCalled(topic)
	}
}

// SetMaxMessagesForTopic -
func (t *TopicFloodPreventerStub) SetMaxMessagesForTopic(topic string, maxNum uint32) {
	if t.SetMaxMessagesForTopicCalled != nil {
		t.SetMaxMessagesForTopicCalled(topic, maxNum)
	}
}

// IsInterfaceNil -
func (t *TopicFloodPreventerStub) IsInterfaceNil() bool {
	return t == nil
}
