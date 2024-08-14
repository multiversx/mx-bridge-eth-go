package core

const (
	// Executed is the Executed with success status value
	Executed = byte(3)
	// Rejected is the Rejected status value
	Rejected = byte(4)

	// Uint32ArgBytes is the constant used for the number of bytes to encode an Uint32 value
	Uint32ArgBytes = 4

	// Uint64ArgBytes is the constant used for the number of bytes to encode an Uint64 value
	Uint64ArgBytes = 8

	// MissingDataProtocolMarker defines the marker for missing data (simple transfers)
	MissingDataProtocolMarker byte = 0x00

	// DataPresentProtocolMarker defines the marker for existing data (transfers with SC calls)
	DataPresentProtocolMarker byte = 0x01
)

const (
	// EthFastGasPrice represents the fast gas price value
	EthFastGasPrice EthGasPriceSelector = "FastGasPrice"

	// EthSafeGasPrice represents the safe gas price value
	EthSafeGasPrice EthGasPriceSelector = "SafeGasPrice"

	// EthProposeGasPrice represents the proposed gas price value
	EthProposeGasPrice EthGasPriceSelector = "ProposeGasPrice"

	// WebServerOffString represents the constant used to switch off the web server
	WebServerOffString = "off"
)

const (
	// MetricNumBatches represents the metric used for counting the number of executed batches
	MetricNumBatches = "num batches"

	// MetricLastError represents the metric used to store the last encountered error
	MetricLastError = "last encountered error"

	// MetricCurrentStateMachineStep represents the metric used to store the current running machine step
	MetricCurrentStateMachineStep = "current state machine step"

	// MetricNumEthClientRequests represents the metric used to count the number of ethereum client requests
	MetricNumEthClientRequests = "num ethereum client requests"

	// MetricNumEthClientTransactions represents the metric used to count the number of ethereum sent transactions
	MetricNumEthClientTransactions = "num ethereum client transactions"

	// MetricLastQueriedEthereumBlockNumber represents the metric used to store the last ethereum block number that was
	// fetched from the ethereum client
	MetricLastQueriedEthereumBlockNumber = "ethereum last queried block number"

	// MetricEthereumClientStatus represents the metric used to store the status of the ethereum client
	MetricEthereumClientStatus = "ethereum client status"

	// MetricLastEthereumClientError represents the metric used to store the last encountered error from the ethereum client
	MetricLastEthereumClientError = "ethereum client last encountered error"

	// MetricLastQueriedMultiversXBlockNumber represents the metric used to store the last MultiversX block number that was
	// fetched from the MultiversX client
	MetricLastQueriedMultiversXBlockNumber = "multiversx last queried block number"

	// MetricMultiversXClientStatus represents the metric used to store the status of the MultiversX client
	MetricMultiversXClientStatus = "multiversx client status"

	// MetricLastMultiversXClientError represents the metric used to store the last encountered error from the MultiversX client
	MetricLastMultiversXClientError = "multiversx client last encountered error"

	// MetricRelayerP2PAddresses represents the metric used to store all the P2P addresses the messenger has bound to
	MetricRelayerP2PAddresses = "relayer P2P addresses"

	// MetricConnectedP2PAddresses represents the metric used to store all the P2P addresses the messenger has connected to
	MetricConnectedP2PAddresses = "connected P2P addresses"

	// MetricLastBlockNonce represents the last block nonce queried
	MetricLastBlockNonce = "last block nonce"
)

// PersistedMetrics represents the array of metrics that should be persisted
var PersistedMetrics = []string{MetricNumBatches, MetricNumEthClientRequests, MetricNumEthClientTransactions,
	MetricLastQueriedEthereumBlockNumber, MetricLastQueriedMultiversXBlockNumber, MetricEthereumClientStatus,
	MetricMultiversXClientStatus, MetricLastEthereumClientError, MetricLastMultiversXClientError, MetricLastBlockNonce}

const (
	// EthClientStatusHandlerName is the Ethereum client status handler name
	EthClientStatusHandlerName = "eth-client"

	// MultiversXClientStatusHandlerName is the MultiversX client status handler name
	MultiversXClientStatusHandlerName = "multiversx-client"
)
