package sui

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"sync"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/transaction"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"golang.org/x/crypto/blake2b"
)

const (
	minQuorumValue                  = uint64(1)
	minClientAvailabilityAllowDelta = 1
)

type ArgsSuiClient struct {
	Proxy                      Proxy
	TxHandler                  txHandler
	Log                        chainCore.Logger
	Signer                     *signer.Signer
	PackageId                  string
	SafeObjectId               string
	SafeInitialSharedVersion   uint64
	BridgeObjectId             string
	BridgeInitialSharedVersion uint64
	TokensMapper               TokensMapper
	StatusHandler              bridgeCore.StatusHandler
	Broadcaster                Broadcaster
	SignatureHolder            SignaturesHolder

	ClientAvailabilityAllowDelta uint64
}

type client struct {
	*suiClientDataGetter
	proxy            Proxy
	txHandler        txHandler
	signer           *signer.Signer
	packageId        string
	safeObjectId     string
	bridgeObjectId   string
	log              chainCore.Logger
	tokensMapper     TokensMapper
	addressConverter bridgeCore.AddressConverter
	statusHandler    bridgeCore.StatusHandler
	broadcaster      Broadcaster
	signatureHolder  SignaturesHolder

	lastCheckpoint               uint64
	retriesAvailabilityCheck     uint64
	clientAvailabilityAllowDelta uint64
	mut                          sync.RWMutex
}

func NewSuiClient(args ArgsSuiClient) (*client, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	argsSuiClientDataGetter := ArgsSuiClientDataGetter{
		PackageId:                  args.PackageId,
		SafeObjectId:               args.SafeObjectId,
		SafeInitialSharedVersion:   args.SafeInitialSharedVersion,
		BridgeObjectId:             args.BridgeObjectId,
		BridgeInitialSharedVersion: args.BridgeInitialSharedVersion,
		RelayerAddress:             args.Signer.Address,
		Proxy:                      args.Proxy,
		Log:                        args.Log,
	}
	getter, err := NewSuiClientDataGetter(argsSuiClientDataGetter)
	if err != nil {
		return nil, err
	}

	addressConverter, err := converters.NewAddressConverter()
	if err != nil {
		return nil, clients.ErrNilAddressConverter
	}

	c := &client{
		proxy:                        args.Proxy,
		signer:                       args.Signer,
		txHandler:                    args.TxHandler,
		suiClientDataGetter:          getter,
		packageId:                    args.PackageId,
		safeObjectId:                 args.SafeObjectId,
		bridgeObjectId:               args.BridgeObjectId,
		log:                          args.Log,
		addressConverter:             addressConverter,
		broadcaster:                  args.Broadcaster,
		tokensMapper:                 args.TokensMapper,
		statusHandler:                args.StatusHandler,
		signatureHolder:              args.SignatureHolder,
		clientAvailabilityAllowDelta: args.ClientAvailabilityAllowDelta,
	}

	c.log.Info("NewSuiClient",
		"relayer address", c.signer.Address,
		"package ID", c.packageId,
		"bridge object ID", c.bridgeObjectId,
		"safe object ID", c.safeObjectId)

	return c, err
}

func checkArgs(args ArgsSuiClient) error {
	if args.Proxy == nil {
		return errNilProxy
	}
	if check.IfNil(args.TxHandler) {
		return errNilTxHandler
	}
	if args.Signer == nil {
		return errNilSigner
	}
	if len(args.PackageId) == 0 {
		return fmt.Errorf("%w for the PackageId argument", errNilPackageId)
	}
	if len(args.BridgeObjectId) == 0 {
		return fmt.Errorf("%w for the BridgeObjectId argument", errNilObjectId)
	}
	if len(args.SafeObjectId) == 0 {
		return fmt.Errorf("%w for the SafeObjectId argument", errNilObjectId)
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}
	if check.IfNil(args.TokensMapper) {
		return clients.ErrNilTokensMapper
	}
	if check.IfNil(args.Broadcaster) {
		return clients.ErrNilBroadcaster
	}
	if check.IfNil(args.SignatureHolder) {
		return clients.ErrNilSignaturesHolder
	}
	if check.IfNil(args.StatusHandler) {
		return clients.ErrNilStatusHandler
	}
	if args.ClientAvailabilityAllowDelta < minClientAvailabilityAllowDelta {
		return fmt.Errorf("%w for args.AllowedDelta, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.ClientAvailabilityAllowDelta, minClientAvailabilityAllowDelta)
	}

	return nil
}

func (c *client) getSuiClient() (*sui.Client, bool) {
	suiClient, ok := c.proxy.(*sui.Client)
	return suiClient, ok
}

// GetBatch returns the transfer batch by providing the nonce
func (c *client) GetBatch(ctx context.Context, nonce uint64) (*bridgeCore.TransferBatch, bool, error) {
	c.log.Info("Getting batch", "nonce", nonce)
	batch, isFinalBatch, err := c.GetBatchByNonce(ctx, nonce)
	if err != nil {
		return nil, false, err
	}
	deposits, areFinalDeposits, err := c.GetBatchDeposits(ctx, nonce)
	if err != nil {
		return nil, false, err
	}
	if int(batch.DepositsCount) != len(deposits) {
		return nil, false, fmt.Errorf("%w, batch.DepositsCount: %d, fetched deposits len: %d",
			clients.ErrDepositsAndBatchDepositsCountDiffer, batch.DepositsCount, len(deposits))
	}

	transferBatch := &bridgeCore.TransferBatch{
		ID:          batch.Nonce,
		BlockNumber: batch.TimestampMs,
		Deposits:    make([]*bridgeCore.DepositTransfer, 0, batch.DepositsCount),
	}
	cachedTokens := make(map[string][]byte)
	for i := range deposits {
		deposit := deposits[i]
		toBytes := deposit.Recipient
		fromBytes := deposit.Sender[:]
		tokenId := deposit.TokenTypeBytes

		depositTransfer := &bridgeCore.DepositTransfer{
			Nonce:            deposit.Nonce,
			ToBytes:          toBytes,
			DisplayableTo:    c.addressConverter.ToBech32StringSilent(toBytes),
			FromBytes:        fromBytes,
			DisplayableFrom:  suiAddressFromBytes(fromBytes),
			SourceTokenBytes: tokenId,
			DisplayableToken: "0x" + string(tokenId),
			Amount:           big.NewInt(0).SetUint64(deposit.Amount),
		}
		storedConvertedTokenBytes, exists := cachedTokens[depositTransfer.DisplayableToken]
		if !exists {
			coinTypeBytes := []byte(depositTransfer.DisplayableToken)
			coinTypeBytesWithLength := AppendLengthToData(coinTypeBytes)
			depositTransfer.DestinationTokenBytes, err = c.tokensMapper.ConvertToken(ctx, coinTypeBytesWithLength)
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

// WasExecuted returns true if the MultiversX batch ID was executed
func (c *client) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
	return c.WasBatchExecuted(ctx, batchID)
}

// BroadcastSignatureForMessageHash will send the signature for the provided message hash
func (c *client) BroadcastSignatureForMessageHash(msgHash []byte) {
	if len(msgHash)%32 != 0 {
		c.log.Error("invalid message hash length", "msg hash", msgHash, "length", len(msgHash))
		return
	}

	var signatures []byte
	numberOfHashes := len(msgHash) / 32
	for i := 0; i < numberOfHashes; i++ {
		start := i * 32
		end := start + 32
		currentHash := msgHash[start:end]
		resp, err := c.signer.SignPersonalMessageV1(string(currentHash))
		if err != nil {
			c.log.Error("error generating signature", "msh hash", currentHash, "error", err)
			return
		}

		signatures = append(signatures, []byte(resp.Signature)...)
	}

	c.broadcaster.BroadcastSignature(signatures, msgHash)
}

// GenerateMessageHash will generate the message hash based on the provided batch
func (c *client) GenerateMessageHash(batch *batchProcessor.ArgListsBatch, batchId uint64) ([]byte, error) {
	if batch == nil {
		return nil, clients.ErrNilBatch
	}

	var hash []byte
	groups := c.groupTransfersByTokenType(batch)
	for _, group := range groups {
		groupHash, err := c.getHashForTokenGroupData(group, batchId)
		if err != nil {
			return nil, fmt.Errorf("error getting hash for token group data: %v", err)
		}
		hash = append(hash, groupHash...)
	}

	return hash, nil
}

func (c *client) getHashForTokenGroupData(batch *TokenTransferGroup, batchId uint64) ([]byte, error) {
	message, err := c.constructBatchMessage(batchId, batch)
	if err != nil {
		return nil, fmt.Errorf("error constructing batch message: %v", err)
	}

	hash := blake2b.Sum256(message)
	return hash[:], nil
}

func (c *client) constructBatchMessage(
	batchId uint64,
	batch *TokenTransferGroup,
) ([]byte, error) {
	var message bytes.Buffer

	bcsEncoder := mystenbcs.NewEncoder(&message)
	err := bcsEncoder.Encode(batchId)
	if err != nil {
		return nil, fmt.Errorf("error encoding batch_id: %v", err)
	}

	for i := 0; i < len(batch.Tokens); i++ {
		tokenBuf := bytes.Buffer{}
		tokenEncoder := mystenbcs.NewEncoder(&tokenBuf)
		err = tokenEncoder.Encode(batch.Tokens[i])
		if err != nil {
			return nil, fmt.Errorf("error encoding token: %v", err)
		}
		message.Write(tokenBuf.Bytes())

		recipientBuf := bytes.Buffer{}
		recipientEncoder := mystenbcs.NewEncoder(&recipientBuf)
		err = recipientEncoder.Encode(batch.Recipients[i])
		if err != nil {
			return nil, fmt.Errorf("error encoding recipient: %v", err)
		}
		message.Write(recipientBuf.Bytes())

		amountBuf := bytes.Buffer{}
		amountEncoder := mystenbcs.NewEncoder(&amountBuf)
		err = amountEncoder.Encode(batch.Amounts[i])
		if err != nil {
			return nil, fmt.Errorf("error encoding amount: %v", err)
		}
		message.Write(amountBuf.Bytes())

		nonceBuf := bytes.Buffer{}
		nonceEncoder := mystenbcs.NewEncoder(&nonceBuf)
		err = nonceEncoder.Encode(batch.Nonces[i])
		if err != nil {
			return nil, fmt.Errorf("error encoding nonce: %v", err)
		}
		message.Write(nonceBuf.Bytes())
	}

	return message.Bytes(), nil
}

// ExecuteTransfer will initiate and send the transaction from the transfer batch struct
func (c *client) ExecuteTransfer(
	ctx context.Context,
	msgHash []byte,
	argLists *batchProcessor.ArgListsBatch,
	batchID uint64,
	quorum int,
) (string, error) {
	if argLists == nil {
		return "", clients.ErrNilBatch
	}

	isPaused, err := c.IsPaused(ctx)
	if err != nil {
		return "", fmt.Errorf("%w in client.ExecuteTransfer", err)
	}
	if isPaused {
		return "", fmt.Errorf("%w in client.ExecuteTransfer", clients.ErrMultisigContractPaused)
	}

	serializedSignatures := c.signatureHolder.Signatures(msgHash)
	if len(serializedSignatures) < quorum {
		return "", fmt.Errorf("%w num signatures: %d, quorum: %d", clients.ErrQuorumNotReached, len(serializedSignatures), quorum)
	}
	if len(serializedSignatures) > quorum {
		c.log.Debug("reducing the size of the signatures set",
			"quorum", quorum, "total signatures", len(serializedSignatures))
		serializedSignatures = serializedSignatures[:quorum]
	}

	tokenGroups := c.groupTransfersByTokenType(argLists)

	err = c.processSignaturesOfRelayers(tokenGroups, serializedSignatures)
	if err != nil {
		return "", err
	}

	gasCoin, err := c.getGasCoinOfRelayer(ctx)
	if err != nil {
		return "", err
	}

	calls, err := c.prepareExecuteTransferCallArgs(batchID, tokenGroups)
	if err != nil {
		return "", err
	}

	return c.txHandler.SendTransactionReturnHash(ctx, gasCoin, calls)
}

func (c *client) groupTransfersByTokenType(argLists *batchProcessor.ArgListsBatch) map[string]*TokenTransferGroup {
	groups := make(map[string]*TokenTransferGroup)

	for i := 0; i < len(argLists.PeerTokens); i++ {
		tokenTypeStr := string(argLists.PeerTokens[i])

		if groups[tokenTypeStr] == nil {
			groups[tokenTypeStr] = &TokenTransferGroup{
				Recipients: make([]models.SuiAddressBytes, 0),
				Amounts:    make([]uint64, 0),
				Tokens:     make([][]byte, 0),
				Nonces:     make([]uint64, 0),
			}
		}

		suiAddress := suiAddressFromBytes(argLists.Recipients[i])
		suiAddressBytes, err := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(suiAddress))
		if err != nil {
			return nil
		}

		groups[tokenTypeStr].Recipients = append(groups[tokenTypeStr].Recipients, *suiAddressBytes)
		groups[tokenTypeStr].Amounts = append(groups[tokenTypeStr].Amounts, argLists.Amounts[i].Uint64())
		groups[tokenTypeStr].Tokens = append(groups[tokenTypeStr].Tokens, argLists.PeerTokens[i])
		groups[tokenTypeStr].Nonces = append(groups[tokenTypeStr].Nonces, argLists.Nonces[i].Uint64())
	}

	return groups
}

func (c *client) processSignaturesOfRelayers(tokenGroups map[string]*TokenTransferGroup, serializedSignatures [][]byte) error {
	var tokenTypes []string
	for key := range tokenGroups {
		tokenTypes = append(tokenTypes, key)
	}

	for _, serializedSigsOfRelayer := range serializedSignatures {
		n := len(serializedSigsOfRelayer) / 132
		for i := 0; i < n; i++ {
			start := i * 132
			end := start + 132
			signature := serializedSigsOfRelayer[start:end]

			_bytes, err := base64.StdEncoding.DecodeString(string(signature))
			if err != nil {
				return fmt.Errorf("error decoding signature: %w", err)
			}

			sig := [96]byte(_bytes[1:]) // remove the signature scheme byte
			tokenGroups[tokenTypes[i]].Signatures = append(tokenGroups[tokenTypes[i]].Signatures, sig)
		}
	}

	return nil
}

func (c *client) getGasCoinOfRelayer(ctx context.Context) (*transaction.SuiObjectRef, error) {
	coins, err := c.GetCoinsForAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting gas coin for relayer %s: %w", c.relayerAddress, err)
	}

	if len(coins.Data) == 0 {
		return nil, fmt.Errorf("no coins found for relayer %s to pay gas", c.relayerAddress)
	}
	coinObj := coins.Data[0]

	return transaction.NewSuiObjectRef(
		models.SuiAddress(coinObj.CoinObjectId),
		coinObj.Version,
		models.ObjectDigest(coinObj.Digest),
	)
}

func (c *client) prepareExecuteTransferCallArgs(batchID uint64, tokenGroups map[string]*TokenTransferGroup) ([]bridgeCore.SuiPTBOperation, error) {
	i := 0
	n := len(tokenGroups)
	calls := make([]bridgeCore.SuiPTBOperation, 0, n)
	for coinType, group := range tokenGroups {
		i++
		isBatchComplete := i == n

		// COPY loop vars to local variables to avoid closure-capture bug
		localGroup := group
		localIsBatchComplete := isBatchComplete

		coinParts, err := parseCoinType(coinType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse coin type %s: %w", coinType, err)
		}

		coinIdBytes, err := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(coinParts[0]))
		if err != nil {
			return nil, fmt.Errorf("failed to convert coin type %s: %w", coinType, err)
		}

		// copy dereferenced address value if you'll use inside TypeTags/ArgsFn
		coinIdBytesVal := *coinIdBytes

		calls = append(calls, bridgeCore.SuiPTBOperation{
			Package:  models.SuiAddress(c.packageId),
			Module:   "bridge",
			Function: "execute_transfer",
			TypeTags: []transaction.TypeTag{
				{
					Struct: &transaction.StructTag{
						Address: coinIdBytesVal,
						Module:  coinParts[1],
						Name:    coinParts[2],
					},
				},
			},
			ArgsFn: func(tx *transaction.Transaction) []transaction.Argument {
				// use localGroup, localIsBatchComplete, coinIdBytesVal (if needed) – not the loop vars
				return []transaction.Argument{
					tx.Object(transaction.CallArg{Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             c.bridgeObjectIdBytes,
							InitialSharedVersion: c.bridgeInitialSharedVersion,
							Mutable:              true,
						},
					}}),
					tx.Object(transaction.CallArg{Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             c.safeObjectIdBytes,
							InitialSharedVersion: c.safeInitialSharedVersion,
							Mutable:              true,
						},
					}}),
					tx.Pure(localGroup.Recipients),
					tx.Pure(localGroup.Amounts),
					tx.Pure(localGroup.Tokens),
					tx.Pure(localGroup.Nonces),
					tx.Pure(batchID),
					tx.Pure(localGroup.Signatures),
					tx.Pure(localIsBatchComplete),
					tx.Object(transaction.CallArg{Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             c.clockIdBytes,
							InitialSharedVersion: clockInitialSharedVersion,
							Mutable:              false,
						},
					}}),
				}
			},
		})
	}

	return calls, nil
}

func (c *client) CheckClientAvailability(ctx context.Context) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	currentCheckpoint, err := c.GetLatestCheckpoint(ctx)
	if err != nil {
		c.setStatusForAvailabilityCheck(bridgeCore.Unavailable, err.Error(), currentCheckpoint)

		return err
	}

	if currentCheckpoint != c.lastCheckpoint {
		c.retriesAvailabilityCheck = 0
		c.lastCheckpoint = currentCheckpoint
	}

	// if we reached this point we will need to increment the retries counter
	defer c.incrementRetriesAvailabilityCheck()

	if c.retriesAvailabilityCheck > c.clientAvailabilityAllowDelta {
		message := fmt.Sprintf("block %d fetched for %d times in a row", currentCheckpoint, c.retriesAvailabilityCheck)
		c.setStatusForAvailabilityCheck(bridgeCore.Unavailable, message, currentCheckpoint)

		return nil
	}

	c.setStatusForAvailabilityCheck(bridgeCore.Available, "", currentCheckpoint)

	return nil
}

func (c *client) setStatusForAvailabilityCheck(status bridgeCore.ClientStatus, message string, nonce uint64) {
	c.statusHandler.SetStringMetric(bridgeCore.MetricMultiversXClientStatus, status.String())
	c.statusHandler.SetStringMetric(bridgeCore.MetricLastMultiversXClientError, message)
	c.statusHandler.SetIntMetric(bridgeCore.MetricLastBlockNonce, int(nonce))
}

func (c *client) incrementRetriesAvailabilityCheck() {
	c.retriesAvailabilityCheck++
}

// CheckRequiredBalance will check if the safe has enough balance for the transfer
func (c *client) CheckRequiredBalance(ctx context.Context, coinType []byte, value *big.Int) error {
	coinTypeStr := string(coinType)
	existingBalance, err := c.GetBalance(ctx, coinTypeStr)
	if err != nil {
		return fmt.Errorf("%w for owner %s for coin %s", err, c.safeObjectId, coinTypeStr)
	}

	totalExistingBalance := big.NewInt(0).SetUint64(existingBalance)
	if value.Cmp(totalExistingBalance) > 0 {
		return fmt.Errorf("%w, existing: %s, required: %s for coin %s and owner %s",
			errInsufficientCoinBalance, totalExistingBalance.String(), value.String(), coinTypeStr, c.safeObjectId)
	}

	c.log.Debug("checked coin balance",
		"Coin type", coinTypeStr,
		"owner address", c.safeObjectId,
		"existing balance", totalExistingBalance.String(),
		"needed", value.String())

	return nil
}

// TotalBalances returns the total balance of the given token
func (c *client) TotalBalances(ctx context.Context, coinType []byte) (*big.Int, error) {
	coinTypeStr := string(coinType)
	balance, err := c.GetTotalBalanceFromSafe(ctx, coinTypeStr)
	if err != nil {
		return nil, err
	}
	return big.NewInt(0).SetUint64(balance), nil
}

// WhitelistedTokens returns true if the token is whitelisted
func (c *client) WhitelistedTokens(ctx context.Context, coinType []byte) (bool, error) {
	coinTypeStr := string(coinType)
	return c.IsTokenWhitelisted(ctx, coinTypeStr)
}

// MintBalances returns nil every time
func (c *client) MintBalances(_ context.Context, _ []byte) (*big.Int, error) {
	return nil, nil
}

// BurnBalances returns the burn balance of the given token
func (c *client) BurnBalances(_ context.Context, _ []byte) (*big.Int, error) {
	return nil, nil
}

// MintBurnTokens returns false every time
func (c *client) MintBurnTokens(_ context.Context, _ []byte) (bool, error) {
	return false, nil
}

// NativeTokens returns true every time
func (c *client) NativeTokens(_ context.Context, _ []byte) (bool, error) {
	return true, nil
}

// GetTransactionsStatuses will return the transactions statuses from the batch
func (c *client) GetTransactionsStatuses(ctx context.Context, batchId uint64) ([]byte, error) {
	buff, isFinal, err := c.GetStatusesAfterExecution(ctx, batchId)
	if err != nil {
		return nil, err
	}
	if !isFinal {
		return nil, clients.ErrStatusIsNotFinal
	}

	return buff, nil
}

// GetQuorumSize returns the size of the quorum
func (c *client) GetQuorumSize(ctx context.Context) (*big.Int, error) {
	quorum, err := c.Quorum(ctx)
	if err != nil {
		return nil, err
	}
	return big.NewInt(0).SetUint64(quorum), nil
}

// IsQuorumReached returns true if the number of signatures is at least the size of quorum
func (c *client) IsQuorumReached(ctx context.Context, msg []byte) (bool, error) {
	signatures := c.signatureHolder.Signatures(msg)
	quorum, err := c.Quorum(ctx)
	if err != nil {
		return false, fmt.Errorf("%w in IsQuorumReached, Quorum call", err)
	}
	if quorum < minQuorumValue {
		return false, fmt.Errorf("%w in IsQuorumReached, minQuorum %d, got: %d", clients.ErrInvalidValue, minQuorumValue, quorum)
	}

	return len(signatures) >= int(quorum), nil
}

// GetBatchSCMetadata returns nil every time
func (c *client) GetBatchSCMetadata(_ context.Context, _ uint64, _ int64) ([]*contract.ERC20SafeERC20SCDeposit, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}
