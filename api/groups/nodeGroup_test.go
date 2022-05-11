package groups

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	mockFacade "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/facade"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	elrondApiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var marshalizer = &marshal.JsonMarshalizer{}

func equalStructsThroughJsonSerialization(t *testing.T, expected interface{}, got interface{}) {
	expectedBuff, err := marshalizer.Marshal(expected)
	require.Nil(t, err)

	gotBuff, err := marshalizer.Marshal(got)
	require.Nil(t, err)

	assert.Equal(t, string(expectedBuff), string(gotBuff))
}

func TestNewNodeGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		ng, err := NewNodeGroup(nil)

		assert.True(t, check.IfNil(ng))
		assert.True(t, errors.Is(err, elrondApiErrors.ErrNilFacadeHandler))
	})
	t.Run("should work", func(t *testing.T) {
		ng, err := NewNodeGroup(&mockFacade.RelayerFacadeStub{})

		assert.False(t, check.IfNil(ng))
		assert.Nil(t, err)
	})
}

func TestGetStatus_Errors(t *testing.T) {
	t.Parallel()

	facade := mockFacade.RelayerFacadeStub{
		GetMetricsCalled: func(name string) (core.GeneralMetrics, error) {
			return nil, expectedError
		},
	}

	ng, err := NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(ng, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := generalResponse{}
	loadResponse(resp.Body, &statusRsp)

	assert.Nil(t, statusRsp.Data)
	assert.True(t, strings.Contains(statusRsp.Error, expectedError.Error()))
	assert.True(t, strings.Contains(statusRsp.Error, ErrGettingMetrics.Error()))
	require.Equal(t, resp.Code, http.StatusInternalServerError)
}

func TestGetStatus_ShouldWork(t *testing.T) {
	t.Parallel()

	response := make(core.GeneralMetrics)
	response["metric"] = "value1"
	facade := mockFacade.RelayerFacadeStub{
		GetMetricsCalled: func(name string) (core.GeneralMetrics, error) {
			return response, nil
		},
	}

	ng, err := NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(ng, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := generalResponse{}
	loadResponse(resp.Body, &statusRsp)

	equalStructsThroughJsonSerialization(t, response, statusRsp.Data)

	require.Equal(t, resp.Code, http.StatusOK)
	assert.Empty(t, statusRsp.Error)
}

func TestGetStatusList(t *testing.T) {
	t.Parallel()

	response := make(core.GeneralMetrics)
	response["metric"] = []string{"value1", "value2"}
	facade := mockFacade.RelayerFacadeStub{
		GetMetricsListCalled: func() core.GeneralMetrics {
			return response
		},
	}

	ng, err := NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(ng, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/status/list", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := generalResponse{}
	loadResponse(resp.Body, &statusRsp)

	equalStructsThroughJsonSerialization(t, response, statusRsp.Data)

	require.Equal(t, resp.Code, http.StatusOK)
	assert.Empty(t, statusRsp.Error)
}

func TestNodeGroup_UpdateFacade(t *testing.T) {
	t.Parallel()

	t.Run("nil facade should error", func(t *testing.T) {
		ng, _ := NewNodeGroup(&mockFacade.RelayerFacadeStub{})

		err := ng.UpdateFacade(nil)
		assert.Equal(t, elrondApiErrors.ErrNilFacadeHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		ng, _ := NewNodeGroup(&mockFacade.RelayerFacadeStub{})

		newFacade := &mockFacade.RelayerFacadeStub{}

		err := ng.UpdateFacade(newFacade)
		assert.Nil(t, err)
		assert.True(t, ng.facade == newFacade) // pointer testing
	})
}
