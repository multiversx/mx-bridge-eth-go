package bridge

// ElrondConfig represents the Elrond Config parameters
type ElrondConfig struct {
	NetworkAddress               string
	BridgeAddress                string
	PrivateKeyFile               string
	IntervalToResendTxsInSeconds uint64
	GasLimit                     uint64
}

// EthereumConfig represents the Ethereum Config parameters
type EthereumConfig struct {
	NetworkAddress               string
	BridgeAddress                string
	PrivateKeyFile               string
	IntervalToResendTxsInSeconds uint64
	GasLimit                     uint64
	ERC20Contracts               []string
	GasStation                   GasStationConfig
}

// GasStationConfig represents the configuration for the gas station handler
type GasStationConfig struct {
	Enabled                  bool
	URL                      string
	PollingIntervalInSeconds int
	MaximumAllowedGasPrice   int
	GasPriceSelector         string
}
