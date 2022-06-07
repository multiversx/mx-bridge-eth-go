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
		ProposeStatusBase:      104,
		ProposeStatusForEach:   105,
		PerformActionBase:      106,
		PerformActionForEach:   107,
	}
}
