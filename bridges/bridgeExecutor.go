package bridges

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// splits - represent the number of times we split the maximum interval
// we wait for the transfer confirmation on peer client
const splits = 10
const minRetries = 1

// ArgsBridgeExecutor is the arguments DTO struct used in both bridges
type ArgsBridgeExecutor struct {
	Log                          logger.Logger
	TopologyProvider             TopologyProvider
	MultiversXClient             MultiversXClient
	PeerChainClient              PeerChainClient
	TimeForWaitOnPeerClient      time.Duration
	StatusHandler                core.StatusHandler
	SignaturesHolder             SignaturesHolder
	BalanceValidator             BalanceValidator
	MaxQuorumRetriesOnPeerClient uint64
	MaxQuorumRetriesOnMultiversX uint64
	MaxRestriesOnWasProposed     uint64
}

type bridgeExecutor struct {
	log                          logger.Logger
	topologyProvider             TopologyProvider
	multiversXClient             MultiversXClient
	peerChainClient              PeerChainClient
	timeForWaitOnPeerClient      time.Duration
	statusHandler                core.StatusHandler
	sigsHolder                   SignaturesHolder
	balanceValidator             BalanceValidator
	maxQuorumRetriesOnPeerClient uint64
	maxQuorumRetriesOnMultiversX uint64
	maxRetriesOnWasProposed      uint64

	batch                     *bridgeCore.TransferBatch
	actionID                  uint64
	msgHash                   []byte
	quorumRetriesOnPeerClient uint64
	quorumRetriesOnMultiversX uint64
	retriesOnWasProposed      uint64
}

// NewBridgeExecutor creates a bridge executor, which can be used for both half-bridges
func NewBridgeExecutor(args ArgsBridgeExecutor) (*bridgeExecutor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	executor := createBridgeExecutor(args)
	return executor, nil
}

func checkArgs(args ArgsBridgeExecutor) error {
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}
	if check.IfNil(args.MultiversXClient) {
		return ErrNilMultiversXClient
	}
	if check.IfNil(args.PeerChainClient) {
		return ErrNilPeerChainClient
	}
	if check.IfNil(args.TopologyProvider) {
		return ErrNilTopologyProvider
	}
	if check.IfNil(args.StatusHandler) {
		return ErrNilStatusHandler
	}
	if args.TimeForWaitOnPeerClient < durationLimit {
		return ErrInvalidDuration
	}
	if check.IfNil(args.SignaturesHolder) {
		return ErrNilSignaturesHolder
	}
	if check.IfNil(args.BalanceValidator) {
		return ErrNilBalanceValidator
	}
	if args.MaxQuorumRetriesOnPeerClient < minRetries {
		return fmt.Errorf("%w for args.MaxQuorumRetriesOnPeerClient, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.MaxQuorumRetriesOnPeerClient, minRetries)
	}
	if args.MaxQuorumRetriesOnMultiversX < minRetries {
		return fmt.Errorf("%w for args.MaxQuorumRetriesOnMultiversX, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.MaxQuorumRetriesOnMultiversX, minRetries)
	}
	if args.MaxRestriesOnWasProposed < minRetries {
		return fmt.Errorf("%w for args.MaxRestriesOnWasProposed, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.MaxRestriesOnWasProposed, minRetries)
	}
	return nil
}

func createBridgeExecutor(args ArgsBridgeExecutor) *bridgeExecutor {
	return &bridgeExecutor{
		log:                          args.Log,
		multiversXClient:             args.MultiversXClient,
		peerChainClient:              args.PeerChainClient,
		topologyProvider:             args.TopologyProvider,
		statusHandler:                args.StatusHandler,
		timeForWaitOnPeerClient:      args.TimeForWaitOnPeerClient,
		sigsHolder:                   args.SignaturesHolder,
		balanceValidator:             args.BalanceValidator,
		maxQuorumRetriesOnPeerClient: args.MaxQuorumRetriesOnPeerClient,
		maxQuorumRetriesOnMultiversX: args.MaxQuorumRetriesOnMultiversX,
		maxRetriesOnWasProposed:      args.MaxRestriesOnWasProposed,
	}
}

// PrintInfo will print the provided data through the inner logger instance
func (executor *bridgeExecutor) PrintInfo(logLevel logger.LogLevel, message string, extras ...interface{}) {
	executor.log.Log(logLevel, message, extras...)

	switch logLevel {
	case logger.LogWarning, logger.LogError:
		executor.setExecutionMessageInStatusHandler(logLevel, message, extras...)
	}
}

func (executor *bridgeExecutor) setExecutionMessageInStatusHandler(level logger.LogLevel, message string, extras ...interface{}) {
	msg := fmt.Sprintf("%s: %s", level, message)
	for i := 0; i < len(extras)-1; i += 2 {
		msg += fmt.Sprintf(" %s = %s", convertObjectToString(extras[i]), convertObjectToString(extras[i+1]))
	}

	executor.statusHandler.SetStringMetric(core.MetricLastError, msg)
}

// MyTurnAsLeader returns true if the current relayer node is the leader
func (executor *bridgeExecutor) MyTurnAsLeader() bool {
	return executor.topologyProvider.MyTurnAsLeader()
}

// GetBatchFromMultiversX fetches the pending batch from MultiversX
func (executor *bridgeExecutor) GetBatchFromMultiversX(ctx context.Context) (*bridgeCore.TransferBatch, error) {
	batch, err := executor.multiversXClient.GetPendingBatch(ctx)
	if err == nil {
		executor.statusHandler.SetIntMetric(core.MetricNumBatches, int(batch.ID)-1)
	}
	return batch, err
}

// StoreBatchFromMultiversX saves the pending batch from MultiversX
func (executor *bridgeExecutor) StoreBatchFromMultiversX(batch *bridgeCore.TransferBatch) error {
	if batch == nil {
		return ErrNilBatch
	}

	executor.batch = batch
	return nil
}

// GetStoredBatch returns the stored batch
func (executor *bridgeExecutor) GetStoredBatch() *bridgeCore.TransferBatch {
	return executor.batch
}

// GetLastExecutedPeerBatchIDFromMultiversX returns the last executed batch ID that is stored on the MultiversX SC
func (executor *bridgeExecutor) GetLastExecutedPeerBatchIDFromMultiversX(ctx context.Context) (uint64, error) {
	batchID, err := executor.multiversXClient.GetLastExecutedEthBatchID(ctx)
	if err == nil {
		executor.statusHandler.SetIntMetric(core.MetricNumBatches, int(batchID))
	}
	return batchID, err
}

// VerifyLastDepositNonceExecutedOnPeerBatch will check the deposit Nonces from the fetched batch from peer client
func (executor *bridgeExecutor) VerifyLastDepositNonceExecutedOnPeerBatch(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	lastNonce, err := executor.multiversXClient.GetLastExecutedEthTxID(ctx)
	if err != nil {
		return err
	}

	return executor.verifyDepositNonces(lastNonce)
}

func (executor *bridgeExecutor) verifyDepositNonces(lastNonce uint64) error {
	startNonce := lastNonce + 1
	for _, dt := range executor.batch.Deposits {
		if dt.Nonce != startNonce {
			return fmt.Errorf("%w for deposit %s, expected: %d", ErrInvalidDepositNonce, dt.String(), startNonce)
		}

		startNonce++
	}

	return nil
}

// GetAndStoreActionIDForProposeTransferOnMultiversX fetches the action ID for ProposeTransfer by using the stored batch. Stores the action ID and returns it
func (executor *bridgeExecutor) GetAndStoreActionIDForProposeTransferOnMultiversX(ctx context.Context) (uint64, error) {
	if executor.batch == nil {
		return InvalidActionID, ErrNilBatch
	}

	actionID, err := executor.multiversXClient.GetActionIDForProposeTransfer(ctx, executor.batch)
	if err != nil {
		return InvalidActionID, err
	}

	executor.actionID = actionID

	return actionID, nil
}

// GetAndStoreActionIDForProposeSetStatusFromMultiversX fetches the action ID for SetStatus by using the stored batch. Stores the action ID and returns it
func (executor *bridgeExecutor) GetAndStoreActionIDForProposeSetStatusFromMultiversX(ctx context.Context) (uint64, error) {
	if executor.batch == nil {
		return InvalidActionID, ErrNilBatch
	}

	actionID, err := executor.multiversXClient.GetActionIDForSetStatusOnPendingTransfer(ctx, executor.batch)
	if err != nil {
		return InvalidActionID, err
	}

	executor.actionID = actionID

	return actionID, nil
}

// GetStoredActionID returns the stored action ID
func (executor *bridgeExecutor) GetStoredActionID() uint64 {
	return executor.actionID
}

// WasTransferProposedOnMultiversX checks if the transfer was proposed on MultiversX
func (executor *bridgeExecutor) WasTransferProposedOnMultiversX(ctx context.Context) (bool, error) {
	if executor.batch == nil {
		return false, ErrNilBatch
	}

	return executor.multiversXClient.WasProposedTransfer(ctx, executor.batch)
}

// ProposeTransferOnMultiversX propose the transfer on MultiversX
func (executor *bridgeExecutor) ProposeTransferOnMultiversX(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	hash, err := executor.multiversXClient.ProposeTransfer(ctx, executor.batch)
	if err != nil {
		return err
	}

	executor.log.Info("proposed transfer", "hash", hash,
		"batch ID", executor.batch.ID, "action ID", executor.actionID)

	return nil
}

// ProcessMaxRetriesOnWasTransferProposedOnMultiversX checks if the retries on MultiversX were reached and increments the counter
func (executor *bridgeExecutor) ProcessMaxRetriesOnWasTransferProposedOnMultiversX() bool {
	if executor.retriesOnWasProposed < executor.maxRetriesOnWasProposed {
		executor.retriesOnWasProposed++
		return false
	}

	return true
}

// ResetRetriesOnWasTransferProposedOnMultiversX resets the number of retries on was transfer proposed
func (executor *bridgeExecutor) ResetRetriesOnWasTransferProposedOnMultiversX() {
	executor.retriesOnWasProposed = 0
}

// WasSetStatusProposedOnMultiversX checks if set status was proposed on MultiversX
func (executor *bridgeExecutor) WasSetStatusProposedOnMultiversX(ctx context.Context) (bool, error) {
	if executor.batch == nil {
		return false, ErrNilBatch
	}

	return executor.multiversXClient.WasProposedSetStatus(ctx, executor.batch)
}

// ProposeSetStatusOnMultiversX propose set status on MultiversX
func (executor *bridgeExecutor) ProposeSetStatusOnMultiversX(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	hash, err := executor.multiversXClient.ProposeSetStatus(ctx, executor.batch)
	if err != nil {
		return err
	}

	executor.log.Info("proposed set status", "hash", hash,
		"batch ID", executor.batch.ID)

	return nil
}

// WasActionSignedOnMultiversX returns true if the current relayer already signed the action
func (executor *bridgeExecutor) WasActionSignedOnMultiversX(ctx context.Context) (bool, error) {
	return executor.multiversXClient.WasSigned(ctx, executor.actionID)
}

// SignActionOnMultiversX calls the MultiversX client to generate and send the signature
func (executor *bridgeExecutor) SignActionOnMultiversX(ctx context.Context) error {
	hash, err := executor.multiversXClient.Sign(ctx, executor.actionID)
	if err != nil {
		return err
	}

	executor.log.Info("signed proposed transfer", "hash", hash, "action ID", executor.actionID)

	return nil
}

// ProcessQuorumReachedOnMultiversX returns true if the proposed transfer reached the set quorum
func (executor *bridgeExecutor) ProcessQuorumReachedOnMultiversX(ctx context.Context) (bool, error) {
	return executor.multiversXClient.QuorumReached(ctx, executor.actionID)
}

// WaitForTransferConfirmation waits for the confirmation of a transfer
func (executor *bridgeExecutor) WaitForTransferConfirmation(ctx context.Context) {
	wasPerformed := false
	for i := 0; i < splits && !wasPerformed; i++ {
		if executor.waitWithContextSucceeded(ctx) {
			wasPerformed, _ = executor.WasTransferPerformedOnPeerChain(ctx)
		}
	}
}

// WaitAndReturnFinalBatchStatuses waits for the statuses to be final
func (executor *bridgeExecutor) WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte {
	for i := 0; i < splits; i++ {
		if !executor.waitWithContextSucceeded(ctx) {
			return nil
		}

		statuses, err := executor.GetBatchStatusesFromPeerChain(ctx)
		if err != nil {
			executor.log.Debug("got message while fetching batch statuses", "message", err)
			continue
		}
		if len(statuses) == 0 {
			executor.log.Debug("no status available")
			continue
		}

		executor.log.Debug("bridgeExecutor.WaitAndReturnFinalBatchStatuses", "statuses", statuses)
		return statuses
	}

	return nil
}

func (executor *bridgeExecutor) waitWithContextSucceeded(ctx context.Context) bool {
	timer := time.NewTimer(executor.timeForWaitOnPeerClient / splits)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		executor.log.Debug("closing due to context expiration")
		return false
	case <-timer.C:
		return true
	}
}

// GetBatchStatusesFromPeerChain gets statuses for the batch
func (executor *bridgeExecutor) GetBatchStatusesFromPeerChain(ctx context.Context) ([]byte, error) {
	if executor.batch == nil {
		return nil, ErrNilBatch
	}

	statuses, err := executor.peerChainClient.GetTransactionsStatuses(ctx, executor.batch.ID)
	if err != nil {
		return nil, err
	}

	return statuses, nil
}

// WasActionPerformedOnMultiversX returns true if the action was already performed
func (executor *bridgeExecutor) WasActionPerformedOnMultiversX(ctx context.Context) (bool, error) {
	return executor.multiversXClient.WasExecuted(ctx, executor.actionID)
}

// PerformActionOnMultiversX sends the perform-action transaction on the MultiversX chain
func (executor *bridgeExecutor) PerformActionOnMultiversX(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	hash, err := executor.multiversXClient.PerformAction(ctx, executor.actionID, executor.batch)
	if err != nil {
		return err
	}

	executor.log.Info("sent perform action transaction", "hash", hash,
		"batch ID", executor.batch.ID, "action ID", executor.actionID)

	return nil
}

// ResolveNewDepositsStatuses resolves the new deposits statuses for batch
func (executor *bridgeExecutor) ResolveNewDepositsStatuses(numDeposits uint64) {
	executor.batch.ResolveNewDeposits(int(numDeposits))
}

// ProcessMaxQuorumRetriesOnMultiversX checks if the retries on MultiversX were reached and increments the counter
func (executor *bridgeExecutor) ProcessMaxQuorumRetriesOnMultiversX() bool {
	if executor.quorumRetriesOnMultiversX < executor.maxQuorumRetriesOnMultiversX {
		executor.quorumRetriesOnMultiversX++
		return false
	}

	return true
}

// ResetRetriesCountOnMultiversX resets the number of retries on MultiversX
func (executor *bridgeExecutor) ResetRetriesCountOnMultiversX() {
	executor.quorumRetriesOnMultiversX = 0
}

// GetAndStoreBatchFromPeerChain fetches and stores the batch from the peer client
func (executor *bridgeExecutor) GetAndStoreBatchFromPeerChain(ctx context.Context, nonce uint64) error {
	batch, isFinal, err := executor.peerChainClient.GetBatch(ctx, nonce)
	if err != nil {
		return err
	}

	isBatchInvalid := batch.ID != nonce || len(batch.Deposits) == 0 || !isFinal
	if isBatchInvalid {
		return fmt.Errorf("%w, requested nonce: %d, fetched nonce: %d, num deposits: %d, isFinal: %v",
			ErrFinalBatchNotFound, nonce, batch.ID, len(batch.Deposits), isFinal)
	}

	batch, err = executor.addBatchSCMetadata(ctx, batch)
	if err != nil {
		return err
	}
	executor.batch = batch

	return nil
}

// addBatchSCMetadata fetches the logs containing sc calls metadata for the current batch
func (executor *bridgeExecutor) addBatchSCMetadata(ctx context.Context, transfers *bridgeCore.TransferBatch) (*bridgeCore.TransferBatch, error) {
	if transfers == nil {
		return nil, ErrNilBatch
	}

	events, err := executor.peerChainClient.GetBatchSCMetadata(ctx, transfers.ID, int64(transfers.BlockNumber))
	if err != nil {
		return nil, err
	}

	for i, t := range transfers.Deposits {
		transfers.Deposits[i] = executor.addMetadataToTransfer(t, events)
	}

	return transfers, nil
}

func (executor *bridgeExecutor) addMetadataToTransfer(transfer *bridgeCore.DepositTransfer, events []*contract.ERC20SafeERC20SCDeposit) *bridgeCore.DepositTransfer {
	for _, event := range events {
		if event.DepositNonce.Uint64() == transfer.Nonce {
			processData(transfer, event.CallData)
			return transfer
		}
	}

	transfer.Data = []byte{bridgeCore.MissingDataProtocolMarker}
	transfer.DisplayableData = ""

	return transfer
}

func processData(transfer *bridgeCore.DepositTransfer, buff []byte) {
	transfer.Data = buff
	dataLen := len(transfer.Data)
	if dataLen == 0 {
		transfer.Data = []byte{bridgeCore.MissingDataProtocolMarker}
		transfer.DisplayableData = ""
		return
	}
	// this check is optional, but brings an optimisation to reduce the gas used in case of a bad callData
	if dataLen == 1 && buff[0] == bridgeCore.MissingDataProtocolMarker {
		return
	}

	// we have a data field, add the marker & the correct length
	transfer.DisplayableData = hex.EncodeToString(transfer.Data)
	buff32 := make([]byte, bridgeCore.Uint32ArgBytes)
	binary.BigEndian.PutUint32(buff32, uint32(dataLen))

	prefix := append([]byte{bridgeCore.DataPresentProtocolMarker}, buff32...)

	transfer.Data = append(prefix, transfer.Data...)
}

// WasTransferPerformedOnPeerChain returns true if the batch was performed on peer chain
func (executor *bridgeExecutor) WasTransferPerformedOnPeerChain(ctx context.Context) (bool, error) {
	if executor.batch == nil {
		return false, ErrNilBatch
	}

	return executor.peerChainClient.WasExecuted(ctx, executor.batch.ID)
}

// SignTransferOnPeerChain generates the message hash for batch and broadcast the signature
func (executor *bridgeExecutor) SignTransferOnPeerChain() error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	argLists := batchProcessor.ExtractListFromMvx(executor.batch)
	hash, err := executor.peerChainClient.GenerateMessageHash(argLists, executor.batch.ID)
	if err != nil {
		return err
	}

	executor.log.Info("generated message hash on peer chain", "hash", hash,
		"batch ID", executor.batch.ID)

	executor.msgHash = hash
	executor.peerChainClient.BroadcastSignatureForMessageHash(hash)
	return nil
}

// PerformTransferOnPeerChain transfers a batch to peer chain
func (executor *bridgeExecutor) PerformTransferOnPeerChain(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	quorumSize, err := executor.peerChainClient.GetQuorumSize(ctx)
	if err != nil {
		return err
	}

	executor.log.Debug("fetched quorum size", "quorum", quorumSize.Int64())

	argLists := batchProcessor.ExtractListFromMvx(executor.batch)

	executor.log.Info("executing transfer " + executor.batch.String())

	hash, err := executor.peerChainClient.ExecuteTransfer(ctx, executor.msgHash, argLists, executor.batch.ID, int(quorumSize.Int64()))
	if err != nil {
		return err
	}

	executor.log.Info("sent execute transfer", "hash", hash,
		"batch ID", executor.batch.ID)

	return nil
}

func (executor *bridgeExecutor) checkCumulatedTransfers(ctx context.Context, peerTokens [][]byte, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	for i, peerToken := range peerTokens {
		err := executor.balanceValidator.CheckToken(ctx, peerToken, mvxTokens[i], amounts[i], direction)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckAvailableTokens checks the available balances
func (executor *bridgeExecutor) CheckAvailableTokens(ctx context.Context, peerTokens [][]byte, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	peerTokens, mvxTokens, amounts = executor.getCumulatedTransfers(peerTokens, mvxTokens, amounts)

	return executor.checkCumulatedTransfers(ctx, peerTokens, mvxTokens, amounts, direction)
}

func (executor *bridgeExecutor) getCumulatedTransfers(peerTokens [][]byte, mvxTokens [][]byte, amounts []*big.Int) ([][]byte, [][]byte, []*big.Int) {
	cumulatedAmounts := make(map[string]*big.Int)
	tokenMap := make(map[string][]byte)
	uniqueTokensStr := make([]string, 0)
	uniqueConvertedTokens := make([][]byte, 0)

	for i, token := range peerTokens {
		tokenStr := string(token)
		existingValue, exists := cumulatedAmounts[tokenStr]
		if exists {
			existingValue.Add(existingValue, amounts[i])
			continue
		}

		cumulatedAmounts[tokenStr] = big.NewInt(0).Set(amounts[i]) // work on a new pointer
		tokenMap[tokenStr] = token
		uniqueTokensStr = append(uniqueTokensStr, tokenStr)
		uniqueConvertedTokens = append(uniqueConvertedTokens, mvxTokens[i])
	}

	uniqueTokens := make([][]byte, len(uniqueTokensStr))
	finalAmounts := make([]*big.Int, len(uniqueTokensStr))
	for i, tokenStr := range uniqueTokensStr {
		uniqueTokens[i] = tokenMap[tokenStr]
		finalAmounts[i] = cumulatedAmounts[tokenStr]
	}

	return uniqueTokens, uniqueConvertedTokens, finalAmounts
}

// ProcessQuorumReachedOnPeerChain returns true if the proposed transfer reached the set quorum
func (executor *bridgeExecutor) ProcessQuorumReachedOnPeerChain(ctx context.Context) (bool, error) {
	return executor.peerChainClient.IsQuorumReached(ctx, executor.msgHash)
}

// ProcessMaxQuorumRetriesOnPeerChain checks if the retries on peer chain were reached and increments the counter
func (executor *bridgeExecutor) ProcessMaxQuorumRetriesOnPeerChain() bool {
	if executor.quorumRetriesOnPeerClient < executor.maxQuorumRetriesOnPeerClient {
		executor.quorumRetriesOnPeerClient++
		return false
	}

	return true
}

// ResetRetriesCountOnPeerChain resets the number of retries on peer chain
func (executor *bridgeExecutor) ResetRetriesCountOnPeerChain() {
	executor.quorumRetriesOnPeerClient = 0
}

// ClearStoredP2PSignaturesForPeerChain deletes all stored P2P signatures used for peer chain client
func (executor *bridgeExecutor) ClearStoredP2PSignaturesForPeerChain() {
	executor.sigsHolder.ClearStoredSignatures()
	executor.log.Info("cleared stored P2P signatures")
}

// CheckMultiversXClientAvailability trigger a self availability check for the MultiversX client
func (executor *bridgeExecutor) CheckMultiversXClientAvailability(ctx context.Context) error {
	return executor.multiversXClient.CheckClientAvailability(ctx)
}

// CheckPeerClientAvailability trigger a self availability check for the peer client
func (executor *bridgeExecutor) CheckPeerClientAvailability(ctx context.Context) error {
	return executor.peerChainClient.CheckClientAvailability(ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *bridgeExecutor) IsInterfaceNil() bool {
	return executor == nil
}
