package sui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/stretchr/testify/assert"
)

var (
	seedBytes = []byte{
		49, 48, 50, 97, 49, 113, 122, 122, 121, 57, 120, 56, 54, 121, 101, 55,
		107, 121, 112, 116, 103, 57, 117, 51, 115, 97, 113, 117, 103, 48, 104, 100,
	}
	testBridgePackageId = "0x19ecc8df0b93999f87b5e2e531bb80a49d1ab2f4644032cd31f2c53ed624c94a"
)

func createTransactionHandlerWithMockComponents() *transactionHandler {
	return &transactionHandler{
		proxy:         &interactors.SuiProxyStub{},
		relayerSigner: signer.NewSigner(seedBytes),
	}
}

func TestTransactionHandler_SendTransactionReturnHash(t *testing.T) {
	t.Parallel()

	moveCallRequest := models.MoveCallRequest{
		PackageObjectId: testBridgePackageId,
		Module:          "test_module",
		Function:        "test_function",
		TypeArguments:   []interface{}{},
		Arguments:       []interface{}{},
		GasBudget:       "1000000",
	}

	t.Run("MoveCall proxy error", func(t *testing.T) {
		t.Parallel()

		txHandlerInstance := createTransactionHandlerWithMockComponents()
		expectedErr := errors.New("move call error")
		txHandlerInstance.proxy = &interactors.SuiProxyStub{
			MoveCallCalled: func(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error) {
				assert.Equal(t, txHandlerInstance.relayerSigner.Address, req.Signer)
				return models.TxnMetaData{}, expectedErr
			},
		}

		moveCallRequest.Signer = txHandlerInstance.relayerSigner.Address
		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), moveCallRequest)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Equal(t, "", hash)
	})

	t.Run("SignAndExecuteTransactionBlock proxy error", func(t *testing.T) {
		t.Parallel()

		txHandlerInstance := createTransactionHandlerWithMockComponents()
		expectedErr := errors.New("sign and execute error")
		mockTxnMetaData := models.TxnMetaData{
			TxBytes: "mock-tx-bytes",
		}

		txHandlerInstance.proxy = &interactors.SuiProxyStub{
			MoveCallCalled: func(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error) {
				return mockTxnMetaData, nil
			},
			SignAndExecuteTransactionBlockCalled: func(ctx context.Context, req models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				assert.Equal(t, mockTxnMetaData, req.TxnMetaData)
				assert.Equal(t, txHandlerInstance.relayerSigner.PriKey, req.PriKey)
				assert.Equal(t, "WaitForLocalExecution", req.RequestType)
				assert.True(t, req.Options.ShowInput)
				assert.True(t, req.Options.ShowEffects)
				assert.True(t, req.Options.ShowObjectChanges)
				return models.SuiTransactionBlockResponse{}, expectedErr
			},
		}

		moveCallRequest.Signer = txHandlerInstance.relayerSigner.Address
		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), moveCallRequest)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
		assert.Equal(t, "", hash)
	})

	t.Run("transaction failed status", func(t *testing.T) {
		t.Parallel()

		txHandlerInstance := createTransactionHandlerWithMockComponents()
		mockTxnMetaData := models.TxnMetaData{
			TxBytes: "mock-tx-bytes",
		}

		txHandlerInstance.proxy = &interactors.SuiProxyStub{
			MoveCallCalled: func(ctx context.Context, req models.MoveCallRequest) (models.TxnMetaData, error) {
				return mockTxnMetaData, nil
			},
			SignAndExecuteTransactionBlockCalled: func(ctx context.Context, req models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				return models.SuiTransactionBlockResponse{
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "failure",
							Error:  "transaction execution failed",
						},
					},
				}, nil
			},
		}

		moveCallRequest.Signer = txHandlerInstance.relayerSigner.Address
		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), moveCallRequest)
		assert.NotNil(t, err)
		assert.Equal(t, "transaction execution failed", err.Error())
		assert.Equal(t, "", hash)
	})

	t.Run("send transaction successfully", func(t *testing.T) {
		t.Parallel()

		expectedHash := "expected hash"
		mockTxnMetaData := models.TxnMetaData{
			TxBytes: "mock-tx-bytes",
		}

		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHandlerInstance.proxy = &interactors.SuiProxyStub{
			MoveCallCalled: func(ctx context.Context, request models.MoveCallRequest) (models.TxnMetaData, error) {
				assert.Equal(t, txHandlerInstance.relayerSigner.Address, request.Signer)
				return mockTxnMetaData, nil
			},

			SignAndExecuteTransactionBlockCalled: func(ctx context.Context, request models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				assert.Equal(t, mockTxnMetaData, request.TxnMetaData)
				assert.Equal(t, txHandlerInstance.relayerSigner.PriKey, request.PriKey)

				return models.SuiTransactionBlockResponse{
					Digest: expectedHash,
					Effects: models.SuiEffects{
						Status: models.ExecutionStatus{
							Status: "success",
						},
					},
				}, nil
			},
		}

		moveCallRequest.Signer = txHandlerInstance.relayerSigner.Address
		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), moveCallRequest)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
	})
}
