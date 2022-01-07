package antiflood

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("p2p/antiflood")

type antifloodHandler struct {
	topicPreventer topicAntifloodPreventer
}

// NewAntifloodHandler creates a new antiflood handler component
func NewAntifloodHandler(topicFloodPreventer topicAntifloodPreventer) (*antifloodHandler, error) {
	if check.IfNil(topicFloodPreventer) {
		return nil, errNilTopicFloodPreventer
	}

	return &antifloodHandler{
		topicPreventer: topicFloodPreventer,
	}, nil
}

// CanProcessMessagesOnTopic signals if a p2p message can be processed or not for a given topic
func (ah *antifloodHandler) CanProcessMessagesOnTopic(peer core.PeerID, topic string, numMessages uint32) error {
	err := ah.topicPreventer.IncreaseLoad(peer, topic, numMessages)
	if err != nil {
		log.Debug("%w in AntifloodHandler for peer %s", err, peer.Pretty())
		return err
	}

	return nil
}

// ResetForTopic clears all map values for a given topic
func (ah *antifloodHandler) ResetForTopic(topic string) {
	ah.topicPreventer.ResetForTopic(topic)
}

// SetMaxMessagesForTopic will update the maximum number of messages that can be received from a peer in a topic
func (ah *antifloodHandler) SetMaxMessagesForTopic(topic string, maxNum uint32) error {
	if maxNum < topicMinMessages {
		log.Debug("error %w, maxNum: provided %d, minimum %d",
			errInvalidNumberOfMessages, maxNum, topicMinMessages)
		return errInvalidNumberOfMessages
	}
	ah.topicPreventer.SetMaxMessagesForTopic(topic, maxNum)
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ah *antifloodHandler) IsInterfaceNil() bool {
	return ah == nil
}
