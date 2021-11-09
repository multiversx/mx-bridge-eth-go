package factory

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("factory_test")

func TestCreateUnitStorer(t *testing.T) {
	t.Parallel()

	workingDir, err := ioutil.TempDir("", "")
	require.Nil(t, err)

	log.Info("created temporary directory", "directory", workingDir)

	defer func() {
		err = os.RemoveAll(workingDir)
		require.Nil(t, err)

		if err == nil {
			log.Info("removed temporary directory", "directory", workingDir)
		} else {
			log.Error("error while removing temporary directory", "directory", workingDir, "error", err)
		}
	}()

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
