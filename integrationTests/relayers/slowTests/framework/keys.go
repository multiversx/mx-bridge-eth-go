package framework

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"testing"

	suiSigner "github.com/block-vision/sui-go-sdk/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	mxCrypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/stretchr/testify/require"
)

// constants for the keys store
const (
	relayerPemPathFormat         = "multiversx%d.pem"
	SCCallerFilename             = "scCaller.pem"
	projectedShardForBridgeSetup = byte(0)
	projectedShardForDepositor   = byte(1)
	projectedShardForTestKeys    = byte(2)
)

// KeysHolder holds a 2 pk-sk pairs for both chains
type KeysHolder struct {
	MvxAddress *MvxAddress
	MvxSk      []byte
	EthSK      *ecdsa.PrivateKey
	EthAddress common.Address
	SuiAddress []byte
	SuiSK      ed25519.PrivateKey
}

// KeygenOptions holds the options for generating keys
type KeygenOptions struct {
	EthSKHex       string
	SuiSKHex       string
	ProjectedShard byte
}

// KeysStore will hold all the keys used in the test
type KeysStore struct {
	testing.TB
	RelayersKeys   []KeysHolder
	OraclesKeys    []KeysHolder
	SCExecutorKeys KeysHolder
	OwnerKeys      KeysHolder
	DepositorKeys  KeysHolder
	TestKeys       KeysHolder
	workingDir     string
}

const (
	ethOwnerSK     = "b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291"
	ethDepositorSK = "9bb971db41e3815a669a71c3f1bcb24e0b81f21e04bf11faa7a34b9b40e7cfb1"
	ethTestSk      = "dafea2c94bfe5d25f1a508808c2bc2c2e6c6f18b6b010fc841d8eb80755ba27a"

	suiOwnerSK     = "a0ce7a0a55fac7426bee1dfae7e7f6bf60d299f42a1c74091946ef606d426f47"
	suiDepositorSK = "c07f8bcdb7b4cb44fffa47dd510db0d59e775e19c68ad94d49253a7f18b7b1d0"
	suiTestSK      = "596d97615cc01d116a9b7754258da8d14d17dfdc03f6e7953c72c0f47f272423"
)

// NewKeysStore will create a KeysStore instance and generate all keys
func NewKeysStore(
	tb testing.TB,
	workingDir string,
	numRelayers int,
	numOracles int,
) *KeysStore {
	keysStore := &KeysStore{
		TB:             tb,
		RelayersKeys:   make([]KeysHolder, 0, numRelayers),
		SCExecutorKeys: KeysHolder{},
		workingDir:     workingDir,
	}

	keysStore.generateRelayersKeys(numRelayers)
	keysStore.OraclesKeys = keysStore.generateKeys(numOracles, "generated oracle", projectedShardForBridgeSetup)

	keysStore.SCExecutorKeys = keysStore.generateKey(KeygenOptions{
		EthSKHex:       "",
		SuiSKHex:       "",
		ProjectedShard: projectedShardForBridgeSetup,
	})
	keysStore.OwnerKeys = keysStore.generateKey(KeygenOptions{
		EthSKHex:       ethOwnerSK,
		SuiSKHex:       suiOwnerSK,
		ProjectedShard: projectedShardForBridgeSetup,
	})
	log.Info("generated owner",
		"MvX address", keysStore.OwnerKeys.MvxAddress.Bech32(),
		"Eth address", keysStore.OwnerKeys.EthAddress.String(),
		"Sui address", string(keysStore.OwnerKeys.SuiAddress),
	)
	keysStore.DepositorKeys = keysStore.generateKey(KeygenOptions{
		EthSKHex:       ethDepositorSK,
		SuiSKHex:       suiDepositorSK,
		ProjectedShard: projectedShardForDepositor,
	})
	keysStore.TestKeys = keysStore.generateKey(KeygenOptions{
		EthSKHex:       ethTestSk,
		SuiSKHex:       suiTestSK,
		ProjectedShard: projectedShardForTestKeys,
	})

	filename := path.Join(keysStore.workingDir, SCCallerFilename)
	SaveMvxKey(keysStore, filename, keysStore.SCExecutorKeys)

	return keysStore
}

func (keyStore *KeysStore) generateRelayersKeys(numKeys int) {
	for i := 0; i < numKeys; i++ {
		relayerETHSKBytes, err := os.ReadFile(fmt.Sprintf(relayerETHKeyPathFormat, i))
		require.Nil(keyStore, err)

		relayerSuiSKBytes, err := os.ReadFile(fmt.Sprintf(relayerSuiSeedPathFormat, i))
		require.Nil(keyStore, err)

		relayerKeys := keyStore.generateKey(KeygenOptions{
			EthSKHex:       string(relayerETHSKBytes),
			SuiSKHex:       string(relayerSuiSKBytes),
			ProjectedShard: projectedShardForBridgeSetup,
		})
		log.Info("generated relayer", "index", i,
			"MvX address", relayerKeys.MvxAddress.Bech32(),
			"Eth address", relayerKeys.EthAddress.String(),
			"Sui address", string(relayerKeys.SuiAddress),
		)

		keyStore.RelayersKeys = append(keyStore.RelayersKeys, relayerKeys)

		filename := path.Join(keyStore.workingDir, fmt.Sprintf(relayerPemPathFormat, i))

		SaveMvxKey(keyStore, filename, relayerKeys)
	}
}

func (keyStore *KeysStore) generateKeys(numKeys int, message string, projectedShard byte) []KeysHolder {
	keys := make([]KeysHolder, 0, numKeys)

	for i := 0; i < numKeys; i++ {
		ethPrivateKeyBytes := make([]byte, 32)
		_, _ = rand.Read(ethPrivateKeyBytes)

		suiSeedBytes := make([]byte, 32)
		_, _ = rand.Read(suiSeedBytes)

		key := keyStore.generateKey(KeygenOptions{
			EthSKHex:       hex.EncodeToString(ethPrivateKeyBytes),
			SuiSKHex:       hex.EncodeToString(suiSeedBytes),
			ProjectedShard: projectedShard,
		})
		log.Info(message, "index", i,
			"MvX address", key.MvxAddress.Bech32(),
			"Eth address", key.EthAddress.String(),
			"Sui address", string(key.SuiAddress),
		)

		keys = append(keys, key)
	}

	return keys
}

func (keyStore *KeysStore) generateKey(opts KeygenOptions) KeysHolder {
	var err error
	keys := GenerateMvxPrivatePublicKey(keyStore, opts.ProjectedShard)
	if len(opts.EthSKHex) == 0 && len(opts.SuiSKHex) == 0 {
		return keys
	}

	if len(opts.EthSKHex) > 0 {
		keys.EthSK, err = crypto.HexToECDSA(opts.EthSKHex)
		require.Nil(keyStore, err)
		keys.EthAddress = crypto.PubkeyToAddress(keys.EthSK.PublicKey)
	}

	if len(opts.SuiSKHex) > 0 {
		suiSeedBytes, _ := getSeedFromPrivateKey(opts.SuiSKHex)
		relayer := suiSigner.NewSigner(suiSeedBytes)
		keys.SuiSK = relayer.PriKey
		keys.SuiAddress = []byte(relayer.Address)
	}

	return keys
}

func (keyStore *KeysStore) getAllKeys() []KeysHolder {
	allKeys := make([]KeysHolder, 0, len(keyStore.RelayersKeys)+10)
	allKeys = append(allKeys, keyStore.RelayersKeys...)
	allKeys = append(allKeys, keyStore.OraclesKeys...)
	allKeys = append(allKeys, keyStore.SCExecutorKeys, keyStore.OwnerKeys, keyStore.DepositorKeys, keyStore.TestKeys)

	return allKeys
}

// WalletsToFundOnEthereum will return the wallets to fund on Ethereum
func (keyStore *KeysStore) WalletsToFundOnEthereum() []common.Address {
	allKeys := keyStore.getAllKeys()
	walletsToFund := make([]common.Address, 0, len(allKeys))

	for _, key := range allKeys {
		if len(key.MvxSk) == 0 {
			continue
		}

		walletsToFund = append(walletsToFund, key.EthAddress)
	}

	return walletsToFund
}

// WalletsToFundOnSui will return the wallets to fund on Sui
func (keyStore *KeysStore) WalletsToFundOnSui() [][]byte {
	allKeys := keyStore.getAllKeys()
	walletsToFund := make([][]byte, 0, len(allKeys))

	for _, key := range allKeys {
		if len(key.SuiAddress) == 0 {
			continue
		}

		walletsToFund = append(walletsToFund, key.SuiAddress)
	}

	return walletsToFund
}

// WalletsToFundOnMultiversX will return the wallets to fund on MultiversX
func (keyStore *KeysStore) WalletsToFundOnMultiversX() []string {
	allKeys := keyStore.getAllKeys()
	walletsToFund := make([]string, 0, len(allKeys))

	for _, key := range allKeys {
		walletsToFund = append(walletsToFund, key.MvxAddress.Bech32())
	}

	return walletsToFund
}

// GenerateMvxPrivatePublicKey will generate a new keys holder instance that will hold only the MultiversX generated keys
func GenerateMvxPrivatePublicKey(tb testing.TB, projectedShard byte) KeysHolder {
	sk, pkBytes := generateSkPkInShard(tb, projectedShard)

	skBytes, err := sk.ToByteArray()
	require.Nil(tb, err)

	return KeysHolder{
		MvxSk:      skBytes,
		MvxAddress: NewMvxAddressFromBytes(tb, pkBytes),
	}
}

func generateSkPkInShard(tb testing.TB, projectedShard byte) (mxCrypto.PrivateKey, []byte) {
	var sk mxCrypto.PrivateKey
	var pk mxCrypto.PublicKey

	for {
		sk, pk = keyGenerator.GeneratePair()

		pkBytes, err := pk.ToByteArray()
		require.Nil(tb, err)

		if pkBytes[len(pkBytes)-1] == projectedShard {
			return sk, pkBytes
		}
	}
}

// SaveMvxKey will save the MultiversX key
func SaveMvxKey(tb testing.TB, filename string, key KeysHolder) {
	blk := pem.Block{
		Type:  "PRIVATE KEY for " + key.MvxAddress.Bech32(),
		Bytes: []byte(hex.EncodeToString(key.MvxSk)),
	}

	buff := bytes.NewBuffer(make([]byte, 0))
	err := pem.Encode(buff, &blk)
	require.Nil(tb, err)

	err = os.WriteFile(filename, buff.Bytes(), os.ModePerm)
	require.Nil(tb, err)
}

func getSeedFromPrivateKey(data string) ([]byte, error) {
	privKey := converters.TrimWhiteSpaceCharacters(data)
	return hex.DecodeString(privKey)
}
