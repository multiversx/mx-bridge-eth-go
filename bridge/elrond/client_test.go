package elrond

import (
	"context"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/data"
)

var (
	_ = bridge.Bridge(&Client{})
)

type TransactionError string

func (e TransactionError) Error() string {
	return string(e)
}

type testProxy struct {
	transactionHash string
	lastTransaction *data.Transaction
	shouldFail      bool
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

func (p *testProxy) SendTransaction(tx *data.Transaction) (string, error) {
	p.lastTransaction = tx

	if p.shouldFail {
		return "", TransactionError("failed")
	} else {
		return p.transactionHash, nil
	}
}

func TestExecute(t *testing.T) {
	buildTestClient := func(proxy *testProxy) (*Client, *testProxy) {
		privateKey, err := erdgo.LoadPrivateKeyFromPemFile("grace.pem")

		if err != nil {
			t.Fatal(err)
		}

		address, err := erdgo.GetAddressFromPrivateKey(privateKey)
		if err != nil {
			t.Fatal(err)
		}

		client := &Client{
			proxy:         proxy,
			bridgeAddress: "",
			privateKey:    privateKey,
			address:       address,
			nonce:         0,
		}

		return client, proxy
	}
	t.Run("will return the transaction hash", func(t *testing.T) {
		expectedTxHash := "expected hash"
		proxy := &testProxy{expectedTxHash, nil, false}
		client, _ := buildTestClient(proxy)

		hash, _ := client.Execute(context.TODO(), nil)

		if hash != expectedTxHash {
			t.Errorf("Expected %q, got %q", expectedTxHash, hash)
		}
	})
	t.Run("will increase nonce on successive runs", func(t *testing.T) {
		proxy := &testProxy{"", nil, false}
		client, proxy := buildTestClient(proxy)

		_, _ = client.Execute(context.TODO(), nil)
		_, _ = client.Execute(context.TODO(), nil)

		expectedNonce := uint64(1)

		if proxy.lastTransaction.Nonce != expectedNonce {
			t.Errorf("Expected nonce to be %v, but it was %v", expectedNonce, proxy.lastTransaction.Nonce)
		}
	})
	t.Run("will not increment nonce when transactions fails", func(t *testing.T) {
		proxy := &testProxy{"", nil, true}
		client, proxy := buildTestClient(proxy)

		_, _ = client.Execute(context.TODO(), nil)
		_, _ = client.Execute(context.TODO(), nil)

		expectedNonce := uint64(0)

		if proxy.lastTransaction.Nonce != expectedNonce {
			t.Errorf("Expected nonce to be %v, but it was %v", expectedNonce, proxy.lastTransaction.Nonce)
		}
	})
}
