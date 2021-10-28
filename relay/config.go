package relay

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-go/config"
)

// Config general configuration struct
type Config struct {
	Eth          bridge.EthereumConfig
	Elrond       bridge.ElrondConfig
	P2P          ConfigP2P
	StateMachine map[string]ConfigStateMachine
	Relayer      ConfigRelayer
	Logs         LogsConfig
}

// ConfigP2P configuration for the P2P communication
type ConfigP2P struct {
	Port            string
	Seed            string
	InitialPeerList []string
	ProtocolID      string
}

// ConfigRelayer configuration for general relayer configuration
type ConfigRelayer struct {
	Marshalizer  config.MarshalizerConfig
	RoleProvider RoleProviderConfig
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

// ContextFlagsConfig the configuration for flags
type ContextFlagsConfig struct {
	WorkingDir        string
	LogLevel          string
	DisableAnsiColor  bool
	ConfigurationFile string
	SaveLogFile       bool
	EnableLogName     bool
	RestApiInterface  string
	EnablePprof       bool
}

// LogsConfig will hold settings related to the logging sub-system
type LogsConfig struct {
	LogFileLifeSpanInSec int
}

// RoleProviderConfig is the configuration for the role provider component
type RoleProviderConfig struct {
	PollingIntervalInMillis uint64
}
