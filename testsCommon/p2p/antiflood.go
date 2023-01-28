package p2p

import chainConfig "github.com/multiversx/mx-chain-go/config"

// CreateAntifloodConfig will create a testing AntifloodConfig
func CreateAntifloodConfig() chainConfig.AntifloodConfig {
	cfg := chainConfig.AntifloodConfig{
		Enabled:                   true,
		NumConcurrentResolverJobs: 10,
		OutOfSpecs:                createAntiFloodPreventerConfig(),
		FastReacting:              createAntiFloodPreventerConfig(),
		SlowReacting:              createAntiFloodPreventerConfig(),
		PeerMaxOutput: chainConfig.AntifloodLimitsConfig{
			BaseMessagesPerInterval: 100,
			TotalSizePerInterval:    2000,
			IncreaseFactor: chainConfig.IncreaseFactorConfig{
				Threshold: 10,
				Factor:    1,
			},
		},
		Cache: chainConfig.CacheConfig{
			Type:     "LRU",
			Capacity: 100,
			Shards:   2,
		},
		Topic: chainConfig.TopicAntifloodConfig{
			DefaultMaxMessagesPerSec: 10,
			MaxMessages: []chainConfig.TopicMaxMessagesConfig{
				{
					Topic:             "test_join",
					NumMessagesPerSec: 10,
				},
				{
					Topic:             "test_sign",
					NumMessagesPerSec: 10,
				},
			},
		},
	}

	return cfg
}

func createAntiFloodPreventerConfig() chainConfig.FloodPreventerConfig {
	return chainConfig.FloodPreventerConfig{
		IntervalInSeconds: 1,
		PeerMaxInput: chainConfig.AntifloodLimitsConfig{
			BaseMessagesPerInterval: 100,
			TotalSizePerInterval:    2000,
		},
		BlackList: chainConfig.BlackListConfig{
			ThresholdNumMessagesPerInterval: 100,
			ThresholdSizePerInterval:        2000,
			NumFloodingRounds:               10,
			PeerBanDurationInSeconds:        10,
		},
	}
}
