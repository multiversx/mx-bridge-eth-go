package mock

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ethereum/go-ethereum/common"
)

// tokensRegistryMock is not concurrent safe
type tokensRegistryMock struct {
	ethToElrond map[common.Address]string
	elrondToEth map[string]common.Address
}

func (mock *tokensRegistryMock) addTokensPair(erc20Address common.Address, ticker string) {
	integrationTests.Log.Info("added tokens pair", "ticker", ticker, "erc20 address", erc20Address.String())

	mock.ethToElrond[erc20Address] = ticker

	hexedTicker := hex.EncodeToString([]byte(ticker))
	mock.elrondToEth[hexedTicker] = erc20Address
}

func (mock *tokensRegistryMock) clearTokens() {
	mock.ethToElrond = make(map[common.Address]string)
	mock.elrondToEth = make(map[string]common.Address)
}

func (mock *tokensRegistryMock) getTicker(erc20Address common.Address) string {
	ticker, found := mock.ethToElrond[erc20Address]
	if !found {
		panic("tiker for erc20 address " + erc20Address.String() + " not found")
	}

	return ticker
}

func (mock *tokensRegistryMock) getErc20Address(ticker string) common.Address {
	addr, found := mock.elrondToEth[ticker]
	if !found {
		panic("erc20 address for ticker " + ticker + " not found")
	}

	return addr
}
