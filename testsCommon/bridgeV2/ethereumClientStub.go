package bridgeV2

import (
    "context"

    "github.com/ElrondNetwork/elrond-eth-bridge/clients"
    "github.com/ethereum/go-ethereum/common"
)

type EthereumClientStub struct {
    GetBatchCalled func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error)
    WasExecutedCalled func(ctx context.Context, batchID uint64) (bool, error)
    GenerateMessageHashCalled func(batch *clients.TransferBatch) (common.Hash, error)

    BroadcastSignatureForMessageHashCalled func(msgHash common.Hash)
    ExecuteTransferCalled func(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error)
}

func (stub *EthereumClientStub) GetBatch(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
    if stub.GetBatchCalled != nil {
        return stub.GetBatchCalled(ctx, nonce)
    }

    return nil, errNotImplemented
}

func (stub *EthereumClientStub) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
    if stub.WasExecutedCalled != nil {
        return stub.WasExecutedCalled(ctx, batchID)
    }

    return false, errNotImplemented
}

func (stub *EthereumClientStub) GenerateMessageHash(batch *clients.TransferBatch) (common.Hash, error) {
    if stub.GenerateMessageHashCalled != nil {
        return stub.GenerateMessageHashCalled(batch)
    }

    return common.Hash{}, errNotImplemented
}

func (stub *EthereumClientStub) BroadcastSignatureForMessageHash(msgHash common.Hash) {
    if stub.BroadcastSignatureForMessageHashCalled != nil {
        stub.BroadcastSignatureForMessageHashCalled(msgHash)
    }
}

func (stub *EthereumClientStub) ExecuteTransfer(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error) {
    if stub.ExecuteTransferCalled != nil {
        return stub.ExecuteTransferCalled(ctx, msgHash, batch, quorum)
    }

    return "", errNotImplemented
}

func (stub *EthereumClientStub) IsInterfaceNil() bool {
    return stub == nil
}
