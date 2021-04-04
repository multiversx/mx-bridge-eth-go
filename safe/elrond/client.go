package elrond

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/data"
	"math/big"
)

type elrondProxy interface {
	GetNetworkConfig() (*data.NetworkConfig, error)
	SendTransaction(tx *data.Transaction) (string, error)
}

type Client struct {
	proxy       elrondProxy
	safeAddress string
	privateKey  []byte

	address *elrondAddress
	account *data.Account
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

func NewClient(rawUrl, safeAddress, privateKeyPath string) (*Client, error) {
	proxy := blockchain.NewElrondProxy(rawUrl)

	privateKey, err := erdgo.LoadPrivateKeyFromPemFile(privateKeyPath)
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

	return &Client{
		proxy:       proxy,
		safeAddress: safeAddress,
		privateKey:  privateKey,

		address: address,
		account: account,
	}, nil
}

// Bridge broadcasts a transaction to the network and returns the txhash if successful
func (c *Client) Bridge(*safe.DepositTransaction) (string, error) {
	networkConfig, _ := c.proxy.GetNetworkConfig()

	tx := c.buildTransaction(networkConfig)
	err := erdgo.SignTransaction(&tx, c.privateKey)

	if err != nil {
		return "", err
	}

	return c.proxy.SendTransaction(&tx)
}

func (c *Client) GetTransactions(context.Context, *big.Int, safe.SafeTxChan) {
	// TODO: follow the pattern in eth to get blocks -> transactions to the safe contract
}

func (c *Client) buildTransaction(networkConfig *data.NetworkConfig) data.Transaction {
	return data.Transaction{
		ChainID: networkConfig.ChainID,
		Version: networkConfig.MinTransactionVersion,
		// TODO: /transaction/cost to estimate tx cost
		GasLimit: networkConfig.MinGasLimit * 2 * 10,
		GasPrice: networkConfig.MinGasPrice,
		Nonce:    c.account.Nonce,
		Data:     []byte("increment"),
		SndAddr:  c.address.AddressAsBech32String(),
		RcvAddr:  c.safeAddress,
		Value:    "0",
	}
}
