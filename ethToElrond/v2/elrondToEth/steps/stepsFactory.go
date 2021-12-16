package steps

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	v2 "github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/elrondToEth"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

// CreateSteps creates all machine states providing the bridge executor
func CreateSteps(executor elrondToEth.ElrondToEthBridge) (core.MachineStates, error) {
	if check.IfNil(executor) {
		return nil, v2.ErrNilExecutor
	}

	return createMachineStates(executor)
}

func createMachineStates(executor elrondToEth.ElrondToEthBridge) (core.MachineStates, error) {
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
			return nil, fmt.Errorf("%w for identifier '%s'", v2.ErrDuplicatedStepIdentifier, s.Identifier())
		}

		machineStates[s.Identifier()] = s
	}

	return machineStates, nil
}
