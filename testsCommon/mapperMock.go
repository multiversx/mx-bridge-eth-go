package testsCommon

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
)

// MapperMock -
type MapperMock struct {
	mut         sync.RWMutex
	erc20Ticker map[common.Address]string
	tickerErc20 map[string]common.Address
}

// NewMapperMock -
func NewMapperMock() *MapperMock {
	return &MapperMock{
		erc20Ticker: make(map[common.Address]string),
		tickerErc20: make(map[string]common.Address),
	}
}

// AddPair -
func (mock *MapperMock) AddPair(erc20 common.Address, ticker string) {
	mock.mut.Lock()
	mock.erc20Ticker[erc20] = ticker
	mock.tickerErc20[ticker] = erc20
	mock.mut.Unlock()
}

// GetTokenId -
func (mock *MapperMock) GetTokenId(s string) string {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	addr := common.HexToAddress(s)

	return mock.erc20Ticker[addr]
}

// GetErc20Address -
func (mock *MapperMock) GetErc20Address(s string) string {
	mock.mut.RLock()
	defer mock.mut.RUnlock()

	return mock.tickerErc20[s].Hex()
}

// IsInterfaceNil -
func (mock *MapperMock) IsInterfaceNil() bool {
	return mock == nil
}
