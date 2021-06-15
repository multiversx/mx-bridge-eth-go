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
	GetNextPendingBatch(opts *bind.CallOpts) (Batch, error)
	FinishCurrentPendingBatch(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error)
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	WasBatchExecuted(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	WasBatchFinished(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
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

	lastProposedStatus uint8
	lastTransferBatch  *bridge.Batch

	log logger.Logger
}

func NewClient(config bridge.Config, broadcaster bridge.Broadcaster, mapper bridge.Mapper) (*Client, error) {
	log := logger.GetOrCreate("EthClient")

	ethClient, err := ethclient.Dial(config.NetworkAddress)
	if err != nil {
		return nil, err
	}

	instance, err := NewBridge(common.HexToAddress(config.BridgeAddress), ethClient)
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

func (c *Client) GetPending(ctx context.Context) *bridge.Batch {
	batch, err := c.bridgeContract.GetNextPendingBatch(&bind.CallOpts{Context: ctx})
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	var result *bridge.Batch
	if !reflect.DeepEqual(batch.Nonce, bridge.NewBatchId(0)) {
		var transactions []*bridge.DepositTransaction
		for _, deposit := range batch.Deposits {
			tx := &bridge.DepositTransaction{
				To:           string(deposit.Recipient),
				From:         deposit.Depositor.String(),
				TokenAddress: deposit.TokenAddress.String(),
				Amount:       deposit.Amount,
				DepositNonce: deposit.Nonce,
			}
			transactions = append(transactions, tx)
		}

		result = &bridge.Batch{
			Id:           batch.Nonce,
			Transactions: transactions,
		}
	}

	return result
}

func (c *Client) ProposeSetStatus(_ context.Context, _ *bridge.Batch) {
	// TODO: revisit this
	c.broadcastSignatureForFinishCurrentPendingTransaction(c.lastProposedStatus, nil)
}

func (c *Client) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	c.lastTransferBatch = batch
	//c.broadcastSignatureForTransfer(tx.To, c.getErc20AddressFromTokenId(tx.TokenAddress), tx.Amount, tx.DepositNonce)
	return "", nil
}

func (c *Client) WasProposedTransfer(context.Context, bridge.BatchId) bool {
	return true
}

func (c *Client) GetActionIdForProposeTransfer(context.Context, bridge.BatchId) bridge.ActionId {
	return bridge.NewActionId(0)
}

func (c *Client) WasProposedSetStatus(context.Context, *bridge.Batch) bool {
	return true
}

func (c *Client) GetActionIdForSetStatusOnPendingTransfer(context.Context) bridge.ActionId {
	return bridge.NewActionId(0)
}

func (c *Client) WasExecuted(ctx context.Context, _ bridge.ActionId, batchId bridge.BatchId) bool {
	var wasExecuted bool
	var err error = nil

	if c.lastTransferBatch == nil {
		wasExecuted, err = c.bridgeContract.WasBatchExecuted(&bind.CallOpts{Context: ctx}, batchId)
	} else {
		wasExecuted, err = c.bridgeContract.WasBatchFinished(&bind.CallOpts{Context: ctx}, c.lastTransferBatch.Id)
	}
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	c.cleanState(wasExecuted)

	return wasExecuted
}

func (c *Client) Sign(context.Context, bridge.ActionId) (string, error) {
	return "", nil
}

func (c *Client) Execute(ctx context.Context, _ bridge.ActionId, batchId bridge.BatchId) (string, error) {
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

	//var transaction *types.Transaction
	//
	//signatures := c.broadcaster.Signatures()
	//if c.lastTransferBatch == nil {
	//	transaction, err = c.bridgeContract.FinishCurrentPendingBatch(auth, batchId, c.lastProposedStatus, signatures)
	//} else {
	//	tx := c.lastTransferBatch
	//	tokenAddress := common.HexToAddress(c.getErc20AddressFromTokenId(tx.TokenAddress))
	//	toAddress := common.HexToAddress(tx.To)
	//	transaction, err = c.bridgeContract.ExecuteTransfer(auth, tokenAddress, toAddress, tx.Amount, tx.DepositNonce, signatures)
	//}
	//if err != nil {
	//	return "", err
	//}

	//return transaction.Hash().String(), err
	return "todo", err
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
	return c.mapper.GetErc20Address(tokenId[2:])
}

func (c *Client) cleanState(wasExecuted bool) {
	if wasExecuted && c.lastTransferBatch != nil {
		c.lastTransferBatch = nil
	} else if wasExecuted {
		c.lastProposedStatus = 0
	}
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
