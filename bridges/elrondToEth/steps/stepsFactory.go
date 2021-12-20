package steps

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// CreateSteps creates all machine states providing the bridge executor
func CreateSteps(executor bridge.Executor) (core.MachineStates, error) {
	if check.IfNil(executor) {
		return nil, bridges.ErrNilExecutor
	}

	return createMachineStates(executor)
}

func createMachineStates(executor bridge.Executor) (core.MachineStates, error) {
	machineStates := make(core.MachineStates)

	steps := []core.Step{
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

	for _, s := range steps {
		_, found := machineStates[s.Identifier()]
		if found {
			return nil, fmt.Errorf("%w for identifier '%s'", bridges.ErrDuplicatedStepIdentifier, s.Identifier())
		}

		machineStates[s.Identifier()] = s
	}

	return machineStates, nil
}
