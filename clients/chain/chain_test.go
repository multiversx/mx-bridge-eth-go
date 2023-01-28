package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ethToMultiversXName(t *testing.T) {
	assert.Equal(t, Ethereum.EvmCompatibleChainToMultiversXName(), "EthereumToMultiversX")
	assert.Equal(t, Bsc.EvmCompatibleChainToMultiversXName(), "BscToMultiversX")
}

func Test_multiversXToEthName(t *testing.T) {
	assert.Equal(t, Ethereum.MultiversXToEvmCompatibleChainName(), "MultiversXToEthereum")
	assert.Equal(t, Bsc.MultiversXToEvmCompatibleChainName(), "MultiversXToBsc")
}

func Test_baseLogId(t *testing.T) {
	assert.Equal(t, Ethereum.BaseLogId(), "EthereumMultiversX-Base")
	assert.Equal(t, Bsc.BaseLogId(), "BscMultiversX-Base")
}

func Test_multiversXClientLogId(t *testing.T) {
	assert.Equal(t, Ethereum.MultiversXClientLogId(), "EthereumMultiversX-MultiversXClient")
	assert.Equal(t, Bsc.MultiversXClientLogId(), "BscMultiversX-MultiversXClient")
}

func Test_multiversXDataGetterLogId(t *testing.T) {
	assert.Equal(t, Ethereum.MultiversXDataGetterLogId(), "EthereumMultiversX-MultiversXDataGetter")
	assert.Equal(t, Bsc.MultiversXDataGetterLogId(), "BscMultiversX-MultiversXDataGetter")
}

func Test_ethClientLogId(t *testing.T) {
	assert.Equal(t, Ethereum.EvmCompatibleChainClientLogId(), "EthereumMultiversX-EthereumClient")
	assert.Equal(t, Bsc.EvmCompatibleChainClientLogId(), "BscMultiversX-BscClient")
}

func Test_multiversXRoleProviderLogId(t *testing.T) {
	assert.Equal(t, Ethereum.MultiversXRoleProviderLogId(), "EthereumMultiversX-MultiversXRoleProvider")
	assert.Equal(t, Bsc.MultiversXRoleProviderLogId(), "BscMultiversX-MultiversXRoleProvider")
}

func Test_ethRoleProviderLogId(t *testing.T) {
	assert.Equal(t, Ethereum.EvmCompatibleChainRoleProviderLogId(), "EthereumMultiversX-EthereumRoleProvider")
	assert.Equal(t, Bsc.EvmCompatibleChainRoleProviderLogId(), "BscMultiversX-BscRoleProvider")
}

func Test_broadcasterLogId(t *testing.T) {
	assert.Equal(t, Ethereum.BroadcasterLogId(), "EthereumMultiversX-Broadcaster")
	assert.Equal(t, Bsc.BroadcasterLogId(), "BscMultiversX-Broadcaster")
}

func TestToLower(t *testing.T) {
	assert.Equal(t, MultiversX.ToLower(), "msx")
	assert.Equal(t, Ethereum.ToLower(), "ethereum")
	assert.Equal(t, Bsc.ToLower(), "bsc")
}
