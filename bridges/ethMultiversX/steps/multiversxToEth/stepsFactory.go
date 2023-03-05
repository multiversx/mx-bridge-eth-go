package multiversxtoeth

import (
	"fmt"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX/steps"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// CreateSteps creates all machine states providing the bridge executor
func CreateSteps(executor steps.Executor) (core.MachineStates, error) {
	if check.IfNil(executor) {
		return nil, ethmultiversx.ErrNilExecutor
	}

	return createMachineStates(executor)
}

func createMachineStates(executor steps.Executor) (core.MachineStates, error) {
	machineStates := make(core.MachineStates)

	stepsSlice := []core.Step{
		&getPendingStep{
			bridge: executor,
		},
		&signProposedTransferStep{
			bridge: executor,
		},
		&waitForQuorumOnTransferStep{
			bridge: executor,
		},
		&performTransferStep{
			bridge: executor,
		},
		&waitTransferConfirmationStep{
			bridge: executor,
		},
		&resolveSetStatusStep{
			bridge: executor,
		},
		&proposeSetStatusStep{
			bridge: executor,
		},
		&signProposedSetStatusStep{
			bridge: executor,
		},
		&waitForQuorumOnSetStatusStep{
			bridge: executor,
		},
		&performSetStatusStep{
			bridge: executor,
		},
	}

	for _, s := range stepsSlice {
		_, found := machineStates[s.Identifier()]
		if found {
			return nil, fmt.Errorf("%w for identifier '%s'", ethmultiversx.ErrDuplicatedStepIdentifier, s.Identifier())
		}

		machineStates[s.Identifier()] = s
	}

	return machineStates, nil
}
