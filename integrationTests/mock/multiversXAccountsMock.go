package mock

import (
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

type multiversXAccountsMock struct {
	accounts map[string]*data.Account
}

func newMultiversXAccountsMock() *multiversXAccountsMock {
	return &multiversXAccountsMock{
		accounts: make(map[string]*data.Account),
	}
}

func (mock *multiversXAccountsMock) getOrCreate(address core.AddressHandler) *data.Account {
	addrAsString := string(address.AddressBytes())
	acc, found := mock.accounts[addrAsString]
	if !found {
		acc = &data.Account{}
		mock.accounts[addrAsString] = acc
	}

	return acc
}

func (mock *multiversXAccountsMock) updateNonce(address core.AddressHandler, nonce uint64) {
	acc := mock.getOrCreate(address)
	acc.Nonce = nonce
}
