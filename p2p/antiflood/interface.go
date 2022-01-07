package antiflood

import elrondCore "github.com/ElrondNetwork/elrond-go-core/core"

type topicAntifloodPreventer interface {
	IncreaseLoad(pid elrondCore.PeerID, topic string, numMessages uint32) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32)
	IsInterfaceNil() bool
}

// AntifloodHandler defines a component able to signal that the system is too busy (flooded) processing messages
type AntifloodHandler interface {
	CanProcessMessagesOnTopic(peer elrondCore.PeerID, topic string, numMessages uint32) error
	ResetForTopic(topic string)
	SetMaxMessagesForTopic(topic string, maxNum uint32) error
	IsInterfaceNil() bool
}
