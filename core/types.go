package core

import "context"

// StepIdentifier defines a step name
type StepIdentifier string

// MachineStates defines all available steps for a state machine to run
type MachineStates map[StepIdentifier]Step

// Step defines a state machine step
type Step interface {
	Execute(ctx context.Context) (StepIdentifier, error)
	Identifier() StepIdentifier
	IsInterfaceNil() bool
}
