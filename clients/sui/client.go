package sui

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/multiversx/mx-bridge-eth-go/clients/sui/dtos"
	"math/big"
	"sync"

	"github.com/block-vision/sui-go-sdk/common/keypair"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/signer"
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
	Log                        chainCore.Logger
	RelayerPrivateKey          ed25519.PrivateKey
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
	txHandler        txHandler
	tokensMapper     TokensMapper
	relayerPublicKey ed25519.PublicKey
	relayerAddress   string
	packageId        string
	safeObjectId     string
	bridgeObjectId   string
	log              chainCore.Logger
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

	relayerPubKey, relayerAddress := generatePubKeyAndAddressFromPriKey(args.RelayerPrivateKey)
	relayerSigner := &signer.Signer{
		PriKey:  args.RelayerPrivateKey,
		PubKey:  relayerPubKey,
		Address: relayerAddress,
	}

	argsSuiClientDataGetter := ArgsSuiClientDataGetter{
		PackageId:                  args.PackageId,
		SafeObjectId:               args.SafeObjectId,
		SafeInitialSharedVersion:   args.SafeInitialSharedVersion,
		BridgeObjectId:             args.BridgeObjectId,
		BridgeInitialSharedVersion: args.BridgeInitialSharedVersion,
		RelayerAddress:             relayerAddress,
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
		txHandler: &transactionHandler{
			proxy:         args.Proxy,
			relayerSigner: relayerSigner,
		},
		suiClientDataGetter:          getter,
		relayerPublicKey:             relayerPubKey,
		relayerAddress:               relayerAddress,
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

	c.log.Info("NewSuiClient")
	c.log.Info("NewSuiClient",
		"relayer address", relayerAddress,
		"package ID", c.packageId,
		"bridge object ID", c.bridgeObjectId,
		"safe object ID", c.safeObjectId)

	return c, err
}

func checkArgs(args ArgsSuiClient) error {
	if args.Proxy == nil {
		return errNilProxy
	}
	if len(args.RelayerPrivateKey) == 0 {
		return clients.ErrNilPrivateKey
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

func generatePubKeyAndAddressFromPriKey(priKey ed25519.PrivateKey) (ed25519.PublicKey, string) {
	pubKey := priKey.Public().(ed25519.PublicKey)

	tmp := []byte{byte(keypair.Ed25519Flag)}
	tmp = append(tmp, pubKey...)
	addrBytes := blake2b.Sum256(tmp)
	addr := "0x" + hex.EncodeToString(addrBytes[:])[:64]

	return pubKey, addr
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
		toBytes := deposit.Recipient[:]
		fromBytes := deposit.Sender
		tokenId := deposit.TokenTypeBytes

		depositTransfer := &bridgeCore.DepositTransfer{
			Nonce:            deposit.Nonce,
			ToBytes:          toBytes,
			DisplayableTo:    c.addressConverter.ToBech32StringSilent(toBytes),
			FromBytes:        getDataWithAppendedLength(fromBytes[:]),
			DisplayableFrom:  AddressFromBytes(fromBytes),
			SourceTokenBytes: tokenId,
			DisplayableToken: "0x" + string(tokenId),
			Amount:           big.NewInt(0).SetUint64(deposit.Amount),
		}
		storedConvertedTokenBytes, exists := cachedTokens[depositTransfer.DisplayableToken]
		if !exists {
			coinTypeBytes := []byte(depositTransfer.DisplayableToken)
			lenBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(lenBytes, uint32(len(coinTypeBytes)))
			encoded := append(lenBytes, coinTypeBytes...)
			depositTransfer.DestinationTokenBytes, err = c.tokensMapper.ConvertToken(ctx, encoded)
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

func getDataWithAppendedLength(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(data)))
	encoded := append(lenBytes, data...)

	return encoded
}

// WasExecuted returns true if the MultiversX batch ID was executed
func (c *client) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
	return c.WasBatchExecuted(ctx, batchID)
}

// BroadcastSignatureForMessageHash will send the signature for the provided message hash
func (c *client) BroadcastSignatureForMessageHash(msgHash []byte) {
	resp, err := c.txHandler.Sign(msgHash)
	if err != nil {
		c.log.Error("error generating signature", "msh hash", msgHash, "error", err)
		return
	}

	c.broadcaster.BroadcastSignature([]byte(resp.Signature), msgHash)
}

// GenerateMessageHash will generate the message hash based on the provided batch
func (c *client) GenerateMessageHash(batch *batchProcessor.ArgListsBatch, batchId uint64) ([]byte, error) {
	if batch == nil {
		return nil, clients.ErrNilBatch
	}

	formattedRecipients := make([]models.SuiAddressBytes, 0, len(batch.Recipients))
	for _, recipient := range batch.Recipients {
		formattedRecipients = append(formattedRecipients, models.SuiAddressBytes(recipient[4:]))
	}

	uint64Amounts := make([]uint64, 0, len(batch.Amounts))
	for _, amount := range batch.Amounts {
		uint64Amounts = append(uint64Amounts, amount.Uint64())
	}

	uint64Nonces := make([]uint64, 0, len(batch.Nonces))
	for _, nonce := range batch.Nonces {
		uint64Nonces = append(uint64Nonces, nonce.Uint64())
	}
	message, err := c.ConstructBatchMessage(batchId, batch.PeerTokens, formattedRecipients, uint64Amounts, uint64Nonces)
	if err != nil {
		return nil, fmt.Errorf("error constructing batch message: %v", err)
	}

	hash := blake2b.Sum256(message)

	return hash[:], nil
}

func (c *client) ConstructBatchMessage(
	batchId uint64,
	tokens [][]byte,
	recipients []models.SuiAddressBytes,
	amounts []uint64,
	depositNonces []uint64,
) ([]byte, error) {
	var message bytes.Buffer

	bcsEncoder := mystenbcs.NewEncoder(&message)
	err := bcsEncoder.Encode(batchId)
	if err != nil {
		return nil, fmt.Errorf("error encoding batch_id: %v", err)
	}

	for i := 0; i < len(tokens); i++ {
		tokenBuf := bytes.Buffer{}
		tokenEncoder := mystenbcs.NewEncoder(&tokenBuf)
		err := tokenEncoder.Encode(tokens[i])
		if err != nil {
			return nil, fmt.Errorf("error encoding token: %v", err)
		}
		message.Write(tokenBuf.Bytes())

		recipientBuf := bytes.Buffer{}
		recipientEncoder := mystenbcs.NewEncoder(&recipientBuf)
		err = recipientEncoder.Encode(recipients[i])
		if err != nil {
			return nil, fmt.Errorf("error encoding recipient: %v", err)
		}
		message.Write(recipientBuf.Bytes())

		amountBuf := bytes.Buffer{}
		amountEncoder := mystenbcs.NewEncoder(&amountBuf)
		err = amountEncoder.Encode(amounts[i])
		if err != nil {
			return nil, fmt.Errorf("error encoding amount: %v", err)
		}
		message.Write(amountBuf.Bytes())

		nonceBuf := bytes.Buffer{}
		nonceEncoder := mystenbcs.NewEncoder(&nonceBuf)
		err = nonceEncoder.Encode(depositNonces[i])
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
	batchId uint64,
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

	var signatures [][96]byte
	for _, serializedSig := range serializedSignatures {
		_bytes, err := base64.StdEncoding.DecodeString(string(serializedSig))
		if err != nil {
			return "", fmt.Errorf("error decoding signature: %w", err)
		}
		var sig [96]byte
		copy(sig[:], _bytes[1:]) // remove the first byte which is the signature scheme
		signatures = append(signatures, sig)
	}

	var hash string
	var allErrors []error
	tokenGroups := c.groupTransfersByTokenType(argLists)

	for tokenType, group := range tokenGroups {
		hash, err = c.executeTransferForTokenType(ctx, tokenType, group, batchId, signatures)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to execute transfer for token type %s: %w", tokenType, err))
			continue
		}
		c.log.Info("Executed transfer transaction", "tokenType", tokenType, "batchID", batchId, "hash", hash)
	}

	if len(allErrors) > 0 {
		return hash, fmt.Errorf("some transfers failed: %v", allErrors)
	}

	return hash, err
}

func (c *client) groupTransfersByTokenType(argLists *batchProcessor.ArgListsBatch) map[string]*dtos.TokenTransferGroup {
	groups := make(map[string]*dtos.TokenTransferGroup)

	for i := 0; i < len(argLists.PeerTokens); i++ {
		tokenTypeStr := string(argLists.PeerTokens[i])

		if groups[tokenTypeStr] == nil {
			groups[tokenTypeStr] = &dtos.TokenTransferGroup{
				Recipients: make([]string, 0),
				Amounts:    make([]string, 0),
				Nonces:     make([]string, 0),
			}
		}

		hexStr := hex.EncodeToString(argLists.Recipients[i][4:])
		suiAddress := "0x" + hexStr

		groups[tokenTypeStr].Recipients = append(groups[tokenTypeStr].Recipients, suiAddress)
		groups[tokenTypeStr].Amounts = append(groups[tokenTypeStr].Amounts, argLists.Amounts[i].String())
		groups[tokenTypeStr].Nonces = append(groups[tokenTypeStr].Nonces, argLists.Nonces[i].String())
	}

	return groups
}

func (c *client) executeTransferForTokenType(
	ctx context.Context,
	tokenType string,
	group *dtos.TokenTransferGroup,
	batchId uint64,
	signatures [][96]byte,
) (string, error) {
	moveCallReq := models.MoveCallRequest{
		Signer:          c.relayerAddress,
		PackageObjectId: c.packageId,
		Module:          "bridge",
		Function:        "execute_transfer",
		TypeArguments:   []interface{}{tokenType},
		Arguments: []interface{}{
			c.bridgeObjectId,
			c.safeObjectId,
			group.Recipients,
			group.Amounts,
			group.Nonces,
			big.NewInt(0).SetUint64(batchId).String(),
			signatures,
			clockId,
		},
		GasBudget: "100000000",
	}

	return c.txHandler.SendTransactionReturnHash(ctx, moveCallReq)
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
	existingBalance, err := c.GetBalance(ctx, c.safeObjectId, coinTypeStr)
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
func (c *client) TotalBalances(ctx context.Context, token []byte) (*big.Int, error) {
	tokenAddr := AddressBytesToString(token)
	balance, err := c.GetTotalBalanceFromSafe(ctx, tokenAddr)
	if err != nil {
		return nil, err
	}
	return big.NewInt(0).SetUint64(balance), nil
}

// WhitelistedTokens returns true if the token is whitelisted
func (c *client) WhitelistedTokens(ctx context.Context, token []byte) (bool, error) {
	tokenAddr := AddressBytesToString(token)
	return c.IsTokenWhitelisted(ctx, tokenAddr)
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
