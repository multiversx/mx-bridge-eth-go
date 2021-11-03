package bridgeExecutors

// TopologyProvider is able to manage the current relayers topology
type TopologyProvider interface {
	AmITheLeader() bool
	IsInterfaceNil() bool
}
