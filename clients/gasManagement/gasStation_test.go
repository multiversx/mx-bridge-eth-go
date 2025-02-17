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

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
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
	// synchronize the process loop & the testing go routine with an unbuffered channel
	chanOk := make(chan struct{})
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-chanOk
		// simulating that the operation takes a lot of time

		time.Sleep(time.Second * 3)

		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(nil)
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	chanOk <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())
	_ = gs.Close()

	time.Sleep(time.Millisecond * 500)

	assert.False(t, gs.loopStatus.IsSet())
}

func TestGasStation_InvalidJsonResponse(t *testing.T) {
	t.Parallel()

	args := createMockArgsGasStation()
	// synchronize the process loop & the testing go routine with an unbuffered channel
	chanNok := make(chan struct{})
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-chanNok
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("invalid json response"))
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	chanNok <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())
	_ = gs.Close()

	time.Sleep(time.Millisecond * 500)
	assert.False(t, gs.loopStatus.IsSet())
	assert.Equal(t, float64(-1), gs.GetLatestGasPrice())
	gasPrice, err := gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)
}

func TestGasStation_GoodResponseShouldSave(t *testing.T) {
	t.Parallel()

	gsResponse := createMockGasStationResponse()
	args := createMockArgsGasStation()
	// synchronize the process loop & the testing go routine with an unbuffered channel
	chanOk := make(chan struct{})
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-chanOk
		rw.WriteHeader(http.StatusOK)

		resp, _ := json.Marshal(&gsResponse)
		_, _ = rw.Write(resp)
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	chanOk <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())
	_ = gs.Close()

	time.Sleep(time.Millisecond * 500)
	assert.False(t, gs.loopStatus.IsSet())
	var expectedPrice = float64(-1)
	_, err = fmt.Sscanf(gsResponse.Result.SafeGasPrice, "%f", &expectedPrice)
	require.Nil(t, err)
	assert.Equal(t, expectedPrice, gs.GetLatestGasPrice())
}

func TestGasStation_RetryMechanism_FailsFirstRequests(t *testing.T) {
	t.Parallel()

	args := createMockArgsGasStation()
	args.RequestRetryDelay = time.Second
	args.RequestPollingInterval = 2 * time.Second
	args.MaximumFetchRetries = 3

	// synchronize the process loop & the testing go routine with unbuffered channels
	chanOk := make(chan struct{})
	chanNok := make(chan struct{})
	gsResponse := createMockGasStationResponse()
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		select {
		case <-chanOk:
			resp, _ := json.Marshal(&gsResponse)
			_, _ = rw.Write(resp)
		case <-chanNok:
			_, _ = rw.Write([]byte("invalid json response"))
		}
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	chanNok <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, float64(-1), gs.GetLatestGasPrice())
	gasPrice, err := gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)

	chanNok <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, float64(-1), gs.GetLatestGasPrice())
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)

	chanNok <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, float64(-1), gs.GetLatestGasPrice())
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)

	chanOk <- struct{}{} // response is now ok
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())
	assert.Equal(t, float64(81), gs.GetLatestGasPrice())
	gasPrice, err = gs.GetCurrentGasPrice()
	assert.Equal(t, big.NewInt(int64(gs.GetLatestGasPrice()*float64(args.GasPriceMultiplier))), gasPrice)
	assert.Nil(t, err)
	_ = gs.Close()

	time.Sleep(args.RequestPollingInterval + 1)
	assert.False(t, gs.loopStatus.IsSet())
}

func TestGasStation_RetryMechanism_IntermittentFails(t *testing.T) {
	t.Parallel()

	args := createMockArgsGasStation()
	args.RequestRetryDelay = time.Second
	args.RequestPollingInterval = 2 * time.Second

	// synchronize the process loop & the testing go routine with unbuffered channels
	chanOk := make(chan struct{})
	chanNok := make(chan struct{})
	gsResponse := createMockGasStationResponse()
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Println("http server go routine")
		select {
		case <-chanOk:
			resp, _ := json.Marshal(&gsResponse)
			_, _ = rw.Write(resp)
		case <-chanNok:
			_, _ = rw.Write([]byte("invalid json response"))
		}
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	for i := 0; i < 6; i++ {
		shouldFail := i > 0 && i%3 == 0
		if shouldFail {
			chanNok <- struct{}{}
		} else {
			chanOk <- struct{}{}
		}
		time.Sleep(time.Millisecond * 100)

		assert.True(t, gs.loopStatus.IsSet())
		assert.Equal(t, float64(81), gs.GetLatestGasPrice())
		gasPrice, errGet := gs.GetCurrentGasPrice()
		assert.Equal(t, big.NewInt(int64(gs.GetLatestGasPrice()*float64(args.GasPriceMultiplier))), gasPrice)
		assert.Nil(t, errGet)
	}

	_ = gs.Close()

	time.Sleep(args.RequestPollingInterval + 1)
	assert.False(t, gs.loopStatus.IsSet())
}

func TestGasStation_GetCurrentGasPriceShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should work with int values", func(t *testing.T) {
		gsResponse := createMockGasStationResponse()
		args := createMockArgsGasStation()

		expectedFast := big.NewInt(83000000000)
		expectedPropose := big.NewInt(82000000000)
		expectedSafe := big.NewInt(81000000000)

		testGetCurrentGasPrice(t, gsResponse, args, expectedFast, expectedPropose, expectedSafe)
	})
	t.Run("should work with float64 values and no trim", func(t *testing.T) {
		gsResponse := createMockGasStationResponseWithFloatValues()
		args := createMockArgsGasStation()

		expectedFast := big.NewInt(1460784306)
		expectedPropose := big.NewInt(1327985733)
		expectedSafe := big.NewInt(1289774824)

		testGetCurrentGasPrice(t, gsResponse, args, expectedFast, expectedPropose, expectedSafe)
	})
	t.Run("should work with float64 values with trim", func(t *testing.T) {
		gsResponse := createMockGasStationResponseWithFloatValues()
		args := createMockArgsGasStation()
		args.GasPriceMultiplier = 10000

		expectedFast := big.NewInt(14607)
		expectedPropose := big.NewInt(13279)
		expectedSafe := big.NewInt(12897)

		testGetCurrentGasPrice(t, gsResponse, args, expectedFast, expectedPropose, expectedSafe)
	})
}

func testGetCurrentGasPrice(
	t *testing.T,
	gsResponse gasStationResponse,
	args ArgsGasStation,
	expectedFast *big.Int,
	expectedPropose *big.Int,
	expectedSafe *big.Int,
) {
	// synchronize the process loop & the testing go routine with an unbuffered channel
	chanOk := make(chan struct{})
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-chanOk
		rw.WriteHeader(http.StatusOK)

		resp, _ := json.Marshal(&gsResponse)
		_, _ = rw.Write(resp)
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	chanOk <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())

	gs.SetSelector(core.EthFastGasPrice)
	chanOk <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	price, err := gs.GetCurrentGasPrice()
	require.Nil(t, err)
	assert.Equal(t, expectedFast, price)

	gs.SetSelector(core.EthProposeGasPrice)
	chanOk <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	price, err = gs.GetCurrentGasPrice()
	require.Nil(t, err)
	assert.Equal(t, expectedPropose, price)

	gs.SetSelector(core.EthSafeGasPrice)
	chanOk <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	price, err = gs.GetCurrentGasPrice()
	require.Nil(t, err)
	assert.Equal(t, expectedSafe, price)

	gs.SetSelector("invalid")
	chanOk <- struct{}{}
	time.Sleep(time.Millisecond * 100)
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
	// synchronize the process loop & the testing go routine with an unbuffered channel
	chanOk := make(chan struct{})
	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-chanOk
		rw.WriteHeader(http.StatusOK)

		resp, _ := json.Marshal(&gsResponse)
		_, _ = rw.Write(resp)
	}))
	defer httpServer.Close()

	args.RequestURL = httpServer.URL

	gs, err := NewGasStation(args)
	require.Nil(t, err)

	chanOk <- struct{}{}
	time.Sleep(time.Millisecond * 100)
	assert.True(t, gs.loopStatus.IsSet())

	price, err := gs.GetCurrentGasPrice()
	require.ErrorIs(t, err, ErrGasPriceIsHigherThanTheMaximumSet)
	require.Contains(t, err.Error(), "set maximum value: 100.000")
	require.Contains(t, err.Error(), "fetched value: 101.000")
	require.Contains(t, err.Error(), "gas price selector: SafeGasPrice")
	assert.Equal(t, big.NewInt(0), price)
	_ = gs.Close()
}

func TestGasStation_GetCurrentGasPriceBelowMin(t *testing.T) {
	t.Parallel()

	gsResponse := createMockGasStationResponse()
	gsResponse.Result.SafeGasPrice = "0.944851822"
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
	expectedPrice := gs.minGasPriceValue * gs.gasPriceMultiplier

	time.Sleep(time.Second * 2)
	assert.True(t, gs.loopStatus.IsSet())

	price, err := gs.GetCurrentGasPrice()
	require.Nil(t, err)
	assert.Equal(t, big.NewInt(0).SetInt64(int64(expectedPrice)), price)
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

func createMockGasStationResponseWithFloatValues() gasStationResponse {
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
			LastBlock:       "21779373",
			SafeGasPrice:    "1.289774824",
			ProposeGasPrice: "1.327985733",
			FastGasPrice:    "1.460784306",
			SuggestBaseFee:  "1.289774824",
			GasUsedRatio:    "0.339906295527105,0.348258876646649,0.41887262757803,0.316123823179043,0.659216773979187",
		},
	}
}
