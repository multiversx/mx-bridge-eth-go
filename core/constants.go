package core

// ElrondAddressLength is the Elrond's address length
const ElrondAddressLength = 32

const (
	// EthFastGasPrice represents the fast gas price value
	EthFastGasPrice EthGasPriceSelector = "fast"

	// EthFastestGasPrice represents the fastest gas price value
	EthFastestGasPrice EthGasPriceSelector = "fastest"

	// EthSafeLowGasPrice represents the lowest safe gat price value
	EthSafeLowGasPrice EthGasPriceSelector = "safeLow"

	// EthAverageGasPrice represents the average gas price value
	EthAverageGasPrice EthGasPriceSelector = "average"

	// WebServerOffString represents the constant used to switch off the web server
	WebServerOffString = "off"
)

const (
	// MetricNumTransactionsSucceeded represents the metric used for counting executed transactions
	MetricNumTransactionsSucceeded = "num transactions succeeded"

	// MetricNumTransactionsRejected represents the metric used for counting rejected transactions
	MetricNumTransactionsRejected = "num transactions rejected"

	// MetricNumBatches represents the metric used for counting the number of executed batches
	MetricNumBatches = "num batches"

	// MetricLastError represents the metric used to store the last encountered error
	MetricLastError = "last encountered error"

	// MetricCurrentStateMachineStep represents the metric used to store the current running machine step
	MetricCurrentStateMachineStep = "current state machine step"

	// MetricErc20Balance represents the metric used for ERC20 balances. It will be suffixed by the ERC20 address
	MetricErc20Balance = "ERC20 balance"

	// MetricNumEthClientRequests represents the metric used to count the number of ethereum client requests
	MetricNumEthClientRequests = "num ethereum client requests"

	// MetricNumEthClientTransactions represents the metric used to count the number of ethereum sent transactions
	MetricNumEthClientTransactions = "num ethereum client transactions"

	// MetricLastQueriedEthereumBlockNumber represents the metric used to store the last ethereum block number that was
	// fetched from the ethereum client
	MetricLastQueriedEthereumBlockNumber = "ethereum last queried block number"

	// MetricRelayerP2PAddresses represents the metric used to store all the P2P addresses the messenger has bound to
	MetricRelayerP2PAddresses = "relayer P2P addresses"

	// MetricConnectedP2PAddresses represents the metric used to store all the P2P addresses the messenger has connected to
	MetricConnectedP2PAddresses = "connected P2P addresses"
)

// PersistedMetrics represents the array of metrics that should be persisted
var PersistedMetrics = []string{MetricNumTransactionsSucceeded, MetricNumTransactionsRejected, MetricNumBatches,
	MetricNumEthClientRequests, MetricNumEthClientTransactions, MetricLastQueriedEthereumBlockNumber}

const (
	// EthClientStatusHandlerName is the ethereum client status handler name
	EthClientStatusHandlerName = "eth-client"

	// EthToElrondStatusHandlerName is the ethereum to elrond bridge status handler name
	EthToElrondStatusHandlerName = "eth-to-elrond"

	// ElrondToEthStatusHandlerName is the elrond to ethereum bridge status handler name
	ElrondToEthStatusHandlerName = "elrond-to-eth"
)
