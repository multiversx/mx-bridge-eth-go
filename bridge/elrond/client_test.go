package elrond

import (
	"context"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"

	"github.com/ElrondNetwork/elrond-go/data/vm"

	"github.com/stretchr/testify/assert"

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

func TestProposeTransfer(t *testing.T) {
	t.Run("it will set proper function and params", func(t *testing.T) {
		proxy := &testProxy{transactionCost: 1024}
		client, _ := buildTestClient(proxy)

		tx := &bridge.DepositTransaction{
			To:           "",
			From:         "",
			TokenAddress: "",
			Amount:       nil,
			DepositNonce: bridge.Nonce(1),
		}
		_, _ = client.ProposeTransfer(context.TODO(), tx)

		assert.Equal(t, []byte("proposeMultiTransferEsdtTransferEsdtToken@01"), proxy.lastTransaction.Data)
	})
}

func TestExecute(t *testing.T) {
	t.Run("will return the transaction hash", func(t *testing.T) {
		expectedTxHash := "expected hash"
		proxy := &testProxy{transactionCost: 1024, transactionHash: expectedTxHash}
		client, _ := buildTestClient(proxy)

		hash, _ := client.Execute(context.TODO(), 42)

		assert.Equal(t, expectedTxHash, hash)
	})
	t.Run("will increase nonce on successive runs", func(t *testing.T) {
		proxy := &testProxy{}
		client, _ := buildTestClient(proxy)

		_, _ = client.Execute(context.TODO(), 42)
		_, _ = client.Execute(context.TODO(), 42)

		expectedNonce := uint64(1)

		assert.Equal(t, expectedNonce, proxy.lastTransaction.Nonce)
	})
	t.Run("will not increment nonce when transactions fails", func(t *testing.T) {
		proxy := &testProxy{shouldFail: true}
		client, _ := buildTestClient(proxy)

		_, _ = client.Execute(context.TODO(), 42)
		_, _ = client.Execute(context.TODO(), 42)

		expectedNonce := uint64(0)

		assert.Equal(t, expectedNonce, proxy.lastTransaction.Nonce)
	})
}

func TestWasProposedTransfer(t *testing.T) {
	t.Run("will return true when response is 1", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedTransfer(context.TODO(), bridge.Nonce(0))
		assert.True(t, got)
	})
	t.Run("will return false when response is 9", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(0)}}}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedTransfer(context.TODO(), bridge.Nonce(0))
		assert.False(t, got)
	})
	t.Run("will return false when response code is not ok", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "not ok", queryResponseData: nil}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedTransfer(context.TODO(), bridge.Nonce(0))
		assert.False(t, got)
	})
}

func TestWasProposedSetStatusSuccessOnPendingTransfer(t *testing.T) {
	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
	client, _ := buildTestClient(proxy)

	got := client.WasProposedSetStatusSuccessOnPendingTransfer(context.TODO())
	assert.True(t, got)
}

func TestWasProposedSetStatusFailedOnPendingTransfer(t *testing.T) {
	t.Run("will return true when response is 1", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedSetStatusFailedOnPendingTransfer(context.TODO())
		assert.True(t, got)
	})
	t.Run("will return false when response is empty", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{}}}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedSetStatusFailedOnPendingTransfer(context.TODO())
		assert.False(t, got)
	})
}

func TestGetActionIdForEthTxNonce(t *testing.T) {
	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(42)}}}
	client, _ := buildTestClient(proxy)

	got := client.GetActionIdForProposeTransfer(context.TODO(), bridge.Nonce(41))
	assert.Equal(t, got, bridge.ActionId(42))
}

func TestGetActionIdForSetStatusOnPendingTransfer(t *testing.T) {
	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(43)}}}
	client, _ := buildTestClient(proxy)

	got := client.GetActionIdForSetStatusOnPendingTransfer(context.TODO())
	assert.Equal(t, got, bridge.ActionId(43))
}

func TestWasExecuted(t *testing.T) {
	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
	client, _ := buildTestClient(proxy)

	got := client.WasExecuted(context.TODO(), bridge.ActionId(42))
	assert.True(t, got)
}

func TestSign(t *testing.T) {
	t.Run("it will set proper transaction cost", func(t *testing.T) {
		expect := uint64(1024)
		proxy := &testProxy{transactionCost: expect}
		client, _ := buildTestClient(proxy)

		_, _ = client.Sign(context.TODO(), bridge.ActionId(42))

		assert.Equal(t, expect, proxy.lastTransaction.GasLimit)
	})
	t.Run("it will set proper function and params", func(t *testing.T) {
		proxy := &testProxy{transactionCost: 1024}
		client, _ := buildTestClient(proxy)

		_, _ = client.Sign(context.TODO(), bridge.ActionId(42))

		assert.Equal(t, []byte("sign@2a"), proxy.lastTransaction.Data)
	})
}

func buildTestClient(proxy *testProxy) (*Client, error) {
	privateKey, err := erdgo.LoadPrivateKeyFromPemFile("grace.pem")
	if err != nil {
		return nil, err
	}

	address, err := erdgo.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	client := &Client{
		log:           logger.GetOrCreate("testHelpers"),
		proxy:         proxy,
		bridgeAddress: "",
		privateKey:    privateKey,
		address:       address,
		nonce:         0,
	}

	return client, nil
}

type testProxy struct {
	transactionHash string
	lastTransaction *data.Transaction
	shouldFail      bool

	queryResponseData [][]byte
	queryResponseCode string

	transactionCost      uint64
	transactionCostError error
}

func (p *testProxy) GetNetworkConfig() (*data.NetworkConfig, error) {
	return &data.NetworkConfig{
		ChainID:                  "testHelpers-chain",
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

func (p *testProxy) GetTransactionInfoWithResults(string) (*data.TransactionInfo, error) {
	return nil, nil
}

func (p *testProxy) RequestTransactionCost(*data.Transaction) (*data.TxCostResponseData, error) {
	return &data.TxCostResponseData{
		TxCost:     p.transactionCost,
		RetMessage: "",
	}, p.transactionCostError
}

func (p *testProxy) ExecuteVMQuery(*data.VmValueRequest) (*data.VmValuesResponseData, error) {
	return &data.VmValuesResponseData{Data: &vm.VMOutputApi{ReturnCode: p.queryResponseCode, ReturnData: p.queryResponseData}}, nil
}
