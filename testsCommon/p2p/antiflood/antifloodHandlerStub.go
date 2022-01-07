package antiflood

import "github.com/ElrondNetwork/elrond-go-core/core"

type AntiFloodHandlerStub struct {
	CanProcessMessagesOnTopicCalled func(peer core.PeerID, topic string, numMessages uint32) error
	ResetForTopicCalled             func(topic string)
	SetMaxMessagesForTopicCalled    func(topic string, maxNum uint32) error
	IsInterfaceNilCalled            func() bool
}

// CanProcessMessagesOnTopic -
func (a *AntiFloodHandlerStub) CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32) error {
	if a.CanProcessMessagesOnTopicCalled != nil {
		return a.CanProcessMessagesOnTopicCalled(peer, topic, numMessages)
	}
	return nil
}

// ResetForTopic -
func (a *AntiFloodHandlerStub) ResetForTopic(topic string) {
	if a.ResetForTopicCalled != nil {
		a.ResetForTopicCalled(topic)
	}
}

// SetMaxMessagesForTopic -
func (a *AntiFloodHandlerStub) SetMaxMessagesForTopic(topic string, maxNum uint32) error {
	if a.SetMaxMessagesForTopicCalled != nil {
		return a.SetMaxMessagesForTopicCalled(topic, maxNum)
	}
	return nil
}

// IsInterfaceNil -
func (a *AntiFloodHandlerStub) IsInterfaceNil() bool {
	return a == nil
}
