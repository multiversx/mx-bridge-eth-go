package steps

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// CreateSteps creates all machine states providing the bridge executor
func CreateSteps(executor ethToElrond.EthToElrondBridge) (core.MachineStates, error) {
	if check.IfNil(executor) {
		return nil, ethToElrond.ErrNilExecutor
	}

	return createMachineStates(executor)
}

func createMachineStates(executor ethToElrond.EthToElrondBridge) (core.MachineStates, error) {
	machineStates := make(core.MachineStates)

	steps := []core.Step{
		&getPendingStep{
			bridge: executor,
		},
		&proposeTransferStep{
			bridge: executor,
		},
		&signProposedTransferStep{
			bridge: executor,
		},
		&waitForQuorumStep{
			bridge: executor,
		},
		&performActionIDStep{
			bridge: executor,
		},
	}

	for _, s := range steps {
		_, found := machineStates[s.Identifier()]
		if found {
			return nil, fmt.Errorf("%w for identifier '%s'", ethToElrond.ErrDuplicatedStepIdentifier, s.Identifier())
		}

		machineStates[s.Identifier()] = s
	}

	return machineStates, nil
}
