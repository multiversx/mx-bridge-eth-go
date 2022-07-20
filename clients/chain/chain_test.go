package chain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ethToElrondName(t *testing.T) {
	assert.Equal(t, Ethereum.EthToElrondName(), "EthereumToElrond")
	assert.Equal(t, Bsc.EthToElrondName(), "BscToElrond")
}

func Test_elrondToEthName(t *testing.T) {
	assert.Equal(t, Ethereum.ElrondToEthName(), "ElrondToEthereum")
	assert.Equal(t, Bsc.ElrondToEthName(), "ElrondToBsc")
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
	assert.Equal(t, Ethereum.EthClientLogId(), "EthereumElrond-EthereumClient")
	assert.Equal(t, Bsc.EthClientLogId(), "BscElrond-BscClient")
}

func Test_elrondRoleProviderLogId(t *testing.T) {
	assert.Equal(t, Ethereum.ElrondRoleProviderLogId(), "EthereumElrond-ElrondRoleProvider")
	assert.Equal(t, Bsc.ElrondRoleProviderLogId(), "BscElrond-ElrondRoleProvider")
}

func Test_ethRoleProviderLogId(t *testing.T) {
	assert.Equal(t, Ethereum.EthRoleProviderLogId(), "EthereumElrond-EthereumRoleProvider")
	assert.Equal(t, Bsc.EthRoleProviderLogId(), "BscElrond-BscRoleProvider")
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
