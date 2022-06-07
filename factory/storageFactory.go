package factory

import (
	"path"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// CreateUnitStorer based on the config and the working directory
func CreateUnitStorer(config config.StorageConfig, workingDir string) (core.Storer, error) {
	statusMetricsDbConfig := factory.GetDBFromConfig(config.DB)
	dbPath := path.Join(workingDir, config.DB.FilePath)
	statusMetricsDbConfig.FilePath = dbPath

	return storageUnit.NewStorageUnitFromConf(
		factory.GetCacherFromConfig(config.Cache),
		statusMetricsDbConfig)
}
