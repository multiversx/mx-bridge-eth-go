package elrond

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/data"
)

type elrondProxy interface {
	GetNetworkConfig() (*data.NetworkConfig, error)
	SendTransaction(*data.Transaction) (string, error)
}

type Client struct {
	proxy         elrondProxy
	bridgeAddress string
	privateKey    []byte
	address       *elrondAddress
	nonce         uint64
}

// TODO: remove this when Stringer bug is fixes
type elrondAddress struct {
	addressString string
}

func (a *elrondAddress) AddressAsBech32String() string {
	return a.addressString
}

func (a *elrondAddress) AddressBytes() []byte {
	return nil
}

func (a *elrondAddress) IsValid() bool {
	return true
}

func (a *elrondAddress) IsInterfaceNil() bool {
	return false
}

func (a *elrondAddress) String() string {
	return a.addressString
}

// end here

func NewClient(config bridge.Config) (*Client, error) {
	proxy := blockchain.NewElrondProxy(config.NetworkAddress)

	privateKey, err := erdgo.LoadPrivateKeyFromPemFile(config.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	addressString, err := erdgo.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	address := &elrondAddress{addressString: addressString}

	account, err := proxy.GetAccount(address)
	if err != nil {
		return nil, err
	}
	initialNonce := account.Nonce

	return &Client{
		proxy:         proxy,
		bridgeAddress: config.BridgeAddress,
		privateKey:    privateKey,
		address:       address,
		nonce:         initialNonce,
	}, nil
}

func (c *Client) GetPendingDepositTransaction(context.Context) *bridge.DepositTransaction {
	return nil
}

func (c *Client) Propose(*bridge.DepositTransaction) {
}

func (c *Client) WasProposed(*bridge.DepositTransaction) bool {
	return false
}

func (c *Client) WasExecuted(*bridge.DepositTransaction) bool {
	return false
}

func (c *Client) Sign(*bridge.DepositTransaction) {
}

func (c *Client) Execute(*bridge.DepositTransaction) (string, error) {
	tx, err := c.buildTransaction()
	if err != nil {
		return "", nil
	}

	hash, err := c.proxy.SendTransaction(&tx)
	if err == nil {
		c.incrementNonce()
	}

	return hash, err
}

func (c *Client) SignersCount(*bridge.DepositTransaction) uint {
	return 0
}

func (c *Client) buildTransaction() (data.Transaction, error) {
	networkConfig, err := c.proxy.GetNetworkConfig()
	if err != nil {
		return data.Transaction{}, err
	}

	tx := data.Transaction{
		ChainID: networkConfig.ChainID,
		Version: networkConfig.MinTransactionVersion,
		// TODO: /transaction/cost to estimate tx cost
		GasLimit: networkConfig.MinGasLimit * 4 * 10,
		GasPrice: networkConfig.MinGasPrice,
		Nonce:    c.nonce,
		Data:     []byte("increment"),
		SndAddr:  c.address.AddressAsBech32String(),
		RcvAddr:  c.bridgeAddress,
		Value:    "0",
	}

	err = erdgo.SignTransaction(&tx, c.privateKey)
	if err != nil {
		return data.Transaction{}, err
	}

	return tx, nil
}

func (c *Client) incrementNonce() {
	c.nonce++
}
