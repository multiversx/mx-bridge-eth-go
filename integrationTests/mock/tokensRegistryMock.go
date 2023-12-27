package mock

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
)

// tokensRegistryMock is not concurrent safe
type tokensRegistryMock struct {
	ethToMultiversX     map[common.Address]string
	multiversXToEth     map[string]common.Address
	nativeTokensBalance map[string]*big.Int
}

func (mock *tokensRegistryMock) addTokensPair(erc20Address common.Address, ticker string, isNativeToken bool, nativeBalance *big.Int) {
	integrationTests.Log.Info("added tokens pair", "ticker", ticker,
		"erc20 address", erc20Address.String(), "is native token", isNativeToken, "native balance", nativeBalance)

	mock.ethToMultiversX[erc20Address] = ticker

	hexedTicker := hex.EncodeToString([]byte(ticker))
	mock.multiversXToEth[hexedTicker] = erc20Address

	if isNativeToken {
		mock.nativeTokensBalance[hexedTicker] = nativeBalance
	}
}

func (mock *tokensRegistryMock) clearTokens() {
	mock.ethToMultiversX = make(map[common.Address]string)
	mock.multiversXToEth = make(map[string]common.Address)
	mock.nativeTokensBalance = make(map[string]*big.Int)
}

func (mock *tokensRegistryMock) getTicker(erc20Address common.Address) string {
	ticker, found := mock.ethToMultiversX[erc20Address]
	if !found {
		panic("tiker for erc20 address " + erc20Address.String() + " not found")
	}

	return ticker
}

func (mock *tokensRegistryMock) getErc20Address(ticker string) common.Address {
	addr, found := mock.multiversXToEth[ticker]
	if !found {
		panic("erc20 address for ticker " + ticker + " not found")
	}

	return addr
}

func (mock *tokensRegistryMock) isNativeToken(ticker string) bool {
	_, found := mock.nativeTokensBalance[ticker]

	return found
}

func (mock *tokensRegistryMock) getAccumulatedBurn(ticker string) *big.Int {
	return mock.nativeTokensBalance[ticker]
}
