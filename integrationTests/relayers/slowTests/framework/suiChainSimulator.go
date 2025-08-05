package framework

import (
	"testing"

	"github.com/block-vision/sui-go-sdk/sui"
)

type suiChainSimulatorWrapper struct {
	testing.TB
	networkUrl string
}

type ArgSuiChainSimulatorWrapper struct {
	TB testing.TB
}

func CreateSuiChainSimulatorWrapper(args ArgSuiChainSimulatorWrapper) *suiChainSimulatorWrapper {
	wrapper := &suiChainSimulatorWrapper{
		TB:         args.TB,
		networkUrl: "http://localhost:9000",
	}
	return wrapper
}

func (s *suiChainSimulatorWrapper) GenerateBlocks() {}

func (s *suiChainSimulatorWrapper) FundWallets(wallets []string) {
	for _, wallet := range wallets {

		header := map[string]string{}
		err := sui.RequestSuiFromFaucet(s.networkUrl, wallet, header)
		if err != nil {
			log.Error("error in suiChainSimulatorWrapper.FundWallets", "error", err)
		}
		log.Info("Funded wallet: " + wallet)
	}
}

func (s *suiChainSimulatorWrapper) GetNetworkUrl() string {
	return s.networkUrl
}
