package batchValidatorManagement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/chain"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsBatchValidator() ArgsBatchValidator {
	return ArgsBatchValidator{
		SourceChain:      chain.Ethereum,
		DestinationChain: chain.MultiversX,
		RequestURL:       "",
		RequestTime:      time.Second,
	}
}

func TestNewBatchValidator(t *testing.T) {
	t.Parallel()

	t.Run("invalid SourceChain", func(t *testing.T) {
		args := createMockArgsBatchValidator()
		args.SourceChain = ""

		bv, err := NewBatchValidator(args)
		assert.True(t, check.IfNil(bv))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
	})
	t.Run("invalid DestinationChain", func(t *testing.T) {
		args := createMockArgsBatchValidator()
		args.DestinationChain = ""

		bv, err := NewBatchValidator(args)
		assert.True(t, check.IfNil(bv))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
	})
	t.Run("invalid request time", func(t *testing.T) {
		args := createMockArgsBatchValidator()
		args.RequestTime = time.Duration(minRequestTime.Nanoseconds() - 1)

		bv, err := NewBatchValidator(args)
		assert.True(t, check.IfNil(bv))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value RequestTime"))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsBatchValidator()

		bv, err := NewBatchValidator(args)
		assert.False(t, check.IfNil(bv))
		assert.Nil(t, err)
	})
}

func TestBatchValidator_ValidateBatch(t *testing.T) {
	t.Parallel()

	largeValue, ok := big.NewInt(0).SetString("1000000000000000000000", 10) // 1000 units (10^18 denominated)
	require.True(t, ok)
	batch := &clients.TransferBatch{
		ID: 1,
		Deposits: []*clients.DepositTransfer{
			{
				Nonce:            1,
				DisplayableTo:    "to1",
				DisplayableFrom:  "from1",
				DisplayableToken: "token1",
				Amount:           big.NewInt(0).Add(largeValue, big.NewInt(1)),
			},
			{
				Nonce:            2,
				DisplayableTo:    "to2",
				DisplayableFrom:  "from2",
				DisplayableToken: "token2",
				Amount:           big.NewInt(0).Add(largeValue, big.NewInt(2)),
			},
		},
		Statuses: []byte{0x3, 0x4},
	}
	expectedJsonString := `{"batchId":1,"deposits":[{"nonce":1,"to":"to1","from":"from1","token":"token1","amount":1000000000000000000001,"data":""},{"nonce":2,"to":"to2","from":"from2","token":"token2","amount":1000000000000000000002,"data":""}],"statuses":"AwQ="}`

	t.Run("server errors with Bad Request, but no reason", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBatchValidator()
		responseHandler := &testsCommon.HTTPHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				expectedURL := fmt.Sprintf("/%s/%s", args.SourceChain.ToLower(), args.DestinationChain.ToLower())
				require.Equal(t, expectedURL, request.URL.String())

				writer.WriteHeader(http.StatusBadRequest)
			},
		}

		server := httptest.NewServer(responseHandler)
		defer server.Close()

		args.RequestURL = server.URL
		bv, _ := NewBatchValidator(args)

		isValid, err := bv.ValidateBatch(context.Background(), batch)
		assert.False(t, isValid)
		assert.NotNil(t, err)
		assert.Equal(t, "got status 400 Bad Request while executing request", err.Error())
	})
	t.Run("server errors with Bad Request and reason", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBatchValidator()
		bodyJson := microserviceBadRequestBody{
			StatusCode: 400,
			Message:    "different number of swaps. given: 2, stored: 3",
			Error:      "Bad Request",
		}
		responseHandler := &testsCommon.HTTPHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				expectedURL := fmt.Sprintf("/%s/%s", args.SourceChain.ToLower(), args.DestinationChain.ToLower())
				require.Equal(t, expectedURL, request.URL.String())

				writer.WriteHeader(http.StatusBadRequest)
				body, _ := json.Marshal(bodyJson)
				_, _ = writer.Write(body)
			},
		}

		server := httptest.NewServer(responseHandler)
		defer server.Close()

		args.RequestURL = server.URL
		bv, _ := NewBatchValidator(args)

		isValid, err := bv.ValidateBatch(context.Background(), batch)
		assert.False(t, isValid)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("got status 400 Bad Request: %s while executing request", bodyJson.Message), err.Error())
	})
	t.Run("server errors with other than Bad Request", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBatchValidator()
		responseHandler := &testsCommon.HTTPHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				expectedURL := fmt.Sprintf("/%s/%s", args.SourceChain.ToLower(), args.DestinationChain.ToLower())
				require.Equal(t, expectedURL, request.URL.String())

				writer.WriteHeader(http.StatusForbidden)
			},
		}

		server := httptest.NewServer(responseHandler)
		defer server.Close()

		args.RequestURL = server.URL
		bv, _ := NewBatchValidator(args)

		isValid, err := bv.ValidateBatch(context.Background(), batch)
		assert.False(t, isValid)
		assert.NotNil(t, err)
		assert.Equal(t, "got status 403 Forbidden while executing request", err.Error())
	})
	t.Run("empty response", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBatchValidator()
		responseHandler := &testsCommon.HTTPHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				expectedURL := fmt.Sprintf("/%s/%s", args.SourceChain.ToLower(), args.DestinationChain.ToLower())
				require.Equal(t, expectedURL, request.URL.String())

				writer.WriteHeader(http.StatusOK)
				_, _ = writer.Write(nil)
			},
		}

		server := httptest.NewServer(responseHandler)
		defer server.Close()

		args.RequestURL = server.URL
		bv, _ := NewBatchValidator(args)

		isValid, err := bv.ValidateBatch(context.Background(), batch)
		assert.False(t, isValid)
		assert.NotNil(t, err)
		assert.Equal(t, "empty response", err.Error())
	})
	t.Run("improper response", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBatchValidator()
		responseHandler := &testsCommon.HTTPHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				expectedURL := fmt.Sprintf("/%s/%s", args.SourceChain.ToLower(), args.DestinationChain.ToLower())
				require.Equal(t, expectedURL, request.URL.String())

				writer.WriteHeader(http.StatusOK)
				_, _ = writer.Write([]byte("garbage"))
			},
		}

		server := httptest.NewServer(responseHandler)
		defer server.Close()

		args.RequestURL = server.URL
		bv, _ := NewBatchValidator(args)

		isValid, err := bv.ValidateBatch(context.Background(), batch)
		assert.False(t, isValid)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "value during response unmarshal"))
	})
	t.Run("context deadline exceeded", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBatchValidator()
		responseHandler := &testsCommon.HTTPHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				expectedURL := fmt.Sprintf("/%s/%s", args.SourceChain.ToLower(), args.DestinationChain.ToLower())
				require.Equal(t, expectedURL, request.URL.String())
			},
		}

		server := httptest.NewServer(responseHandler)
		defer server.Close()

		args.RequestURL = server.URL
		bv, _ := NewBatchValidator(args)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		isValid, err := bv.ValidateBatch(ctx, batch)
		assert.False(t, isValid)
		assert.NotNil(t, err)
		assert.True(t, strings.Contains(err.Error(), "context canceled while executing request"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsBatchValidator()
		responseHandler := &testsCommon.HTTPHandlerStub{
			ServeHTTPCalled: func(writer http.ResponseWriter, request *http.Request) {
				expectedURL := fmt.Sprintf("/%s/%s", args.SourceChain.ToLower(), args.DestinationChain.ToLower())
				require.Equal(t, expectedURL, request.URL.String())

				defer func() {
					_ = request.Body.Close()
				}()

				body, err := io.ReadAll(request.Body)
				require.Nil(t, err)
				require.Equal(t, expectedJsonString, string(body))

				resp := &microserviceResponse{
					Valid: true,
				}
				writer.WriteHeader(http.StatusOK)
				respBytes, _ := json.Marshal(resp)
				_, _ = writer.Write(respBytes)
			},
		}

		server := httptest.NewServer(responseHandler)
		defer server.Close()

		args.RequestURL = server.URL
		bv, _ := NewBatchValidator(args)

		isValid, err := bv.ValidateBatch(context.Background(), batch)
		assert.True(t, isValid)
		assert.Nil(t, err)
	})
}
