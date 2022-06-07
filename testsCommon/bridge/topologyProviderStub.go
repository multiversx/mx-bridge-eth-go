package bridge

// TopologyProviderStub -
type TopologyProviderStub struct {
	MyTurnAsLeaderCalled func() bool
}

// MyTurnAsLeader -
func (stub *TopologyProviderStub) MyTurnAsLeader() bool {
	if stub.MyTurnAsLeaderCalled != nil {
		return stub.MyTurnAsLeaderCalled()
	}

	return false
}

// IsInterfaceNil -
func (stub *TopologyProviderStub) IsInterfaceNil() bool {
	return stub == nil
}
