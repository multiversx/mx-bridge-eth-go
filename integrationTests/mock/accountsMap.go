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

func newAccountsMap(accounts map[string]*api.AccountResponse, contracts map[string]*Contract) *accountsMap {
	return &accountsMap{
		accounts:  accounts,
		contracts: contracts,
	}
}

// GetOrCreateAccount returns a stored account based on the provided address or creates a new one if the account does not exist
func (am *accountsMap) GetOrCreateAccount(address string) *api.AccountResponse {
	am.mutAccounts.Lock()
	defer am.mutAccounts.Unlock()

	account, exists := am.accounts[address]
	if !exists {
		account = &api.AccountResponse{
			Address: address,
		}

		am.accounts[address] = account
	}

	return account
}

// SetAccount will set the account & contract with the provided values
func (am *accountsMap) SetAccount(account *api.AccountResponse, contract *Contract) {
	am.mutAccounts.Lock()
	defer am.mutAccounts.Unlock()

	if account == nil && contract == nil {
		log.Error("programming error in accountsMap.SetAccount, nil account and contract")
		return
	}

	address := ""
	if account != nil {
		address = account.Address
	}
	if contract != nil {
		address = contract.address
	}

	am.accounts[address] = account
	am.contracts[address] = contract
}

// GetContract will return the contract (if existing)
func (am *accountsMap) GetContract(address string) (*Contract, bool) {
	am.mutAccounts.RLock()
	defer am.mutAccounts.RUnlock()

	contract, found := am.contracts[address]

	return contract, found
}
