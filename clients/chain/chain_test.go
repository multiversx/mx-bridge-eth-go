package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_peerChainToMultiversXName(t *testing.T) {
	assert.Equal(t, "EthereumToMultiversX", Ethereum.PeerChainToMultiversXName())
	assert.Equal(t, "BscToMultiversX", Bsc.PeerChainToMultiversXName())
	assert.Equal(t, "SuiToMultiversX", Sui.PeerChainToMultiversXName())
}

func Test_multiversXToPeerChainName(t *testing.T) {
	assert.Equal(t, "MultiversXToEthereum", Ethereum.MultiversXToPeerChainName())
	assert.Equal(t, "MultiversXToBsc", Bsc.MultiversXToPeerChainName())
	assert.Equal(t, "MultiversXToSui", Sui.MultiversXToPeerChainName())
}

func Test_baseLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-Base", Ethereum.BaseLogId())
	assert.Equal(t, "BscMultiversX-Base", Bsc.BaseLogId())
	assert.Equal(t, "SuiMultiversX-Base", Sui.BaseLogId())
}

func Test_multiversXClientLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-MultiversXClient", Ethereum.MultiversXClientLogId())
	assert.Equal(t, "BscMultiversX-MultiversXClient", Bsc.MultiversXClientLogId())
	assert.Equal(t, "SuiMultiversX-MultiversXClient", Sui.MultiversXClientLogId())
}

func Test_multiversXDataGetterLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-MultiversXDataGetter", Ethereum.MultiversXDataGetterLogId())
	assert.Equal(t, "BscMultiversX-MultiversXDataGetter", Bsc.MultiversXDataGetterLogId())
	assert.Equal(t, "SuiMultiversX-MultiversXDataGetter", Sui.MultiversXDataGetterLogId())
}

func Test_peerChainClientLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-EthereumClient", Ethereum.PeerChainClientLogId())
	assert.Equal(t, "BscMultiversX-BscClient", Bsc.PeerChainClientLogId())
	assert.Equal(t, "SuiMultiversX-SuiClient", Sui.PeerChainClientLogId())
}

func Test_multiversXRoleProviderLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-MultiversXRoleProvider", Ethereum.MultiversXRoleProviderLogId())
	assert.Equal(t, "BscMultiversX-MultiversXRoleProvider", Bsc.MultiversXRoleProviderLogId())
	assert.Equal(t, "SuiMultiversX-MultiversXRoleProvider", Sui.MultiversXRoleProviderLogId())
}

func Test_peerChainRoleProviderLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-EthereumRoleProvider", Ethereum.PeerChainRoleProviderLogId())
	assert.Equal(t, "BscMultiversX-BscRoleProvider", Bsc.PeerChainRoleProviderLogId())
	assert.Equal(t, "SuiMultiversX-SuiRoleProvider", Sui.PeerChainRoleProviderLogId())
}

func Test_broadcasterLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-Broadcaster", Ethereum.BroadcasterLogId())
	assert.Equal(t, "BscMultiversX-Broadcaster", Bsc.BroadcasterLogId())
	assert.Equal(t, "SuiMultiversX-Broadcaster", Sui.BroadcasterLogId())
}

func TestToLower(t *testing.T) {
	assert.Equal(t, "msx", MultiversX.ToLower())
	assert.Equal(t, "ethereum", Ethereum.ToLower())
	assert.Equal(t, "bsc", Bsc.ToLower())
	assert.Equal(t, "sui", Sui.ToLower())
}
