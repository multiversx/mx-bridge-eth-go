package mock

import (
	"context"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
)

type BridgeStub struct {
	pendingBatchCallIndex          int
	pendingBatches                 []*bridge.Batch
	wasProposedTransfer            bool
	lastProposedBatch              *bridge.Batch
	lastWasProposedTransferBatchId bridge.BatchId
	lastSignedActionId             bridge.ActionId
	signersCount                   uint
	lastExecutedActionId           bridge.ActionId
	wasExecuted                    bool
	proposeTransferActionId        bridge.ActionId
	proposeTransferError           error
	proposedStatusBatch            *bridge.Batch
	proposeSetStatusActionId       bridge.ActionId

	proposeSetStatusMutex sync.Mutex
	proposeTransferMutex  sync.Mutex
	signMutex             sync.Mutex
	executeMutex          sync.Mutex
}

func (b BridgeStub) GetPending(ctx context.Context) *bridge.Batch {
	panic("implement me")
}

func (b BridgeStub) ProposeSetStatus(ctx context.Context, batch *bridge.Batch) {
	panic("implement me")
}

func (b BridgeStub) ProposeTransfer(ctx context.Context, batch *bridge.Batch) (string, error) {
	panic("implement me")
}

func (b BridgeStub) WasProposedTransfer(ctx context.Context, batch *bridge.Batch) bool {
	panic("implement me")
}

func (b BridgeStub) GetActionIdForProposeTransfer(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
	panic("implement me")
}

func (b BridgeStub) WasProposedSetStatus(ctx context.Context, batch *bridge.Batch) bool {
	panic("implement me")
}

func (b BridgeStub) GetActionIdForSetStatusOnPendingTransfer(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
	panic("implement me")
}

func (b BridgeStub) WasExecuted(ctx context.Context, id bridge.ActionId, id2 bridge.BatchId) bool {
	panic("implement me")
}

func (b BridgeStub) Sign(ctx context.Context, id bridge.ActionId) (string, error) {
	panic("implement me")
}

func (b BridgeStub) Execute(ctx context.Context, id bridge.ActionId, batch *bridge.Batch) (string, error) {
	panic("implement me")
}

func (b BridgeStub) SignersCount(ctx context.Context, id bridge.ActionId) uint {
	panic("implement me")
}
