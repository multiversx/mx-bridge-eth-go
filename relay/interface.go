package relay

import "context"

// Startable defines an entity that is able to Start or Stop
type Startable interface {
	Start(context.Context) error
	Stop() error
}

// TopologyProvider defines the topology provider functions
type TopologyProvider interface {
	AmITheLeader() bool
	Clean()
}

// Broadcaster defines a component able to communicate with other such instances and manage signatures and other state related data
type Broadcaster interface {
	BroadcastSignature(signature []byte)
	BroadcastJoinTopic()
	ClearSignatures()
	Signatures() [][]byte
	SortedPublicKeys() [][]byte
	RegisterOnTopics() error
	Close() error
	IsInterfaceNil() bool
}
