package framework

import (
	"context"
	"fmt"
	"testing"

	"github.com/block-vision/sui-go-sdk/sui"
)

const (
	proxyUrl  = "http://127.0.0.1:9000"
	faucetUrl = "http://127.0.0.1:9123"
)

type suiChainSimulatorWrapper struct {
	testing.TB
	suiProxy sui.ISuiAPI
}

type ArgSuiChainSimulatorWrapper struct {
	TB testing.TB
}

func CreateSuiChainSimulatorWrapper(args ArgSuiChainSimulatorWrapper) *suiChainSimulatorWrapper {
	proxy := sui.NewSuiClient(proxyUrl)
	wrapper := &suiChainSimulatorWrapper{
		TB:       args.TB,
		suiProxy: proxy,
	}
	return wrapper
}

func (instance *suiChainSimulatorWrapper) FundWallets(wallets [][]byte) {
	for _, wallet := range wallets {
		header := map[string]string{}
		err := sui.RequestSuiFromFaucet(faucetUrl, string(wallet), header)
		if err != nil {
			log.Error("error in suiChainSimulator.FundWallets", "error", err.Error())
			continue
		}
		log.Info(fmt.Sprintf("Wallet %s funded successfuly", string(wallet)))
	}
}

func (instance *suiChainSimulatorWrapper) GenerateBlocks(ctx context.Context, numBlocks int) {
}

//func (instance *suiChainSimulatorWrapper) GetNetworkUrl() string {
//	return instance.networkUrl
//}
