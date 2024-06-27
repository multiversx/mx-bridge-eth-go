package ethmultiversx

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// splits - represent the number of times we split the maximum interval
// we wait for the transfer confirmation on Ethereum
const splits = 10
const minRetries = 1
const uint32ArgBytes = 4
const uint64ArgBytes = 8

// MissingDataProtocolMarker defines the marker for missing data (simple transfers)
const MissingDataProtocolMarker byte = 0x00

// DataPresentProtocolMarker defines the marker for existing data (transfers with SC calls)
const DataPresentProtocolMarker byte = 0x01

// ArgsBridgeExecutor is the arguments DTO struct used in both bridges
type ArgsBridgeExecutor struct {
	Log                          logger.Logger
	TopologyProvider             TopologyProvider
	MultiversXClient             MultiversXClient
	EthereumClient               EthereumClient
	TimeForWaitOnEthereum        time.Duration
	StatusHandler                core.StatusHandler
	SignaturesHolder             SignaturesHolder
	BalanceValidator             BalanceValidator
	MaxQuorumRetriesOnEthereum   uint64
	MaxQuorumRetriesOnMultiversX uint64
	MaxRestriesOnWasProposed     uint64
}

type bridgeExecutor struct {
	log                          logger.Logger
	topologyProvider             TopologyProvider
	multiversXClient             MultiversXClient
	ethereumClient               EthereumClient
	timeForWaitOnEthereum        time.Duration
	statusHandler                core.StatusHandler
	sigsHolder                   SignaturesHolder
	balanceValidator             BalanceValidator
	maxQuorumRetriesOnEthereum   uint64
	maxQuorumRetriesOnMultiversX uint64
	maxRetriesOnWasProposed      uint64

	batch                     *clients.TransferBatch
	actionID                  uint64
	msgHash                   common.Hash
	quorumRetriesOnEthereum   uint64
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
	if check.IfNil(args.EthereumClient) {
		return ErrNilEthereumClient
	}
	if check.IfNil(args.TopologyProvider) {
		return ErrNilTopologyProvider
	}
	if check.IfNil(args.StatusHandler) {
		return ErrNilStatusHandler
	}
	if args.TimeForWaitOnEthereum < durationLimit {
		return ErrInvalidDuration
	}
	if check.IfNil(args.SignaturesHolder) {
		return ErrNilSignaturesHolder
	}
	if check.IfNil(args.BalanceValidator) {
		return ErrNilBalanceValidator
	}
	if args.MaxQuorumRetriesOnEthereum < minRetries {
		return fmt.Errorf("%w for args.MaxQuorumRetriesOnEthereum, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.MaxQuorumRetriesOnEthereum, minRetries)
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
		ethereumClient:               args.EthereumClient,
		topologyProvider:             args.TopologyProvider,
		statusHandler:                args.StatusHandler,
		timeForWaitOnEthereum:        args.TimeForWaitOnEthereum,
		sigsHolder:                   args.SignaturesHolder,
		balanceValidator:             args.BalanceValidator,
		maxQuorumRetriesOnEthereum:   args.MaxQuorumRetriesOnEthereum,
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
func (executor *bridgeExecutor) GetBatchFromMultiversX(ctx context.Context) (*clients.TransferBatch, error) {
	batch, err := executor.multiversXClient.GetPending(ctx)
	if err == nil {
		executor.statusHandler.SetIntMetric(core.MetricNumBatches, int(batch.ID)-1)
	}
	return batch, err
}

// StoreBatchFromMultiversX saves the pending batch from MultiversX
func (executor *bridgeExecutor) StoreBatchFromMultiversX(batch *clients.TransferBatch) error {
	if batch == nil {
		return ErrNilBatch
	}

	executor.batch = batch
	return nil
}

// GetStoredBatch returns the stored batch
func (executor *bridgeExecutor) GetStoredBatch() *clients.TransferBatch {
	return executor.batch
}

// GetLastExecutedEthBatchIDFromMultiversX returns the last executed batch ID that is stored on the MultiversX SC
func (executor *bridgeExecutor) GetLastExecutedEthBatchIDFromMultiversX(ctx context.Context) (uint64, error) {
	batchID, err := executor.multiversXClient.GetLastExecutedEthBatchID(ctx)
	if err == nil {
		executor.statusHandler.SetIntMetric(core.MetricNumBatches, int(batchID))
	}
	return batchID, err
}

// VerifyLastDepositNonceExecutedOnEthereumBatch will check the deposit Nonces from the fetched batch from Ethereum client
func (executor *bridgeExecutor) VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error {
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
			wasPerformed, _ = executor.WasTransferPerformedOnEthereum(ctx)
		}
	}
}

// WaitAndReturnFinalBatchStatuses waits for the statuses to be final
func (executor *bridgeExecutor) WaitAndReturnFinalBatchStatuses(ctx context.Context) []byte {
	for i := 0; i < splits; i++ {
		if !executor.waitWithContextSucceeded(ctx) {
			return nil
		}

		statuses, err := executor.GetBatchStatusesFromEthereum(ctx)
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
	timer := time.NewTimer(executor.timeForWaitOnEthereum / splits)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		executor.log.Debug("closing due to context expiration")
		return false
	case <-timer.C:
		return true
	}
}

// GetBatchStatusesFromEthereum gets statuses for the batch
func (executor *bridgeExecutor) GetBatchStatusesFromEthereum(ctx context.Context) ([]byte, error) {
	if executor.batch == nil {
		return nil, ErrNilBatch
	}

	statuses, err := executor.ethereumClient.GetTransactionsStatuses(ctx, executor.batch.ID)
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

// GetAndStoreBatchFromEthereum fetches and stores the batch from the ethereum client
func (executor *bridgeExecutor) GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error {
	batch, err := executor.ethereumClient.GetBatch(ctx, nonce)
	if err != nil {
		return err
	}

	isBatchInvalid := batch.ID != nonce || len(batch.Deposits) == 0
	if isBatchInvalid {
		return fmt.Errorf("%w, requested nonce: %d, fetched nonce: %d, num deposits: %d",
			ErrBatchNotFound, nonce, batch.ID, len(batch.Deposits))
	}

	batch, err = executor.addBatchSCMetadata(ctx, batch)
	if err != nil {
		return err
	}
	executor.batch = batch

	return nil
}

// addBatchSCMetadata fetches the logs containing sc calls metadata for the current batch
func (executor *bridgeExecutor) addBatchSCMetadata(ctx context.Context, transfers *clients.TransferBatch) (*clients.TransferBatch, error) {
	if transfers == nil {
		return nil, ErrNilBatch
	}

	events, err := executor.ethereumClient.GetBatchSCMetadata(ctx, transfers.ID)
	if err != nil {
		return nil, err
	}

	for i, t := range transfers.Deposits {
		transfers.Deposits[i] = executor.addMetadataToTransfer(t, events)
	}

	return transfers, nil
}

func (executor *bridgeExecutor) addMetadataToTransfer(transfer *clients.DepositTransfer, events []*contract.SCExecProxyERC20SCDeposit) *clients.DepositTransfer {
	for _, event := range events {
		if event.DepositNonce == transfer.Nonce {
			transfer.Data = []byte(event.CallData)
			var err error
			transfer.DisplayableData, err = ConvertToDisplayableData(transfer.Data)
			if err != nil {
				executor.log.Warn("failed to convert call data to displayable data", "error", err)
			}

			return transfer
		}
	}

	transfer.Data = []byte{MissingDataProtocolMarker}
	transfer.DisplayableData = ""

	return transfer
}

// WasTransferPerformedOnEthereum returns true if the batch was performed on Ethereum
func (executor *bridgeExecutor) WasTransferPerformedOnEthereum(ctx context.Context) (bool, error) {
	if executor.batch == nil {
		return false, ErrNilBatch
	}

	return executor.ethereumClient.WasExecuted(ctx, executor.batch.ID)
}

// SignTransferOnEthereum generates the message hash for batch and broadcast the signature
func (executor *bridgeExecutor) SignTransferOnEthereum() error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	argLists := batchProcessor.ExtractListMvxToEth(executor.batch)
	hash, err := executor.ethereumClient.GenerateMessageHash(argLists, executor.batch.ID)
	if err != nil {
		return err
	}

	executor.log.Info("generated message hash on Ethereum", "hash", hash,
		"batch ID", executor.batch.ID)

	executor.msgHash = hash
	executor.ethereumClient.BroadcastSignatureForMessageHash(hash)
	return nil
}

// PerformTransferOnEthereum transfers a batch to Ethereum
func (executor *bridgeExecutor) PerformTransferOnEthereum(ctx context.Context) error {
	if executor.batch == nil {
		return ErrNilBatch
	}

	quorumSize, err := executor.ethereumClient.GetQuorumSize(ctx)
	if err != nil {
		return err
	}

	executor.log.Debug("fetched quorum size", "quorum", quorumSize.Int64())

	argLists := batchProcessor.ExtractListMvxToEth(executor.batch)

	executor.log.Info("executing transfer " + executor.batch.String())

	hash, err := executor.ethereumClient.ExecuteTransfer(ctx, executor.msgHash, argLists, executor.batch.ID, int(quorumSize.Int64()))
	if err != nil {
		return err
	}

	executor.log.Info("sent execute transfer", "hash", hash,
		"batch ID", executor.batch.ID)

	return nil
}

func (executor *bridgeExecutor) checkCumulatedTransfers(ctx context.Context, ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	for i, ethToken := range ethTokens {
		err := executor.balanceValidator.CheckToken(ctx, ethToken, mvxTokens[i], amounts[i], direction)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckAvailableTokens checks the available balances
func (executor *bridgeExecutor) CheckAvailableTokens(ctx context.Context, ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int, direction batchProcessor.Direction) error {
	ethTokens, mvxTokens, amounts = executor.getCumulatedTransfers(ethTokens, mvxTokens, amounts)

	return executor.checkCumulatedTransfers(ctx, ethTokens, mvxTokens, amounts, direction)
}

func (executor *bridgeExecutor) getCumulatedTransfers(ethTokens []common.Address, mvxTokens [][]byte, amounts []*big.Int) ([]common.Address, [][]byte, []*big.Int) {
	cumulatedAmounts := make(map[common.Address]*big.Int)
	uniqueTokens := make([]common.Address, 0)
	uniqueConvertedTokens := make([][]byte, 0)

	for i, token := range ethTokens {
		existingValue, exists := cumulatedAmounts[token]
		if exists {
			existingValue.Add(existingValue, amounts[i])
			continue
		}

		cumulatedAmounts[token] = big.NewInt(0).Set(amounts[i]) // work on a new pointer
		uniqueTokens = append(uniqueTokens, token)
		uniqueConvertedTokens = append(uniqueConvertedTokens, mvxTokens[i])
	}

	finalAmounts := make([]*big.Int, len(uniqueTokens))
	for i, token := range uniqueTokens {
		finalAmounts[i] = cumulatedAmounts[token]
	}

	return uniqueTokens, uniqueConvertedTokens, finalAmounts
}

// ProcessQuorumReachedOnEthereum returns true if the proposed transfer reached the set quorum
func (executor *bridgeExecutor) ProcessQuorumReachedOnEthereum(ctx context.Context) (bool, error) {
	return executor.ethereumClient.IsQuorumReached(ctx, executor.msgHash)
}

// ProcessMaxQuorumRetriesOnEthereum checks if the retries on Ethereum were reached and increments the counter
func (executor *bridgeExecutor) ProcessMaxQuorumRetriesOnEthereum() bool {
	if executor.quorumRetriesOnEthereum < executor.maxQuorumRetriesOnEthereum {
		executor.quorumRetriesOnEthereum++
		return false
	}

	return true
}

// ResetRetriesCountOnEthereum resets the number of retries on Ethereum
func (executor *bridgeExecutor) ResetRetriesCountOnEthereum() {
	executor.quorumRetriesOnEthereum = 0
}

// ClearStoredP2PSignaturesForEthereum deletes all stored P2P signatures used for Ethereum client
func (executor *bridgeExecutor) ClearStoredP2PSignaturesForEthereum() {
	executor.sigsHolder.ClearStoredSignatures()
	executor.log.Info("cleared stored P2P signatures")
}

// CheckMultiversXClientAvailability trigger a self availability check for the MultiversX client
func (executor *bridgeExecutor) CheckMultiversXClientAvailability(ctx context.Context) error {
	return executor.multiversXClient.CheckClientAvailability(ctx)
}

// CheckEthereumClientAvailability trigger a self availability check for the Ethereum client
func (executor *bridgeExecutor) CheckEthereumClientAvailability(ctx context.Context) error {
	return executor.ethereumClient.CheckClientAvailability(ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *bridgeExecutor) IsInterfaceNil() bool {
	return executor == nil
}

// ConvertToDisplayableData
/** @param callData The encoded data specifying the cross-chain call details. The expected format is:
*        0x01 + endpoint_name_length (4 bytes) + endpoint_name + gas_limit (8 bytes) +
*        num_arguments_length (4 bytes) + [argument_length (4 bytes) + argument]...
*        This payload includes the endpoint name, gas limit for the execution, and the arguments for the call.
 */
func ConvertToDisplayableData(callData []byte) (string, error) {
	if len(callData) == 0 {
		return "", fmt.Errorf("callData too short for protocol indicator")
	}

	marker := callData[0]
	callData = callData[1:]

	switch marker {
	case MissingDataProtocolMarker:
		return "", nil
	case DataPresentProtocolMarker:
		return convertBytesToDisplayableData(callData)
	default:
		return "", fmt.Errorf("callData unexpected protocol indicator: %d", marker)
	}
}

func convertBytesToDisplayableData(callData []byte) (string, error) {
	callData, endpointName, err := extractString(callData)
	if err != nil {
		return "", fmt.Errorf("%w for endpoint", err)
	}

	callData, gasLimit, err := extractGasLimit(callData)
	if err != nil {
		return "", err
	}

	callData, numArgumentsLength, err := extractArgumentsLen(callData)
	if err != nil {
		return "", err
	}

	arguments := make([]string, 0)
	for i := 0; i < numArgumentsLength; i++ {
		var argument string
		callData, argument, err = extractString(callData)
		if err != nil {
			return "", fmt.Errorf("%w for argument %d", err, i)
		}

		arguments = append(arguments, argument)
	}

	return fmt.Sprintf("Endpoint: %s, Gas: %d, Arguments: %s", endpointName, gasLimit, strings.Join(arguments, "@")), nil
}

func extractString(callData []byte) ([]byte, string, error) {
	// Ensure there's enough length for the 4 bytes for length
	if len(callData) < uint32ArgBytes {
		return nil, "", fmt.Errorf("callData too short while extracting the length")
	}
	argumentLength := int(binary.BigEndian.Uint32(callData[:uint32ArgBytes]))
	callData = callData[uint32ArgBytes:] // remove the len bytes

	// Check for the argument data
	if len(callData) < argumentLength {
		return nil, "", fmt.Errorf("callData too short while extracting the string data")
	}
	endpointName := string(callData[:argumentLength])
	callData = callData[argumentLength:] // remove the string bytes

	return callData, endpointName, nil
}

func extractGasLimit(callData []byte) ([]byte, uint64, error) {
	// Check for gas limit
	if len(callData) < uint64ArgBytes { // 8 bytes for gas limit length
		return nil, 0, fmt.Errorf("callData too short for gas limit length")
	}

	gasLimitLength := int(binary.BigEndian.Uint64(callData[:uint64ArgBytes]))
	callData = callData[uint64ArgBytes:]

	// Check for gas limit
	if len(callData) < gasLimitLength {
		return nil, 0, fmt.Errorf("callData too short for gas limit")
	}

	gasLimitBytes := append(bytes.Repeat([]byte{0x00}, 8-int(gasLimitLength)), callData[:gasLimitLength]...)
	gasLimit := binary.BigEndian.Uint64(gasLimitBytes)
	callData = callData[gasLimitLength:] // remove the gas limit bytes

	return callData, gasLimit, nil
}

func extractArgumentsLen(callData []byte) ([]byte, int, error) {
	// Ensure there's enough length for the 4 bytes for endpoint name length
	if len(callData) < uint32ArgBytes {
		return nil, 0, fmt.Errorf("callData too short for numArguments length")
	}
	length := int(binary.BigEndian.Uint32(callData[:uint32ArgBytes]))
	callData = callData[uint32ArgBytes:] // remove the len bytes

	return callData, length, nil
}
