package groups

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	mockFacade "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/facade"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
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

func TestGetStatus(t *testing.T) {
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
