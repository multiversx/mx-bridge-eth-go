package ethToElrond

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

func TestEthToElrond_Succeed(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	//TODO remove this after test is completed
	logger.SetLogLevel("*:DEBUG")

	network := integrationTests.NewMockEthElrondNetwork(t, 1)
	defer network.Close(t)

	//TODO finish test

	time.Sleep(time.Second * 20)
}
