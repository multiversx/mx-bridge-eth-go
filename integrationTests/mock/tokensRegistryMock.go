package mock

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
)

// tokensRegistryMock is not concurrent safe
type tokensRegistryMock struct {
	ethToMultiversX map[common.Address]string
	multiversXToEth map[string]common.Address
	mintBurnTokens  map[string]bool
	nativeTokens    map[string]bool
	totalBalances   map[string]*big.Int
	mintBalances    map[string]*big.Int
	burnBalances    map[string]*big.Int
}

func (mock *tokensRegistryMock) addTokensPair(erc20Address common.Address, ticker string, isNativeToken, isMintBurnToken bool, totalBalance, mintBalances, burnBalances *big.Int) {
	integrationTests.Log.Info("added tokens pair", "ticker", ticker,
		"erc20 address", erc20Address.String(), "is native token", isNativeToken, "is mint burn token", isMintBurnToken,
		"total balance", totalBalance, "mint balances", mintBalances, "burn balances", burnBalances)

	mock.ethToMultiversX[erc20Address] = ticker

	hexedTicker := hex.EncodeToString([]byte(ticker))
	mock.multiversXToEth[hexedTicker] = erc20Address

	if isNativeToken {
		mock.nativeTokens[hexedTicker] = true
	}
	if isMintBurnToken {
		mock.mintBurnTokens[hexedTicker] = true
		mock.mintBalances[hexedTicker] = mintBalances
		mock.burnBalances[hexedTicker] = burnBalances
	} else {
		mock.totalBalances[hexedTicker] = totalBalance
	}
}

func (mock *tokensRegistryMock) clearTokens() {
	mock.ethToMultiversX = make(map[common.Address]string)
	mock.multiversXToEth = make(map[string]common.Address)
	mock.mintBurnTokens = make(map[string]bool)
	mock.nativeTokens = make(map[string]bool)
	mock.totalBalances = make(map[string]*big.Int)
	mock.mintBalances = make(map[string]*big.Int)
	mock.burnBalances = make(map[string]*big.Int)
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

func (mock *tokensRegistryMock) isMintBurnToken(ticker string) bool {
	_, found := mock.mintBurnTokens[ticker]

	return found
}

func (mock *tokensRegistryMock) isNativeToken(ticker string) bool {
	_, found := mock.nativeTokens[ticker]

	return found
}

func (mock *tokensRegistryMock) getTotalBalances(ticker string) *big.Int {
	return mock.totalBalances[ticker]
}

func (mock *tokensRegistryMock) getMintBalances(ticker string) *big.Int {
	return mock.mintBalances[ticker]
}

func (mock *tokensRegistryMock) getBurnBalances(ticker string) *big.Int {
	return mock.burnBalances[ticker]
}
