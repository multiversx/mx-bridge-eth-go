package factory

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/gasManagement/disabled"
	"github.com/stretchr/testify/assert"
)

func createMockArgsGasStation() gasManagement.ArgsGasStation {
	return gasManagement.ArgsGasStation{
		RequestURL:             "",
		RequestPollingInterval: time.Second,
		RequestTime:            time.Second,
		MaximumGasPrice:        1000,
		GasPriceSelector:       "fast",
	}
}

func TestNewGasStation(t *testing.T) {
	t.Parallel()
	args := createMockArgsGasStation()
	t.Run("disabledGasStation", func(t *testing.T) {
		gs, err := CreateGasStation(args, false)

		_, ok := gs.(*disabled.DisabledGasStation)

		assert.True(t, ok)
		assert.Nil(t, err)
	})
	t.Run("normal gasStation", func(t *testing.T) {
		gs, err := CreateGasStation(args, true)

		assert.Equal(t, "*gasManagement.gasStation", fmt.Sprintf("%T", gs))
		assert.Nil(t, err)
	})
}
