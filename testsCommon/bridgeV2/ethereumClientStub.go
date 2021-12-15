package bridgeV2

import (
    "context"
    "math/big"

    "github.com/ElrondNetwork/elrond-eth-bridge/clients"
    "github.com/ethereum/go-ethereum/common"
)

// EthereumClientStub -
type EthereumClientStub struct {
    GetBatchCalled                             func(ctx context.Context, nonce uint64) (*clients.TransferBatch, error)
    WasExecutedCalled                          func(ctx context.Context, batchID uint64) (bool, error)
    GenerateMessageHashCalled                  func(batch *clients.TransferBatch) (common.Hash, error)
    BroadcastSignatureForMessageHashCalled     func(msgHash common.Hash)
    ExecuteTransferCalled                      func(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error)
    GetMaxNumberOfRetriesOnQuorumReachedCalled func() uint64
    GetQuorumSizeCalled                        func() (*big.Int, error)
    IsQuorumReachedCalled                      func() (bool, error)
}

func (stub *EthereumClientStub) GetQuorumSize(ctx context.Context) (*big.Int, error) {
    if stub.GetQuorumSizeCalled != nil {
        return stub.GetQuorumSizeCalled()
    }

    return nil, errNotImplemented
}

func (stub *EthereumClientStub) IsQuorumReached(ctx context.Context, msgHash common.Hash) (bool, error) {
    if stub.IsQuorumReachedCalled != nil {
        return stub.IsQuorumReachedCalled()
    }

    return false, errNotImplemented
}

// GetBatch -
func (stub *EthereumClientStub) GetBatch(ctx context.Context, nonce uint64) (*clients.TransferBatch, error) {
    if stub.GetBatchCalled != nil {
        return stub.GetBatchCalled(ctx, nonce)
    }

    return nil, errNotImplemented
}

//WasExecuted -
func (stub *EthereumClientStub) WasExecuted(ctx context.Context, batchID uint64) (bool, error) {
    if stub.WasExecutedCalled != nil {
        return stub.WasExecutedCalled(ctx, batchID)
    }

    return false, errNotImplemented
}

//GenerateMessageHash -
func (stub *EthereumClientStub) GenerateMessageHash(batch *clients.TransferBatch) (common.Hash, error) {
    if stub.GenerateMessageHashCalled != nil {
        return stub.GenerateMessageHashCalled(batch)
    }

    return common.Hash{}, errNotImplemented
}

//BroadcastSignatureForMessageHash -
func (stub *EthereumClientStub) BroadcastSignatureForMessageHash(msgHash common.Hash) {
    if stub.BroadcastSignatureForMessageHashCalled != nil {
        stub.BroadcastSignatureForMessageHashCalled(msgHash)
    }
}

//ExecuteTransfer -
func (stub *EthereumClientStub) ExecuteTransfer(ctx context.Context, msgHash common.Hash, batch *clients.TransferBatch, quorum int) (string, error) {
    if stub.ExecuteTransferCalled != nil {
        return stub.ExecuteTransferCalled(ctx, msgHash, batch, quorum)
    }

    return "", errNotImplemented
}

// GetMaxNumberOfRetriesOnQuorumReached -
func (stub *EthereumClientStub) GetMaxNumberOfRetriesOnQuorumReached() uint64 {
    if stub.GetMaxNumberOfRetriesOnQuorumReachedCalled != nil {
        return stub.GetMaxNumberOfRetriesOnQuorumReachedCalled()
    }

    return 0
}

//IsInterfaceNil -
func (stub *EthereumClientStub) IsInterfaceNil() bool {
    return stub == nil
}
