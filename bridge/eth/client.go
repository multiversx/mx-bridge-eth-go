package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

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
)

type BridgeContract interface {
	GetNextPendingBatch(opts *bind.CallOpts) (Batch, error)
	FinishCurrentPendingBatch(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error)
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	WasBatchExecuted(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	WasBatchFinished(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	Quorum(opts *bind.CallOpts) (*big.Int, error)
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

	lastProposedStatuses []uint8
	lastTransferBatch    *bridge.Batch
	lastSignatureAction  func()
	gasLimit             uint64

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
		gasLimit:         config.GasLimit,
		privateKey:       privateKey,
		publicKey:        publicKeyECDSA,
		broadcaster:      broadcaster,
		mapper:           mapper,

		log: log,
	}

	return client, nil
}

func (c *Client) GetPending(ctx context.Context) *bridge.Batch {
	c.log.Info("ETH: Getting pending batch")
	batch, err := c.bridgeContract.GetNextPendingBatch(&bind.CallOpts{Context: ctx})
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	var result *bridge.Batch
	if batch.Nonce.Cmp(big.NewInt(0)) != 0 {
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

func (c *Client) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	for _, tx := range batch.Transactions {
		c.lastProposedStatuses = append(c.lastProposedStatuses, tx.Status)
	}
	c.lastSignatureAction = func() {
		c.broadcastSignatureForFinishCurrentPendingTransaction(batch.Id, c.lastProposedStatuses)
	}
	c.log.Info(fmt.Sprintf("ETH: Broadcast status signatures for for batchId %v", batch.Id))
}

func (c *Client) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	c.lastTransferBatch = batch
	c.lastSignatureAction = func() {
		c.broadcastSignatureForTransfer(batch)
	}
	c.log.Info(fmt.Sprintf("ETH: Broadcast transfer signatures for for batchId %v", batch.Id))

	return "", nil
}

func (c *Client) WasProposedTransfer(context.Context, *bridge.Batch) bool {
	return true
}

func (c *Client) GetActionIdForProposeTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
	c.lastTransferBatch = batch
	c.lastSignatureAction = func() {
		c.broadcastSignatureForTransfer(batch)
	}
	return bridge.NewActionId(0)
}

func (c *Client) WasProposedSetStatus(context.Context, *bridge.Batch) bool {
	return true
}

func (c *Client) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
	for _, tx := range batch.Transactions {
		c.lastProposedStatuses = append(c.lastProposedStatuses, tx.Status)
	}
	c.lastSignatureAction = func() {
		c.broadcastSignatureForFinishCurrentPendingTransaction(batch.Id, c.lastProposedStatuses)
	}

	return bridge.NewActionId(0)
}

func (c *Client) WasExecuted(ctx context.Context, _ bridge.ActionId, batchId bridge.BatchId) bool {
	var wasExecuted bool
	var err error = nil

	if c.lastTransferBatch == nil {
		wasExecuted, err = c.bridgeContract.WasBatchFinished(&bind.CallOpts{Context: ctx}, batchId)
	} else {
		wasExecuted, err = c.bridgeContract.WasBatchExecuted(&bind.CallOpts{Context: ctx}, c.lastTransferBatch.Id)
	}
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	c.cleanState(wasExecuted)

	if wasExecuted {
		c.log.Info(fmt.Sprintf("ETH: BatchID %v was executed", batchId))
	} else {
		c.log.Info(fmt.Sprintf("ETH: BatchID %v was not executed", batchId))
	}

	return wasExecuted
}

func (c *Client) Sign(context.Context, bridge.ActionId) (string, error) {
	c.lastSignatureAction()
	return "", nil
}

func (c *Client) Execute(ctx context.Context, _ bridge.ActionId, batch *bridge.Batch) (string, error) {
	fromAddress := crypto.PubkeyToAddress(*c.publicKey)
	batchId := batch.Id

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
	auth.GasLimit = c.gasLimit
	auth.GasPrice = gasPrice
	auth.Context = ctx

	var transaction *types.Transaction

	signatures := c.broadcaster.Signatures()
	if c.lastTransferBatch == nil {
		transaction, err = c.bridgeContract.FinishCurrentPendingBatch(auth, batchId, c.lastProposedStatuses, signatures)
	} else {
		batch := c.lastTransferBatch
		tokens := c.tokenAddresses(batch.Transactions)
		recipients := recipientsAddresses(batch.Transactions)
		amounts := amounts(batch.Transactions)
		transaction, err = c.bridgeContract.ExecuteTransfer(auth, tokens, recipients, amounts, batchId, signatures)
	}

	if err != nil {
		return "", err
	}

	hash := transaction.Hash().String()
	c.log.Info(fmt.Sprintf("ETH: Executed batchId %v with hash %s", batchId, hash))

	return hash, err
}

func (c *Client) SignersCount(context.Context, bridge.ActionId) uint {
	return uint(len(c.broadcaster.Signatures()))
}

// QuorumProvider

func (c *Client) GetQuorum(ctx context.Context) (*big.Int, error) {
	return c.bridgeContract.Quorum(&bind.CallOpts{Context: ctx})
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

func (c *Client) broadcastSignatureForTransfer(batch *bridge.Batch) {
	arguments, err := transferArgs()
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	pack, err := arguments.Pack(recipientsAddresses(batch.Transactions), c.tokenAddresses(batch.Transactions), amounts(batch.Transactions), new(big.Int).Set(batch.Id), "ExecuteBatchedTransfer")
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

func recipientsAddresses(transactions []*bridge.DepositTransaction) []common.Address {
	var result []common.Address

	for _, tx := range transactions {
		result = append(result, common.HexToAddress(tx.To))
	}

	return result
}

func (c *Client) tokenAddresses(transactions []*bridge.DepositTransaction) []common.Address {
	var result []common.Address

	for _, tx := range transactions {
		tokenAddress := c.getErc20AddressFromTokenId(tx.TokenAddress)
		result = append(result, common.HexToAddress(tokenAddress))
	}

	return result
}

func amounts(transactions []*bridge.DepositTransaction) []*big.Int {
	var result []*big.Int

	for _, tx := range transactions {
		result = append(result, tx.Amount)
	}

	return result
}

func (c *Client) broadcastSignatureForFinishCurrentPendingTransaction(batchId bridge.BatchId, statuses []uint8) {
	arguments, err := finishCurrentPendingTransactionArgs()
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	pack, err := arguments.Pack(new(big.Int).Set(batchId), statuses, "CurrentPendingBatch")
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
	} else {
		c.lastProposedStatuses = []uint8{}
	}
}

// helpers

func transferArgs() (abi.Arguments, error) {
	addressesType, err := abi.NewType("address[]", "", nil)
	if err != nil {
		return nil, err
	}

	uint256ArrayType, err := abi.NewType("uint256[]", "", nil)
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
		abi.Argument{Name: "tokens", Type: addressesType},
		abi.Argument{Name: "recipients", Type: addressesType},
		abi.Argument{Name: "amounts", Type: uint256ArrayType},
		abi.Argument{Name: "nonce", Type: uint256Type},
		abi.Argument{Name: "executeTransfer", Type: stringType},
	}, nil

}

func finishCurrentPendingTransactionArgs() (abi.Arguments, error) {
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	uint8Type, err := abi.NewType("uint8[]", "", nil)
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
