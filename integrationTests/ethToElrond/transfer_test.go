package ethToElrond

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/stretchr/testify/require"
)

func TestEthToElrond_Succeed(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	//TODO remove this after test is completed
	logger.SetLogLevel("*:DEBUG")

	network := integrationTests.NewMockEthElrondNetwork(t, 1)
	defer network.Close(t)

	depositor := []byte("12345678901111111111")
	to, err := data.NewAddressFromBech32String("erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede")
	require.Nil(t, err)

	ticker := "TCK-RANDOM"
	tokenAddr := []byte("00000000009999999999")
	network.TokensHandler.AddNewToken(tokenAddr, ticker)
	network.EthereumContract.AddTransferRequest(tokenAddr, depositor, to, big.NewInt(5))

	//TODO finish test

	//TODO make the relayers query faster and reduce this time duration
	time.Sleep(time.Second * 120)
}
