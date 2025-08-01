package framework

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/factory"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

type EthRelayerAdapter struct {
	bc factory.BridgeComponents
}

func NewEthRelayerAdapter(bc factory.BridgeComponents) *EthRelayerAdapter {
	return &EthRelayerAdapter{bc: bc}
}

func (a *EthRelayerAdapter) Start() error {
	return a.bc.Start()
}

func (a *EthRelayerAdapter) Close() error {
	return a.bc.Close()
}

func (a *EthRelayerAdapter) MultiversXRelayerAddress() sdkCore.AddressHandler {
	return a.bc.MultiversXRelayerAddress()
}

func (a *EthRelayerAdapter) PeerChainRelayerAddress() common.Address {
	addrHex := a.bc.PeerChainRelayerAddress()
	return common.HexToAddress(addrHex)
}

// SuiRelayerAdapter wraps factory.BridgeComponents to implement framework.Relayer
type SuiRelayerAdapter struct {
	bc factory.BridgeComponents
}

func NewSuiRelayerAdapter(bc factory.BridgeComponents) *SuiRelayerAdapter {
	return &SuiRelayerAdapter{bc: bc}
}

func (a *SuiRelayerAdapter) Start() error {
	return a.bc.Start()
}

func (a *SuiRelayerAdapter) Close() error {
	return a.bc.Close()
}

func (a *SuiRelayerAdapter) MultiversXRelayerAddress() sdkCore.AddressHandler {
	return a.bc.MultiversXRelayerAddress()
}

func (a *SuiRelayerAdapter) PeerChainRelayerAddress() common.Address {
	// Not applicable for Sui, return zero address
	return common.Address{}
}
