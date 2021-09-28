package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
)

type accountsMap struct {
	mutAccounts sync.RWMutex
	contracts   map[string]*Contract
	accounts    map[string]*api.AccountResponse
}

func newAccountsMap() *accountsMap {
	return &accountsMap{
		accounts:  make(map[string]*api.AccountResponse),
		contracts: make(map[string]*Contract),
	}
}

// GetAccount returns a stored account based on the provided address
func (am *accountsMap) GetAccount(address string) *api.AccountResponse {
	am.mutAccounts.Lock()
	defer am.mutAccounts.Unlock()

	account, exists := am.accounts[address]
	if !exists {
		account = &api.AccountResponse{
			Address:         address,
			Nonce:           0,
			Balance:         "0",
			Username:        "",
			Code:            "",
			CodeHash:        nil,
			RootHash:        nil,
			CodeMetadata:    nil,
			DeveloperReward: "",
			OwnerAddress:    "",
		}

		am.accounts[address] = account
	}

	return account
}

// SetAccount will set the account & contract with the provided value
func (am *accountsMap) SetAccount(address string, account *api.AccountResponse, contract *Contract) {
	am.mutAccounts.Lock()
	defer am.mutAccounts.Unlock()

	am.accounts[address] = account
	am.contracts[address] = contract
}

// GetContract will return the contract (if existing)
func (am *accountsMap) GetContract(address string) (*Contract, bool) {
	am.mutAccounts.Lock()
	defer am.mutAccounts.Unlock()

	contract, found := am.contracts[address]

	return contract, found
}
