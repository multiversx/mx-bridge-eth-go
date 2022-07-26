package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ethToElrondName(t *testing.T) {
	assert.Equal(t, Ethereum.EvmCompatibleChainToElrondName(), "EthereumToElrond")
	assert.Equal(t, Bsc.EvmCompatibleChainToElrondName(), "BscToElrond")
}

func Test_elrondToEthName(t *testing.T) {
	assert.Equal(t, Ethereum.ElrondToEvmCompatibleChainName(), "ElrondToEthereum")
	assert.Equal(t, Bsc.ElrondToEvmCompatibleChainName(), "ElrondToBsc")
}

func Test_baseLogId(t *testing.T) {
	assert.Equal(t, Ethereum.BaseLogId(), "EthereumElrond-Base")
	assert.Equal(t, Bsc.BaseLogId(), "BscElrond-Base")
}

func Test_elrondClientLogId(t *testing.T) {
	assert.Equal(t, Ethereum.ElrondClientLogId(), "EthereumElrond-ElrondClient")
	assert.Equal(t, Bsc.ElrondClientLogId(), "BscElrond-ElrondClient")
}

func Test_elrondDataGetterLogId(t *testing.T) {
	assert.Equal(t, Ethereum.ElrondDataGetterLogId(), "EthereumElrond-ElrondDataGetter")
	assert.Equal(t, Bsc.ElrondDataGetterLogId(), "BscElrond-ElrondDataGetter")
}

func Test_ethClientLogId(t *testing.T) {
	assert.Equal(t, Ethereum.EvmCompatibleChainClientLogId(), "EthereumElrond-EthereumClient")
	assert.Equal(t, Bsc.EvmCompatibleChainClientLogId(), "BscElrond-BscClient")
}

func Test_elrondRoleProviderLogId(t *testing.T) {
	assert.Equal(t, Ethereum.ElrondRoleProviderLogId(), "EthereumElrond-ElrondRoleProvider")
	assert.Equal(t, Bsc.ElrondRoleProviderLogId(), "BscElrond-ElrondRoleProvider")
}

func Test_ethRoleProviderLogId(t *testing.T) {
	assert.Equal(t, Ethereum.EvmCompatibleChainRoleProviderLogId(), "EthereumElrond-EthereumRoleProvider")
	assert.Equal(t, Bsc.EvmCompatibleChainRoleProviderLogId(), "BscElrond-BscRoleProvider")
}

func Test_broadcasterLogId(t *testing.T) {
	assert.Equal(t, Ethereum.BroadcasterLogId(), "EthereumElrond-Broadcaster")
	assert.Equal(t, Bsc.BroadcasterLogId(), "BscElrond-Broadcaster")
}

func TestToLower(t *testing.T) {
	assert.Equal(t, Elrond.ToLower(), "elrond")
	assert.Equal(t, Ethereum.ToLower(), "ethereum")
	assert.Equal(t, Bsc.ToLower(), "bsc")
}
