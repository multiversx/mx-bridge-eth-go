package groups

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	mockFacade "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/facade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	data := statusRsp.Data
	dataAsMap, ok := data.(map[string]interface{})
	require.True(t, ok)

	require.Equal(t, len(dataAsMap), len(response))
	for key, val := range dataAsMap {
		valRequired := response[key]
		assert.Equal(t, valRequired, val)
	}

	require.Equal(t, resp.Code, http.StatusOK)
	assert.Empty(t, statusRsp.Error)
}
