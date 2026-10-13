//go:build !slow

package relayers

import (
	"context"
	"encoding/hex"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/factory"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/mock"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	redirectRecoveryAddress = "0x1234567890abcdef1234567890abcdef12345678"
	otherRecoveryAddress    = "0x00000000000000000000000000000000000000bb"
)

var (
	redirectedRecipientAddress = common.HexToAddress("0x717fb1435f5fa0cc1c7e541e5d92ef71342fcd41")
	bscUsdcAddress             = common.HexToAddress("0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d")
)

func mainnetBatch5956Deposits(t *testing.T) []mock.MultiversXDeposit {
	redirectedSender, err := hex.DecodeString("30adcf1257a0c227befca62f5d073b90a6c9875ee7e80e48ec0c3bbd5f622d13")
	require.Nil(t, err)
	unit := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)

	return []mock.MultiversXDeposit{
		{From: sdkData.NewAddressFromBytes(redirectedSender), To: redirectedRecipientAddress, Ticker: "BSCUSDC-887875", Amount: big.NewInt(0).Mul(big.NewInt(49_999), unit), DepositNonce: 7717},
		{From: sdkData.NewAddressFromBytes(redirectedSender), To: redirectedRecipientAddress, Ticker: "BSCUSDC-887875", Amount: big.NewInt(0).Mul(big.NewInt(30_440), unit), DepositNonce: 7718},
	}
}

type redirectScenarioResult struct {
	ethereumChainMock   *mock.EthereumChainMock
	multiversXChainMock *mock.MultiversXChainMock
	relayerEthAddresses []common.Address
	processFinished     *atomic.Bool
}

func runBatch5956Scenario(t *testing.T, recoveryAddresses []string, timeout time.Duration) redirectScenarioResult {
	deposits := mainnetBatch5956Deposits(t)
	total := big.NewInt(0).Add(deposits[0].Amount, deposits[1].Amount)

	safeContractEthAddress := testsCommon.CreateRandomEthereumAddress()
	erc20ContractsHolder := createMockErc20ContractsHolder([]common.Address{bscUsdcAddress}, safeContractEthAddress, []*big.Int{total})

	numRelayers := len(recoveryAddresses)
	ethereumChainMock := mock.NewEthereumChainMock()
	ethereumChainMock.SetQuorum(numRelayers)
	expectedStatuses := []byte{bridgeCore.Executed, bridgeCore.Executed}
	ethereumChainMock.GetStatusesAfterExecutionHandler = func() ([]byte, bool) {
		if callIsFromBalanceValidator() {
			return expectedStatuses, false
		}

		return expectedStatuses, true
	}
	ethereumChainMock.BalanceAtCalled = func(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
		return relayerEthBalance, nil
	}
	ethereumChainMock.UpdateNativeTokens(bscUsdcAddress, false)
	ethereumChainMock.UpdateMintBurnTokens(bscUsdcAddress, true)
	ethereumChainMock.UpdateMintBalances(bscUsdcAddress, zero)
	ethereumChainMock.UpdateBurnBalances(bscUsdcAddress, zero)

	multiversXChainMock := mock.NewMultiversXChainMock()
	multiversXChainMock.AddTokensPair(bscUsdcAddress, "BSCUSDC-887875", true, true, zero, zero, total)
	multiversXChainMock.SetPendingBatch(&mock.MultiversXPendingBatch{Nonce: big.NewInt(5956), MultiversXDeposits: deposits})
	multiversXChainMock.SetQuorum(numRelayers)

	result := redirectScenarioResult{ethereumChainMock: ethereumChainMock, multiversXChainMock: multiversXChainMock, processFinished: &atomic.Bool{}}
	relayers := make([]bridgeComponents, 0, numRelayers)
	defer func() { closeRelayers(relayers) }()

	messengers := integrationTests.CreateLinkedMessengers(numRelayers)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	multiversXChainMock.ProcessFinishedHandler = func() {
		result.processFinished.Store(true)
		asyncCancelCall(cancel, time.Second*5)
	}

	for i := 0; i < numRelayers; i++ {
		argsBridgeComponents := createMockBridgeComponentsArgs(i, messengers[i], multiversXChainMock, ethereumChainMock)
		argsBridgeComponents.Configs.GeneralConfig.Eth.SafeContractAddress = safeContractEthAddress.Hex()
		argsBridgeComponents.Configs.GeneralConfig.Eth.RedirectRecoveryAddress = recoveryAddresses[i]
		argsBridgeComponents.Erc20ContractsHolder = erc20ContractsHolder
		relayer, err := factory.NewEthMultiversXBridgeComponents(argsBridgeComponents)
		require.Nil(t, err)

		multiversXChainMock.AddRelayer(relayer.MultiversXRelayerAddress())
		ethereumChainMock.AddRelayer(relayer.EthereumRelayerAddress())
		result.relayerEthAddresses = append(result.relayerEthAddresses, relayer.EthereumRelayerAddress())

		go func() {
			errStart := relayer.Start()
			integrationTests.Log.LogIfError(errStart)
		}()

		relayers = append(relayers, relayer)
	}

	<-ctx.Done()
	time.Sleep(time.Second * 5)

	return result
}

func expectedRedirectedDigest(t *testing.T, recoveryAddress string) common.Hash {
	unit := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)
	recovery := common.HexToAddress(recoveryAddress)
	digest, err := ethereum.GenerateMessageHash(&batchProcessor.ArgListsBatch{
		EthTokens:  []common.Address{bscUsdcAddress, bscUsdcAddress},
		Recipients: []common.Address{recovery, recovery},
		Amounts:    []*big.Int{big.NewInt(0).Mul(big.NewInt(49_999), unit), big.NewInt(0).Mul(big.NewInt(30_440), unit)},
		Nonces:     []*big.Int{big.NewInt(7717), big.NewInt(7718)},
	}, 5956)
	require.Nil(t, err)

	return digest
}

func TestRelayersRedirectMainnetBatch5956ToRecoveryAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	result := runBatch5956Scenario(t, []string{redirectRecoveryAddress, redirectRecoveryAddress, redirectRecoveryAddress}, time.Second*600)
	require.True(t, result.processFinished.Load(), "set status on MultiversX was not performed")

	transfer := result.ethereumChainMock.GetLastProposedTransfer()
	require.NotNil(t, transfer)
	assert.Equal(t, big.NewInt(5956), transfer.BatchNonce)
	require.Len(t, transfer.Recipients, 2)
	for i, deposit := range mainnetBatch5956Deposits(t) {
		assert.Equal(t, common.HexToAddress(redirectRecoveryAddress), transfer.Recipients[i])
		assert.NotEqual(t, redirectedRecipientAddress, transfer.Recipients[i])
		assert.Equal(t, bscUsdcAddress, transfer.Tokens[i])
		assert.Equal(t, deposit.Amount, transfer.Amounts[i])
		assert.Equal(t, big.NewInt(0).SetUint64(deposit.DepositNonce), transfer.Nonces[i])
	}

	digest := expectedRedirectedDigest(t, redirectRecoveryAddress)
	assert.Equal(t, "0x2ed413d0cd051d34e31cc307f38e9d1070a1b7a13260f9cc72d6161a61425999", digest.Hex())
	signers := make([]common.Address, 0, len(transfer.Signatures))
	for _, signature := range transfer.Signatures {
		publicKey, err := crypto.SigToPub(digest.Bytes(), signature)
		require.Nil(t, err)
		signers = append(signers, crypto.PubkeyToAddress(*publicKey))
	}
	assert.ElementsMatch(t, result.relayerEthAddresses, signers)
}

func TestRelayersDoNotExecuteBatch5956WhenARelayerHasNoRecoveryAddress(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	result := runBatch5956Scenario(t, []string{redirectRecoveryAddress, redirectRecoveryAddress, ""}, time.Second*120)
	assert.False(t, result.processFinished.Load())
	assert.Nil(t, result.ethereumChainMock.GetLastProposedTransfer())
}

func TestRelayersDoNotExecuteBatch5956WhenRecoveryAddressesDiffer(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	result := runBatch5956Scenario(t, []string{redirectRecoveryAddress, redirectRecoveryAddress, otherRecoveryAddress}, time.Second*120)
	assert.False(t, result.processFinished.Load())
	assert.Nil(t, result.ethereumChainMock.GetLastProposedTransfer())
}
