package gasManagement

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsGasStation() ArgsGasStation {
	return ArgsGasStation{
		RequestURL:                  "",
		MaximumGasPriceInGWeiTenths: 1000,
		GasPriceSelector:            "fast",
	}
}

func TestNewGasStation(t *testing.T) {
	t.Parallel()

	t.Run("invalid gas price selector", func(t *testing.T) {
		args := createMockArgsGasStation()
		args.GasPriceSelector = "invalid"

		gs, err := NewGasStation(args)
		assert.True(t, check.IfNil(gs))
		assert.True(t, errors.Is(err, ErrInvalidGasPriceSelector))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsGasStation()

		gs, err := NewGasStation(args)
		assert.False(t, check.IfNil(gs))
		assert.Nil(t, err, ErrInvalidGasPriceSelector)
	})
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

	err = gs.Execute(context.Background())
	assert.IsType(t, err, &json.SyntaxError{})

	assert.Nil(t, gs.GetLatestResponse())
	gasPrice, err := gs.GetCurrentGasPriceInWei()
	assert.Equal(t, big.NewInt(0), gasPrice)
	assert.Equal(t, ErrLatestGasPricesWereNotFetched, err)
}

func TestGasStation_GoodResponseShouldSave(t *testing.T) {
	t.Parallel()

	gsResponse := gasStationResponse{
		Fast:        1,
		Fastest:     2,
		SafeLow:     3,
		Average:     4,
		BlockTime:   5,
		BlockNum:    6,
		Speed:       7,
		SafeLowWait: 8,
		AvgWait:     9,
		FastWait:    10,
		FastestWait: 11,
	}
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

	err = gs.Execute(context.Background())
	assert.Nil(t, err)

	assert.Equal(t, gsResponse, *gs.GetLatestResponse())
}

func TestGasStation_GetCurrentGasPrice(t *testing.T) {
	t.Parallel()

	gsResponse := gasStationResponse{
		Fast:        1,
		Fastest:     2,
		SafeLow:     3,
		Average:     4,
		BlockTime:   5,
		BlockNum:    6,
		Speed:       7,
		SafeLowWait: 8,
		AvgWait:     9,
		FastWait:    10,
		FastestWait: 11,
	}
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

	_ = gs.Execute(context.Background())

	gs.SetSelector(core.EthFastGasPrice)
	price, err := gs.GetCurrentGasPriceInWei()
	require.Nil(t, err)
	expected := big.NewInt(0).Mul(big.NewInt(int64(gsResponse.Fast)), gasPriceMultiplier)
	assert.Equal(t, expected, price)

	gs.SetSelector(core.EthFastestGasPrice)
	price, err = gs.GetCurrentGasPriceInWei()
	require.Nil(t, err)
	expected = big.NewInt(0).Mul(big.NewInt(int64(gsResponse.Fastest)), gasPriceMultiplier)
	assert.Equal(t, expected, price)

	gs.SetSelector(core.EthAverageGasPrice)
	price, err = gs.GetCurrentGasPriceInWei()
	require.Nil(t, err)
	expected = big.NewInt(0).Mul(big.NewInt(int64(gsResponse.Average)), gasPriceMultiplier)
	assert.Equal(t, expected, price)

	gs.SetSelector(core.EthSafeLowGasPrice)
	price, err = gs.GetCurrentGasPriceInWei()
	require.Nil(t, err)
	expected = big.NewInt(0).Mul(big.NewInt(int64(gsResponse.SafeLow)), gasPriceMultiplier)
	assert.Equal(t, expected, price)

	gs.SetSelector("invalid")
	price, err = gs.GetCurrentGasPriceInWei()
	require.True(t, errors.Is(err, ErrInvalidGasPriceSelector))
	assert.Equal(t, big.NewInt(0), price)
}

func TestGasStation_GetCurrentGasPriceExceededMaximum(t *testing.T) {
	t.Parallel()

	gsResponse := gasStationResponse{
		Fast: 1001,
	}
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

	_ = gs.Execute(context.Background())

	price, err := gs.GetCurrentGasPriceInWei()
	require.True(t, errors.Is(err, ErrGasPriceIsHigherThanTheMaximumSet))
	assert.Equal(t, big.NewInt(0), price)
}
