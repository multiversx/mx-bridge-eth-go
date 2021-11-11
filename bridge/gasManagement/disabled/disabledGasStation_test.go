package disabled

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewDisabledGasStation(t *testing.T) {
	dgs := &DisabledGasStation{}

	assert.False(t, check.IfNil(dgs))

	gasPrice, err := dgs.GetCurrentGasPriceInWei()
	assert.Nil(t, gasPrice)
	assert.Nil(t, err)
}
