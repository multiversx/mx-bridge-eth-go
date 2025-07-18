package sui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/sui/dtos"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

var batchNonce = uint64(7)

func createMockArgsSuiClientDataGetter() ArgsSuiClientDataGetter {
	return ArgsSuiClientDataGetter{
		SafePackageId:              "0x674a8fc0a6b48c8efea86ad7ed962107c5c132a78e7cc79c9c5b9391ba8b6d83",
		SafeObjectId:               "0x32c8ebf5853163472964ce226b194af05aa5f2e4cc47678924ebe538a1c88416",
		SafeInitialSharedVersion:   1,
		BridgePackageId:            "0x19ecc8df0b93999f87b5e2e531bb80a49d1ab2f4644032cd31f2c53ed624c94a",
		BridgeObjectId:             "0xe15513cc93d6efbfbdc7844df141b312bb677ee564a5838b7b22a891f9f05c65",
		BridgeInitialSharedVersion: 2,
		RelayerAddress:             "mock-relayer-address",
		Proxy:                      &interactors.SuiProxyStub{},
		Log:                        logger.GetOrCreate("test"),
	}
}

func createMockProxy(message json.RawMessage) *interactors.SuiProxyStub {
	return &interactors.SuiProxyStub{
		SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
			return models.SuiTransactionBlockResponse{
				Effects: models.SuiEffects{
					Status: models.ExecutionStatus{
						Status: "success",
					},
				},
				Results: message,
			}, nil
		},
	}
}

func createFailMockProxy(err error) *interactors.SuiProxyStub {
	return &interactors.SuiProxyStub{
		SuiDevInspectTransactionBlockCalled: func(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
			return models.SuiTransactionBlockResponse{
				Effects: models.SuiEffects{
					Status: models.ExecutionStatus{
						Status: "failed",
					},
				},
			}, err
		},
	}
}

func createResultsRawFromValues(values ...interface{}) ([]byte, error) {
	var returnValues [][]interface{}

	for _, value := range values {
		marshaled, err := mystenbcs.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value: %w", err)
		}
		returnValues = append(returnValues, []interface{}{marshaled, "test_type"})
	}

	results := []map[string]interface{}{
		{
			"returnValues": returnValues,
		},
	}
	return json.Marshal(results)
}

func TestNewSuiClientDataGetter(t *testing.T) {
	t.Parallel()

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.Log = nil

		dg, err := NewSuiClientDataGetter(args)
		assert.Equal(t, clients.ErrNilLogger, err)
		assert.Nil(t, dg)
	})
	t.Run("nil proxy", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.Proxy = nil

		dg, err := NewSuiClientDataGetter(args)
		assert.Equal(t, errNilProxy, err)
		assert.Nil(t, dg)
	})
	t.Run("nil bridge package id", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.BridgePackageId = ""

		dg, err := NewSuiClientDataGetter(args)
		assert.True(t, errors.Is(err, errNilPackageId))
		assert.True(t, strings.Contains(err.Error(), "BridgePackageId"))
		assert.Nil(t, dg)
	})
	t.Run("nil bridge object id", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.BridgeObjectId = ""

		dg, err := NewSuiClientDataGetter(args)
		assert.True(t, errors.Is(err, errNilObjectId))
		assert.True(t, strings.Contains(err.Error(), "BridgeObjectId"))
		assert.Nil(t, dg)
	})
	t.Run("invalid bridge initial shared version", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.BridgeInitialSharedVersion = 0

		dg, err := NewSuiClientDataGetter(args)
		assert.Equal(t, errInvalidInitialSharedVersion, err)
		assert.Nil(t, dg)
	})
	t.Run("nil safe package id", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.SafePackageId = ""

		dg, err := NewSuiClientDataGetter(args)
		assert.True(t, errors.Is(err, errNilPackageId))
		assert.True(t, strings.Contains(err.Error(), "SafePackageId"))
		assert.Nil(t, dg)
	})
	t.Run("nil safe object id", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.SafeObjectId = ""

		dg, err := NewSuiClientDataGetter(args)
		assert.True(t, errors.Is(err, errNilObjectId))
		assert.True(t, strings.Contains(err.Error(), "SafeObjectId"))
		assert.Nil(t, dg)
	})
	t.Run("invalid safe initial shared version", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.SafeInitialSharedVersion = 0

		dg, err := NewSuiClientDataGetter(args)
		assert.Equal(t, errInvalidInitialSharedVersion, err)
		assert.Nil(t, dg)
	})
	t.Run("nil relayer address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()
		args.RelayerAddress = ""

		dg, err := NewSuiClientDataGetter(args)
		assert.True(t, errors.Is(err, errNilAddress))
		assert.True(t, strings.Contains(err.Error(), "RelayerAddress"))
		assert.Nil(t, dg)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsSuiClientDataGetter()

		dg, err := NewSuiClientDataGetter(args)
		assert.Nil(t, err)
		assert.NotNil(t, dg)
	})
}

func TestSuiClientDataGetter_GetBatchByNonce(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		batch, isFinal, err := dg.GetBatchByNonce(context.Background(), batchNonce)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Equal(t, dtos.Batch{}, batch)
		assert.False(t, isFinal)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)

		batch, isFinal, err := dg.GetBatchByNonce(context.Background(), batchNonce)
		assert.NotNil(t, err)
		assert.Equal(t, dtos.Batch{}, batch)
		assert.False(t, isFinal)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedBatch := dtos.Batch{
			Nonce: 42,
		}
		expectedIsFinal := true

		resultsRaw, err := createResultsRawFromValues(expectedBatch, expectedIsFinal)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		batch, isFinal, err := dg.GetBatchByNonce(context.Background(), batchNonce)
		assert.Nil(t, err)
		assert.Equal(t, expectedBatch, batch)
		assert.Equal(t, expectedIsFinal, isFinal)
	})
}

func TestSuiClientDataGetter_GetBatchDeposits(t *testing.T) {
	t.Parallel()
	args := createMockArgsSuiClientDataGetter()

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		deposits, areFinal, err := dg.GetBatchDeposits(context.Background(), batchNonce)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Nil(t, deposits)
		assert.False(t, areFinal)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)

		deposits, areFinal, err := dg.GetBatchDeposits(context.Background(), batchNonce)
		assert.NotNil(t, err)
		assert.Nil(t, deposits)
		assert.False(t, areFinal)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedDeposits := []dtos.Deposit{
			{
				Nonce:        42,
				TokenAddress: "coin::Coin::0x1",
				Amount:       1000,
				Depositor:    bytes.Repeat([]byte{0x1}, 32),
				Recipient:    bytes.Repeat([]byte{0x2}, 32),
				Status:       uint8(1),
			},
			{
				Nonce:        43,
				TokenAddress: "coin::Coin::0x2",
				Amount:       2000,
				Depositor:    bytes.Repeat([]byte{0x3}, 32),
				Recipient:    bytes.Repeat([]byte{0x4}, 32),
				Status:       uint8(2),
			},
		}
		expectedAreFinal := true

		resultsRaw, err := createResultsRawFromValues(expectedDeposits, expectedAreFinal)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		deposits, areFinal, err := dg.GetBatchDeposits(context.Background(), batchNonce)
		assert.Nil(t, err)
		assert.Equal(t, expectedDeposits, deposits)
		assert.Equal(t, expectedAreFinal, areFinal)
	})
}

func TestSuiClientDataGetter_GetRelayers(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		addresses, err := dg.GetRelayers(context.Background())
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Nil(t, addresses)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)

		addresses, err := dg.GetRelayers(context.Background())
		assert.NotNil(t, err)
		assert.Nil(t, addresses)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedAddresses := []models.SuiAddress{
			"0xc75ff8f7215e3dc9671ae2ef85c519414b886bff4e0f6edb5ff7b8d3d9e648fa",
			"0xfe6d2075696ecea4614d9dd3ac532c84af2ac1962cee43ad99a6d84f3f7e30e0",
		}

		resultsRaw, err := createResultsRawFromValues(expectedAddresses)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		addresses, err := dg.GetRelayers(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedAddresses, addresses)
	})
}

func TestSuiClientDataGetter_WasBatchExecuted(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		wasBatchExecuted, err := dg.WasBatchExecuted(context.Background(), batchNonce)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.False(t, wasBatchExecuted)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)

		wasBatchExecuted, err := dg.WasBatchExecuted(context.Background(), batchNonce)
		assert.NotNil(t, err)
		assert.False(t, wasBatchExecuted)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedBool := true

		resultsRaw, err := createResultsRawFromValues(expectedBool)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		wasBatchExecuted, err := dg.WasBatchExecuted(context.Background(), batchNonce)
		assert.Nil(t, err)
		assert.Equal(t, expectedBool, wasBatchExecuted)
	})
}

func TestSuiClientDataGetter_IsPaused(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		isPaused, err := dg.IsPaused(context.Background())
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.False(t, isPaused)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)

		isPaused, err := dg.IsPaused(context.Background())
		assert.NotNil(t, err)
		assert.False(t, isPaused)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedBool := true
		resultsRaw, err := createResultsRawFromValues(expectedBool)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		isPaused, err := dg.IsPaused(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedBool, isPaused)
	})
}

func TestSuiClientDataGetter_Quorum(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		quorum, err := dg.Quorum(context.Background())
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Equal(t, uint64(0), quorum)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)
		quorum, err := dg.Quorum(context.Background())
		assert.NotNil(t, err)
		assert.Equal(t, uint64(0), quorum)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedQuorum := uint64(5)

		resultsRaw, err := createResultsRawFromValues(expectedQuorum)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		quorum, err := dg.Quorum(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedQuorum, quorum)
	})
}

func TestSuiClientDataGetter_GetStatusesAfterExecution(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		statuses, isFinal, err := dg.GetStatusesAfterExecution(context.Background(), batchNonce)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Nil(t, statuses)
		assert.False(t, isFinal)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)

		statuses, isFinal, err := dg.GetStatusesAfterExecution(context.Background(), batchNonce)
		assert.NotNil(t, err)
		assert.Nil(t, statuses)
		assert.False(t, isFinal)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedStatuses := []uint8{1, 2, 3}
		expectedIsFinal := true

		resultsRaw, err := createResultsRawFromValues(expectedStatuses, expectedIsFinal)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		statuses, isFinal, err := dg.GetStatusesAfterExecution(context.Background(), batchNonce)
		assert.Nil(t, err)
		assert.Equal(t, expectedStatuses, statuses)
		assert.Equal(t, expectedIsFinal, isFinal)
	})
}

func TestSuiClientDataGetter_GetTotalBalanceFromSafe(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()
	coinType := "0x2::sui::SUI"

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		balance, err := dg.GetTotalBalanceFromSafe(context.Background(), coinType)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Equal(t, uint64(0), balance)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)

		balance, err := dg.GetTotalBalanceFromSafe(context.Background(), coinType)
		assert.NotNil(t, err)
		assert.Equal(t, uint64(0), balance)
	})
	t.Run("invalid coin type", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		invalidCoinType := "invalid::cointype"

		balance, err := dg.GetTotalBalanceFromSafe(context.Background(), invalidCoinType)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to parse coin type"))
		assert.Equal(t, uint64(0), balance)
	})
	t.Run("failed to convert coin type", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		invalidAddressCoinType := "invalid-address::sui::SUI"

		balance, err := dg.GetTotalBalanceFromSafe(context.Background(), invalidAddressCoinType)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to convert coin type"))
		assert.Equal(t, uint64(0), balance)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedBalance := uint64(1000000)

		resultsRaw, err := createResultsRawFromValues(expectedBalance)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		balance, err := dg.GetTotalBalanceFromSafe(context.Background(), coinType)
		assert.Nil(t, err)
		assert.Equal(t, expectedBalance, balance)
	})
}

func TestSuiClientDataGetter_IsTokenWhitelisted(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()
	coinType := "0x2::sui::SUI"

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = createFailMockProxy(expectedErr)

		isWhitelisted, err := dg.IsTokenWhitelisted(context.Background(), coinType)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.False(t, isWhitelisted)
	})
	t.Run("failed tx status", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		dg.proxy = createFailMockProxy(nil)

		isWhitelisted, err := dg.IsTokenWhitelisted(context.Background(), coinType)
		assert.NotNil(t, err)
		assert.False(t, isWhitelisted)
	})
	t.Run("invalid coin type", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		invalidCoinType := "invalid::cointype"

		isWhitelisted, err := dg.IsTokenWhitelisted(context.Background(), invalidCoinType)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to parse coin type"))
		assert.False(t, isWhitelisted)
	})
	t.Run("failed to convert coin type", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		invalidAddressCoinType := "invalid-address::sui::SUI"

		isWhitelisted, err := dg.IsTokenWhitelisted(context.Background(), invalidAddressCoinType)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to convert coin type"))
		assert.False(t, isWhitelisted)
	})
	t.Run("should work - token whitelisted", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedIsWhitelisted := true

		resultsRaw, err := createResultsRawFromValues(expectedIsWhitelisted)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		isWhitelisted, err := dg.IsTokenWhitelisted(context.Background(), coinType)
		assert.Nil(t, err)
		assert.Equal(t, expectedIsWhitelisted, isWhitelisted)
	})
	t.Run("should work - token not whitelisted", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)

		expectedIsWhitelisted := false

		resultsRaw, err := createResultsRawFromValues(expectedIsWhitelisted)
		assert.NoError(t, err)

		dg.proxy = createMockProxy(resultsRaw)

		isWhitelisted, err := dg.IsTokenWhitelisted(context.Background(), coinType)
		assert.Nil(t, err)
		assert.Equal(t, expectedIsWhitelisted, isWhitelisted)
	})
}

func TestSuiClientDataGetter_GetLatestCheckpoint(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = &interactors.SuiProxyStub{
			SuiGetLatestCheckpointSequenceNumberCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
		}

		checkpoint, err := dg.GetLatestCheckpoint(context.Background())
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Equal(t, uint64(0), checkpoint)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedCheckpoint := uint64(12345)
		dg.proxy = &interactors.SuiProxyStub{
			SuiGetLatestCheckpointSequenceNumberCalled: func(ctx context.Context) (uint64, error) {
				return expectedCheckpoint, nil
			},
		}

		checkpoint, err := dg.GetLatestCheckpoint(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedCheckpoint, checkpoint)
	})
}

func TestSuiClientDataGetter_GetBalance(t *testing.T) {
	t.Parallel()

	args := createMockArgsSuiClientDataGetter()
	account := "0xc75ff8f7215e3dc9671ae2ef85c519414b886bff4e0f6edb5ff7b8d3d9e648fa"
	coinType := "0x2::sui::SUI"

	t.Run("proxy errors", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = &interactors.SuiProxyStub{
			SuiXGetBalanceCalled: func(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
				return models.CoinBalanceResponse{}, expectedErr
			},
		}

		balance, err := dg.GetBalance(context.Background(), account, coinType)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Equal(t, models.CoinBalanceResponse{}, balance)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dg, _ := NewSuiClientDataGetter(args)
		expectedBalance := models.CoinBalanceResponse{
			CoinType:        coinType,
			CoinObjectCount: 5,
			TotalBalance:    "1000000000",
		}
		dg.proxy = &interactors.SuiProxyStub{
			SuiXGetBalanceCalled: func(ctx context.Context, req models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
				assert.Equal(t, account, req.Owner)
				assert.Equal(t, coinType, req.CoinType)
				return expectedBalance, nil
			},
		}

		balance, err := dg.GetBalance(context.Background(), account, coinType)
		assert.Nil(t, err)
		assert.Equal(t, expectedBalance, balance)
	})
}
