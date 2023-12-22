package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/core"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const (
	messagePrefix   = "\u0019Ethereum Signed Message:\n32"
	minQuorumValue  = uint64(1)
	minAllowedDelta = 1
)

type argListsBatch struct {
	tokens     []common.Address
	recipients []common.Address
	amounts    []*big.Int
	nonces     []*big.Int
}

// ArgsEthereumClient is the DTO used in the ethereum's client constructor
type ArgsEthereumClient struct {
	ClientWrapper           ClientWrapper
	Erc20ContractsHandler   Erc20ContractsHolder
	Log                     chainCore.Logger
	AddressConverter        core.AddressConverter
	Broadcaster             Broadcaster
	PrivateKey              *ecdsa.PrivateKey
	TokensMapper            TokensMapper
	SignatureHolder         SignaturesHolder
	SafeContractAddress     common.Address
	GasHandler              GasHandler
	TransferGasLimitBase    uint64
	TransferGasLimitForEach uint64
	AllowDelta              uint64
}

type client struct {
	clientWrapper           ClientWrapper
	erc20ContractsHandler   Erc20ContractsHolder
	log                     chainCore.Logger
	addressConverter        core.AddressConverter
	broadcaster             Broadcaster
	privateKey              *ecdsa.PrivateKey
	publicKey               *ecdsa.PublicKey
	tokensMapper            TokensMapper
	signatureHolder         SignaturesHolder
	safeContractAddress     common.Address
	gasHandler              GasHandler
	transferGasLimitBase    uint64
	transferGasLimitForEach uint64
	allowDelta              uint64

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

	publicKey := args.PrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errPublicKeyCast
	}

	c := &client{
		clientWrapper:           args.ClientWrapper,
		erc20ContractsHandler:   args.Erc20ContractsHandler,
		log:                     args.Log,
		addressConverter:        args.AddressConverter,
		broadcaster:             args.Broadcaster,
		privateKey:              args.PrivateKey,
		publicKey:               publicKeyECDSA,
		tokensMapper:            args.TokensMapper,
		signatureHolder:         args.SignatureHolder,
		safeContractAddress:     args.SafeContractAddress,
		gasHandler:              args.GasHandler,
		transferGasLimitBase:    args.TransferGasLimitBase,
		transferGasLimitForEach: args.TransferGasLimitForEach,
		allowDelta:              args.AllowDelta,
	}

	c.log.Info("NewEthereumClient",
		"relayer address", crypto.PubkeyToAddress(*publicKeyECDSA),
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
	if args.PrivateKey == nil {
		return clients.ErrNilPrivateKey
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
	if args.AllowDelta < minAllowedDelta {
		return fmt.Errorf("%w for args.AllowedDelta, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.AllowDelta, minAllowedDelta)
	}
	return nil
}

// GetBatch returns the batch (if existing) from the Ethereum contract by providing the nonce
func (c *client) GetBatch(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
	c.log.Info("Getting batch", "nonce", nonce)
	nonceAsBigInt := big.NewInt(0).SetUint64(nonce)
	batch, err := c.clientWrapper.GetBatch(ctx, nonceAsBigInt)
	if err != nil {
		return nil, err
	}
	deposits, err := c.clientWrapper.GetBatchDeposits(ctx, nonceAsBigInt)
	if err != nil {
		return nil, err
	}
	if int(batch.DepositsCount) != len(deposits) {
		return nil, fmt.Errorf("%w, batch.DepositsCount: %d, fetched deposits len: %d",
			errDepositsAndBatchDepositsCountDiffer, batch.DepositsCount, len(deposits))
	}

	transferBatch := &clients.TransferBatch{
		ID:       batch.Nonce.Uint64(),
		Deposits: make([]*clients.DepositTransfer, 0, batch.DepositsCount),
	}
	cachedTokens := make(map[string][]byte)
	for i := range deposits {
		deposit := deposits[i]
		toBytes := deposit.Recipient[:]
		fromBytes := deposit.Depositor[:]
		tokenBytes := deposit.TokenAddress[:]

		depositTransfer := &clients.DepositTransfer{
			Nonce:            deposit.Nonce.Uint64(),
			ToBytes:          toBytes,
			DisplayableTo:    c.addressConverter.ToBech32StringSilent(toBytes),
			FromBytes:        fromBytes,
			DisplayableFrom:  c.addressConverter.ToHexString(fromBytes),
			TokenBytes:       tokenBytes,
			DisplayableToken: c.addressConverter.ToHexString(tokenBytes),
			Amount:           big.NewInt(0).Set(deposit.Amount),
		}
		storedConvertedTokenBytes, exists := cachedTokens[depositTransfer.DisplayableToken]
		if !exists {
			depositTransfer.ConvertedTokenBytes, err = c.tokensMapper.ConvertToken(ctx, depositTransfer.TokenBytes)
			if err != nil {
				return nil, err
			}
			cachedTokens[depositTransfer.DisplayableToken] = depositTransfer.ConvertedTokenBytes
		} else {
			depositTransfer.ConvertedTokenBytes = storedConvertedTokenBytes
		}

		transferBatch.Deposits = append(transferBatch.Deposits, depositTransfer)
	}

	transferBatch.Statuses = make([]byte, len(transferBatch.Deposits))

	return transferBatch, nil
}

// WasExecuted returns true if the batch ID was executed
func (c *client) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
	return c.clientWrapper.WasBatchExecuted(ctx, big.NewInt(0).SetUint64(batchID))
}

// BroadcastSignatureForMessageHash will send the signature for the provided message hash
func (c *client) BroadcastSignatureForMessageHash(msgHash common.Hash) {
	signature, err := crypto.Sign(msgHash.Bytes(), c.privateKey)
	if err != nil {
		c.log.Error("error generating signature", "msh hash", msgHash, "error", err)
		return
	}

	c.broadcaster.BroadcastSignature(signature, msgHash.Bytes())
}

// GenerateMessageHash will generate the message hash based on the provided batch
func (c *client) GenerateMessageHash(batch *clients.TransferBatch) (common.Hash, error) {
	if batch == nil {
		return common.Hash{}, clients.ErrNilBatch
	}

	args, err := generateTransferArgs()
	if err != nil {
		return common.Hash{}, err
	}

	argLists, err := c.extractList(batch)
	if err != nil {
		return common.Hash{}, err
	}

	pack, err := args.Pack(argLists.recipients, argLists.tokens, argLists.amounts, argLists.nonces, big.NewInt(0).SetUint64(batch.ID), "ExecuteBatchedTransfer")
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

func (c *client) extractList(batch *clients.TransferBatch) (argListsBatch, error) {
	arg := argListsBatch{}

	for _, dt := range batch.Deposits {
		recipient := common.BytesToAddress(dt.ToBytes)
		arg.recipients = append(arg.recipients, recipient)

		token := common.BytesToAddress(dt.ConvertedTokenBytes)
		arg.tokens = append(arg.tokens, token)

		amount := big.NewInt(0).Set(dt.Amount)
		arg.amounts = append(arg.amounts, amount)

		nonce := big.NewInt(0).SetUint64(dt.Nonce)
		arg.nonces = append(arg.nonces, nonce)
	}

	return arg, nil
}

// ExecuteTransfer will initiate and send the transaction from the transfer batch struct
func (c *client) ExecuteTransfer(
	ctx context.Context,
	msgHash common.Hash,
	batch *clients.TransferBatch,
	quorum int,
) (string, error) {
	if batch == nil {
		return "", clients.ErrNilBatch
	}

	isPaused, err := c.clientWrapper.IsPaused(ctx)
	if err != nil {
		return "", fmt.Errorf("%w in client.ExecuteTransfer", err)
	}
	if isPaused {
		return "", fmt.Errorf("%w in client.ExecuteTransfer", clients.ErrMultisigContractPaused)
	}

	c.log.Info("executing transfer " + batch.String())

	fromAddress := crypto.PubkeyToAddress(*c.publicKey)

	nonce, err := c.getNonce(ctx, fromAddress)
	if err != nil {
		return "", err
	}

	chainId, err := c.clientWrapper.ChainID(ctx)
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
	auth.GasLimit = c.transferGasLimitBase + uint64(len(batch.Deposits))*c.transferGasLimitForEach
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

	argLists, err := c.extractList(batch)
	if err != nil {
		return "", err
	}

	err = c.checkAvailableTokens(ctx, argLists.tokens, argLists.amounts)
	if err != nil {
		return "", err
	}

	minimumForFee := big.NewInt(int64(auth.GasLimit))
	minimumForFee.Mul(minimumForFee, auth.GasPrice)
	err = c.checkRelayerFundsForFee(ctx, minimumForFee)
	if err != nil {
		return "", err
	}

	batchID := big.NewInt(0).SetUint64(batch.ID)
	tx, err := c.clientWrapper.ExecuteTransfer(auth, argLists.tokens, argLists.recipients, argLists.amounts, argLists.nonces, batchID, signatures)
	if err != nil {
		return "", err
	}

	txHash := tx.Hash().String()
	c.log.Info("Executed transfer transaction", "batchID", batchID, "hash", txHash)

	return txHash, err
}

// CheckClientAvailability will check the client availability and set the metric accordingly
func (c *client) CheckClientAvailability(ctx context.Context) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	currentBlock, err := c.clientWrapper.BlockNumber(ctx)
	if err != nil {
		c.setStatusForAvailabilityCheck(ethmultiversx.Unavailable, err.Error(), currentBlock)

		return err
	}

	if currentBlock != c.lastBlockNumber {
		c.retriesAvailabilityCheck = 0
		c.lastBlockNumber = currentBlock
	}

	// if we reached this point we will need to increment the retries counter
	defer c.incrementRetriesAvailabilityCheck()

	if c.retriesAvailabilityCheck > c.allowDelta {
		message := fmt.Sprintf("block %d fetched for %d times in a row", currentBlock, c.retriesAvailabilityCheck)
		c.setStatusForAvailabilityCheck(ethmultiversx.Unavailable, message, currentBlock)

		return nil
	}

	c.setStatusForAvailabilityCheck(ethmultiversx.Available, "", currentBlock)

	return nil
}

func (c *client) incrementRetriesAvailabilityCheck() {
	c.retriesAvailabilityCheck++
}

func (c *client) setStatusForAvailabilityCheck(status ethmultiversx.ClientStatus, message string, nonce uint64) {
	c.clientWrapper.SetStringMetric(core.MetricMultiversXClientStatus, status.String())
	c.clientWrapper.SetStringMetric(core.MetricLastMultiversXClientError, message)
	c.clientWrapper.SetIntMetric(core.MetricLastBlockNonce, int(nonce))
}

func (c *client) checkAvailableTokens(ctx context.Context, tokens []common.Address, amounts []*big.Int) error {
	transfers := c.getCumulatedTransfers(tokens, amounts)

	return c.checkCumulatedTransfers(ctx, transfers)
}

func (c *client) getCumulatedTransfers(tokens []common.Address, amounts []*big.Int) map[common.Address]*big.Int {
	transfers := make(map[common.Address]*big.Int)
	for i, token := range tokens {
		existing, found := transfers[token]
		if !found {
			existing = big.NewInt(0)
			transfers[token] = existing
		}

		existing.Add(existing, amounts[i])
	}

	return transfers
}

func (c *client) checkCumulatedTransfers(ctx context.Context, transfers map[common.Address]*big.Int) error {
	for erc20Address, value := range transfers {
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
	}

	return nil
}

func (c *client) checkRelayerFundsForFee(ctx context.Context, transferFee *big.Int) error {

	ethereumRelayerAddress := crypto.PubkeyToAddress(*c.publicKey)

	existingBalance, err := c.clientWrapper.BalanceAt(ctx, ethereumRelayerAddress, nil)
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
	return c.clientWrapper.GetStatusesAfterExecution(ctx, big.NewInt(0).SetUint64(batchId))
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
