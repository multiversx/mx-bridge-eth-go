package broadcaster

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	relayP2P "github.com/ElrondNetwork/elrond-eth-bridge/relay/p2p"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	mockRoleProviders "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/roleProviders"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/p2p"
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
	broadcasters, signaturesHolders := createBroadcasters(t, numBroadcasters, messengers, roleProvider, privateKeys)

	time.Sleep(time.Second)

	expectedPkInOrder := copyAndSortBytesSlices(publicKeysBytes)

	messageHash := []byte("message hash")
	joinBroadcasters(broadcasters)
	signatures := createSignatures(numBroadcasters, "mock signature - try 1")
	sendSignatures(broadcasters, signatures, messageHash)
	checkBroadcasterState(t, broadcasters, signaturesHolders, signatures, expectedPkInOrder, messageHash)

	// clear test
	clearSignatures(signaturesHolders)
	checkBroadcasterState(t, broadcasters, signaturesHolders, make([][]byte, 0), expectedPkInOrder, messageHash)

	messageHash = []byte("message hash 1")
	signatures = createSignatures(numBroadcasters, "mock signature - try 2")
	sendSignatures(broadcasters, signatures, messageHash)
	checkBroadcasterState(t, broadcasters, signaturesHolders, signatures, expectedPkInOrder, messageHash)

	// overwrite test
	messageHash = []byte("message hash 2")
	signatures = createSignatures(numBroadcasters, "mock signature - try 3")
	sendSignatures(broadcasters, signatures, messageHash)
	checkBroadcasterState(t, broadcasters, signaturesHolders, signatures, expectedPkInOrder, messageHash)
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
	broadcasters, signaturesHolders := createBroadcasters(t, numBroadcasters, messengers, roleProvider, privateKeys)

	time.Sleep(time.Second)

	expectedPkInOrder := copyAndSortBytesSlices(publicKeysBytes[1:])
	messageHash := []byte("message hash")
	joiningBroadcasters := broadcasters[1:]
	joinBroadcasters(joiningBroadcasters)
	signatures := createSignatures(numBroadcasters, "mock signature - try 1")
	sendSignatures(joiningBroadcasters, signatures[1:], messageHash)
	checkBroadcasterState(t, joiningBroadcasters, signaturesHolders, signatures[1:], expectedPkInOrder, messageHash)

	lateBroadcasters := []integrationTests.Broadcaster{broadcasters[0]}
	checkBroadcasterState(t, lateBroadcasters, signaturesHolders, signatures[1:], expectedPkInOrder, messageHash)
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
	broadcasters, signaturesHolders := createBroadcasters(t, numBroadcasters-1, messengers, roleProvider, privateKeys)

	time.Sleep(time.Second)

	expectedPkInOrder := copyAndSortBytesSlices(publicKeysBytes[:len(publicKeysBytes)-1])
	messageHash := []byte("message hash")

	joinBroadcasters(broadcasters)
	signatures := createSignatures(numBroadcasters-1, "mock signature - try 1")
	sendSignatures(broadcasters, signatures, messageHash)
	checkBroadcasterState(t, broadcasters, signaturesHolders, signatures, expectedPkInOrder, messageHash)

	expectedPkInOrder = copyAndSortBytesSlices(publicKeysBytes)

	integrationTests.Log.Info("creating the late broadcaster")
	lateBroadcaster, lateSigHolder := createBroadcaster(t, messengers[len(messengers)-1], roleProvider, privateKeys[len(privateKeys)-1])

	time.Sleep(time.Second)
	lateBroadcaster.BroadcastJoinTopic()
	time.Sleep(time.Second)

	lateBroadcasters := []integrationTests.Broadcaster{lateBroadcaster}
	lateSigHolders := []*testsCommon.SignaturesHolderMock{lateSigHolder}
	checkBroadcasterState(t, lateBroadcasters, lateSigHolders, signatures, expectedPkInOrder, messageHash)
	checkBroadcasterState(t, broadcasters, signaturesHolders, signatures, expectedPkInOrder, messageHash)
}

func createBroadcasters(
	t *testing.T,
	numBroadcasters int,
	messengers []p2p.Messenger,
	roleProvider *mockRoleProviders.ElrondRoleProviderStub,
	privateKeys []crypto.PrivateKey,
) ([]integrationTests.Broadcaster, []*testsCommon.SignaturesHolderMock) {
	broadcasters := make([]integrationTests.Broadcaster, 0, numBroadcasters)
	signaturesHolders := make([]*testsCommon.SignaturesHolderMock, 0, numBroadcasters)
	for i := 0; i < numBroadcasters; i++ {
		b, sigHolder := createBroadcaster(t, messengers[i], roleProvider, privateKeys[i])

		broadcasters = append(broadcasters, b)
		signaturesHolders = append(signaturesHolders, sigHolder)
	}

	return broadcasters, signaturesHolders
}

func createBroadcaster(
	t *testing.T,
	messenger p2p.Messenger,
	roleProvider *mockRoleProviders.ElrondRoleProviderStub,
	privateKey crypto.PrivateKey,
) (integrationTests.Broadcaster, *testsCommon.SignaturesHolderMock) {
	args := relayP2P.ArgsBroadcaster{
		Messenger:          messenger,
		Log:                integrationTests.Log,
		ElrondRoleProvider: roleProvider,
		KeyGen:             integrationTests.TestKeyGenerator,
		SingleSigner:       integrationTests.TestSingleSigner,
		PrivateKey:         privateKey,
		SignatureProcessor: &testsCommon.SignatureProcessorStub{},
		Name:               "test",
	}

	b, err := relayP2P.NewBroadcaster(args)
	require.Nil(t, err)

	err = b.RegisterOnTopics()
	require.Nil(t, err)

	sigHolder := testsCommon.NewSignaturesHolderMock()
	err = b.AddBroadcastClient(sigHolder)
	require.Nil(t, err)

	return b, sigHolder
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
	signatureHolders []*testsCommon.SignaturesHolderMock,
	expectedSigs [][]byte,
	expectedPublicKeys [][]byte,
	messageHash []byte,
) {
	integrationTests.Log.Info("checking received signatures",
		"num broadcasters", len(broadcasters), "num expected signatures", len(expectedSigs))
	for i, b := range broadcasters {
		checkStateOnBroadcaster(t, b, signatureHolders[i], expectedSigs, expectedPublicKeys, messageHash)
	}
}

func checkStateOnBroadcaster(
	t *testing.T,
	b integrationTests.Broadcaster,
	sh *testsCommon.SignaturesHolderMock,
	expectedSigs [][]byte,
	expectedPublicKeys [][]byte,
	messageHash []byte,
) {
	sigs := sh.Signatures(messageHash)
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

func clearSignatures(signatureHolders []*testsCommon.SignaturesHolderMock) {
	integrationTests.Log.Info("clearing signatures...")
	for _, sh := range signatureHolders {
		sh.ClearStoredSignatures()
	}

	time.Sleep(time.Second)
}
