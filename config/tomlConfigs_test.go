package config

import (
	"testing"

	"github.com/multiversx/mx-chain-communication-go/p2p/config"
	chainConfig "github.com/multiversx/mx-chain-go/config"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
)

func TestConfigs(t *testing.T) {
	t.Parallel()

	expectedConfig := Config{
		Eth: EthereumConfig{
			Chain:                              "Ethereum",
			NetworkAddress:                     "http://127.0.0.1:8545",
			MultisigContractAddress:            "3009d97FfeD62E57d444e552A9eDF9Ee6Bc8644c",
			SafeContractAddress:                "A6504Cc508889bbDBd4B748aFf6EA6b5D0d2684c",
			PrivateKeyFile:                     "keys/ethereum.sk",
			IntervalToWaitForTransferInSeconds: 600,
			GasLimitBase:                       350000,
			GasLimitForEach:                    30000,
			GasStation: GasStationConfig{
				Enabled:                    true,
				URL:                        "https://api.etherscan.io/api?module=gastracker&action=gasoracle",
				PollingIntervalInSeconds:   60,
				RequestRetryDelayInSeconds: 5,
				MaxFetchRetries:            3,
				RequestTimeInSeconds:       2,
				MaximumAllowedGasPrice:     300,
				GasPriceSelector:           "SafeGasPrice",
				GasPriceMultiplier:         1000000000,
			},
			MaxRetriesOnQuorumReached:    3,
			ClientAvailabilityAllowDelta: 10,
			EventsBlockRangeFrom:         -100,
			EventsBlockRangeTo:           400,
		},
		MultiversX: MultiversXConfig{
			NetworkAddress:               "https://devnet-gateway.multiversx.com",
			MultisigContractAddress:      "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
			SafeContractAddress:          "erd1qqqqqqqqqqqqqpgqtvnswnzxxz8susupesys0hvg7q2z5nawrcjq06qdus",
			PrivateKeyFile:               "keys/multiversx.pem",
			IntervalToResendTxsInSeconds: 60,
			GasMap: MultiversXGasMapConfig{
				Sign:                   8000000,
				ProposeTransferBase:    11000000,
				ProposeTransferForEach: 5500000,
				ProposeStatusBase:      10000000,
				ProposeStatusForEach:   7000000,
				PerformActionBase:      40000000,
				PerformActionForEach:   5500000,
			},
			MaxRetriesOnQuorumReached:       3,
			MaxRetriesOnWasTransferProposed: 3,
			ClientAvailabilityAllowDelta:    10,
			Proxy: ProxyConfig{
				CacherExpirationSeconds: 600,
				RestAPIEntityType:       "observer",
				MaxNoncesDelta:          7,
				FinalityCheck:           true,
			},
		},
		P2P: ConfigP2P{
			Port:            "10010",
			InitialPeerList: make([]string, 0),
			ProtocolID:      "/erd/relay/1.0.0",
			Transports: p2pConfig.P2PTransportConfig{
				TCP: config.TCPProtocolConfig{
					ListenAddress:    "/ip4/0.0.0.0/tcp/%d",
					PreventPortReuse: false,
				},
				QUICAddress:         "",
				WebSocketAddress:    "",
				WebTransportAddress: "",
			},
			ResourceLimiter: config.ResourceLimiterConfig{
				Type:                   "default autoscale",
				ManualSystemMemoryInMB: 1,
				ManualMaximumFD:        2,
			},
			AntifloodConfig: chainConfig.AntifloodConfig{
				Enabled:                   true,
				NumConcurrentResolverJobs: 50,
				OutOfSpecs: chainConfig.FloodPreventerConfig{
					IntervalInSeconds: 1,
					ReservedPercent:   0,
					PeerMaxInput: chainConfig.AntifloodLimitsConfig{
						BaseMessagesPerInterval: 140,
						TotalSizePerInterval:    4194304,
						IncreaseFactor: chainConfig.IncreaseFactorConfig{
							Threshold: 0,
							Factor:    0,
						},
					},
					BlackList: chainConfig.BlackListConfig{
						ThresholdNumMessagesPerInterval: 200,
						ThresholdSizePerInterval:        6291456,
						NumFloodingRounds:               2,
						PeerBanDurationInSeconds:        3600,
					},
				},
				FastReacting: chainConfig.FloodPreventerConfig{
					IntervalInSeconds: 1,
					ReservedPercent:   20,
					PeerMaxInput: chainConfig.AntifloodLimitsConfig{
						BaseMessagesPerInterval: 10,
						TotalSizePerInterval:    1048576,
						IncreaseFactor: chainConfig.IncreaseFactorConfig{
							Threshold: 10,
							Factor:    1,
						},
					},
					BlackList: chainConfig.BlackListConfig{
						ThresholdNumMessagesPerInterval: 70,
						ThresholdSizePerInterval:        2097154,
						NumFloodingRounds:               10,
						PeerBanDurationInSeconds:        300,
					},
				},
				SlowReacting: chainConfig.FloodPreventerConfig{
					IntervalInSeconds: 30,
					ReservedPercent:   20,
					PeerMaxInput: chainConfig.AntifloodLimitsConfig{
						BaseMessagesPerInterval: 400,
						TotalSizePerInterval:    10485760,
						IncreaseFactor: chainConfig.IncreaseFactorConfig{
							Threshold: 10,
							Factor:    0,
						},
					},
					BlackList: chainConfig.BlackListConfig{
						ThresholdNumMessagesPerInterval: 800,
						ThresholdSizePerInterval:        20971540,
						NumFloodingRounds:               2,
						PeerBanDurationInSeconds:        3600,
					},
				},
				PeerMaxOutput: chainConfig.AntifloodLimitsConfig{
					BaseMessagesPerInterval: 5,
					TotalSizePerInterval:    524288,
					IncreaseFactor:          chainConfig.IncreaseFactorConfig{},
				},
				Cache: chainConfig.CacheConfig{
					Name:     "Antiflood",
					Type:     "LRU",
					Capacity: 7000,
				},
				Topic: chainConfig.TopicAntifloodConfig{
					DefaultMaxMessagesPerSec: 300,
					MaxMessages: []chainConfig.TopicMaxMessagesConfig{
						{
							Topic:             "EthereumToMultiversX_join",
							NumMessagesPerSec: 100,
						},
						{
							Topic:             "EthereumToMultiversX_sign",
							NumMessagesPerSec: 100,
						},
					},
				},
				TxAccumulator: chainConfig.TxAccumulatorConfig{},
			},
		},
		StateMachine: map[string]ConfigStateMachine{
			"EthereumToMultiversX": {
				StepDurationInMillis:       12000,
				IntervalForLeaderInSeconds: 120,
			},
			"MultiversXToEthereum": {
				StepDurationInMillis:       12000,
				IntervalForLeaderInSeconds: 720,
			},
		},
		Relayer: ConfigRelayer{
			Marshalizer: chainConfig.MarshalizerConfig{
				Type:           "gogo protobuf",
				SizeCheckDelta: 10,
			},
			RoleProvider: RoleProviderConfig{
				PollingIntervalInMillis: 60000,
			},
			StatusMetricsStorage: chainConfig.StorageConfig{
				Cache: chainConfig.CacheConfig{
					Name:     "StatusMetricsStorage",
					Type:     "LRU",
					Capacity: 1000,
				},
				DB: chainConfig.DBConfig{
					FilePath:          "StatusMetricsStorageDB",
					Type:              "LvlDBSerial",
					BatchDelaySeconds: 2,
					MaxBatchSize:      100,
					MaxOpenFiles:      10,
				},
			},
		},
		Logs: LogsConfig{
			LogFileLifeSpanInSec: 86400,
			LogFileLifeSpanInMB:  1024,
		},
		WebAntiflood: WebAntifloodConfig{
			Enabled: true,
			WebServer: WebServerAntifloodConfig{
				SimultaneousRequests:         100,
				SameSourceRequests:           10000,
				SameSourceResetIntervalInSec: 1,
			},
		},
		PeersRatingConfig: PeersRatingConfig{
			TopRatedCacheCapacity: 5000,
			BadRatedCacheCapacity: 5000,
		},
	}

	testString := `
[Eth]
    Chain = "Ethereum"
    NetworkAddress = "http://127.0.0.1:8545" # a network address
    MultisigContractAddress = "3009d97FfeD62E57d444e552A9eDF9Ee6Bc8644c" # the eth address for the bridge contract
    SafeContractAddress = "A6504Cc508889bbDBd4B748aFf6EA6b5D0d2684c"
    PrivateKeyFile = "keys/ethereum.sk" # the path to the file containing the relayer eth private key
    GasLimitBase = 350000
    GasLimitForEach = 30000
    IntervalToWaitForTransferInSeconds = 600 #10 minutes
    MaxRetriesOnQuorumReached = 3
    ClientAvailabilityAllowDelta = 10
    EventsBlockRangeFrom = -100
    EventsBlockRangeTo = 400
    [Eth.GasStation]
        Enabled = true
        URL = "https://api.etherscan.io/api?module=gastracker&action=gasoracle" # gas station URL. Suggestion to provide the api-key here
        GasPriceMultiplier = 1000000000 # the value to be multiplied with the fetched value. Useful in test chains. On production chain should be 1000000000
        PollingIntervalInSeconds = 60 # number of seconds between gas price polling
        RequestRetryDelayInSeconds = 5 # number of seconds of delay after one failed request
        MaxFetchRetries = 3 # number of fetch retries before printing an error
        RequestTimeInSeconds = 2 # maximum timeout (in seconds) for the gas price request
        MaximumAllowedGasPrice = 300 # maximum value allowed for the fetched gas price value
        # GasPriceSelector available options: "SafeGasPrice", "ProposeGasPrice", "FastGasPrice"
        GasPriceSelector = "SafeGasPrice" # selector used to provide the gas price

[MultiversX]
    NetworkAddress = "https://devnet-gateway.multiversx.com" # the network address
    MultisigContractAddress = "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf" # the multiversx address for the bridge contract
    SafeContractAddress = "erd1qqqqqqqqqqqqqpgqtvnswnzxxz8susupesys0hvg7q2z5nawrcjq06qdus" # the multiversx address for the safe contract
    PrivateKeyFile = "keys/multiversx.pem" # the path to the pem file containing the relayer multiversx wallet
    IntervalToResendTxsInSeconds = 60 # the time in seconds between nonce reads
    MaxRetriesOnQuorumReached = 3
    MaxRetriesOnWasTransferProposed = 3
    ClientAvailabilityAllowDelta = 10
    [MultiversX.Proxy]
        CacherExpirationSeconds = 600 # the caching time in seconds

        # valid options for ProxyRestAPIEntityType are "observer" and "proxy". Any other value will trigger an error.
        # "observer" is useful when querying an observer, directly and "proxy" is useful when querying a squad's proxy (gateway)
        RestAPIEntityType = "observer"
        FinalityCheck = true
        MaxNoncesDelta = 7 # the number of maximum blocks allowed to be "in front" of what the metachain has notarized
    [MultiversX.GasMap]
        Sign = 8000000
        ProposeTransferBase = 11000000
        ProposeTransferForEach = 5500000
        ProposeStatusBase = 10000000
        ProposeStatusForEach = 7000000
        PerformActionBase = 40000000
        PerformActionForEach = 5500000

[P2P]
    Port = "10010"
    InitialPeerList = []
    ProtocolID = "/erd/relay/1.0.0"
    [P2P.Transports]
        QUICAddress = "" # optional QUIC address. If this transport should be activated, should be in this format: /ip4/0.0.0.0/udp/%d/quic-v1
        WebSocketAddress = "" # optional WebSocket address. If this transport should be activated, should be in this format: /ip4/0.0.0.0/tcp/%d/ws
        WebTransportAddress = "" # optional WebTransport address. If this transport should be activated, should be in this format: /ip4/0.0.0.0/udp/%d/quic-v1/webtransport
        [P2P.Transports.TCP]
            ListenAddress = "/ip4/0.0.0.0/tcp/%d" # TCP listen address
            PreventPortReuse = false
        [P2P.ResourceLimiter]
			Type = "default autoscale" #available options "default autoscale", "infinite", "default with manual scale".
			ManualSystemMemoryInMB = 1 # not taken into account if the type is not "default with manual scale"
			ManualMaximumFD = 2 # not taken into account if the type is not "default with manual scale"
    [P2P.AntifloodConfig]
        Enabled = true
        NumConcurrentResolverJobs = 50
        [P2P.AntifloodConfig.FastReacting]
            IntervalInSeconds = 1
            ReservedPercent   = 20.0
            [P2P.AntifloodConfig.FastReacting.PeerMaxInput]
                BaseMessagesPerInterval  = 10
                TotalSizePerInterval = 1048576 #1MB/s
                [P2P.AntifloodConfig.FastReacting.PeerMaxInput.IncreaseFactor]
                    Threshold = 10 #if consensus size will exceed this value, then
                    Factor = 1.0   #increase the base value with [factor*consensus size]
            [P2P.AntifloodConfig.FastReacting.BlackList]
                ThresholdNumMessagesPerInterval = 70
                ThresholdSizePerInterval = 2097154 #2MB/s
                NumFloodingRounds = 10
                PeerBanDurationInSeconds = 300

        [P2P.AntifloodConfig.SlowReacting]
            IntervalInSeconds = 30
            ReservedPercent   = 20.0
            [P2P.AntifloodConfig.SlowReacting.PeerMaxInput]
                BaseMessagesPerInterval = 400
                TotalSizePerInterval = 10485760 #10MB/interval
                [P2P.AntifloodConfig.SlowReacting.PeerMaxInput.IncreaseFactor]
                    Threshold = 10 #if consensus size will exceed this value, then
                    Factor = 0.0   #increase the base value with [factor*consensus size]
            [P2P.AntifloodConfig.SlowReacting.BlackList]
                ThresholdNumMessagesPerInterval = 800
                ThresholdSizePerInterval = 20971540 #20MB/interval
                NumFloodingRounds = 2
                PeerBanDurationInSeconds = 3600

        [P2P.AntifloodConfig.OutOfSpecs]
            IntervalInSeconds = 1
            ReservedPercent   = 0.0
            [P2P.AntifloodConfig.OutOfSpecs.PeerMaxInput]
                BaseMessagesPerInterval = 140
                TotalSizePerInterval = 4194304 #4MB/s
                [P2P.AntifloodConfig.OutOfSpecs.PeerMaxInput.IncreaseFactor]
                    Threshold = 0 #if consensus size will exceed this value, then
                    Factor = 0.0     #increase the base value with [factor*consensus size]
            [P2P.AntifloodConfig.OutOfSpecs.BlackList]
                ThresholdNumMessagesPerInterval = 200
                ThresholdSizePerInterval = 6291456 #6MB/s
                NumFloodingRounds = 2
                PeerBanDurationInSeconds = 3600

        [P2P.AntifloodConfig.PeerMaxOutput]
            BaseMessagesPerInterval  = 5
            TotalSizePerInterval     = 524288 #512KB/s

        [P2P.AntifloodConfig.Cache]
            Name = "Antiflood"
            Capacity = 7000
            Type = "LRU"
        [P2P.AntifloodConfig.Topic]
            DefaultMaxMessagesPerSec = 300 # default number of messages per interval for a topic
            MaxMessages = [{ Topic = "EthereumToMultiversX_join", NumMessagesPerSec = 100 },
                           { Topic = "EthereumToMultiversX_sign", NumMessagesPerSec = 100 }]

[Relayer]
    [Relayer.Marshalizer]
        Type = "gogo protobuf"
        SizeCheckDelta = 10
    [Relayer.RoleProvider]
        PollingIntervalInMillis = 60000 # 1 minute
    [Relayer.StatusMetricsStorage]
        [Relayer.StatusMetricsStorage.Cache]
            Name = "StatusMetricsStorage"
            Capacity = 1000
            Type = "LRU"
        [Relayer.StatusMetricsStorage.DB]
            FilePath = "StatusMetricsStorageDB"
            Type = "LvlDBSerial"
            BatchDelaySeconds = 2
            MaxBatchSize = 100
            MaxOpenFiles = 10

[StateMachine]
    [StateMachine.EthereumToMultiversX]
        StepDurationInMillis = 12000 #12 seconds
        IntervalForLeaderInSeconds = 120 #2 minutes

    [StateMachine.MultiversXToEthereum]
        StepDurationInMillis = 12000 #12 seconds
        IntervalForLeaderInSeconds = 720 #12 minutes

[Logs]
    LogFileLifeSpanInSec = 86400 # 24h
    LogFileLifeSpanInMB = 1024 # 1GB

[WebAntiflood]
    Enabled = true
    [WebAntiflood.WebServer]
            # SimultaneousRequests represents the number of concurrent requests accepted by the web server
            # this is a global throttler that acts on all http connections regardless of the originating source
            SimultaneousRequests = 100
            # SameSourceRequests defines how many requests are allowed from the same source in the specified
            # time frame (SameSourceResetIntervalInSec)
            SameSourceRequests = 10000
            # SameSourceResetIntervalInSec time frame between counter reset, in seconds
            SameSourceResetIntervalInSec = 1

[PeersRatingConfig]
    TopRatedCacheCapacity = 5000
    BadRatedCacheCapacity = 5000

`

	cfg := Config{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	require.Nil(t, err)
	require.Equal(t, expectedConfig, cfg)
}

func TestScCallsExecutorConfigs(t *testing.T) {
	t.Parallel()

	expectedConfig := ScCallsModuleConfig{
		General: GeneralScCallsModuleConfig{
			ScProxyBech32Addresses: []string{
				"erd1qqqqqqqqqqqqqpgqnef5f5aq32d63kljld8w5vnvz4gk5sy9hrrq2ld08s",
				"erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
			},
			NetworkAddress:               "127.0.0.1:8085",
			ProxyMaxNoncesDelta:          7,
			ProxyFinalityCheck:           true,
			ProxyCacherExpirationSeconds: 600,
			ProxyRestAPIEntityType:       "observer",
			IntervalToResendTxsInSeconds: 60,
			PrivateKeyFile:               "keys/multiversx.pem",
		},
		ScCallsExecutor: ScCallsExecutorConfig{
			ExtraGasToExecute:               50000000,
			MaxGasLimitToUse:                249999999,
			GasLimitForOutOfGasTransactions: 30000000,
			PollingIntervalInMillis:         6000,
		},
		RefundExecutor: RefundExecutorConfig{
			GasToExecute:            20000000,
			PollingIntervalInMillis: 6000,
		},
		Filter: PendingOperationsFilterConfig{
			AllowedEthAddresses: []string{"*"},
			AllowedMvxAddresses: []string{"*"},
			AllowedTokens:       []string{"MEME-a43fa1"},
		},
		Logs: LogsConfig{
			LogFileLifeSpanInSec: 86400,
			LogFileLifeSpanInMB:  1024,
		},
		TransactionChecks: TransactionChecksConfig{
			CheckTransactionResults:    true,
			TimeInSecondsBetweenChecks: 6,
			ExecutionTimeoutInSeconds:  120,
			CloseAppOnError:            false,
			ExtraDelayInSecondsOnError: 120,
		},
	}

	testString := `
[General]
	ScProxyBech32Addresses = [
		"erd1qqqqqqqqqqqqqpgqnef5f5aq32d63kljld8w5vnvz4gk5sy9hrrq2ld08s",
    	"erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
	]
	NetworkAddress = "127.0.0.1:8085"
	ProxyMaxNoncesDelta = 7
	ProxyFinalityCheck = true
	ProxyCacherExpirationSeconds = 600
	ProxyRestAPIEntityType = "observer"
	IntervalToResendTxsInSeconds = 60
	PrivateKeyFile = "keys/multiversx.pem"

[ScCallsExecutor]
	ExtraGasToExecute = 50000000
	MaxGasLimitToUse = 249999999 # this is a safe max gas limit to use both intra-shard & cross-shard
	GasLimitForOutOfGasTransactions = 30000000 # this value will be used when a transaction specified a gas limit > 249999999 
	PollingIntervalInMillis = 6000

[RefundExecutor]
	GasToExecute = 20000000
	PollingIntervalInMillis = 6000

[Filter]
	AllowedEthAddresses = ["*"]		# execute SC calls from all ETH addresses
	AllowedMvxAddresses = ["*"]     # execute SC calls to all MvX contracts
	AllowedTokens = ["MEME-a43fa1"] # execute SC calls for this token only

[Logs]
	LogFileLifeSpanInSec = 86400 # 24h
    LogFileLifeSpanInMB = 1024 # 1GB

[TransactionChecks]
	CheckTransactionResults = true     # enable or disable the transaction execution checking
	TimeInSecondsBetweenChecks = 6     # the number of seconds to recheck the status of the transaction
	ExecutionTimeoutInSeconds  = 120   # the number of seconds after the transaction is considered failed if it was not seen by the blockchain 
	CloseAppOnError            = false # enable or disable if the executor should automatically close on a transaction execution error  
	ExtraDelayInSecondsOnError = 120   # extra delay in seconds if the transaction execution errored 
`

	cfg := ScCallsModuleConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	require.Nil(t, err)
	require.Equal(t, expectedConfig, cfg)

}

func TestMigrationToolConfig(t *testing.T) {
	t.Parallel()

	expectedConfig := MigrationToolConfig{
		Eth: EthereumConfig{
			Chain:               "Ethereum",
			NetworkAddress:      "http://127.0.0.1:8545",
			SafeContractAddress: "A6504Cc508889bbDBd4B748aFf6EA6b5D0d2684c",
			PrivateKeyFile:      "keys/ethereum.sk",
			GasLimitBase:        350000,
			GasLimitForEach:     30000,
			GasStation: GasStationConfig{
				Enabled:                    true,
				URL:                        "https://api.etherscan.io/api?module=gastracker&action=gasoracle",
				PollingIntervalInSeconds:   60,
				RequestRetryDelayInSeconds: 5,
				MaxFetchRetries:            3,
				RequestTimeInSeconds:       2,
				MaximumAllowedGasPrice:     300,
				GasPriceSelector:           "SafeGasPrice",
				GasPriceMultiplier:         1000000000,
			},
		},
		MultiversX: MultiversXConfig{
			NetworkAddress:          "https://devnet-gateway.multiversx.com",
			MultisigContractAddress: "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
			SafeContractAddress:     "erd1qqqqqqqqqqqqqpgqtvnswnzxxz8susupesys0hvg7q2z5nawrcjq06qdus",
			Proxy: ProxyConfig{
				CacherExpirationSeconds: 600,
				RestAPIEntityType:       "observer",
				MaxNoncesDelta:          7,
				FinalityCheck:           true,
			},
		},
		Logs: LogsConfig{
			LogFileLifeSpanInSec: 86400,
			LogFileLifeSpanInMB:  1024,
		},
	}

	testString := `
[Eth]
    Chain = "Ethereum"
    NetworkAddress = "http://127.0.0.1:8545" # a network address
    PrivateKeyFile = "keys/ethereum.sk" # the path to the file containing the relayer eth private key
    SafeContractAddress = "A6504Cc508889bbDBd4B748aFf6EA6b5D0d2684c"
    GasLimitBase = 350000
    GasLimitForEach = 30000
    [Eth.GasStation]
        Enabled = true
        URL = "https://api.etherscan.io/api?module=gastracker&action=gasoracle" # gas station URL. Suggestion to provide the api-key here
        GasPriceMultiplier = 1000000000 # the value to be multiplied with the fetched value. Useful in test chains. On production chain should be 1000000000
        PollingIntervalInSeconds = 60 # number of seconds between gas price polling
        RequestRetryDelayInSeconds = 5 # number of seconds of delay after one failed request
        MaxFetchRetries = 3 # number of fetch retries before printing an error
        RequestTimeInSeconds = 2 # maximum timeout (in seconds) for the gas price request
        MaximumAllowedGasPrice = 300 # maximum value allowed for the fetched gas price value
        # GasPriceSelector available options: "SafeGasPrice", "ProposeGasPrice", "FastGasPrice"
        GasPriceSelector = "SafeGasPrice" # selector used to provide the gas price

[MultiversX]
    NetworkAddress = "https://devnet-gateway.multiversx.com" # the network address
    MultisigContractAddress = "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf" # the multiversx address for the bridge contract
    SafeContractAddress = "erd1qqqqqqqqqqqqqpgqtvnswnzxxz8susupesys0hvg7q2z5nawrcjq06qdus" # the multiversx address for the safe contract
    [MultiversX.Proxy]
        CacherExpirationSeconds = 600 # the caching time in seconds

        # valid options for ProxyRestAPIEntityType are "observer" and "proxy". Any other value will trigger an error.
        # "observer" is useful when querying an observer, directly and "proxy" is useful when querying a squad's proxy (gateway)
        RestAPIEntityType = "observer"
        FinalityCheck = true
        MaxNoncesDelta = 7 # the number of maximum blocks allowed to be "in front" of what the metachain has notarized

[Logs]
    LogFileLifeSpanInSec = 86400 # 24h
    LogFileLifeSpanInMB = 1024 # 1GB

[WhitelistedTokens]
    List = [
        "ETHUSDC-220753",
        "ETHUTK-8cdf7a",
        "ETHUSDT-9c73c6",
        "ETHBUSD-450923",
        "ETHHMT-18538a",
        "ETHCGG-ee4e0c",
        "ETHINFRA-60a3bf",
        "ETHWBTC-74e282",
        "ETHWETH-e1c126",
        "ETHWSDAI-572803",
        "ETHWDAI-bd65f9",
        "ETHUMB-291202",
    ]
`

	cfg := MigrationToolConfig{}

	err := toml.Unmarshal([]byte(testString), &cfg)

	require.Nil(t, err)
	require.Equal(t, expectedConfig, cfg)
}
