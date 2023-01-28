package factory

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/status"
	"github.com/stretchr/testify/assert"
)

func TestStartWebServer(t *testing.T) {
	t.Parallel()

	cfg := config.Configs{
		GeneralConfig:   config.Config{},
		ApiRoutesConfig: config.ApiRoutesConfig{},
		FlagsConfig: config.ContextFlagsConfig{
			RestApiInterface: core.WebServerOffString,
		},
	}

	webServer, err := StartWebServer(cfg, status.NewMetricsHolder())
	assert.Nil(t, err)
	assert.NotNil(t, webServer)

	err = webServer.Close()
	assert.Nil(t, err)
}
