package config

import (
	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-chain-go/config"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
)

// Configs is a holder for the relayer configuration parameters
type Configs struct {
	GeneralConfig   Config
	ApiRoutesConfig ApiRoutesConfig
	FlagsConfig     ContextFlagsConfig
}

// Config general configuration struct
type Config struct {
	Eth               EthereumConfig
	MultiversX        MultiversXConfig
	P2P               ConfigP2P
	StateMachine      map[string]ConfigStateMachine
	Relayer           ConfigRelayer
	Logs              LogsConfig
	WebAntiflood      WebAntifloodConfig
	BatchValidator    BatchValidatorConfig
	PeersRatingConfig PeersRatingConfig
}

// EthereumConfig represents the Ethereum Config parameters
type EthereumConfig struct {
	Chain                              chain.Chain
	NetworkAddress                     string
	MultisigContractAddress            string
	SafeContractAddress                string
	PrivateKeyFile                     string
	IntervalToResendTxsInSeconds       uint64
	GasLimitBase                       uint64
	GasLimitForEach                    uint64
	GasStation                         GasStationConfig
	MaxRetriesOnQuorumReached          uint64
	IntervalToWaitForTransferInSeconds uint64
	MaxBlocksDelta                     uint64
}

// GasStationConfig represents the configuration for the gas station handler
type GasStationConfig struct {
	Enabled                    bool
	URL                        string
	PollingIntervalInSeconds   int
	RequestRetryDelayInSeconds int
	MaxFetchRetries            int
	RequestTimeInSeconds       int
	MaximumAllowedGasPrice     int
	GasPriceSelector           string
	GasPriceMultiplier         int
}

// ConfigP2P configuration for the P2P communication
type ConfigP2P struct {
	Port            string
	InitialPeerList []string
	ProtocolID      string
	Transports      p2pConfig.P2PTransportConfig
	AntifloodConfig config.AntifloodConfig
	Transport       p2pConfig.P2PTCPTransport
}

// ConfigRelayer configuration for general relayer configuration
type ConfigRelayer struct {
	Marshalizer          config.MarshalizerConfig
	RoleProvider         RoleProviderConfig
	StatusMetricsStorage config.StorageConfig
}

// ConfigStateMachine the configuration for the state machine
type ConfigStateMachine struct {
	StepDurationInMillis       uint64
	IntervalForLeaderInSeconds uint64
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

// WebAntifloodConfig will hold all web antiflood parameters
type WebAntifloodConfig struct {
	Enabled   bool
	WebServer WebServerAntifloodConfig
}

// BatchValidatorConfig represents the configuration for the batch validator
type BatchValidatorConfig struct {
	Enabled              bool
	URL                  string
	RequestTimeInSeconds int
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
	LogFileLifeSpanInMB  int
}

// RoleProviderConfig is the configuration for the role provider component
type RoleProviderConfig struct {
	PollingIntervalInMillis uint64
}

// MultiversXConfig represents the MultiversX Config parameters
type MultiversXConfig struct {
	NetworkAddress                  string
	MultisigContractAddress         string
	PrivateKeyFile                  string
	IntervalToResendTxsInSeconds    uint64
	GasMap                          MultiversXGasMapConfig
	MaxRetriesOnQuorumReached       uint64
	MaxRetriesOnWasTransferProposed uint64
	ProxyCacherExpirationSeconds    uint64
	ProxyRestAPIEntityType          string
	ProxyMaxNoncesDelta             int
	ProxyFinalityCheck              bool
}

// MultiversXGasMapConfig represents the gas limits for MultiversX operations
type MultiversXGasMapConfig struct {
	Sign                   uint64
	ProposeTransferBase    uint64
	ProposeTransferForEach uint64
	ProposeStatusBase      uint64
	ProposeStatusForEach   uint64
	PerformActionBase      uint64
	PerformActionForEach   uint64
}

// PeersRatingConfig will hold settings related to peers rating
type PeersRatingConfig struct {
	TopRatedCacheCapacity int
	BadRatedCacheCapacity int
}
