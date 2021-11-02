package mock

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type elrondAccountsMock struct {
	accounts map[string]*data.Account
}

func newElrondAccountsMock() *elrondAccountsMock {
	return &elrondAccountsMock{
		accounts: make(map[string]*data.Account),
	}
}

func (mock *elrondAccountsMock) getOrCreate(address core.AddressHandler) *data.Account {
	addrAsString := string(address.AddressBytes())
	acc, found := mock.accounts[addrAsString]
	if !found {
		acc = &data.Account{}
		mock.accounts[addrAsString] = acc
	}

	return acc
}

func (mock *elrondAccountsMock) updateNonce(address core.AddressHandler, nonce uint64) {
	acc := mock.getOrCreate(address)
	acc.Nonce = nonce
}
