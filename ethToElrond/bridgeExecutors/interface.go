package bridgeExecutors

import "time"

// TopologyProvider is able to manage the current relayers topology
type TopologyProvider interface {
	AmITheLeader() bool
	Clean()
}

// Timer defines operations related to time
type Timer interface {
	After(d time.Duration) <-chan time.Time
	NowUnix() int64
	Start()
	Close() error
}
