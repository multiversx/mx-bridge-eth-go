package bridge

// ElrondConfig represents the Elrond Config parameters
type ElrondConfig struct {
	NetworkAddress               string
	BridgeAddress                string
	PrivateKey                   string
	IntervalToResendTxsInSeconds uint64
	GasLimit                     uint64
}

// EthereumConfig represents the Ethereum Config parameters
type EthereumConfig struct {
	NetworkAddress               string
	BridgeAddress                string
	PrivateKey                   string
	IntervalToResendTxsInSeconds uint64
	GasLimit                     uint64
	GasStation                   GasStationConfig
}

// GasStationConfig represents the configuration for the gas station handler
type GasStationConfig struct {
	URL                      string
	PollingIntervalInSeconds int
	RequestTimeInSeconds     int
	MaximumAllowedGasPrice   int
	GasPriceSelector         string
}
