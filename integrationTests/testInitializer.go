package integrationTests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock/contracts"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/require"
)

const ethContractAddress = "5DdDe022a65F8063eE9adaC54F359CBF46166068"
const ethPrivateKey = "9b7971db47e3815a669a91c3f1bcb21e0b81f2de04bf11faa7a34b9b10e7cfbb"
const elrondContractAdress = "erd1qqqqqqqqqqqqqpgqgftcwj09u0nhmskrw7xxqcqh8qmzwyexd8ss7ftcxx"

var log = logger.GetOrCreate("integrationTests")

// MockEthElrondNetwork represents a network of relayers that communicate with mocked elrond & eth clients
type MockEthElrondNetwork struct {
	Seeder           p2p.Messenger
	Relayers         []*relay.Relay //TODO: replace this with an interface
	ElrondClient     *mock.ElrondMockClient
	ElrondContract   *contracts.ElrondContract
	EthereumClient   *mock.EthereumMockClient
	EthereumContract *contracts.EthereumContract
	TokensHandler    contracts.TokensHandler
	cancelFunc       func()
}

// NewMockEthElrondNetwork creates a mocked eth elrond network with mocked clients
func NewMockEthElrondNetwork(tb testing.TB, numRelayers int) *MockEthElrondNetwork {
	var err error

	network := &MockEthElrondNetwork{
		Seeder:         integrationTests.CreateMessengerWithKadDht(""),
		ElrondClient:   mock.NewElrondMockClient(),
		EthereumClient: mock.NewEthereumMockClient(),
		TokensHandler:  mock.NewTokensHolder(),
	}
	elrondContractAddress := "erd1qqqqqqqqqqqqqpgqgftcwj09u0nhmskrw7xxqcqh8qmzwyexd8ss7ftcxx" //TODO remove this hardcoded value
	network.ElrondContract, err = contracts.NewElrondContract(elrondContractAddress, network.TokensHandler)
	require.Nil(tb, err)
	network.ElrondClient.SetAccount(nil, network.ElrondContract.Contract)
	network.ElrondContract.WhiteListAddress("erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede") //TODO remove this hardcoded value

	network.EthereumContract, err = contracts.NewEthereumContract("0x" + strings.ToLower(ethContractAddress))
	require.Nil(tb, err)
	network.EthereumClient.SetAccount(nil, network.EthereumContract.Contract)

	log.Info("Elrond mock client", "URL", network.ElrondClient.URL())

	var ctx context.Context
	ctx, network.cancelFunc = context.WithCancel(context.Background())

	for i := 0; i < numRelayers; i++ {
		name := fmt.Sprintf("relayer_%d", i)

		cfg := &relay.Config{
			Eth: bridge.Config{
				NetworkAddress: network.EthereumClient.URL(),
				BridgeAddress:  ethContractAddress,
				PrivateKey:     ethPrivateKey,
				GasLimit:       500000,
			},
			Elrond: bridge.Config{
				NetworkAddress:               network.ElrondClient.URL(),
				BridgeAddress:                elrondContractAdress,
				PrivateKey:                   "../testdata/grace.pem", //TODO replace here with a crypto.PrivateKey
				IntervalToResendTxsInSeconds: 5,
			},
			P2P: relay.ConfigP2P{
				Port:            "0",
				Seed:            "",
				InitialPeerList: network.Seeder.Addresses(),
				ProtocolID:      "/erd/kad/1.0.0",
			},
		}

		var r *relay.Relay
		r, err = relay.NewRelay(cfg, name)
		require.Nil(tb, err)
		go func() {
			errStart := r.Start(ctx)
			require.Nil(tb, errStart)
		}()

		network.Relayers = append(network.Relayers, r)
	}

	return network
}

// Close will close any relayers/clients opened
func (meen *MockEthElrondNetwork) Close(tb testing.TB) {
	err := meen.Seeder.Close()
	require.Nil(tb, err)

	meen.cancelFunc()

	for _, relayer := range meen.Relayers {
		relayer.Clean()
	}

	meen.ElrondClient.Close()
}
