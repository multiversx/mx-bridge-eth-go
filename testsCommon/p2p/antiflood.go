package p2p

import elrondConfig "github.com/ElrondNetwork/elrond-go/config"

// CreateAntifloodConfig will create a testing AntifloodConfig
func CreateAntifloodConfig() elrondConfig.AntifloodConfig {
	cfg := elrondConfig.AntifloodConfig{
		Enabled:                   true,
		NumConcurrentResolverJobs: 10,
		OutOfSpecs:                createAntiFloodPreventerConfig(),
		FastReacting:              createAntiFloodPreventerConfig(),
		SlowReacting:              createAntiFloodPreventerConfig(),
		PeerMaxOutput: elrondConfig.AntifloodLimitsConfig{
			BaseMessagesPerInterval: 100,
			TotalSizePerInterval:    1000,
			IncreaseFactor: elrondConfig.IncreaseFactorConfig{
				Threshold: 10,
				Factor:    1,
			},
		},
		Cache: elrondConfig.CacheConfig{
			Type:     "LRU",
			Capacity: 100,
			Shards:   2,
		},
		Topic: elrondConfig.TopicAntifloodConfig{
			DefaultMaxMessagesPerSec: 10,
			MaxMessages: []elrondConfig.TopicMaxMessagesConfig{
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

func createAntiFloodPreventerConfig() elrondConfig.FloodPreventerConfig {
	return elrondConfig.FloodPreventerConfig{
		IntervalInSeconds: 1,
		PeerMaxInput: elrondConfig.AntifloodLimitsConfig{
			BaseMessagesPerInterval: 100,
			TotalSizePerInterval:    1000,
		},
		BlackList: elrondConfig.BlackListConfig{
			ThresholdNumMessagesPerInterval: 100,
			ThresholdSizePerInterval:        1000,
			NumFloodingRounds:               10,
			PeerBanDurationInSeconds:        10,
		},
	}
}
