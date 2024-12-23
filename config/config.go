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
	ClientAvailabilityAllowDelta       uint64
	EventsBlockRangeFrom               int64
	EventsBlockRangeTo                 int64
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
	ResourceLimiter p2pConfig.P2PResourceLimiterConfig
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
	SafeContractAddress             string
	PrivateKeyFile                  string
	IntervalToResendTxsInSeconds    uint64
	GasMap                          MultiversXGasMapConfig
	MaxRetriesOnQuorumReached       uint64
	MaxRetriesOnWasTransferProposed uint64
	ClientAvailabilityAllowDelta    uint64
	Proxy                           ProxyConfig
}

// ProxyConfig represents the configuration for the MultiversX proxy
type ProxyConfig struct {
	CacherExpirationSeconds uint64
	RestAPIEntityType       string
	MaxNoncesDelta          int
	FinalityCheck           bool
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
	ScCallPerByte          uint64
	ScCallPerformForEach   uint64
}

// PeersRatingConfig will hold settings related to peers rating
type PeersRatingConfig struct {
	TopRatedCacheCapacity int
	BadRatedCacheCapacity int
}

// PendingOperationsFilterConfig defines the filter structure
type PendingOperationsFilterConfig struct {
	DeniedEthAddresses  []string
	AllowedEthAddresses []string
	DeniedMvxAddresses  []string
	AllowedMvxAddresses []string
	DeniedTokens        []string
	AllowedTokens       []string
}

// ScCallsModuleConfig will hold the settings for the SC calls module
type ScCallsModuleConfig struct {
	General           GeneralScCallsModuleConfig
	ScCallsExecutor   ScCallsExecutorConfig
	RefundExecutor    RefundExecutorConfig
	Filter            PendingOperationsFilterConfig
	Logs              LogsConfig
	TransactionChecks TransactionChecksConfig
}

// GeneralScCallsModuleConfig will hold the general settings for the SC calls module
type GeneralScCallsModuleConfig struct {
	ScProxyBech32Addresses       []string
	NetworkAddress               string
	ProxyMaxNoncesDelta          int
	ProxyFinalityCheck           bool
	ProxyCacherExpirationSeconds uint64
	ProxyRestAPIEntityType       string
	IntervalToResendTxsInSeconds uint64
	PrivateKeyFile               string
}

// ScCallsExecutorConfig will hold the settings for the SC calls executor
type ScCallsExecutorConfig struct {
	ExtraGasToExecute               uint64
	MaxGasLimitToUse                uint64
	GasLimitForOutOfGasTransactions uint64
	PollingIntervalInMillis         uint64
}

// RefundExecutorConfig will hold the settings for the refund executor
type RefundExecutorConfig struct {
	GasToExecute            uint64
	PollingIntervalInMillis uint64
}

// TransactionChecksConfig will hold the setting for how to handle the transaction execution
type TransactionChecksConfig struct {
	CheckTransactionResults    bool
	TimeInSecondsBetweenChecks uint64
	ExecutionTimeoutInSeconds  uint64
	CloseAppOnError            bool
	ExtraDelayInSecondsOnError uint64
}

// MigrationToolConfig is the migration tool config struct
type MigrationToolConfig struct {
	Eth        EthereumConfig
	MultiversX MultiversXConfig
	Logs       LogsConfig
}
