package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ethToMultiversXName(t *testing.T) {
	assert.Equal(t, "EthereumToMultiversX", Ethereum.EvmCompatibleChainToMultiversXName())
	assert.Equal(t, "BscToMultiversX", Bsc.EvmCompatibleChainToMultiversXName())
}

func Test_multiversXToEthName(t *testing.T) {
	assert.Equal(t, "MultiversXToEthereum", Ethereum.MultiversXToEvmCompatibleChainName())
	assert.Equal(t, "MultiversXToBsc", Bsc.MultiversXToEvmCompatibleChainName())
}

func Test_baseLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-Base", Ethereum.BaseLogId())
	assert.Equal(t, "BscMultiversX-Base", Bsc.BaseLogId())
}

func Test_multiversXClientLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-MultiversXClient", Ethereum.MultiversXClientLogId())
	assert.Equal(t, "BscMultiversX-MultiversXClient", Bsc.MultiversXClientLogId())
}

func Test_multiversXDataGetterLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-MultiversXDataGetter", Ethereum.MultiversXDataGetterLogId())
	assert.Equal(t, "BscMultiversX-MultiversXDataGetter", Bsc.MultiversXDataGetterLogId())
}

func Test_ethClientLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-EthereumClient", Ethereum.EvmCompatibleChainClientLogId())
	assert.Equal(t, "BscMultiversX-BscClient", Bsc.EvmCompatibleChainClientLogId())
}

func Test_multiversXRoleProviderLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-MultiversXRoleProvider", Ethereum.MultiversXRoleProviderLogId())
	assert.Equal(t, "BscMultiversX-MultiversXRoleProvider", Bsc.MultiversXRoleProviderLogId())
}

func Test_ethRoleProviderLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-EthereumRoleProvider", Ethereum.EvmCompatibleChainRoleProviderLogId())
	assert.Equal(t, "BscMultiversX-BscRoleProvider", Bsc.EvmCompatibleChainRoleProviderLogId())
}

func Test_broadcasterLogId(t *testing.T) {
	assert.Equal(t, "EthereumMultiversX-Broadcaster", Ethereum.BroadcasterLogId())
	assert.Equal(t, "BscMultiversX-Broadcaster", Bsc.BroadcasterLogId())
}

func TestToLower(t *testing.T) {
	assert.Equal(t, "msx", MultiversX.ToLower())
	assert.Equal(t, "ethereum", Ethereum.ToLower())
	assert.Equal(t, "bsc", Bsc.ToLower())
}
