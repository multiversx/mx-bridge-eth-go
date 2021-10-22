package steps

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// CreateSteps creates all machine states providing the bridge executor
func CreateSteps(executor BridgeExecutor) (core.MachineStates, error) {
	if check.IfNil(executor) {
		return nil, ErrNilBridgeExecutor
	}

	return createMachineStates(executor)
}

func createMachineStates(executor BridgeExecutor) (core.MachineStates, error) {
	machineStates := make(core.MachineStates)

	steps := []core.Step{
		&getPendingStep{
			bridge: executor,
		},
		&proposeTransferStep{
			bridge: executor,
		},
		&waitForSignaturesForProposeTransferStep{
			bridge: executor,
		},
		&executeTransferStep{
			bridge: executor,
		},
		&proposeSetStatusStep{
			bridge: executor,
		},
		&waitForSignaturesForProposeSetStatusStep{
			bridge: executor,
		},
		&executeSetStatusStep{
			bridge: executor,
		},
	}

	for _, s := range steps {
		_, found := machineStates[s.Identifier()]
		if found {
			return nil, fmt.Errorf("%w for identifier '%s'", ErrDuplicatedStepIdentifier, s.Identifier())
		}

		machineStates[s.Identifier()] = s
	}

	return machineStates, nil
}
