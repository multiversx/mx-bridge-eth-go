package elrond

import (
	"context"
	"encoding/hex"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/elrond/mock"
	"github.com/ElrondNetwork/elrond-eth-bridge/testHelpers"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ = bridge.Bridge(&client{})
)

type TransactionError string

func (e TransactionError) Error() string {
	return string(e)
}

func createMockArguments() ClientArgs {
	return ClientArgs{
		Config: bridge.Config{
			BridgeAddress:        "bridge address",
			PrivateKey:           "grace.pem",
			NonceUpdateInSeconds: 1,
		},
		Proxy: &mock.ElrondProxyStub{},
	}
}

func TestNewClient(t *testing.T) {
	t.Run("wrong NonceUpdateInSeconds value", func(t *testing.T) {
		args := createMockArguments()
		args.Config.NonceUpdateInSeconds = 0
		c, err := NewClient(args)
		require.Nil(t, c)
		require.Equal(t, ErrInvalidNonceUpdateInterval, err)
	})
	t.Run("nil proxy", func(t *testing.T) {
		args := createMockArguments()
		args.Proxy = nil
		c, err := NewClient(args)
		require.Nil(t, c)
		require.Equal(t, ErrNilProxy, err)
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArguments()
		c, err := NewClient(args)
		require.Nil(t, err)
		require.NotNil(t, c)
	})
}

func TestClient_PollsNonce(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	args.Config.NonceUpdateInSeconds = 1
	numCalled := uint64(0)
	args.Proxy = &mock.ElrondProxyStub{
		GetAccountCalled: func(address core.AddressHandler) (*data.Account, error) {
			atomic.AddUint64(&numCalled, 1)

			return &data.Account{
				Nonce: atomic.LoadUint64(&numCalled),
			}, nil
		},
	}
	c, _ := NewClient(args)
	require.NotNil(t, c)

	// sleep for 3.5 seconds (do not use a "round number" of seconds as it will make the
	// test flaky)
	time.Sleep(time.Second*3 + time.Millisecond*500)

	err := c.Close()
	require.Nil(t, err)

	// the go routine should have stopped, preventing calls to the proxy
	time.Sleep(time.Second*3 + time.Millisecond*500)

	assert.Equal(t, uint64(4), atomic.LoadUint64(&numCalled))
	assert.Equal(t, uint64(4), atomic.LoadUint64(&c.nonce))
}

func TestGetPending(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("when there is a current transaction", func(t *testing.T) {
		batchId, _ := hex.DecodeString("01")
		blockNonce1, _ := hex.DecodeString("025d43")
		nonce1, _ := hex.DecodeString("01")
		from1, _ := hex.DecodeString("b4b6b2377f786d9dd3745695bb839434f94acb47a027a66f0069b8b8389551a5")
		to1, _ := hex.DecodeString("264eeffe37aa569bec16a951c51ba25a98e07dab")
		tokenIdentifier1, _ := hex.DecodeString("574554482d656366316331")
		amount1, _ := hex.DecodeString("01")
		blockNonce2, _ := hex.DecodeString("025d43")
		nonce2, _ := hex.DecodeString("02")
		from2, _ := hex.DecodeString("b4b6b2377f786d9dd3745695bb839434f94acb47a027a66f0069b8b8389551a5")
		to2, _ := hex.DecodeString("264eeffe37aa569bec16a951c51ba25a98e07dab")
		tokenIdentifier2, _ := hex.DecodeString("574554482d656366316331")
		amount2, _ := hex.DecodeString("02")
		responseData := [][]byte{
			batchId,
			blockNonce1,
			nonce1,
			from1,
			to1,
			tokenIdentifier1,
			amount1,
			blockNonce2,
			nonce2,
			from2,
			to2,
			tokenIdentifier2,
			amount2,
		}

		proxy := &testProxy{
			transactionCost:   1024,
			queryResponseCode: "ok",
			queryResponseData: responseData,
		}
		c, _ := buildTestClient(proxy)

		actual := c.GetPending(context.TODO())
		tx1 := &bridge.DepositTransaction{
			To:           "0x264eeffe37aa569bec16a951c51ba25a98e07dab",
			From:         "erd1kjmtydml0pkem5m5262mhqu5xnu54j685qn6vmcqdxutswy42xjskgdla5",
			TokenAddress: "0x574554482d656366316331",
			Amount:       big.NewInt(1),
			DepositNonce: bridge.NewNonce(1),
			BlockNonce:   bridge.NewNonce(154947),
			Status:       0,
			Error:        nil,
		}
		tx2 := &bridge.DepositTransaction{
			To:           "0x264eeffe37aa569bec16a951c51ba25a98e07dab",
			From:         "erd1kjmtydml0pkem5m5262mhqu5xnu54j685qn6vmcqdxutswy42xjskgdla5",
			TokenAddress: "0x574554482d656366316331",
			Amount:       big.NewInt(2),
			DepositNonce: bridge.NewNonce(2),
			BlockNonce:   bridge.NewNonce(154947),
			Status:       0,
			Error:        nil,
		}
		expected := &bridge.Batch{
			Id:           bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{tx1, tx2},
		}

		assert.Equal(t, expected, actual)
	})
	t.Run("when there is no current transaction it will call get pending", func(t *testing.T) {
		batchId, _ := hex.DecodeString("01")
		blockNonce, _ := hex.DecodeString("0564a7")
		nonce, _ := hex.DecodeString("01")
		from, _ := hex.DecodeString("04aa6d6029b4e136d04848f5b588c2951185666cc871982994f7ef1654282fa3")
		to, _ := hex.DecodeString("cf95254084ab772696643f0e05ac4711ed674ac1")
		tokenIdentifier, _ := hex.DecodeString("574554482d386538333666")
		amount, _ := hex.DecodeString("01")
		responseData := [][]byte{
			batchId,
			blockNonce,
			nonce,
			from,
			to,
			tokenIdentifier,
			amount,
		}

		proxy := &testProxy{
			transactionCost:                   1024,
			queryResponseCode:                 "ok",
			queryResponseData:                 [][]byte{{}},
			afterTransactionQueryResponseData: responseData,
		}

		c, _ := buildTestClient(proxy)
		actual := c.GetPending(context.TODO())
		expected := &bridge.Batch{
			Id: bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{
				{
					To:           "0xcf95254084ab772696643f0e05ac4711ed674ac1",
					From:         "erd1qj4x6cpfknsnd5zgfr6mtzxzj5gc2envepces2v57lh3v4pg973sqtm427",
					TokenAddress: "0x574554482d386538333666",
					Amount:       big.NewInt(1),
					DepositNonce: bridge.NewNonce(1),
					BlockNonce:   bridge.NewNonce(353447),
					Status:       0,
					Error:        nil,
				},
			},
		}

		assert.Equal(t, expected, actual)
		assert.Equal(t, uint64(260_000_000), proxy.lastTransaction.GasLimit)
	})
	t.Run("where there is no pending transaction it will return nil", func(t *testing.T) {
		proxy := &testProxy{
			transactionCost:   1024,
			queryResponseCode: "ok",
			shouldFail:        true,
		}

		c, _ := buildTestClient(proxy)
		actual := c.GetPending(context.TODO())

		assert.Nil(t, actual)
	})
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
		c, _ := buildTestClient(proxy)

		batch := &bridge.Batch{
			Id: bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{
				{
					To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
					From:         "0x132A150926691F08a693721503a38affeD18d524",
					TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
					Amount:       big.NewInt(42),
					DepositNonce: bridge.NewNonce(1),
				},
			},
		}

		_, _ = c.ProposeTransfer(context.TODO(), batch)
		expected := "proposeMultiTransferEsdtBatch@01@b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba@574554482d393761323662@2a"

		assert.Equal(t, []byte(expected), proxy.lastTransaction.Data)
		assert.Equal(t, uint64(45_000_000+len(batch.Transactions)*25_000_000), proxy.lastTransaction.GasLimit)
	})
}

func TestProposeSetStatus(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("it will set proper function and params", func(t *testing.T) {
		proxy := &testProxy{
			transactionCost:   1024,
			queryResponseCode: "ok",
			queryResponseData: [][]byte{},
		}
		c, _ := buildTestClient(proxy)

		batch := &bridge.Batch{
			Id: bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{
				{
					To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
					From:         "0x132A150926691F08a693721503a38affeD18d524",
					TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
					Amount:       big.NewInt(42),
					DepositNonce: bridge.NewNonce(1),
					Status:       bridge.Executed,
				},
				{
					To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
					From:         "0x132A150926691F08a693721503a38affeD18d524",
					TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
					Amount:       big.NewInt(42),
					DepositNonce: bridge.NewNonce(1),
					Status:       bridge.Rejected,
				},
			},
		}

		c.ProposeSetStatus(context.TODO(), batch)
		expected := "proposeEsdtSafeSetCurrentTransactionBatchStatus@01@03@04"

		assert.Equal(t, []byte(expected), proxy.lastTransaction.Data)
		assert.Equal(t, uint64(60_000_000), proxy.lastTransaction.GasLimit)
	})
}

func TestExecute(t *testing.T) {
	testHelpers.SetTestLogLevel()

	expectedTxHash := "expected hash"
	proxy := &testProxy{transactionCost: 1024, transactionHash: expectedTxHash}
	c, _ := buildTestClient(proxy)

	batch := &bridge.Batch{
		Id: bridge.NewBatchId(1),
		Transactions: []*bridge.DepositTransaction{
			{
				To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
				From:         "0x132A150926691F08a693721503a38affeD18d524",
				TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
				Amount:       big.NewInt(42),
				DepositNonce: bridge.NewNonce(1),
				Status:       bridge.Executed,
			},
			{
				To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
				From:         "0x132A150926691F08a693721503a38affeD18d524",
				TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
				Amount:       big.NewInt(42),
				DepositNonce: bridge.NewNonce(1),
				Status:       bridge.Rejected,
			},
		},
	}
	hash, _ := c.Execute(context.TODO(), bridge.NewActionId(42), batch)

	assert.Equal(t, expectedTxHash, hash)
	assert.Equal(t, uint64(70_000_000+len(batch.Transactions)*30_000_000), proxy.lastTransaction.GasLimit)
	assert.Equal(t, uint64(43), c.nonce)
}

func TestWasProposedTransfer(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("will return true when response is 1", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
		c, _ := buildTestClient(proxy)

		batch := &bridge.Batch{
			Id: bridge.NewBatchId(12),
			Transactions: []*bridge.DepositTransaction{
				{
					To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
					From:         "0x132A150926691F08a693721503a38affeD18d524",
					TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
					Amount:       big.NewInt(42),
					DepositNonce: bridge.NewNonce(1),
					Status:       bridge.Executed,
				},
			},
		}

		got := c.WasProposedTransfer(context.TODO(), batch)
		assert.True(t, got)
	})
	t.Run("will return false when response is 9", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(0)}}}
		c, _ := buildTestClient(proxy)

		batch := &bridge.Batch{
			Id: bridge.NewBatchId(41),
			Transactions: []*bridge.DepositTransaction{
				{
					To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
					From:         "0x132A150926691F08a693721503a38affeD18d524",
					TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
					Amount:       big.NewInt(42),
					DepositNonce: bridge.NewNonce(1),
				},
			},
		}

		got := c.WasProposedTransfer(context.TODO(), batch)
		assert.False(t, got)
	})
	t.Run("will send tx's as arguments", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
		c, _ := buildTestClient(proxy)

		batch := &bridge.Batch{
			Id: bridge.NewBatchId(41),
			Transactions: []*bridge.DepositTransaction{
				{
					To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
					From:         "0x132A150926691F08a693721503a38affeD18d524",
					TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
					Amount:       big.NewInt(42),
					DepositNonce: bridge.NewNonce(1),
				},
			},
		}

		_ = c.WasProposedTransfer(context.TODO(), batch)

		assert.Equal(t, 4, len(proxy.lastQueryArgs))
		// batchID
		assert.Equal(t, "29", proxy.lastQueryArgs[0])
		// tx to address
		assert.Equal(t, "b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba", proxy.lastQueryArgs[1])
		// tokenId
		assert.Equal(t, "01", proxy.lastQueryArgs[2])
		// amount
		assert.Equal(t, "2a", proxy.lastQueryArgs[3])
	})
	t.Run("will return false when response code is not ok", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "not ok", queryResponseData: nil}
		c, _ := buildTestClient(proxy)

		batch := &bridge.Batch{
			Id: bridge.NewBatchId(41),
			Transactions: []*bridge.DepositTransaction{
				{
					To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
					From:         "0x132A150926691F08a693721503a38affeD18d524",
					TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
					Amount:       big.NewInt(42),
					DepositNonce: bridge.NewNonce(1),
				},
			},
		}

		got := c.WasProposedTransfer(context.TODO(), batch)
		assert.False(t, got)
	})
}

func TestSignersCount(t *testing.T) {
	testHelpers.SetTestLogLevel()

	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(42)}}}
	c, _ := buildTestClient(proxy)

	got := c.SignersCount(context.TODO(), bridge.NewActionId(0))

	assert.Equal(t, uint(42), got)
}

func TestWasProposedSetStatus(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("will return true when response is 1", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
		c, _ := buildTestClient(proxy)

		batch := &bridge.Batch{
			Id: bridge.NewBatchId(1),
			Transactions: []*bridge.DepositTransaction{
				{
					Status: bridge.Rejected,
				},
			},
		}
		got := c.WasProposedSetStatus(context.TODO(), batch)

		assert.True(t, got)
		assert.Equal(t, "01", proxy.lastQueryArgs[0])
		assert.Equal(t, "04", proxy.lastQueryArgs[1])
	})
	t.Run("will return false when response is empty", func(t *testing.T) {
		proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{}}}
		c, _ := buildTestClient(proxy)

		batch := &bridge.Batch{
			Id:           bridge.NewBatchId(0),
			Transactions: []*bridge.DepositTransaction{},
		}
		got := c.WasProposedSetStatus(context.TODO(), batch)

		assert.False(t, got)
	})
}

func TestGetActionIdForProposeTransfer(t *testing.T) {
	testHelpers.SetTestLogLevel()

	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(42)}}}
	c, _ := buildTestClient(proxy)

	batch := &bridge.Batch{
		Id: bridge.NewBatchId(41),
		Transactions: []*bridge.DepositTransaction{
			{
				To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
				From:         "0x132A150926691F08a693721503a38affeD18d524",
				TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
				Amount:       big.NewInt(42),
				DepositNonce: bridge.NewNonce(1),
			},
		},
	}

	got := c.GetActionIdForProposeTransfer(context.TODO(), batch)

	assert.Equal(t, bridge.NewActionId(42), got)
	assert.Equal(t, 4, len(proxy.lastQueryArgs))
	// batchID
	assert.Equal(t, "29", proxy.lastQueryArgs[0])
	// tx to address
	assert.Equal(t, "b2a11555ce521e4944e09ab17549d85b487dcd26c84b5017a39e31a3670889ba", proxy.lastQueryArgs[1])
	// tokenId
	assert.Equal(t, "2a", proxy.lastQueryArgs[2])
	// amount
	assert.Equal(t, "2a", proxy.lastQueryArgs[3])
}

func TestGetActionIdForSetStatusOnPendingTransfer(t *testing.T) {
	testHelpers.SetTestLogLevel()

	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(43)}}}
	c, _ := buildTestClient(proxy)

	batch := &bridge.Batch{
		Id: bridge.NewBatchId(12),
		Transactions: []*bridge.DepositTransaction{
			{
				To:           "erd1k2s324ww2g0yj38qn2ch2jwctdy8mnfxep94q9arncc6xecg3xaq6mjse8",
				From:         "0x132A150926691F08a693721503a38affeD18d524",
				TokenAddress: "0x3a41ed2dD119E44B802c87E84840F7C85206f4f1",
				Amount:       big.NewInt(42),
				DepositNonce: bridge.NewNonce(1),
				Status:       bridge.Executed,
			},
		},
	}

	got := c.GetActionIdForSetStatusOnPendingTransfer(context.TODO(), batch)
	assert.Equal(t, got, bridge.NewActionId(43))
	assert.Equal(t, "0c", proxy.lastQueryArgs[0])
	assert.Equal(t, "03", proxy.lastQueryArgs[1])
}

func TestWasExecuted(t *testing.T) {
	testHelpers.SetTestLogLevel()

	proxy := &testProxy{queryResponseCode: "ok", queryResponseData: [][]byte{{byte(1)}}}
	c, _ := buildTestClient(proxy)

	got := c.WasExecuted(context.TODO(), bridge.NewActionId(42), bridge.NewBatchId(0))
	assert.True(t, got)
}

func TestSign(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("it will set proper transaction cost", func(t *testing.T) {
		expect := uint64(45_000_000)
		proxy := &testProxy{}
		c, _ := buildTestClient(proxy)

		_, _ = c.Sign(context.TODO(), bridge.NewActionId(42))

		assert.Equal(t, expect, proxy.lastTransaction.GasLimit)
	})
	t.Run("it will set proper function and params", func(t *testing.T) {
		proxy := &testProxy{transactionCost: 1024}
		c, _ := buildTestClient(proxy)

		_, _ = c.Sign(context.TODO(), bridge.NewActionId(42))

		assert.Equal(t, []byte("sign@2a"), proxy.lastTransaction.Data)
	})
}

func TestIsWhitelisted(t *testing.T) {
	testHelpers.SetTestLogLevel()

	t.Run("where role is 2 it will return true", func(t *testing.T) {
		role, _ := hex.DecodeString("02")
		responseData := [][]byte{role}
		proxy := &testProxy{
			transactionCost:   1024,
			queryResponseCode: "ok",
			queryResponseData: responseData,
		}
		c, _ := buildTestClient(proxy)

		isWhitelisted := c.IsWhitelisted("some address")

		assert.True(t, isWhitelisted)
	})
	t.Run("where role is 1 it will return false", func(t *testing.T) {
		role, _ := hex.DecodeString("01")
		responseData := [][]byte{role}
		proxy := &testProxy{
			transactionCost:   1024,
			queryResponseCode: "ok",
			queryResponseData: responseData,
		}
		c, _ := buildTestClient(proxy)

		isWhitelisted := c.IsWhitelisted("some address")

		assert.False(t, isWhitelisted)
	})
}

func buildTestClient(proxy *testProxy) (*client, error) {
	wallet := interactors.NewWallet()

	privateKey, err := wallet.LoadPrivateKeyFromPemFile("grace.pem")
	if err != nil {
		return nil, err
	}

	address, err := wallet.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	c := &client{
		log:           logger.GetOrCreate("testHelpers"),
		proxy:         proxy,
		nonce:         42,
		bridgeAddress: "",
		privateKey:    privateKey,
		address:       address,
	}

	return c, nil
}

//TODO move this in mock package
type testProxy struct {
	transactionHash string
	lastTransaction *data.Transaction
	shouldFail      bool

	queryResponseData                 [][]byte
	afterTransactionQueryResponseData [][]byte
	queryResponseCode                 string
	lastQueryArgs                     []string

	transactionCost uint64
}

// GetNetworkConfig -
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

// SendTransaction -
func (p *testProxy) SendTransaction(tx *data.Transaction) (string, error) {
	p.lastTransaction = tx
	p.queryResponseData = p.afterTransactionQueryResponseData

	if p.shouldFail {
		return "", TransactionError("failed")
	} else {
		return p.transactionHash, nil
	}
}

// GetTransactionInfoWithResults -
func (p *testProxy) GetTransactionInfoWithResults(string) (*data.TransactionInfo, error) {
	return nil, nil
}

// ExecuteVMQuery -
func (p *testProxy) ExecuteVMQuery(valueRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
	p.lastQueryArgs = valueRequest.Args
	return &data.VmValuesResponseData{Data: &vm.VMOutputApi{ReturnCode: p.queryResponseCode, ReturnData: p.queryResponseData}}, nil
}

// GetAccount -
func (p *testProxy) GetAccount(_ core.AddressHandler) (*data.Account, error) {
	return &data.Account{}, nil
}

// IsInterfaceNil -
func (p *testProxy) IsInterfaceNil() bool {
	return p == nil
}
