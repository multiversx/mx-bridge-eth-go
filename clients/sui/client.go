package sui

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"github.com/block-vision/sui-go-sdk/common/keypair"
	signer "github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"golang.org/x/crypto/blake2b"
	"math/big"
	"sync"
)

const (
	minQuorumValue                  = uint64(1)
	minClientAvailabilityAllowDelta = 1
)

type ArgsSuiClient struct {
	suiClient             sui.ISuiAPI
	Log                   chainCore.Logger
	RelayerPrivateKey     ed25519.PrivateKey
	SafeContractAddress   string
	BridgeContractAddress string
	TokensMapper          TokensMapper
	StatusHandler         bridgeCore.StatusHandler
	Broadcaster           Broadcaster
	SignatureHolder       SignaturesHolder

	ClientAvailabilityAllowDelta uint64
}

type client struct {
	*suiClientDataGetter
	txHandler             txHandler
	tokensMapper          TokensMapper
	relayerPublicKey      ed25519.PublicKey
	relayerAddress        string
	safeContractAddress   string
	bridgeContractAddress string
	log                   chainCore.Logger
	addressConverter      bridgeCore.AddressConverter
	statusHandler         bridgeCore.StatusHandler
	broadcaster           Broadcaster
	signatureHolder       SignaturesHolder

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

	relayerPubKey, relayerAddress := generateRelayerPubKeyAndAddress(args.RelayerPrivateKey)
	signerRelayer := &signer.Signer{
		PriKey:  args.RelayerPrivateKey,
		PubKey:  relayerPubKey,
		Address: relayerAddress,
	}

	argsSuiClientDataGetter := ArgsSuiClientDataGetter{
		BridgeContractAddress: args.BridgeContractAddress,
		SafeContractAddress:   args.SafeContractAddress,
		RelayerAddress:        signerRelayer,
		Client:                args.suiClient,
		Log:                   args.Log,
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
			client:                args.suiClient,
			relayerAddress:        relayerAddress,
			relayerPrivateKey:     args.RelayerPrivateKey,
			bridgeContractAddress: args.BridgeContractAddress,
		},
		suiClientDataGetter:          getter,
		relayerPublicKey:             relayerPubKey,
		relayerAddress:               relayerAddress,
		bridgeContractAddress:        args.BridgeContractAddress,
		safeContractAddress:          args.SafeContractAddress,
		log:                          args.Log,
		addressConverter:             addressConverter,
		broadcaster:                  args.Broadcaster,
		tokensMapper:                 args.TokensMapper,
		statusHandler:                args.StatusHandler,
		signatureHolder:              args.SignatureHolder,
		clientAvailabilityAllowDelta: args.ClientAvailabilityAllowDelta,
	}

	c.log.Info("NewSuiClient")

	return c, err
}

func checkArgs(args ArgsSuiClient) error {
	if check.IfNil(args.suiClient) {
		return errNilClient
	}
	if len(args.RelayerPrivateKey) == 0 {
		return clients.ErrNilPrivateKey
	}
	if args.BridgeContractAddress == "" {
		return fmt.Errorf("%w for the MultisigContractAddress argument", errNilBridgeContract)
	}
	if args.SafeContractAddress == "" {
		return fmt.Errorf("%w for the SafeContractAddress argument", errNilSafeContract)
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

func generateRelayerPubKeyAndAddress(priKey ed25519.PrivateKey) (ed25519.PublicKey, string) {
	pubKey := priKey.Public().(ed25519.PublicKey)

	tmp := []byte{byte(keypair.Ed25519Flag)}
	tmp = append(tmp, pubKey...)
	addrBytes := blake2b.Sum256(tmp)
	addr := "0x" + hex.EncodeToString(addrBytes[:])[:64]

	return pubKey, addr
}

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
			clients.ErrDepositsAndBatchDepositsCountDiffer, batch.DepositsCount, len(deposits))
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
		tokenId := deposit.TokenAddress

		depositTransfer := &bridgeCore.DepositTransfer{
			Nonce:            deposit.Nonce.Uint64(),
			ToBytes:          toBytes,
			DisplayableTo:    c.addressConverter.ToBech32StringSilent(toBytes),
			FromBytes:        fromBytes,
			DisplayableFrom:  c.addressConverter.ToHexString(fromBytes),
			SourceTokenBytes: []byte(tokenId),
			DisplayableToken: tokenId,
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

// WasExecuted returns true if the MultiversX batch ID was executed
func (c *client) WasExecuted(ctx context.Context, mvxBatchID uint64) (bool, error) {
	return c.clientWrapper.WasBatchExecuted(ctx, big.NewInt(0).SetUint64(mvxBatchID))
}

// BroadcastSignatureForMessageHash will send the signature for the provided message hash
func (c *client) BroadcastSignatureForMessageHash(msgHash []byte) {
	signature, err := c.cryptoHandler.Sign(msgHash)
	if err != nil {
		c.log.Error("error generating signature", "msh hash", msgHash, "error", err)
		return
	}

	c.broadcaster.BroadcastSignature(signature, msgHash)
}

// GenerateMessageHash will generate the message hash based on the provided batch
func (c *client) GenerateMessageHash(batch *batchProcessor.ArgListsBatch, batchId uint64) (common.Hash, error) {
	panic("not implemented yet")
}

// ExecuteTransfer will initiate and send the transaction from the transfer batch struct
func (c *client) ExecuteTransfer(
	ctx context.Context,
	msgHash []byte,
	argLists *batchProcessor.ArgListsBatchSui,
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

	signatures := c.signatureHolder.Signatures(msgHash)
	if len(signatures) < quorum {
		return "", fmt.Errorf("%w num signatures: %d, quorum: %d", clients.ErrQuorumNotReached, len(signatures), quorum)
	}
	if len(signatures) > quorum {
		c.log.Debug("reducing the size of the signatures set",
			"quorum", quorum, "total signatures", len(signatures))
		signatures = signatures[:quorum]
	}

	//minimumForFee := big.NewInt(int64(auth.GasLimit))
	//minimumForFee.Mul(minimumForFee, auth.GasPrice)
	//err = c.checkRelayerFundsForFee(ctx, minimumForFee)
	//if err != nil {
	//	return "", err
	//}

	batchID := big.NewInt(0).SetUint64(batchId)
	tx, err := c.clientWrapper.ExecuteTransfer(argLists.SuiTokens, argLists.Recipients, argLists.Amounts, argLists.Nonces, batchID, signatures)
	if err != nil {
		return "", err
	}

	txHash := tx.Hash().String()
	c.log.Info("Executed transfer transaction", "batchID", batchID, "hash", txHash)

	return txHash, err
}

func (c *client) CheckClientAvailability(ctx context.Context) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	currentCheckpoint, err := c.clientWrapper.GetLatestCheckpoint(ctx)
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
	c.clientWrapper.SetStringMetric(bridgeCore.MetricMultiversXClientStatus, status.String())
	c.clientWrapper.SetStringMetric(bridgeCore.MetricLastMultiversXClientError, message)
	c.clientWrapper.SetIntMetric(bridgeCore.MetricLastBlockNonce, int(nonce))
}

func (c *client) incrementRetriesAvailabilityCheck() {
	c.retriesAvailabilityCheck++
}

// CheckRequiredBalance will check if the safe has enough balance for the transfer
func (c *client) CheckRequiredBalance(ctx context.Context, coinType string, value *big.Int) error {
	isMintBurn, err := c.MintBurnTokens(ctx, coinType)
	if err != nil {
		return err
	}

	if isMintBurn {
		return nil
	}
	existingBalance, err := c.clientWrapper.GetBalance(ctx, c.safeContractId, coinType)
	if err != nil {
		return fmt.Errorf("%w for owner %s for coin %s", err, c.safeContractId, coinType)
	}

	totalExistingBalanceStr := existingBalance.TotalBalance
	totalExistingBalance := new(big.Int)
	_, ok := totalExistingBalance.SetString(totalExistingBalanceStr, 10)
	if !ok {
		return fmt.Errorf("invalid balance string: %s", totalExistingBalanceStr)
	}
	if value.Cmp(totalExistingBalance) > 0 {
		return fmt.Errorf("%w, existing: %s, required: %s for coin %s and owner %s",
			errInsufficientCoinBalance, totalExistingBalanceStr, value.String(), coinType, c.safeContractId)
	}

	c.log.Debug("checked coin balance",
		"Coin type", coinType,
		"owner address", c.safeContractId,
		"existing balance", totalExistingBalanceStr,
		"needed", value.String())

	return nil
}

// TotalBalances returns the total balance of the given token
func (c *client) TotalBalances(ctx context.Context, token string) (*big.Int, error) {
	return c.clientWrapper.TotalBalances(ctx, token)
}

// MintBalances returns the mint balance of the given token
func (c *client) MintBalances(ctx context.Context, token string) (*big.Int, error) {
	return c.clientWrapper.MintBalances(ctx, token)
}

// BurnBalances returns the burn balance of the given token
func (c *client) BurnBalances(ctx context.Context, token string) (*big.Int, error) {
	return c.clientWrapper.BurnBalances(ctx, token)
}

// MintBurnTokens returns true if the token is mintBurn token
func (c *client) MintBurnTokens(ctx context.Context, token string) (bool, error) {
	return c.clientWrapper.MintBurnTokens(ctx, token)
}

// NativeTokens returns true if the token is native
func (c *client) NativeTokens(ctx context.Context, token string) (bool, error) {
	return c.clientWrapper.NativeTokens(ctx, token)
}

// WhitelistedTokens returns true if the token is whitelisted
func (c *client) WhitelistedTokens(ctx context.Context, token string) (bool, error) {
	return c.clientWrapper.WhitelistedTokens(ctx, token)
}

// GetTransactionsStatuses will return the transactions statuses from the batch
func (c *client) GetTransactionsStatuses(ctx context.Context, batchId uint64) ([]byte, error) {
	buff, isFinal, err := c.clientWrapper.GetStatusesAfterExecution(ctx, big.NewInt(0).SetUint64(batchId))
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
	return c.clientWrapper.Quorum(ctx)
}

// IsQuorumReached returns true if the number of signatures is at least the size of quorum
func (c *client) IsQuorumReached(ctx context.Context, msg []byte) (bool, error) {
	signatures := c.signatureHolder.Signatures(msg)
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
