package elrond

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testHelpers"

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

func TestGetPendingDepositTransaction(t *testing.T) {
	testHelpers.SetTestLogLevel()

	blockNonce, _ := hex.DecodeString("0564a7")
	nonce, _ := hex.DecodeString("01")
	from, _ := hex.DecodeString("04aa6d6029b4e136d04848f5b588c2951185666cc871982994f7ef1654282fa3")
	to, _ := hex.DecodeString("cf95254084ab772696643f0e05ac4711ed674ac1000000000000000000000000")
	tokenIdentifier, _ := hex.DecodeString("574554482d386538333666")
	amount, _ := hex.DecodeString("01")
	responseData := [][]byte{
		blockNonce,
		nonce,
		from,
		to,
		tokenIdentifier,
		amount,
	}

	proxy := &testProxy{
		transactionCost:   1024,
		queryResponseCode: "ok",
		queryResponseData: responseData,
	}
	client, _ := buildTestClient(proxy)

	actual := client.GetPendingDepositTransaction(context.TODO())
	expected := &bridge.DepositTransaction{
		To:           "0xcf95254084ab772696643f0e05ac4711ed674ac1",
		From:         "erd1qj4x6cpfknsnd5zgfr6mtzxzj5gc2envepces2v57lh3v4pg973sqtm427",
		TokenAddress: "0x574554482d386538333666",
		Amount:       big.NewInt(1),
		DepositNonce: bridge.NewNonce(1),
	}

	assert.Equal(t, expected, actual)
}

func TestProposeTransfer(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("it will set proper function and params", func(t *testing.T) {
		tokenId, _ := hex.DecodeString("574554482d393761323662")
		proxy := &testProxy{
			transactionCost:   1024,
			queryResponseCode: "ok",
			queryResponseData: [][]byte{tokenId},
		}
		client, _ := buildTestClient(proxy)

		tx := &bridge.DepositTransaction{
			To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
			From:         "0x132A150926691F08a693721503a38affeD18d524",
			TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
			Amount:       big.NewInt(42),
			DepositNonce: bridge.NewNonce(1),
		}

		_, _ = client.ProposeTransfer(context.TODO(), tx)
		expected := "proposeMultiTransferEsdtTransferEsdtToken@01@b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba@574554482d393761323662@2a"

		assert.Equal(t, []byte(expected), proxy.lastTransaction.Data)
	})
}

func TestExecute(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("will return the transaction hash", func(t *testing.T) {
		expectedTxHash := "expected hash"
		proxy := &testProxy{transactionCost: 1024, transactionHash: expectedTxHash}
		client, _ := buildTestClient(proxy)

		hash, _ := client.Execute(context.TODO(), bridge.NewActionId(42), bridge.NewNonce(0))

		assert.Equal(t, expectedTxHash, hash)
	})
	t.Run("will increase nonce on successive runs", func(t *testing.T) {
		proxy := &testProxy{}
		client, _ := buildTestClient(proxy)

		_, _ = client.Execute(context.TODO(), bridge.NewActionId(42), bridge.NewNonce(0))
		_, _ = client.Execute(context.TODO(), bridge.NewActionId(42), bridge.NewNonce(0))

		expectedNonce := uint64(1)

		assert.Equal(t, expectedNonce, proxy.lastTransaction.Nonce)
	})
	t.Run("will not increment nonce when transactions fails", func(t *testing.T) {
		proxy := &testProxy{shouldFail: true}
		client, _ := buildTestClient(proxy)

		_, _ = client.Execute(context.TODO(), bridge.NewActionId(42), bridge.NewNonce(0))
		_, _ = client.Execute(context.TODO(), bridge.NewActionId(42), bridge.NewNonce(0))

		expectedNonce := uint64(0)

		assert.Equal(t, expectedNonce, proxy.lastTransaction.Nonce)
	})
}

func TestWasProposedTransfer(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("will return true when response is 1", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedTransfer(context.TODO(), bridge.NewNonce(0))
		assert.True(t, got)
	})
	t.Run("will return false when response is 9", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(0)}}}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedTransfer(context.TODO(), bridge.NewNonce(0))
		assert.False(t, got)
	})
	t.Run("will return false when response code is not ok", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "not ok", queryResponseData: nil}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedTransfer(context.TODO(), bridge.NewNonce(0))
		assert.False(t, got)
	})
}

func TestSignersCount(t *testing.T) {
	testHelpers.SetTestLogLevel()

	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(42)}}}
	client, _ := buildTestClient(proxy)

	got := client.SignersCount(context.TODO(), bridge.NewActionId(0))

	assert.Equal(t, uint(42), got)
}

func TestWasProposedSetStatusOnPendingTransfer(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("will return true when response is 1", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedSetStatusOnPendingTransfer(context.TODO(), bridge.Executed)
		assert.True(t, got)
	})
	t.Run("will return false when response is empty", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{}}}
		client, _ := buildTestClient(proxy)

		got := client.WasProposedSetStatusOnPendingTransfer(context.TODO(), bridge.Rejected)
		assert.False(t, got)
	})
}

func TestGetActionIdForEthTxNonce(t *testing.T) {
	testHelpers.SetTestLogLevel()

	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(42)}}}
	client, _ := buildTestClient(proxy)

	got := client.GetActionIdForProposeTransfer(context.TODO(), bridge.NewNonce(41))

	assert.Equal(t, bridge.NewActionId(42), got)
}

func TestGetActionIdForSetStatusOnPendingTransfer(t *testing.T) {
	testHelpers.SetTestLogLevel()

	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(43)}}}
	client, _ := buildTestClient(proxy)

	got := client.GetActionIdForSetStatusOnPendingTransfer(context.TODO())
	assert.Equal(t, got, bridge.NewActionId(43))
}

func TestWasExecuted(t *testing.T) {
	testHelpers.SetTestLogLevel()

	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
	client, _ := buildTestClient(proxy)

	got := client.WasExecuted(context.TODO(), bridge.NewActionId(42), bridge.NewNonce(0))
	assert.True(t, got)
}

func TestSign(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("it will set proper transaction cost", func(t *testing.T) {
		expect := uint64(1024)
		proxy := &testProxy{transactionCost: expect}
		client, _ := buildTestClient(proxy)

		_, _ = client.Sign(context.TODO(), bridge.NewActionId(42))

		assert.Equal(t, expect, proxy.lastTransaction.GasLimit)
	})
	t.Run("it will set proper function and params", func(t *testing.T) {
		proxy := &testProxy{transactionCost: 1024}
		client, _ := buildTestClient(proxy)

		_, _ = client.Sign(context.TODO(), bridge.NewActionId(42))

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
