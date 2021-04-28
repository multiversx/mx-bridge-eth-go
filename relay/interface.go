package relay

import "context"

type Startable interface {
	Start(context.Context) error
	Stop() error
}

type TopologyProvider interface {
	PeerCount() int
	AmITheLeader() bool
	Clean()
}
