package v2

import (
	"context"
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// ArgsEthToElrondBridgeExecutor is the arguments DTO struct used in the ethToElrondBridgeExecutor constructor
type ArgsEthToElrondBridgeExecutor struct {
	Log              logger.Logger
	TopologyProvider TopologyProvider
	ElrondClient     ElrondClient
	EthereumClient   EthereumClient
	StatusHandler    core.StatusHandler
}

type ethToElrondBridgeExecutor struct {
	log              logger.Logger
	topologyProvider TopologyProvider
	elrondClient     ElrondClient
	ethereumClient   EthereumClient
	batch            *clients.TransferBatch
	actionID         uint64
	retriesOnElrond  uint64
	statusHandler    core.StatusHandler
}

// NewEthToElrondBridgeExecutor will create a bridge executor for the Ethereum -> Elrond flow
func NewEthToElrondBridgeExecutor(args ArgsEthToElrondBridgeExecutor) (*ethToElrondBridgeExecutor, error) {
	if check.IfNil(args.Log) {
		return nil, errNilLogger
	}
	if check.IfNil(args.ElrondClient) {
		return nil, errNilElrondClient
	}
	if check.IfNil(args.EthereumClient) {
		return nil, errNilEthereumClient
	}
	if check.IfNil(args.TopologyProvider) {
		return nil, errNilTopologyProvider
	}
	if check.IfNil(args.StatusHandler) {
		return nil, errNilStatusHandler
	}

	return &ethToElrondBridgeExecutor{
		log:              args.Log,
		topologyProvider: args.TopologyProvider,
		elrondClient:     args.ElrondClient,
		ethereumClient:   args.EthereumClient,
		statusHandler:    args.StatusHandler,
	}, nil
}

// GetLogger returns the logger implementation
func (executor *ethToElrondBridgeExecutor) GetLogger() logger.Logger {
	return executor.log
}

// MyTurnAsLeader returns true if the current relayer node is the leader
func (executor *ethToElrondBridgeExecutor) MyTurnAsLeader() bool {
	return executor.topologyProvider.MyTurnAsLeader()
}

// GetAndStoreActionIDFromElrond fetches the action ID by using the stored batch and will return and store the action ID
func (executor *ethToElrondBridgeExecutor) GetAndStoreActionIDFromElrond(ctx context.Context) (uint64, error) {
	if executor.batch == nil {
		return 0, errNilBatch
	}

	actionID, err := executor.elrondClient.GetActionIDForProposeTransfer(ctx, executor.batch)
	if err != nil {
		return 0, err
	}

	executor.actionID = actionID

	return actionID, nil
}

// TODO(next PR) use & integrate this
func (executor *ethToElrondBridgeExecutor) setExecutionMessageInStatusHandler(level logger.LogLevel, message string, extras ...interface{}) {
	msg := fmt.Sprintf("%s: %s", level, message)
	for i := 0; i < len(extras)-1; i += 2 {
		msg += fmt.Sprintf(" %s = %s", convertObjectToString(extras[i]), convertObjectToString(extras[i+1]))
	}

	executor.statusHandler.SetStringMetric(core.MetricLastError, msg)
}

// GetStoredActionID will return the stored action ID
func (executor *ethToElrondBridgeExecutor) GetStoredActionID() uint64 {
	return executor.actionID
}

// GetAndStoreBatchFromEthereum will fetch and store the batch from the ethereum client
func (executor *ethToElrondBridgeExecutor) GetAndStoreBatchFromEthereum(ctx context.Context, nonce uint64) error {
	batch, err := executor.ethereumClient.GetBatch(ctx, nonce)
	// TODO add error filtering here
	if err != nil {
		return err
	}

	executor.batch = batch

	return nil
}

// GetStoredBatch returns the stored batch
func (executor *ethToElrondBridgeExecutor) GetStoredBatch() *clients.TransferBatch {
	return executor.batch
}

// GetLastExecutedEthBatchIDFromElrond returns the last executed batch ID that is stored on the Elrond SC
func (executor *ethToElrondBridgeExecutor) GetLastExecutedEthBatchIDFromElrond(ctx context.Context) (uint64, error) {
	return executor.elrondClient.GetLastExecutedEthBatchID(ctx)
}

// VerifyLastDepositNonceExecutedOnEthereumBatch will check the deposit nonces from the fetched batch from Ethereum client
func (executor *ethToElrondBridgeExecutor) VerifyLastDepositNonceExecutedOnEthereumBatch(ctx context.Context) error {
	if executor.batch == nil {
		return errNilBatch
	}

	lastNonce, err := executor.elrondClient.GetLastExecutedEthTxID(ctx)
	if err != nil {
		return err
	}

	return executor.verifyDepositNonces(lastNonce)
}

func (executor *ethToElrondBridgeExecutor) verifyDepositNonces(lastNonce uint64) error {
	startNonce := lastNonce + 1
	for _, dt := range executor.batch.Deposits {
		if dt.Nonce != startNonce {
			return fmt.Errorf("%w for deposit %s, expected: %d", errInvalidDepositNonce, dt.String(), startNonce)
		}

		startNonce++
	}

	return nil
}

// WasTransferProposedOnElrond checks if the transfer was proposed on Elrond
func (executor *ethToElrondBridgeExecutor) WasTransferProposedOnElrond(ctx context.Context) (bool, error) {
	if executor.batch == nil {
		return false, errNilBatch
	}

	return executor.elrondClient.WasProposedTransfer(ctx, executor.batch)
}

// ProposeTransferOnElrond will propose the transfer on Elrond
func (executor *ethToElrondBridgeExecutor) ProposeTransferOnElrond(ctx context.Context) error {
	if executor.batch == nil {
		return errNilBatch
	}

	hash, err := executor.elrondClient.ProposeTransfer(ctx, executor.batch)
	if err != nil {
		return err
	}

	executor.log.Info("proposed transfer", "hash", hash,
		"batch ID", executor.batch.ID, "action ID", executor.actionID)

	return nil
}

// WasProposedTransferSignedOnElrond returns true if the current relayer already signed the proposed transfer
func (executor *ethToElrondBridgeExecutor) WasProposedTransferSignedOnElrond(ctx context.Context) (bool, error) {
	return executor.elrondClient.WasExecuted(ctx, executor.actionID)
}

// SignProposedTransferOnElrond will call the Elrond client to generate and send the signature
func (executor *ethToElrondBridgeExecutor) SignProposedTransferOnElrond(ctx context.Context) error {
	hash, err := executor.elrondClient.Sign(ctx, executor.actionID)
	if err != nil {
		return err
	}

	executor.log.Info("signed proposed transfer", "hash", hash, "action ID", executor.actionID)

	return nil
}

// IsQuorumReachedOnElrond will return true if the proposed transfer reached the set quorum
func (executor *ethToElrondBridgeExecutor) IsQuorumReachedOnElrond(ctx context.Context) (bool, error) {
	return executor.elrondClient.QuorumReached(ctx, executor.actionID)
}

// WasActionIDPerformedOnElrond will return true if the action ID was already performed
func (executor *ethToElrondBridgeExecutor) WasActionIDPerformedOnElrond(ctx context.Context) (bool, error) {
	return executor.elrondClient.WasExecuted(ctx, executor.actionID)
}

// PerformActionIDOnElrond will send the perform-action transaction on the Elrond chain
func (executor *ethToElrondBridgeExecutor) PerformActionIDOnElrond(ctx context.Context) error {
	if executor.batch == nil {
		return errNilBatch
	}

	hash, err := executor.elrondClient.PerformAction(ctx, executor.actionID, executor.batch)
	if err != nil {
		return err
	}

	executor.log.Info("sent perform action transaction", "hash", hash,
		"batch ID", executor.batch.ID, "action ID", executor.actionID)

	return nil
}

// ProcessMaxRetriesOnElrond checks if the retries on waiting were reached and increments the counter
func (executor *ethToElrondBridgeExecutor) ProcessMaxRetriesOnElrond() bool {
	maxNumberOfRetries := executor.elrondClient.GetMaxNumberOfRetriesOnQuorumReached()
	if executor.retriesOnElrond < maxNumberOfRetries {
		executor.retriesOnElrond++
		return false
	}

	return true
}

// ResetRetriesCountOnElrond resets the number of retries
func (executor *ethToElrondBridgeExecutor) ResetRetriesCountOnElrond() {
	executor.retriesOnElrond = 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *ethToElrondBridgeExecutor) IsInterfaceNil() bool {
	return executor == nil
}
