package eth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	messagePrefix   = "\u0019Ethereum Signed Message:\n32"
	transferAction  = int64(0)
	setStatusAction = int64(1)
	addressLength   = 32
)

// BridgeContract defines the supported Ethereum contract operations
type BridgeContract interface {
	GetNextPendingBatch(opts *bind.CallOpts) (contract.Batch, error)
	FinishCurrentPendingBatch(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error)
	ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	WasBatchExecuted(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	WasBatchFinished(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	Quorum(opts *bind.CallOpts) (*big.Int, error)
	GetStatusesAfterExecution(opts *bind.CallOpts, batchNonceElrondETH *big.Int) ([]uint8, error)
	GetRelayers(opts *bind.CallOpts) ([]common.Address, error)
}

// BlockchainClient defines the RPC operations on the Ethereum node
type BlockchainClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainID(ctx context.Context) (*big.Int, error)
}

type client struct {
	bridgeContract   BridgeContract
	blockchainClient BlockchainClient
	addressConverter core.PubkeyConverter
	privateKey       *ecdsa.PrivateKey
	publicKey        *ecdsa.PublicKey
	broadcaster      bridge.Broadcaster
	mapper           bridge.Mapper
	gasLimit         uint64
	log              logger.Logger
	gasHandler       bridge.GasHandler
}

// ArgsClient is the DTO used in the client constructor
type ArgsClient struct {
	Config      bridge.EthereumConfig
	Broadcaster bridge.Broadcaster
	Mapper      bridge.Mapper
	GasHandler  bridge.GasHandler
	EthClient   BlockchainClient
	EthInstance BridgeContract
}

func NewClient(args ArgsClient) (*client, error) {

	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	log := logger.GetOrCreate("EthClient")

	if check.IfNil(args.GasHandler) {
		return nil, ErrNilGasHandler
	}
	privateKeyBytes, err := ioutil.ReadFile(args.Config.PrivateKeyFile)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.HexToECDSA(string(privateKeyBytes))
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("error casting public key to ECDSA")
	}

	c := &client{
		bridgeContract:   args.EthInstance,
		blockchainClient: args.EthClient,
		gasLimit:         args.Config.GasLimit,
		privateKey:       privateKey,
		publicKey:        publicKeyECDSA,
		broadcaster:      args.Broadcaster,
		mapper:           args.Mapper,
		log:              log,
		gasHandler:       args.GasHandler,
	}
	c.addressConverter, err = pubkeyConverter.NewBech32PubkeyConverter(addressLength, log)
	if err != nil {
		return nil, err
	}

	ethAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	log.Info("Ethereum: NewClient", "address", ethAddress)

	return c, nil
}

func checkArgs(args ArgsClient) error {
	if check.IfNilReflect(args.Config) {
		return ErrNilConfig
	}
	if check.IfNil(args.Broadcaster) {
		return ErrNilBroadcaster
	}
	if check.IfNil(args.Mapper) {
		return ErrNilMapper
	}
	if check.IfNil(args.GasHandler) {
		return ErrNilGasHandler
	}
	if check.IfNilReflect(args.EthClient) {
		return ErrNilBlockchainClient
	}
	if check.IfNilReflect(args.EthInstance) {
		return ErrNilBrdgeContract
	}
	return nil
}

// GetPending returns the pending batch in the Ethereum contract
func (c *client) GetPending(ctx context.Context) *bridge.Batch {
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
				To:            string(deposit.Recipient),
				DisplayableTo: c.addressConverter.Encode(deposit.Recipient),
				From:          deposit.Depositor.String(),
				TokenAddress:  deposit.TokenAddress.String(),
				Amount:        deposit.Amount,
				DepositNonce:  deposit.Nonce,
			}
			c.log.Trace("created deposit transaction: " + tx.String())
			transactions = append(transactions, tx)
		}

		result = &bridge.Batch{
			Id:           batch.Nonce,
			Transactions: transactions,
		}
	}

	return result
}

// ProposeSetStatus will propose the status of an executed batch of transactions
func (c *client) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	// Nothing needs to get proposed, simply gather signatures
	c.log.Info("ETH: Broadcast status signatures for for batchId", batch.Id)
}

// ProposeTransfer will propose the transfer coming from other client
func (c *client) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	// Nothing needs to get proposed, simply gather signatures
	c.log.Info("ETH: Broadcast transfer signatures for for batchId", batch.Id)

	return "", nil
}

// WasProposedTransfer returns true
func (c *client) WasProposedTransfer(context.Context, *bridge.Batch) bool {
	return true
}

// GetActionIdForProposeTransfer returns a hardcoded value fot the transfer action ID
func (c *client) GetActionIdForProposeTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionId {
	return bridge.NewActionId(transferAction)
}

// WasProposedSetStatus returns true
func (c *client) WasProposedSetStatus(context.Context, *bridge.Batch) bool {
	return true
}

// GetActionIdForSetStatusOnPendingTransfer a hardcoded value fot the set status action ID
func (c *client) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, _ *bridge.Batch) bridge.ActionId {
	return bridge.NewActionId(setStatusAction)
}

// GetRelayers returns the current registered relayers from the Ethereum SC
func (c *client) GetRelayers(ctx context.Context) ([]common.Address, error) {
	return c.bridgeContract.GetRelayers(&bind.CallOpts{Context: ctx})
}

// WasExecuted returns true if the action ID was executed
func (c *client) WasExecuted(ctx context.Context, actionId bridge.ActionId, batchId bridge.BatchId) bool {
	var wasExecuted bool
	var err error = nil

	switch int64FromActionId(actionId) {
	case transferAction:
		wasExecuted, err = c.bridgeContract.WasBatchExecuted(&bind.CallOpts{Context: ctx}, batchId)
	case setStatusAction:
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

// GetTransactionsStatuses will return the transactions statuses from the batch ID
func (c *client) GetTransactionsStatuses(ctx context.Context, batchId bridge.BatchId) ([]uint8, error) {
	return c.bridgeContract.GetStatusesAfterExecution(&bind.CallOpts{Context: ctx}, batchId)
}

// Sign will sign upon the provided batch and send the signatures through the broadcaster to other relayers
func (c *client) Sign(_ context.Context, action bridge.ActionId, batch *bridge.Batch) (string, error) {
	switch int64FromActionId(action) {
	case transferAction:
		c.broadcastSignatureForTransfer(batch)
	case setStatusAction:
		c.broadcastSignatureForFinish(batch)
	}

	return "", nil
}

// Execute will pack and send a transaction providing the batch data and received signatures from the other relayers
func (c *client) Execute(ctx context.Context, action bridge.ActionId, batch *bridge.Batch) (string, error) {
	fromAddress := crypto.PubkeyToAddress(*c.publicKey)
	batchId := batch.Id

	nonce, err := c.getNonce(ctx, fromAddress)
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

	gasPrice, err := c.gasHandler.GetCurrentGasPrice()
	if err != nil {
		return "", err
	}

	auth.Nonce = big.NewInt(nonce)
	auth.Value = big.NewInt(0)
	auth.GasLimit = c.gasLimit
	auth.Context = ctx
	auth.GasPrice = gasPrice

	var transaction *types.Transaction

	hash, err := c.generateHash(batch, action)
	if err != nil {
		return "", fmt.Errorf("ETH: %w", err)
	}

	signatures := c.broadcaster.Signatures(hash.Bytes())
	// TODO optimize this: no need to re-fetch the quorum, can be provided by the bridge executor
	quorum, err := c.GetQuorum(ctx)
	if err != nil {
		return "", fmt.Errorf("%w while getting the quorum in client.Execute", err)
	}
	if len(signatures) > int(quorum) {
		c.log.Debug("reducing the size of the signatures set",
			"quorum", quorum, "total signatures", len(signatures))
		signatures = signatures[:quorum]
	}

	switch int64FromActionId(action) {
	case transferAction:
		transaction, err = c.transfer(auth, signatures, batch)
	case setStatusAction:
		transaction, err = c.finish(auth, signatures, batch)
	}

	if err != nil {
		return "", fmt.Errorf("ETH: %w", err)
	}

	txHash := transaction.Hash().String()
	c.log.Info(fmt.Sprintf("ETH: Executed batchId %v with hash %s", batchId, txHash))

	return txHash, err
}

func (c *client) getNonce(ctx context.Context, fromAddress common.Address) (int64, error) {
	blockNonce, err := c.blockchainClient.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w in getNonce, BlockNumber call", err)
	}

	nonce, err := c.blockchainClient.NonceAt(ctx, fromAddress, big.NewInt(int64(blockNonce)))

	return int64(nonce), err
}

func (c *client) transfer(auth *bind.TransactOpts, signatures [][]byte, batch *bridge.Batch) (*types.Transaction, error) {
	tokens := c.tokenAddresses(batch.Transactions)
	recipients := recipientsAddresses(batch.Transactions)
	amountsValues := amounts(batch.Transactions)

	c.log.Debug("client.transfer", "auth", transactOptsToString(auth),
		"batchId", batch.Id, "tokens", tokens, "recipients", recipients, "amounts", amountsValues,
		"num signatures", len(signatures))

	return c.bridgeContract.ExecuteTransfer(auth, tokens, recipients, amountsValues, batch.Id, signatures)
}

func (c *client) finish(auth *bind.TransactOpts, signatures [][]byte, batch *bridge.Batch) (*types.Transaction, error) {
	var proposedStatuses []uint8
	for _, tx := range batch.Transactions {
		proposedStatuses = append(proposedStatuses, tx.Status)
	}

	c.log.Debug("client.finish", "auth", transactOptsToString(auth),
		"batchId", batch.Id, "proposed statuses", proposedStatuses, "num signatures", len(signatures))

	return c.bridgeContract.FinishCurrentPendingBatch(auth, batch.Id, proposedStatuses, signatures)
}

// SignersCount will return the total signers number that sent the signatures on the required message hash
func (c *client) SignersCount(_ context.Context, batch *bridge.Batch, actionId bridge.ActionId) uint {
	hash, err := c.generateHash(batch, actionId)
	if err != nil {
		c.log.Error(err.Error())

		return 0
	}

	return uint(len(c.broadcaster.Signatures(hash.Bytes())))
}

// QuorumProvider implementation

// GetQuorum returns the Quorum value from the Ethereum SC
func (c *client) GetQuorum(ctx context.Context) (uint, error) {
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

func (c *client) signHash(hash common.Hash) ([]byte, error) {
	valueToSign := crypto.Keccak256Hash(append([]byte(messagePrefix), hash.Bytes()...))
	signature, err := crypto.Sign(valueToSign.Bytes(), c.privateKey)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (c *client) broadcastSignatureForTransfer(batch *bridge.Batch) {
	hash, err := c.generateHashForTransfer(batch)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	signature, err := c.signHash(hash)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	c.broadcaster.SendSignature(signature, hash.Bytes())
}

func recipientsAddresses(transactions []*bridge.DepositTransaction) []common.Address {
	var result []common.Address

	for _, tx := range transactions {
		result = append(result, common.HexToAddress(tx.To))
	}

	return result
}

func (c *client) tokenAddresses(transactions []*bridge.DepositTransaction) []common.Address {
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

func (c *client) generateHash(batch *bridge.Batch, actionId bridge.ActionId) (common.Hash, error) {
	switch int64FromActionId(actionId) {
	case transferAction:
		return c.generateHashForTransfer(batch)
	case setStatusAction:
		return c.generateHashForFinish(batch)
	}

	return common.Hash{}, fmt.Errorf("Client.generateHash not implemented for action ID %v", actionId)
}

func (c *client) generateHashForTransfer(batch *bridge.Batch) (common.Hash, error) {
	arguments, err := transferArgs()
	if err != nil {
		return common.Hash{}, err
	}

	pack, err := arguments.Pack(recipientsAddresses(batch.Transactions), c.tokenAddresses(batch.Transactions), amounts(batch.Transactions), new(big.Int).Set(batch.Id), "ExecuteBatchedTransfer")
	if err != nil {
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(pack), nil
}

func (c *client) generateHashForFinish(batch *bridge.Batch) (common.Hash, error) {
	var statuses []uint8
	for _, tx := range batch.Transactions {
		statuses = append(statuses, tx.Status)
	}

	arguments, err := finishCurrentPendingTransactionArgs()
	if err != nil {
		return common.Hash{}, err
	}

	pack, err := arguments.Pack(new(big.Int).Set(batch.Id), statuses, "CurrentPendingBatch")
	if err != nil {
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(pack), nil
}

func (c *client) broadcastSignatureForFinish(batch *bridge.Batch) {
	hash, err := c.generateHashForFinish(batch)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	signature, err := c.signHash(hash)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	c.broadcaster.SendSignature(signature, hash.Bytes())
}

func (c *client) getErc20AddressFromTokenId(tokenId string) string {
	return c.mapper.GetErc20Address(tokenId[2:])
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
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

func transactOptsToString(opts *bind.TransactOpts) string {
	if opts == nil {
		return "<nil>"
	}

	return fmt.Sprintf("from: %v, nonce: %v, value: %v, gas price: %v, gas limit: %v",
		opts.From,
		opts.Nonce,
		opts.Value,
		opts.GasPrice,
		opts.GasLimit,
	)
}
