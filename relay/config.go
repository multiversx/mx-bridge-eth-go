package relay

import "github.com/ElrondNetwork/elrond-eth-bridge/bridge"

// Config general configuration struct
type Config struct {
	Eth          bridge.EthereumConfig
	Elrond       bridge.ElrondConfig
	P2P          ConfigP2P
	StateMachine ConfigStateMachine
}

// ConfigP2P configuration for the P2P communication
type ConfigP2P struct {
	Port            string
	Seed            string
	InitialPeerList []string
	ProtocolID      string
}

// ConfigStateMachine the configuration for the state machine
type ConfigStateMachine struct {
	StepDurationInMillis uint64
	Steps                []StepConfig
}

// StepConfig defines a step configuration
type StepConfig struct {
	Name             string
	DurationInMillis uint64
}
