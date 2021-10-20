package stateMachine

// LoopStatus -
func (sm *stateMachine) LoopStatus() bool {
	return sm.loopStatus.IsSet()
}
