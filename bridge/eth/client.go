package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	logger "github.com/ElrondNetwork/elrond-go-logger"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	SignDatName   = "CurrentPendingTransaction"
	MessagePrefix = "\u0019Ethereum Signed Message:\n%d%s"
	GasLimit      = uint64(300000)
)

type BridgeContract interface {
	GetNextPendingTransaction(opts *bind.CallOpts) (Deposit, error)
	FinishCurrentPendingTransaction(opts *bind.TransactOpts, signData string, signatures [][]byte) (*types.Transaction, error)
}

type BlockchainClient interface {
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	ChainID(ctx context.Context) (*big.Int, error)
}

type Client struct {
	bridgeContract   BridgeContract
	blockchainClient BlockchainClient

	privateKey  *ecdsa.PrivateKey
	publicKey   *ecdsa.PublicKey
	broadcaster bridge.Broadcaster

	log logger.Logger
}

func NewClient(config bridge.Config, broadcaster bridge.Broadcaster) (*Client, error) {
	log := logger.GetOrCreate("EthClient")

	ethClient, err := ethclient.Dial(config.NetworkAddress)
	if err != nil {
		return nil, err
	}

	instance, err := NewContract(common.HexToAddress(config.BridgeAddress), ethClient)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	client := &Client{
		bridgeContract:   instance,
		blockchainClient: ethClient,

		privateKey:  privateKey,
		publicKey:   publicKeyECDSA,
		broadcaster: broadcaster,

		log: log,
	}

	return client, nil
}

func (c *Client) GetPendingDepositTransaction(ctx context.Context) *bridge.DepositTransaction {
	deposit, err := c.bridgeContract.GetNextPendingTransaction(&bind.CallOpts{Context: ctx})
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	var result *bridge.DepositTransaction
	if !reflect.DeepEqual(deposit.Depositor, common.Address{}) {
		result = &bridge.DepositTransaction{
			To:           string(deposit.Recipient),
			From:         deposit.Depositor.String(),
			TokenAddress: deposit.TokenAddress.String(),
			Amount:       deposit.Amount,
			DepositNonce: bridge.Nonce(deposit.Nonce.Uint64()),
		}
	}

	return result
}

func (c *Client) ProposeTransfer(context.Context, *bridge.DepositTransaction) (string, error) {
	return "", nil
}

func (c *Client) ProposeSetStatusSuccessOnPendingTransfer(context.Context) {
	data := fmt.Sprintf("%s:%d", SignDatName, bridge.Executed)
	msg := fmt.Sprintf(MessagePrefix, len(data), data)
	signature, err := c.signHash(msg)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	c.broadcaster.SendSignature(msg, signature)
}

func (c *Client) ProposeSetStatusFailedOnPendingTransfer(context.Context) {
	data := fmt.Sprintf("%s:%d", SignDatName, bridge.Rejected)
	msg := fmt.Sprintf(MessagePrefix, len(data), data)
	signature, err := c.signHash(msg)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	c.broadcaster.SendSignature(msg, signature)
}

func (c *Client) WasProposedTransfer(context.Context, bridge.Nonce) bool {
	return true
}

func (c *Client) GetActionIdForProposeTransfer(context.Context, bridge.Nonce) bridge.ActionId {
	return bridge.ActionId(0)
}

func (c *Client) WasProposedSetStatusSuccessOnPendingTransfer(context.Context) bool {
	return true
}

func (c *Client) WasProposedSetStatusFailedOnPendingTransfer(context.Context) bool {
	return true
}

func (c *Client) GetActionIdForSetStatusOnPendingTransfer(context.Context) bridge.ActionId {
	return bridge.ActionId(0)
}

func (c *Client) WasExecuted(context.Context, bridge.ActionId) bool {
	return false
}

func (c *Client) Sign(context.Context, bridge.ActionId) (string, error) {
	return "", nil
}

func (c *Client) Execute(ctx context.Context, _ bridge.ActionId) (string, error) {
	fromAddress := crypto.PubkeyToAddress(*c.publicKey)

	nonce, err := c.blockchainClient.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", err
	}

	gasPrice, err := c.blockchainClient.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	chainId, err := c.blockchainClient.ChainID(ctx)
	if err != nil {
		return "", err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, chainId)
	if err != nil {
		return "", err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = GasLimit
	auth.GasPrice = gasPrice

	transaction, err := c.bridgeContract.FinishCurrentPendingTransaction(auth, c.broadcaster.SignData(), c.broadcaster.Signatures())
	if err != nil {
		return "", err
	}

	return transaction.Hash().String(), err
}

func (c *Client) SignersCount(context.Context, bridge.ActionId) uint {
	return uint(len(c.broadcaster.Signatures()))
}

func (c *Client) signHash(msg string) ([]byte, error) {
	hash := crypto.Keccak256Hash([]byte(msg))

	signature, err := crypto.Sign(hash.Bytes(), c.privateKey)
	if err != nil {
		return nil, err
	}

	return signature, nil
}
