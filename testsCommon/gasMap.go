package testsCommon

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
)

// CreateTestElrondGasMap will create a testing gas map for Elrond client
func CreateTestElrondGasMap() config.ElrondGasMapConfig {
	return config.ElrondGasMapConfig{
		Sign:                   101,
		ProposeTransferBase:    102,
		ProposeTransferForEach: 103,
		ProposeStatus:          104,
		PerformActionBase:      105,
		PerformActionForEach:   106,
	}
}
