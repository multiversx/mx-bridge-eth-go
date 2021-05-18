package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	logger "github.com/ElrondNetwork/elrond-go-logger"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	MessagePrefix = "\u0019Ethereum Signed Message:\n32"
	GasLimit      = uint64(300000)
)

type BridgeContract interface {
	GetNextPendingTransaction(opts *bind.CallOpts) (Deposit, error)
	FinishCurrentPendingTransaction(opts *bind.TransactOpts, depositNonce *big.Int, newDepositStatus uint8, signatures [][]byte) (*types.Transaction, error)
	ExecuteTransfer(opts *bind.TransactOpts, token common.Address, recipient common.Address, amount *big.Int, depositNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	WasTransactionExecuted(opts *bind.CallOpts, nonceId *big.Int) (bool, error)
	WasTransferExecuted(opts *bind.CallOpts, nonceId *big.Int) (bool, error)
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
	mapper      bridge.Mapper

	lastProposedStatus      uint8
	lastTransferTransaction *bridge.DepositTransaction

	log logger.Logger
}

func NewClient(config bridge.Config, broadcaster bridge.Broadcaster, mapper bridge.Mapper) (*Client, error) {
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
		mapper:      mapper,

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
			DepositNonce: deposit.Nonce,
		}
	}

	return result
}

func (c *Client) ProposeTransfer(_ context.Context, tx *bridge.DepositTransaction) (string, error) {
	c.lastTransferTransaction = tx
	c.broadcastSignatureForTransfer(tx.To, c.getErc20AddressFromTokenId(tx.TokenAddress), tx.Amount, tx.DepositNonce)
	return "", nil
}

func (c *Client) ProposeSetStatus(_ context.Context, status uint8, nonce bridge.Nonce) {
	c.lastProposedStatus = status
	c.broadcastSignatureForFinishCurrentPendingTransaction(c.lastProposedStatus, nonce)
}

func (c *Client) WasProposedTransfer(context.Context, bridge.Nonce) bool {
	return true
}

func (c *Client) GetActionIdForProposeTransfer(context.Context, bridge.Nonce) bridge.ActionId {
	return bridge.NewActionId(0)
}

func (c *Client) WasProposedSetStatusOnPendingTransfer(context.Context, uint8) bool {
	return true
}

func (c *Client) GetActionIdForSetStatusOnPendingTransfer(context.Context) bridge.ActionId {
	return bridge.NewActionId(0)
}

func (c *Client) WasExecuted(ctx context.Context, _ bridge.ActionId, nonce bridge.Nonce) bool {
	var wasExecuted bool
	var err error = nil

	if c.lastTransferTransaction == nil {
		wasExecuted, err = c.bridgeContract.WasTransactionExecuted(&bind.CallOpts{Context: ctx}, nonce)
	} else {
		wasExecuted, err = c.bridgeContract.WasTransferExecuted(&bind.CallOpts{Context: ctx}, c.lastTransferTransaction.DepositNonce)
	}
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	if wasExecuted {
		c.cleanState()
	}

	return wasExecuted
}

func (c *Client) Sign(context.Context, bridge.ActionId) (string, error) {
	return "", nil
}

func (c *Client) Execute(ctx context.Context, _ bridge.ActionId, nonce bridge.Nonce) (string, error) {
	fromAddress := crypto.PubkeyToAddress(*c.publicKey)

	blockNonce, err := c.blockchainClient.PendingNonceAt(ctx, fromAddress)
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

	auth.Nonce = big.NewInt(int64(blockNonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = GasLimit
	auth.GasPrice = gasPrice
	auth.Context = ctx

	var transaction *types.Transaction

	signatures := c.broadcaster.Signatures()
	if c.lastTransferTransaction == nil {
		transaction, err = c.bridgeContract.FinishCurrentPendingTransaction(auth, nonce, c.lastProposedStatus, signatures)
	} else {
		tx := c.lastTransferTransaction
		tokenAddress := common.HexToAddress(c.getErc20AddressFromTokenId(tx.TokenAddress))
		toAddress := common.HexToAddress(tx.To)
		transaction, err = c.bridgeContract.ExecuteTransfer(auth, tokenAddress, toAddress, tx.Amount, tx.DepositNonce, signatures)
	}
	if err != nil {
		return "", err
	}

	return transaction.Hash().String(), err
}

func (c *Client) SignersCount(context.Context, bridge.ActionId) uint {
	return uint(len(c.broadcaster.Signatures()))
}

// utils

func (c *Client) signHash(hash common.Hash) ([]byte, error) {
	valueToSign := crypto.Keccak256Hash(append([]byte(MessagePrefix), hash.Bytes()...))
	signature, err := crypto.Sign(valueToSign.Bytes(), c.privateKey)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (c *Client) broadcastSignatureForTransfer(to, tokenAddress string, amount *big.Int, depositNonce *big.Int) {
	arguments, err := transferArgs()
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	pack, err := arguments.Pack(common.HexToAddress(to), common.HexToAddress(tokenAddress), amount, depositNonce, "ExecuteTransfer")
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	hash := crypto.Keccak256Hash(pack)
	signature, err := c.signHash(hash)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	c.broadcaster.SendSignature(signature)
}

func (c *Client) broadcastSignatureForFinishCurrentPendingTransaction(status uint8, nonce bridge.Nonce) {
	arguments, err := finishCurrentPendingTransactionArgs()
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	pack, err := arguments.Pack(new(big.Int).Set(nonce), status, "CurrentPendingTransaction")
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	hash := crypto.Keccak256Hash(pack)
	signature, err := c.signHash(hash)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	c.broadcaster.SendSignature(signature)
}

func (c *Client) getErc20AddressFromTokenId(tokenId string) string {
	return c.mapper.GetErc20Address(tokenId[2:])[:40]
}

func (c *Client) cleanState() {
	c.lastProposedStatus = 0
	c.lastTransferTransaction = nil
}

// helpers

func transferArgs() (abi.Arguments, error) {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	stringType, err := abi.NewType("string", "", nil)
	if err != nil {
		return nil, err
	}

	return abi.Arguments{
		abi.Argument{Name: "recipientAddress", Type: addressType},
		abi.Argument{Name: "tokenAddress", Type: addressType},
		abi.Argument{Name: "blockNonce", Type: uint256Type},
		abi.Argument{Name: "amount", Type: uint256Type},
		abi.Argument{Name: "executeTransfer", Type: stringType},
	}, nil

}

func finishCurrentPendingTransactionArgs() (abi.Arguments, error) {
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	uint8Type, err := abi.NewType("uint8", "", nil)
	if err != nil {
		return nil, err
	}

	stringType, err := abi.NewType("string", "", nil)
	if err != nil {
		return nil, err
	}

	return abi.Arguments{
		abi.Argument{Name: "depositNonce", Type: uint256Type},
		abi.Argument{Name: "newDepositStatus", Type: uint8Type},
		abi.Argument{Name: "currentPendingTransaction", Type: stringType},
	}, nil
}
