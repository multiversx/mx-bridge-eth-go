package state

// LoopStatus -
func (sm *stateMachine) LoopStatus() bool {
	return sm.loopStatus.IsSet()
}
