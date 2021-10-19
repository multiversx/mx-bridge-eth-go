package relay

import "context"

type StepIdentifier string

// Step defines a state machine step
type Step interface {
	Execute(ctx context.Context) (StepIdentifier, error)
	Identifier() StepIdentifier
	IsInterfaceNil() bool
}

// MachineStates defines all available steps for a state machine to run
type MachineStates map[StepIdentifier]Step
