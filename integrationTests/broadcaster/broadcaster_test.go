package broadcaster

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ElrondNetwork/elrond-eth-bridge/relay/p2p"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	mockRoleProviders "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/roleProviders"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/stretchr/testify/require"
)

func TestNetworkOfBroadcastersShouldPassTheSignatures(t *testing.T) {
	numBroadcasters := 5

	integrationTests.Log.Info("creating & linking network messengers...")
	messengers := integrationTests.CreateLinkedMessengers(numBroadcasters)
	defer func() {
		for _, m := range messengers {
			_ = m.Close()
		}
	}()

	privateKeys, publicKeysBytes := createKeys(t, numBroadcasters)

	roleProvider := &mockRoleProviders.ElrondRoleProviderStub{
		IsWhitelistedCalled: func(address core.AddressHandler) bool {
			for _, pkBytes := range publicKeysBytes {
				if bytes.Equal(address.AddressBytes(), pkBytes) {
					return true
				}
			}

			return false
		},
	}

	integrationTests.Log.Info("creating broadcasters...")
	broadcasters := make([]integrationTests.Broadcaster, 0, numBroadcasters)
	for i := 0; i < numBroadcasters; i++ {
		args := p2p.ArgsBroadcaster{
			Messenger:          messengers[i],
			Log:                integrationTests.Log,
			ElrondRoleProvider: roleProvider,
			KeyGen:             integrationTests.TestKeyGenerator,
			SingleSigner:       integrationTests.TestSingleSigner,
			PrivateKey:         privateKeys[i],
			SignatureProcessor: &testsCommon.SignatureProcessorStub{},
		}

		b, err := p2p.NewBroadcaster(args)
		require.Nil(t, err)

		err = b.RegisterOnTopics()
		require.Nil(t, err)

		broadcasters = append(broadcasters, b)
	}

	time.Sleep(time.Second)

	expectedPkInOrder := copyAndSortBytesSlices(publicKeysBytes)

	messageHash := []byte("message hash")
	joinBroadcasters(broadcasters)
	signatures := createSignatures(numBroadcasters, "mock signature - try 1")
	sendSignatures(broadcasters, signatures, messageHash)
	checkBroadcasterState(t, broadcasters, signatures, expectedPkInOrder, messageHash)

	// clear test
	clearSignatures(broadcasters)
	checkBroadcasterState(t, broadcasters, make([][]byte, 0), expectedPkInOrder, messageHash)

	messageHash = []byte("message hash 1")
	signatures = createSignatures(numBroadcasters, "mock signature - try 2")
	sendSignatures(broadcasters, signatures, messageHash)
	checkBroadcasterState(t, broadcasters, signatures, expectedPkInOrder, messageHash)

	// overwrite test
	messageHash = []byte("message hash 2")
	signatures = createSignatures(numBroadcasters, "mock signature - try 3")
	sendSignatures(broadcasters, signatures, messageHash)
	checkBroadcasterState(t, broadcasters, signatures, expectedPkInOrder, messageHash)
}

func TestNetworkOfBroadcastersShouldBootstrapOnLateBroadcasterWhenNotJoining(t *testing.T) {
	numBroadcasters := 5

	integrationTests.Log.Info("creating & linking network messengers...")
	messengers := integrationTests.CreateLinkedMessengers(numBroadcasters)
	defer func() {
		for _, m := range messengers {
			_ = m.Close()
		}
	}()

	privateKeys, publicKeysBytes := createKeys(t, numBroadcasters)

	roleProvider := &mockRoleProviders.ElrondRoleProviderStub{
		IsWhitelistedCalled: func(address core.AddressHandler) bool {
			for _, pkBytes := range publicKeysBytes {
				if bytes.Equal(address.AddressBytes(), pkBytes) {
					return true
				}
			}

			return false
		},
	}

	integrationTests.Log.Info("creating broadcasters...")
	broadcasters := make([]integrationTests.Broadcaster, 0, numBroadcasters)
	for i := 0; i < numBroadcasters; i++ {
		args := p2p.ArgsBroadcaster{
			Messenger:          messengers[i],
			Log:                integrationTests.Log,
			ElrondRoleProvider: roleProvider,
			KeyGen:             integrationTests.TestKeyGenerator,
			SingleSigner:       integrationTests.TestSingleSigner,
			PrivateKey:         privateKeys[i],
			SignatureProcessor: &testsCommon.SignatureProcessorStub{},
		}

		b, err := p2p.NewBroadcaster(args)
		require.Nil(t, err)

		err = b.RegisterOnTopics()
		require.Nil(t, err)

		broadcasters = append(broadcasters, b)
	}

	time.Sleep(time.Second)

	expectedPkInOrder := copyAndSortBytesSlices(publicKeysBytes[1:])
	messageHash := []byte("message hash")
	joiningBroadcasters := broadcasters[1:]
	joinBroadcasters(joiningBroadcasters)
	signatures := createSignatures(numBroadcasters, "mock signature - try 1")
	sendSignatures(joiningBroadcasters, signatures[1:], messageHash)
	checkBroadcasterState(t, joiningBroadcasters, signatures[1:], expectedPkInOrder, messageHash)

	lateBroadcasters := []integrationTests.Broadcaster{broadcasters[0]}
	checkBroadcasterState(t, lateBroadcasters, signatures[1:], expectedPkInOrder, messageHash)
}

func TestNetworkOfBroadcastersShouldBootstrapOnLateBroadcasterWhenLateConnecting(t *testing.T) {
	numBroadcasters := 5

	integrationTests.Log.Info("creating & linking network messengers...")
	messengers := integrationTests.CreateLinkedMessengers(numBroadcasters)
	defer func() {
		for _, m := range messengers {
			_ = m.Close()
		}
	}()

	privateKeys, publicKeysBytes := createKeys(t, numBroadcasters)

	roleProvider := &mockRoleProviders.ElrondRoleProviderStub{
		IsWhitelistedCalled: func(address core.AddressHandler) bool {
			for _, pkBytes := range publicKeysBytes {
				if bytes.Equal(address.AddressBytes(), pkBytes) {
					return true
				}
			}

			return false
		},
	}

	integrationTests.Log.Info("creating broadcasters...")
	broadcasters := make([]integrationTests.Broadcaster, 0, numBroadcasters-1)
	for i := 0; i < numBroadcasters-1; i++ {
		args := p2p.ArgsBroadcaster{
			Messenger:          messengers[i],
			Log:                integrationTests.Log,
			ElrondRoleProvider: roleProvider,
			KeyGen:             integrationTests.TestKeyGenerator,
			SingleSigner:       integrationTests.TestSingleSigner,
			PrivateKey:         privateKeys[i],
			SignatureProcessor: &testsCommon.SignatureProcessorStub{},
		}

		b, err := p2p.NewBroadcaster(args)
		require.Nil(t, err)

		err = b.RegisterOnTopics()
		require.Nil(t, err)

		broadcasters = append(broadcasters, b)
	}

	time.Sleep(time.Second)

	expectedPkInOrder := copyAndSortBytesSlices(publicKeysBytes[:len(publicKeysBytes)-1])
	messageHash := []byte("message hash")

	joinBroadcasters(broadcasters)
	signatures := createSignatures(numBroadcasters-1, "mock signature - try 1")
	sendSignatures(broadcasters, signatures, messageHash)
	checkBroadcasterState(t, broadcasters, signatures, expectedPkInOrder, messageHash)

	expectedPkInOrder = copyAndSortBytesSlices(publicKeysBytes)

	integrationTests.Log.Info("creating the late broadcaster")
	args := p2p.ArgsBroadcaster{
		Messenger:          messengers[len(messengers)-1],
		Log:                integrationTests.Log,
		ElrondRoleProvider: roleProvider,
		KeyGen:             integrationTests.TestKeyGenerator,
		SingleSigner:       integrationTests.TestSingleSigner,
		PrivateKey:         privateKeys[len(privateKeys)-1],
		SignatureProcessor: &testsCommon.SignatureProcessorStub{},
	}

	lateBroadcaster, err := p2p.NewBroadcaster(args)
	require.Nil(t, err)

	err = lateBroadcaster.RegisterOnTopics()
	require.Nil(t, err)

	time.Sleep(time.Second)
	lateBroadcaster.BroadcastJoinTopic()
	time.Sleep(time.Second)

	lateBroadcasters := []integrationTests.Broadcaster{lateBroadcaster}
	checkBroadcasterState(t, lateBroadcasters, signatures, expectedPkInOrder, messageHash)
	checkBroadcasterState(t, broadcasters, signatures, expectedPkInOrder, messageHash)
}

func createKeys(t *testing.T, numKeys int) ([]crypto.PrivateKey, [][]byte) {
	privateKeys := make([]crypto.PrivateKey, 0, numKeys)
	publicKeysBytes := make([][]byte, 0, numKeys)
	for i := 0; i < numKeys; i++ {
		sk, pk := integrationTests.TestKeyGenerator.GeneratePair()
		pkBytes, err := pk.ToByteArray()
		require.Nil(t, err)
		publicKeysBytes = append(publicKeysBytes, pkBytes)
		privateKeys = append(privateKeys, sk)
	}

	return privateKeys, publicKeysBytes
}

func copyAndSortBytesSlices(src [][]byte) [][]byte {
	dst := make([][]byte, 0, len(src))
	for _, srcBuff := range src {
		dstBuff := make([]byte, len(srcBuff))
		copy(dstBuff, srcBuff)
		dst = append(dst, dstBuff)
	}

	sort.Slice(dst, func(i, j int) bool {
		return bytes.Compare(dst[i], dst[j]) < 0
	})

	return dst
}

func joinBroadcasters(broadcasters []integrationTests.Broadcaster) {
	integrationTests.Log.Info("joining the broadcasters...")
	for _, b := range broadcasters {
		b.BroadcastJoinTopic()
	}

	time.Sleep(time.Second)
}

func createSignatures(numSignatures int, suffix string) [][]byte {
	integrationTests.Log.Info("creating signatures...")
	signatures := make([][]byte, 0, numSignatures)
	for i := 0; i < numSignatures; i++ {
		signatures = append(signatures, []byte(fmt.Sprintf("%s%d", suffix, i)))
	}

	return signatures
}

func sendSignatures(broadcasters []integrationTests.Broadcaster, signatures [][]byte, messageHash []byte) {
	integrationTests.Log.Info("sending signatures...")
	for i, b := range broadcasters {
		b.BroadcastSignature(signatures[i], messageHash)
	}

	time.Sleep(time.Second)
}

func checkBroadcasterState(
	t *testing.T,
	broadcasters []integrationTests.Broadcaster,
	expectedSigs [][]byte,
	expectedPublicKeys [][]byte,
	messageHash []byte,
) {
	integrationTests.Log.Info("checking received signatures",
		"num broadcasters", len(broadcasters), "num expected signatures", len(expectedSigs))
	for _, b := range broadcasters {
		checkStateOnBroadcaster(t, b, expectedSigs, expectedPublicKeys, messageHash)
	}
}

func checkStateOnBroadcaster(
	t *testing.T,
	b integrationTests.Broadcaster,
	expectedSigs [][]byte,
	expectedPublicKeys [][]byte,
	messageHash []byte,
) {
	sigs := b.Signatures(messageHash)
	require.Equal(t, len(expectedSigs), len(sigs))
	require.Equal(t, expectedPublicKeys, b.SortedPublicKeys())

	// the order is random, using a map
	sigMap := make(map[string]int)
	for _, sig := range expectedSigs {
		sigMap[string(sig)] = 0
	}
	for _, sig := range sigs {
		sigMap[string(sig)]++
	}

	for sig, num := range sigMap {
		require.Equal(t, 1, num, fmt.Sprintf("%s got %d sigs", sig, num))
	}
}

func clearSignatures(broadcasters []integrationTests.Broadcaster) {
	integrationTests.Log.Info("clearing signatures...")
	for _, b := range broadcasters {
		b.ClearSignatures()
	}

	time.Sleep(time.Second)
}
