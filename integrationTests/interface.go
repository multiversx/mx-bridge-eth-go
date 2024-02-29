package integrationTests

import (
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/process"
)

// chainSimulatorHandler defines what a chain simulator should be able to do
type chainSimulatorHandler interface {
	GetNodeHandler(shardID uint32) process.NodeHandler
	SetStateMultiple(stateSlice []*dtos.AddressState) error
	IsInterfaceNil() bool
}
