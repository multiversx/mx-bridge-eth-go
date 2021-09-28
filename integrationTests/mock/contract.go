package mock

import "sync"

// ContractHandler represents a handler of a contract
type ContractHandler func(caller string, value string, arguments ...string) ([][]byte, error)

// Contract represents a VM contract
type Contract struct {
	address    string
	mutHandler sync.RWMutex
	handlers   map[string]ContractHandler
}

// NewContract returns a new Contract instance
func NewContract(address string) *Contract {
	return &Contract{
		address:  address,
		handlers: make(map[string]ContractHandler),
	}
}

// AddHandler will add a handler to the contract
func (c *Contract) AddHandler(functionName string, handler ContractHandler) {
	c.mutHandler.Lock()
	defer c.mutHandler.Unlock()

	c.handlers[functionName] = handler
}

// GetHandler will return the handler associated to the provided function. Can return nil.
func (c *Contract) GetHandler(functionName string) ContractHandler {
	c.mutHandler.RLock()
	defer c.mutHandler.RUnlock()

	return c.handlers[functionName]
}
