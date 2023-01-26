package gasManagement

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsGasStation() ArgsGasStation {
	return ArgsGasStation{
		RequestURL:             "",
		RequestPollingInterval: time.Second,
		RequestRetryDelay:      time.Second,
		MaximumFetchRetries:    3,
		RequestTime:            time.Second,
		MaximumGasPrice:        100,
		GasPriceSelector:       "SafeGasPrice",
		GasPriceMultiplier:     1000000000,
	}
}

func TestNewGasStation(t *testing.T) {
	t.Parallel()

	t.Run("invalid polling time", func(t *testing.T) {
		args := createMockArgsGasStation()
		args.RequestPollingInterval = time.Duration(minPollingInterval.Nanoseconds() - 1)

		gs, err := NewGasStation(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value RequestPollingInterval"))
	})
	t.Run("invalid polling time for retry delay", func(t *testing.T) {
		args := createMockArgsGasStation()
		args.RequestRetryDelay = time.Duration(minPollingInterval.Nanoseconds() - 1)

		gs, err := NewGasStation(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value RequestRetryDelay"))
	})
	t.Run("invalid maximum fetch retries", func(t *testing.T) {
		args := createMockArgsGasStation()
		args.MaximumFetchRetries = 0

		gs, err := NewGasStation(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value MaximumFetchRetries"))
	})
	t.Run("invalid request time", func(t *testing.T) {
		args := createMockArgsGasStation()
		args.RequestTime = time.Duration(minRequestTime.Nanoseconds() - 1)

		gs, err := NewGasStation(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value RequestTime"))
	})
	t.Run("invalid gas price selector", func(t *testing.T) {
		args := createMockArgsGasStation()
		args.GasPriceSelector = "invalid"

		gs, err := NewGasStation(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, ErrInvalidGasPriceSelector))
	})
	t.Run("invalid gas price multiplier", func(t *testing.T) {
		args := createMockArgsGasStation()
		args.GasPriceMultiplier = 0

		gs, err := NewGasStation(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, clients.ErrInvalidValue))
		assert.True(t, strings.Contains(err.Error(), "checkArgs for value GasPriceMultiplier"))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsGasStation()

		gs, err := NewGasStation(args)
		assert.False(t, check.IfNil(gs))
		assert.Nil(t, err)

		_ = gs.Close()
	})
}

func TestGasStation_CloseWhileDoingRequest(t *testing.T) {
	t.Parallel()

	args := createMockArgsGasStation()
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// simulating that the operation takes a lot of time

		time.Sleep(time.Second * 3)

		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(nil)
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	time.Sleep(time.Second)
	assert.True(t, gs.loopStatus.IsSet())
	_ = gs.Close()

	time.Sleep(time.Millisecond * 500)

	assert.False(t, gs.loopStatus.IsSet())
}

func TestGasStation_InvalidJsonResponse(t *testing.T) {
	t.Parallel()

	args := createMockArgsGasStation()
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("invalid json response"))
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	time.Sleep(time.Second * 2)
	assert.True(t, gs.loopStatus.IsSet())
	_ = gs.Close()

	time.Sleep(time.Millisecond * 500)
	assert.False(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), -1)
	gasPrice, err := gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)
}

func TestGasStation_GoodResponseShouldSave(t *testing.T) {
	t.Parallel()

	gsResponse := createMockGasStationResponse()
	args := createMockArgsGasStation()
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		resp, _ := json.Marshal(&gsResponse)
		_, _ = rw.Write(resp)
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	time.Sleep(time.Second * 2)
	assert.True(t, gs.loopStatus.IsSet())
	_ = gs.Close()

	time.Sleep(time.Millisecond * 500)
	assert.False(t, gs.loopStatus.IsSet())
	var expectedPrice = -1
	_, err = fmt.Sscanf(gsResponse.Result.SafeGasPrice, "%d", &expectedPrice)
	require.Nil(t, err)
	assert.NotEqual(t, expectedPrice, -1)
	assert.Equal(t, gs.GetLatestGasPrice(), expectedPrice)
}

func TestGasStation_RetryMechanism_FailsFirstRequests(t *testing.T) {
	t.Parallel()

	args := createMockArgsGasStation()
	args.RequestRetryDelay = time.Second
	args.RequestPollingInterval = 2 * time.Second
	numCalled := 0
	gsResponse := createMockGasStationResponse()
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		if numCalled <= args.MaximumFetchRetries {
			_, _ = rw.Write([]byte("invalid json response"))
		} else {
			resp, _ := json.Marshal(&gsResponse)
			_, _ = rw.Write(resp)
		}
		numCalled++
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)
	time.Sleep(args.RequestRetryDelay + 1)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), -1)
	assert.Equal(t, numCalled, 1)
	assert.Equal(t, gs.fetchRetries, 1)
	gasPrice, err := gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)

	time.Sleep(args.RequestRetryDelay + 1)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), -1)
	assert.Equal(t, numCalled, 2)
	assert.Equal(t, gs.fetchRetries, 2)
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)

	time.Sleep(args.RequestRetryDelay + 1)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), -1)
	assert.Equal(t, numCalled, 3)
	assert.Equal(t, gs.fetchRetries, 3)
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)

	time.Sleep(args.RequestRetryDelay + 1)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), -1)
	assert.Equal(t, numCalled, 4)
	assert.Equal(t, gs.fetchRetries, 0)
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)

	time.Sleep(args.RequestRetryDelay + 1)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), -1)
	assert.Equal(t, numCalled, 4)
	assert.Equal(t, gs.fetchRetries, 0)
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)

	time.Sleep(args.RequestRetryDelay + 1)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), 81)
	assert.Equal(t, numCalled, 5)
	assert.Equal(t, gs.fetchRetries, 0)
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(int64(gs.GetLatestGasPrice()*args.GasPriceMultiplier)), gasPrice)
	assert.Nil(t, err)
	_ = gs.Close()

	time.Sleep(args.RequestPollingInterval + 1)
	assert.False(t, gs.loopStatus.IsSet())
}

func TestGasStation_RetryMechanism_IntermitentFails(t *testing.T) {
	t.Parallel()

	args := createMockArgsGasStation()
	args.RequestRetryDelay = time.Second
	args.RequestPollingInterval = 2 * time.Second
	numCalled := 0
	gsResponse := createMockGasStationResponse()
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		if numCalled != 0 && numCalled%3 == 0 {
			_, _ = rw.Write([]byte("invalid json response"))
		} else {
			resp, _ := json.Marshal(&gsResponse)
			_, _ = rw.Write(resp)
		}
		numCalled++
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	time.Sleep(args.RequestPollingInterval*3 + args.RequestRetryDelay + 1)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), 81)
	assert.Equal(t, numCalled, 4)
	assert.Equal(t, gs.fetchRetries, 1)
	gasPrice, err := gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(int64(gs.GetLatestGasPrice()*args.GasPriceMultiplier)), gasPrice)
	assert.Nil(t, err)

	time.Sleep(args.RequestRetryDelay + 1)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, gs.GetLatestGasPrice(), 81)
	assert.Equal(t, numCalled, 5)
	assert.Equal(t, gs.fetchRetries, 0)
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(int64(gs.GetLatestGasPrice()*args.GasPriceMultiplier)), gasPrice)
	assert.Nil(t, err)

	_ = gs.Close()

	time.Sleep(args.RequestPollingInterval + 1)
	assert.False(t, gs.loopStatus.IsSet())
}

func TestGasStation_GetCurrentGasPrice(t *testing.T) {
	t.Parallel()

	gsResponse := createMockGasStationResponse()
	args := createMockArgsGasStation()
	gasPriceMultiplier := big.NewInt(int64(args.GasPriceMultiplier))
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		resp, _ := json.Marshal(&gsResponse)
		_, _ = rw.Write(resp)
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	time.Sleep(time.Millisecond * 1100)
	assert.True(t, gs.loopStatus.IsSet())

	gs.SetSelector(core.EthFastGasPrice)
	time.Sleep(time.Millisecond * 1100)
	price, err := gs.GetCurrentGasPrice()
	require.Nil(t, err)
	expectedPrice := -1
	_, err = fmt.Sscanf(gsResponse.Result.FastGasPrice, "%d", &expectedPrice)
	require.Nil(t, err)
	assert.NotEqual(t, expectedPrice, -1)
	expected := big.NewInt(0).Mul(big.NewInt(int64(expectedPrice)), gasPriceMultiplier)
	assert.Equal(t, expected, price)

	gs.SetSelector(core.EthProposeGasPrice)
	time.Sleep(time.Millisecond * 1100)
	price, err = gs.GetCurrentGasPrice()
	require.Nil(t, err)
	expectedPrice = -1
	_, err = fmt.Sscanf(gsResponse.Result.ProposeGasPrice, "%d", &expectedPrice)
	require.Nil(t, err)
	assert.NotEqual(t, expectedPrice, -1)
	expected = big.NewInt(0).Mul(big.NewInt(int64(expectedPrice)), gasPriceMultiplier)
	assert.Equal(t, expected, price)

	gs.SetSelector(core.EthSafeGasPrice)
	time.Sleep(time.Millisecond * 1100)
	price, err = gs.GetCurrentGasPrice()
	require.Nil(t, err)
	expectedPrice = -1
	_, err = fmt.Sscanf(gsResponse.Result.SafeGasPrice, "%d", &expectedPrice)
	require.Nil(t, err)
	assert.NotEqual(t, expectedPrice, -1)
	expected = big.NewInt(0).Mul(big.NewInt(int64(expectedPrice)), gasPriceMultiplier)
	assert.Equal(t, expected, price)

	gs.SetSelector("invalid")
	time.Sleep(time.Millisecond * 1100)
	price, err = gs.GetCurrentGasPrice()
	require.True(t, errors.Is(err, ErrLatestGasPricesWereNotFetched))
	assert.Equal(t, big.NewInt(0), price)
	_ = gs.Close()
}

func TestGasStation_GetCurrentGasPriceExceededMaximum(t *testing.T) {
	t.Parallel()

	gsResponse := createMockGasStationResponse()
	gsResponse.Result.SafeGasPrice = "101"
	args := createMockArgsGasStation()
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		resp, _ := json.Marshal(&gsResponse)
		_, _ = rw.Write(resp)
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	time.Sleep(time.Second * 2)
	assert.True(t, gs.loopStatus.IsSet())

	price, err := gs.GetCurrentGasPrice()
	require.True(t, errors.Is(err, ErrGasPriceIsHigherThanTheMaximumSet))
	assert.Equal(t, big.NewInt(0), price)
	_ = gs.Close()
}

func createMockGasStationResponse() gasStationResponse {
	return gasStationResponse{
		Status:  "1",
		Message: "OK-Missing/Invalid API Key, rate limit of 1/5sec applied",
		Result: struct {
			LastBlock       string `json:"LastBlock"`
			SafeGasPrice    string `json:"SafeGasPrice"`
			ProposeGasPrice string `json:"ProposeGasPrice"`
			FastGasPrice    string `json:"FastGasPrice"`
			SuggestBaseFee  string `json:"suggestBaseFee"`
			GasUsedRatio    string `json:"gasUsedRatio"`
		}{
			LastBlock:       "14836699",
			SafeGasPrice:    "81",
			ProposeGasPrice: "82",
			FastGasPrice:    "83",
			SuggestBaseFee:  "80.856621497",
			GasUsedRatio:    "0.0422401857919075,0.636178148305543,0.399708304558626,0.212555933333333,0.645151576152554",
		},
	}
}
