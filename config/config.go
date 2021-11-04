package config

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-go/config"
)

// Configs is a holder for the relayer configuration parameters
type Configs struct {
	GeneralConfig   *Config
	ApiRoutesConfig *ApiRoutesConfig
	FlagsConfig     *ContextFlagsConfig
}

// Config general configuration struct
type Config struct {
	Eth          bridge.EthereumConfig
	Elrond       bridge.ElrondConfig
	P2P          ConfigP2P
	StateMachine map[string]ConfigStateMachine
	Relayer      ConfigRelayer
	Logs         LogsConfig
	Antiflood    AntifloodConfig
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
	WorkingDir           string
	LogLevel             string
	DisableAnsiColor     bool
	ConfigurationFile    string
	ConfigurationApiFile string
	SaveLogFile          bool
	EnableLogName        bool
	RestApiInterface     string
	EnablePprof          bool
}

// WebServerAntifloodConfig will hold the anti-flooding parameters for the web server
type WebServerAntifloodConfig struct {
	SimultaneousRequests         uint32
	SameSourceRequests           uint32
	SameSourceResetIntervalInSec uint32
}

// AntifloodConfig will hold all p2p antiflood parameters
type AntifloodConfig struct {
	Enabled   bool
	WebServer WebServerAntifloodConfig
}

// ApiRoutesConfig holds the configuration related to Rest API routes
type ApiRoutesConfig struct {
	Logging     ApiLoggingConfig
	APIPackages map[string]APIPackageConfig
}

// ApiLoggingConfig holds the configuration related to API requests logging
type ApiLoggingConfig struct {
	LoggingEnabled          bool
	ThresholdInMicroSeconds int
}

// APIPackageConfig holds the configuration for the routes of each package
type APIPackageConfig struct {
	Routes []RouteConfig
}

// RouteConfig holds the configuration for a single route
type RouteConfig struct {
	Name string
	Open bool
}

// LogsConfig will hold settings related to the logging sub-system
type LogsConfig struct {
	LogFileLifeSpanInSec int
}

// RoleProviderConfig is the configuration for the role provider component
type RoleProviderConfig struct {
	PollingIntervalInMillis uint64
}
