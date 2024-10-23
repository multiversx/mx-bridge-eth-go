package factory

import (
	"path"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

var log = logger.GetOrCreate("factory_test")

func TestCreateUnitStorer(t *testing.T) {
	t.Parallel()

	workingDir := t.TempDir()

	log.Info("created temporary directory", "directory", workingDir)

	cfg := config.StorageConfig{
		Cache: config.CacheConfig{
			Name:     "StatusMetricsStorage",
			Type:     "LRU",
			Capacity: 1000,
		},
		DB: config.DBConfig{
			FilePath:          "StatusMetricsStorageDB",
			Type:              "LvlDBSerial",
			BatchDelaySeconds: 1,
			MaxBatchSize:      100,
			MaxOpenFiles:      10,
		},
	}

	storer, err := CreateUnitStorer(cfg, path.Join(workingDir, "db"))
	assert.Nil(t, err)
	assert.False(t, check.IfNil(storer))

	log.Info("writing data in storer")
	err = storer.Put([]byte("key"), []byte("value"))
	assert.Nil(t, err)

	time.Sleep(time.Second * 2)

	log.Info("closing the storer")
	err = storer.Close()
	assert.Nil(t, err)
}
