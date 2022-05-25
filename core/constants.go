package core

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

	// MetricLastQueriedElrondBlockNumber represents the metric used to store the last elrond block number that was
	// fetched from the elrond client
	MetricLastQueriedElrondBlockNumber = "elrond last queried block number"

	// MetricElrondClientStatus represents the metric used to store the status of the elrond client
	MetricElrondClientStatus = "elrond client status"

	// MetricLastElrondClientError represents the metric used to store the last encountered error from the elrond client
	MetricLastElrondClientError = "elrond client last encountered error"

	// MetricRelayerP2PAddresses represents the metric used to store all the P2P addresses the messenger has bound to
	MetricRelayerP2PAddresses = "relayer P2P addresses"

	// MetricConnectedP2PAddresses represents the metric used to store all the P2P addresses the messenger has connected to
	MetricConnectedP2PAddresses = "connected P2P addresses"

	// MetricLastBlockNonce represents the last block nonce queried
	MetricLastBlockNonce = "last block nonce"
)

// PersistedMetrics represents the array of metrics that should be persisted
var PersistedMetrics = []string{MetricNumBatches, MetricNumEthClientRequests, MetricNumEthClientTransactions,
	MetricLastQueriedEthereumBlockNumber, MetricLastQueriedElrondBlockNumber, MetricEthereumClientStatus,
	MetricElrondClientStatus, MetricLastEthereumClientError, MetricLastElrondClientError, MetricLastBlockNonce}

const (
	// EthClientStatusHandlerName is the ethereum client status handler name
	EthClientStatusHandlerName = "eth-client"

	// ElrondClientStatusHandlerName is the elrond client status handler name
	ElrondClientStatusHandlerName = "elrond-client"
)
