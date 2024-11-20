package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const (
	messagePrefix                   = "\u0019Ethereum Signed Message:\n32"
	minQuorumValue                  = uint64(1)
	minClientAvailabilityAllowDelta = 1
)

// ArgsEthereumClient is the DTO used in the ethereum's client constructor
type ArgsEthereumClient struct {
	ClientWrapper                ClientWrapper
	Erc20ContractsHandler        Erc20ContractsHolder
	Log                          chainCore.Logger
	AddressConverter             core.AddressConverter
	Broadcaster                  Broadcaster
	CryptoHandler                CryptoHandler
	TokensMapper                 TokensMapper
	SignatureHolder              SignaturesHolder
	SafeContractAddress          common.Address
	GasHandler                   GasHandler
	TransferGasLimitBase         uint64
	TransferGasLimitForEach      uint64
	ClientAvailabilityAllowDelta uint64
	EventsBlockRangeFrom         int64
	EventsBlockRangeTo           int64
}

type client struct {
	clientWrapper                ClientWrapper
	erc20ContractsHandler        Erc20ContractsHolder
	log                          chainCore.Logger
	addressConverter             core.AddressConverter
	broadcaster                  Broadcaster
	cryptoHandler                CryptoHandler
	tokensMapper                 TokensMapper
	signatureHolder              SignaturesHolder
	safeContractAddress          common.Address
	gasHandler                   GasHandler
	transferGasLimitBase         uint64
	transferGasLimitForEach      uint64
	clientAvailabilityAllowDelta uint64
	eventsBlockRangeFrom         int64
	eventsBlockRangeTo           int64

	lastBlockNumber          uint64
	retriesAvailabilityCheck uint64
	mut                      sync.RWMutex
}

// NewEthereumClient will create a new Ethereum client
func NewEthereumClient(args ArgsEthereumClient) (*client, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	c := &client{
		clientWrapper:                args.ClientWrapper,
		erc20ContractsHandler:        args.Erc20ContractsHandler,
		log:                          args.Log,
		addressConverter:             args.AddressConverter,
		broadcaster:                  args.Broadcaster,
		cryptoHandler:                args.CryptoHandler,
		tokensMapper:                 args.TokensMapper,
		signatureHolder:              args.SignatureHolder,
		safeContractAddress:          args.SafeContractAddress,
		gasHandler:                   args.GasHandler,
		transferGasLimitBase:         args.TransferGasLimitBase,
		transferGasLimitForEach:      args.TransferGasLimitForEach,
		clientAvailabilityAllowDelta: args.ClientAvailabilityAllowDelta,
		eventsBlockRangeFrom:         args.EventsBlockRangeFrom,
		eventsBlockRangeTo:           args.EventsBlockRangeTo,
	}

	c.log.Info("NewEthereumClient",
		"relayer address", c.cryptoHandler.GetAddress(),
		"safe contract address", c.safeContractAddress.String())

	return c, err
}

func checkArgs(args ArgsEthereumClient) error {
	if check.IfNil(args.ClientWrapper) {
		return errNilClientWrapper
	}
	if check.IfNil(args.Erc20ContractsHandler) {
		return errNilERC20ContractsHandler
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}
	if check.IfNil(args.AddressConverter) {
		return clients.ErrNilAddressConverter
	}
	if check.IfNil(args.Broadcaster) {
		return errNilBroadcaster
	}
	if check.IfNil(args.CryptoHandler) {
		return clients.ErrNilCryptoHandler
	}
	if check.IfNil(args.TokensMapper) {
		return clients.ErrNilTokensMapper
	}
	if check.IfNil(args.SignatureHolder) {
		return errNilSignaturesHolder
	}
	if check.IfNil(args.GasHandler) {
		return errNilGasHandler
	}
	if args.TransferGasLimitBase == 0 {
		return errInvalidGasLimit
	}
	if args.TransferGasLimitForEach == 0 {
		return errInvalidGasLimit
	}
	if args.ClientAvailabilityAllowDelta < minClientAvailabilityAllowDelta {
		return fmt.Errorf("%w for args.AllowedDelta, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.ClientAvailabilityAllowDelta, minClientAvailabilityAllowDelta)
	}
	if args.EventsBlockRangeFrom > args.EventsBlockRangeTo {
		return fmt.Errorf("%w, args.EventsBlockRangeFrom: %d, args.EventsBlockRangeTo: %d",
			clients.ErrInvalidValue, args.EventsBlockRangeFrom, args.EventsBlockRangeTo)
	}
	return nil
}

// GetBatch returns the batch (if existing) from the Ethereum contract by providing the nonce
func (c *client) GetBatch(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
	c.log.Info("Getting batch", "nonce", nonce)
	nonceAsBigInt := big.NewInt(0).SetUint64(nonce)
	batch, isFinalBatch, err := c.clientWrapper.GetBatch(ctx, nonceAsBigInt)
	if err != nil {
		return nil, false, err
	}
	deposits, areFinalDeposits, err := c.clientWrapper.GetBatchDeposits(ctx, nonceAsBigInt)
	if err != nil {
		return nil, false, err
	}
	if int(batch.DepositsCount) != len(deposits) {
		return nil, false, fmt.Errorf("%w, batch.DepositsCount: %d, fetched deposits len: %d",
			errDepositsAndBatchDepositsCountDiffer, batch.DepositsCount, len(deposits))
	}

	transferBatch := &bridgeCore.TransferBatch{
		ID:          batch.Nonce.Uint64(),
		BlockNumber: batch.BlockNumber,
		Deposits:    make([]*bridgeCore.DepositTransfer, 0, batch.DepositsCount),
	}
	cachedTokens := make(map[string][]byte)
	for i := range deposits {
		deposit := deposits[i]
		toBytes := deposit.Recipient[:]
		fromBytes := deposit.Depositor[:]
		tokenBytes := deposit.TokenAddress[:]

		depositTransfer := &bridgeCore.DepositTransfer{
			Nonce:            deposit.Nonce.Uint64(),
			ToBytes:          toBytes,
			DisplayableTo:    c.addressConverter.ToBech32StringSilent(toBytes),
			FromBytes:        fromBytes,
			DisplayableFrom:  c.addressConverter.ToHexString(fromBytes),
			SourceTokenBytes: tokenBytes,
			DisplayableToken: c.addressConverter.ToHexString(tokenBytes),
			Amount:           big.NewInt(0).Set(deposit.Amount),
		}
		storedConvertedTokenBytes, exists := cachedTokens[depositTransfer.DisplayableToken]
		if !exists {
			depositTransfer.DestinationTokenBytes, err = c.tokensMapper.ConvertToken(ctx, depositTransfer.SourceTokenBytes)
			if err != nil {
				return nil, false, err
			}
			cachedTokens[depositTransfer.DisplayableToken] = depositTransfer.DestinationTokenBytes
		} else {
			depositTransfer.DestinationTokenBytes = storedConvertedTokenBytes
		}

		transferBatch.Deposits = append(transferBatch.Deposits, depositTransfer)
	}

	transferBatch.Statuses = make([]byte, len(transferBatch.Deposits))

	return transferBatch, isFinalBatch && areFinalDeposits, nil
}

// GetBatchSCMetadata returns the emitted logs in a batch that hold metadata for SC execution on MVX
func (c *client) GetBatchSCMetadata(ctx context.Context, nonce uint64, blockNumber int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
	scExecAbi, err := contract.ERC20SafeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{c.safeContractAddress},
		Topics: [][]common.Hash{
			{scExecAbi.Events["ERC20SCDeposit"].ID},
			{common.BytesToHash(new(big.Int).SetUint64(nonce).Bytes())},
		},
		FromBlock: big.NewInt(blockNumber + c.eventsBlockRangeFrom),
		ToBlock:   big.NewInt(blockNumber + c.eventsBlockRangeTo),
	}

	logs, err := c.clientWrapper.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	depositEvents := make([]*contract.ERC20SafeERC20SCDeposit, 0)
	for _, vLog := range logs {
		event := new(contract.ERC20SafeERC20SCDeposit)
		err = scExecAbi.UnpackIntoInterface(event, "ERC20SCDeposit", vLog.Data)
		if err != nil {
			return nil, err
		}

		// Add this manually since UnpackIntoInterface only unpacks non-indexed arguments
		event.BatchId = big.NewInt(0).SetUint64(nonce)
		depositEvents = append(depositEvents, event)
	}

	return depositEvents, nil
}

// WasExecuted returns true if the batch ID was executed
func (c *client) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
	return c.clientWrapper.WasBatchExecuted(ctx, big.NewInt(0).SetUint64(batchID))
}

// BroadcastSignatureForMessageHash will send the signature for the provided message hash
func (c *client) BroadcastSignatureForMessageHash(msgHash common.Hash) {
	signature, err := c.cryptoHandler.Sign(msgHash)
	if err != nil {
		c.log.Error("error generating signature", "msh hash", msgHash, "error", err)
		return
	}

	c.broadcaster.BroadcastSignature(signature, msgHash.Bytes())
}

// GenerateMessageHash will generate the message hash based on the provided batch
func (c *client) GenerateMessageHash(batch *batchProcessor.ArgListsBatch, batchId uint64) (common.Hash, error) {
	return GenerateMessageHash(batch, batchId)
}

// GenerateMessageHash will generate the message hash based on the provided batch
func GenerateMessageHash(batch *batchProcessor.ArgListsBatch, batchId uint64) (common.Hash, error) {
	if batch == nil {
		return common.Hash{}, clients.ErrNilBatch
	}

	args, err := generateTransferArgs()
	if err != nil {
		return common.Hash{}, err
	}

	pack, err := args.Pack(batch.Recipients, batch.EthTokens, batch.Amounts, batch.Nonces, big.NewInt(0).SetUint64(batchId), "ExecuteBatchedTransfer")
	if err != nil {
		return common.Hash{}, err
	}

	hash := crypto.Keccak256Hash(pack)
	return crypto.Keccak256Hash(append([]byte(messagePrefix), hash.Bytes()...)), nil
}

func generateTransferArgs() (abi.Arguments, error) {
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
		abi.Argument{Name: "recipients", Type: addressesType},
		abi.Argument{Name: "tokens", Type: addressesType},
		abi.Argument{Name: "amounts", Type: uint256ArrayType},
		abi.Argument{Name: "nonces", Type: uint256ArrayType},
		abi.Argument{Name: "nonce", Type: uint256Type},
		abi.Argument{Name: "executeTransfer", Type: stringType},
	}, nil
}

// ExecuteTransfer will initiate and send the transaction from the transfer batch struct
func (c *client) ExecuteTransfer(
	ctx context.Context,
	msgHash common.Hash,
	argLists *batchProcessor.ArgListsBatch,
	batchId uint64,
	quorum int,
) (string, error) {
	if argLists == nil {
		return "", clients.ErrNilBatch
	}

	isPaused, err := c.clientWrapper.IsPaused(ctx)
	if err != nil {
		return "", fmt.Errorf("%w in client.ExecuteTransfer", err)
	}
	if isPaused {
		return "", fmt.Errorf("%w in client.ExecuteTransfer", clients.ErrMultisigContractPaused)
	}

	nonce, err := c.getNonce(ctx, c.cryptoHandler.GetAddress())
	if err != nil {
		return "", err
	}

	chainId, err := c.clientWrapper.ChainID(ctx)
	if err != nil {
		return "", err
	}

	auth, err := c.cryptoHandler.CreateKeyedTransactor(chainId)
	if err != nil {
		return "", err
	}

	gasPrice, err := c.gasHandler.GetCurrentGasPrice()
	if err != nil {
		return "", err
	}

	auth.Nonce = big.NewInt(nonce)
	auth.Value = big.NewInt(0)
	auth.GasLimit = c.transferGasLimitBase + uint64(len(argLists.EthTokens))*c.transferGasLimitForEach
	auth.Context = ctx
	auth.GasPrice = gasPrice

	signatures := c.signatureHolder.Signatures(msgHash.Bytes())
	if len(signatures) < quorum {
		return "", fmt.Errorf("%w num signatures: %d, quorum: %d", errQuorumNotReached, len(signatures), quorum)
	}
	if len(signatures) > quorum {
		c.log.Debug("reducing the size of the signatures set",
			"quorum", quorum, "total signatures", len(signatures))
		signatures = signatures[:quorum]
	}

	minimumForFee := big.NewInt(int64(auth.GasLimit))
	minimumForFee.Mul(minimumForFee, auth.GasPrice)
	err = c.checkRelayerFundsForFee(ctx, minimumForFee)
	if err != nil {
		return "", err
	}

	batchID := big.NewInt(0).SetUint64(batchId)

	mvxTransactions, err := convertArgsListBatchToMvxTransactions(argLists)
	if err != nil {
		return "", err
	}

	tx, err := c.clientWrapper.ExecuteTransfer(auth, mvxTransactions, batchID, signatures)
	if err != nil {
		return "", err
	}

	txHash := tx.Hash().String()
	c.log.Info("Executed transfer transaction", "batchID", batchID, "hash", txHash)

	return txHash, err
}

func convertArgsListBatchToMvxTransactions(argLists *batchProcessor.ArgListsBatch) ([]contract.MvxTransaction, error) {
	numTokens := len(argLists.EthTokens)
	if len(argLists.Recipients) != numTokens {
		return nil, fmt.Errorf("%w for argLists.Recipients", errInternalErrorValidatingLists)
	}
	if len(argLists.Amounts) != numTokens {
		return nil, fmt.Errorf("%w for argLists.Amounts", errInternalErrorValidatingLists)
	}
	if len(argLists.Nonces) != numTokens {
		return nil, fmt.Errorf("%w for argLists.Nonces", errInternalErrorValidatingLists)
	}
	if len(argLists.Senders) != numTokens {
		return nil, fmt.Errorf("%w for argLists.Senders", errInternalErrorValidatingLists)
	}
	if len(argLists.ScCalls) != numTokens {
		return nil, fmt.Errorf("%w for argLists.ScCalls", errInternalErrorValidatingLists)
	}

	mvxTransactions := make([]contract.MvxTransaction, numTokens)
	for i := range argLists.MvxTokenBytes {
		mvxTransactions[i] = contract.MvxTransaction{
			Token:        argLists.EthTokens[i],
			Sender:       argLists.Senders[i],
			Recipient:    argLists.Recipients[i],
			Amount:       argLists.Amounts[i],
			DepositNonce: argLists.Nonces[i],
			CallData:     argLists.ScCalls[i],
		}
	}

	return mvxTransactions, nil
}

// CheckClientAvailability will check the client availability and set the metric accordingly
func (c *client) CheckClientAvailability(ctx context.Context) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	currentBlock, err := c.clientWrapper.BlockNumber(ctx)
	if err != nil {
		c.setStatusForAvailabilityCheck(bridgeCore.Unavailable, err.Error(), currentBlock)

		return err
	}

	if currentBlock != c.lastBlockNumber {
		c.retriesAvailabilityCheck = 0
		c.lastBlockNumber = currentBlock
	}

	// if we reached this point we will need to increment the retries counter
	defer c.incrementRetriesAvailabilityCheck()

	if c.retriesAvailabilityCheck > c.clientAvailabilityAllowDelta {
		message := fmt.Sprintf("block %d fetched for %d times in a row", currentBlock, c.retriesAvailabilityCheck)
		c.setStatusForAvailabilityCheck(bridgeCore.Unavailable, message, currentBlock)

		return nil
	}

	c.setStatusForAvailabilityCheck(bridgeCore.Available, "", currentBlock)

	return nil
}

func (c *client) incrementRetriesAvailabilityCheck() {
	c.retriesAvailabilityCheck++
}

func (c *client) setStatusForAvailabilityCheck(status bridgeCore.ClientStatus, message string, nonce uint64) {
	c.clientWrapper.SetStringMetric(core.MetricMultiversXClientStatus, status.String())
	c.clientWrapper.SetStringMetric(core.MetricLastMultiversXClientError, message)
	c.clientWrapper.SetIntMetric(core.MetricLastBlockNonce, int(nonce))
}

// CheckRequiredBalance will check if the safe has enough balance for the transfer
func (c *client) CheckRequiredBalance(ctx context.Context, erc20Address common.Address, value *big.Int) error {
	isMintBurn, err := c.MintBurnTokens(ctx, erc20Address)
	if err != nil {
		return err
	}

	if isMintBurn {
		return nil
	}

	existingBalance, err := c.erc20ContractsHandler.BalanceOf(ctx, erc20Address, c.safeContractAddress)
	if err != nil {
		return fmt.Errorf("%w for address %s for ERC20 token %s", err, c.safeContractAddress.String(), erc20Address.String())
	}

	if value.Cmp(existingBalance) > 0 {
		return fmt.Errorf("%w, existing: %s, required: %s for ERC20 token %s and address %s",
			errInsufficientErc20Balance, existingBalance.String(), value.String(), erc20Address.String(), c.safeContractAddress.String())
	}

	c.log.Debug("checked ERC20 balance",
		"ERC20 token", erc20Address.String(),
		"address", c.safeContractAddress.String(),
		"existing balance", existingBalance.String(),
		"needed", value.String())

	return nil
}

// TotalBalances returns the total balance of the given token
func (c *client) TotalBalances(ctx context.Context, token common.Address) (*big.Int, error) {
	return c.clientWrapper.TotalBalances(ctx, token)
}

// MintBalances returns the mint balance of the given token
func (c *client) MintBalances(ctx context.Context, token common.Address) (*big.Int, error) {
	return c.clientWrapper.MintBalances(ctx, token)
}

// BurnBalances returns the burn balance of the given token
func (c *client) BurnBalances(ctx context.Context, token common.Address) (*big.Int, error) {
	return c.clientWrapper.BurnBalances(ctx, token)
}

// MintBurnTokens returns true if the token is mintBurn token
func (c *client) MintBurnTokens(ctx context.Context, token common.Address) (bool, error) {
	return c.clientWrapper.MintBurnTokens(ctx, token)
}

// NativeTokens returns true if the token is native
func (c *client) NativeTokens(ctx context.Context, token common.Address) (bool, error) {
	return c.clientWrapper.NativeTokens(ctx, token)
}

// WhitelistedTokens returns true if the token is whitelisted
func (c *client) WhitelistedTokens(ctx context.Context, token common.Address) (bool, error) {
	return c.clientWrapper.WhitelistedTokens(ctx, token)
}

func (c *client) checkRelayerFundsForFee(ctx context.Context, transferFee *big.Int) error {
	existingBalance, err := c.clientWrapper.BalanceAt(ctx, c.cryptoHandler.GetAddress(), nil)
	if err != nil {
		return err
	}

	if transferFee.Cmp(existingBalance) > 0 {
		return fmt.Errorf("%w, existing: %s, required: %s",
			errInsufficientBalance, existingBalance.String(), transferFee.String())
	}

	c.log.Debug("checked balance",
		"existing balance", existingBalance.String(),
		"needed", transferFee.String())

	return nil
}

func (c *client) getNonce(ctx context.Context, fromAddress common.Address) (int64, error) {
	blockNonce, err := c.clientWrapper.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w in getNonce, BlockNumber call", err)
	}

	nonce, err := c.clientWrapper.NonceAt(ctx, fromAddress, big.NewInt(int64(blockNonce)))

	return int64(nonce), err
}

// GetTransactionsStatuses will return the transactions statuses from the batch
func (c *client) GetTransactionsStatuses(ctx context.Context, batchId uint64) ([]byte, error) {
	buff, isFinal, err := c.clientWrapper.GetStatusesAfterExecution(ctx, big.NewInt(0).SetUint64(batchId))
	if err != nil {
		return nil, err
	}
	if !isFinal {
		return nil, errStatusIsNotFinal
	}

	return buff, nil
}

// GetQuorumSize returns the size of the quorum
func (c *client) GetQuorumSize(ctx context.Context) (*big.Int, error) {
	return c.clientWrapper.Quorum(ctx)
}

// IsQuorumReached returns true if the number of signatures is at least the size of quorum
func (c *client) IsQuorumReached(ctx context.Context, msgHash common.Hash) (bool, error) {
	signatures := c.signatureHolder.Signatures(msgHash.Bytes())
	quorum, err := c.clientWrapper.Quorum(ctx)
	if err != nil {
		return false, fmt.Errorf("%w in IsQuorumReached, Quorum call", err)
	}
	if quorum.Uint64() < minQuorumValue {
		return false, fmt.Errorf("%w in IsQuorumReached, minQuorum %d, got: %s", clients.ErrInvalidValue, minQuorumValue, quorum.String())
	}

	return len(signatures) >= int(quorum.Int64()), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}
