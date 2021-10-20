package bridgeExecutors

// TopologyProvider is able to manage the current relayers topology
type TopologyProvider interface {
	PeerCount() int
	AmITheLeader() bool
	Clean()
}
