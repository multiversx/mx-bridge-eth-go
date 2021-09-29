package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
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
	MessagePrefix   = "\u0019Ethereum Signed Message:\n32"
	TransferAction  = int64(0)
	SetStatusAction = int64(1)
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

	privateKey   *ecdsa.PrivateKey
	publicKey    *ecdsa.PublicKey
	broadcaster  bridge.Broadcaster
	mapper       bridge.Mapper
	pendingBatch *bridge.Batch

	gasLimit uint64

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
	// Nothing needs to get proposed, simply gather signatures
	c.log.Info("ETH: Broadcast status signatures for for batchId", batch.Id)
}

func (c *Client) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	// Nothing needs to get proposed, simply gather signatures
	c.log.Info("ETH: Broadcast transfer signatures for for batchId", batch.Id)

	return "", nil
}

func (c *Client) WasProposedTransfer(context.Context, *bridge.Batch) bool {
	return true
}

func (c *Client) GetActionIdForProposeTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
	c.pendingBatch = batch

	return bridge.NewActionId(TransferAction)
}

func (c *Client) WasProposedSetStatus(context.Context, *bridge.Batch) bool {
	return true
}

func (c *Client) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
	c.pendingBatch = batch

	return bridge.NewActionId(SetStatusAction)
}

func (c *Client) WasExecuted(ctx context.Context, actionId bridge.ActionId, batchId bridge.BatchId) bool {
	var wasExecuted bool
	var err error = nil

	switch int64FromActionId(actionId) {
	case TransferAction:
		wasExecuted, err = c.bridgeContract.WasBatchExecuted(&bind.CallOpts{Context: ctx}, c.pendingBatch.Id)
	case SetStatusAction:
		wasExecuted, err = c.bridgeContract.WasBatchFinished(&bind.CallOpts{Context: ctx}, batchId)
	}
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	if wasExecuted {
		c.log.Info(fmt.Sprintf("ETH: BatchID %v was executed", batchId))
	} else {
		c.log.Info(fmt.Sprintf("ETH: BatchID %v was not executed", batchId))
	}

	return wasExecuted
}

func (c *Client) Sign(_ context.Context, action bridge.ActionId) (string, error) {
	switch int64FromActionId(action) {
	case TransferAction:
		c.broadcastSignatureForTransfer(c.pendingBatch)
	case SetStatusAction:
		var proposedStatuses []uint8
		for _, tx := range c.pendingBatch.Transactions {
			proposedStatuses = append(proposedStatuses, tx.Status)
		}
		c.broadcastSignatureForFinishCurrentPendingTransaction(c.pendingBatch.Id, proposedStatuses)
	}

	return "", nil
}

func (c *Client) Execute(ctx context.Context, action bridge.ActionId, batch *bridge.Batch) (string, error) {
	fromAddress := crypto.PubkeyToAddress(*c.publicKey)
	batchId := batch.Id

	blockNonce, err := c.blockchainClient.PendingNonceAt(ctx, fromAddress)
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
	auth.Context = ctx

	var transaction *types.Transaction

	signatures := c.broadcaster.Signatures()
	switch int64FromActionId(action) {
	case TransferAction:
		transaction, err = c.transfer(auth, signatures)
	case SetStatusAction:
		transaction, err = c.finish(auth, signatures)
	}

	if err != nil {
		return "", err
	}

	hash := transaction.Hash().String()
	c.log.Info(fmt.Sprintf("ETH: Executed batchId %v with hash %s", batchId, hash))

	return hash, err
}

func (c *Client) transfer(auth *bind.TransactOpts, signatures [][]byte) (*types.Transaction, error) {
	batch := c.pendingBatch
	tokens := c.tokenAddresses(batch.Transactions)
	recipients := recipientsAddresses(batch.Transactions)
	amounts := amounts(batch.Transactions)
	return c.bridgeContract.ExecuteTransfer(auth, tokens, recipients, amounts, batch.Id, signatures)
}

func (c *Client) finish(auth *bind.TransactOpts, signatures [][]byte) (*types.Transaction, error) {
	var proposedStatuses []uint8
	for _, tx := range c.pendingBatch.Transactions {
		proposedStatuses = append(proposedStatuses, tx.Status)
	}
	return c.bridgeContract.FinishCurrentPendingBatch(auth, c.pendingBatch.Id, proposedStatuses, signatures)
}

func (c *Client) SignersCount(context.Context, bridge.ActionId) uint {
	return uint(len(c.broadcaster.Signatures()))
}

// QuorumProvider implementation

func (c *Client) GetQuorum(ctx context.Context) (uint, error) {
	n, err := c.bridgeContract.Quorum(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, err
	}

	if n.Cmp(big.NewInt(math.MaxUint32)) > 0 {
		return 0, errors.New("quorum is not a uint")
	}

	return uint(n.Uint64()), nil
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

func int64FromActionId(actionId bridge.ActionId) int64 {
	return (*big.Int)(actionId).Int64()
}
