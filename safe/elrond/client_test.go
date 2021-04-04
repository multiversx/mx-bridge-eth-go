package elrond

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/data"
	"testing"
)

var (
	_ = safe.Safe(&Client{})
)

type testProxy struct {
	expectedTxHash string
}

func (p *testProxy) GetNetworkConfig() (*data.NetworkConfig, error) {
	return &data.NetworkConfig{
		ChainID:                  "test-chain",
		Denomination:             0,
		GasPerDataByte:           0,
		LatestTagSoftwareVersion: "",
		MetaConsensusGroup:       0,
		MinGasLimit:              84,
		MinGasPrice:              12,
		MinTransactionVersion:    42,
		NumMetachainNodes:        0,
		NumNodesInShard:          0,
		NumShardsWithoutMeta:     0,
		RoundDuration:            0,
		ShardConsensusGroupSize:  0,
		StartTime:                0,
	}, nil
}

func (p *testProxy) SendTransaction(*data.Transaction) (string, error) {
	return p.expectedTxHash, nil
}

func TestBridge(t *testing.T) {
	expectedTxHash := "this is the tx hash"

	proxy := &testProxy{
		expectedTxHash,
	}

	privateKey, err := erdgo.LoadPrivateKeyFromPemFile("grace.pem")

	if err != nil {
		t.Fatal(err)
	}

	addressString, err := erdgo.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	address := &elrondAddress{addressString: addressString}
	account := buildMockAccount(addressString)

	client := Client{
		proxy:       proxy,
		safeAddress: "",
		privateKey:  privateKey,

		account: account,
		address: address,
	}

	hash, _ := client.Bridge(nil)

	if hash != expectedTxHash {
		t.Errorf("Expected %q, got %q", expectedTxHash, hash)
	}
}

func buildMockAccount(address string) *data.Account {
	return &data.Account{
		Address:  address,
		Nonce:    42,
		Balance:  "42",
		Code:     "42",
		CodeHash: nil,
		RootHash: nil,
	}
}
