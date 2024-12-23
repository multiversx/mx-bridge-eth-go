package testsCommon

import (
	"github.com/multiversx/mx-bridge-eth-go/config"
)

// CreateTestMultiversXGasMap will create a testing gas map for MultiversX client
func CreateTestMultiversXGasMap() config.MultiversXGasMapConfig {
	return config.MultiversXGasMapConfig{
		Sign:                   101,
		ProposeTransferBase:    102,
		ProposeTransferForEach: 103,
		ProposeStatusBase:      104,
		ProposeStatusForEach:   105,
		PerformActionBase:      106,
		PerformActionForEach:   107,
		ScCallPerByte:          108,
		ScCallPerformForEach:   109,
		AbsoluteMaxGasLimit:    500000,
	}
}
