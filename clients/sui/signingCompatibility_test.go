package sui

import (
	"bytes"
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/transaction"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/stretchr/testify/require"
)

// Golden vectors produced by relayer v3.1.0 (sui-go-sdk v1.0.8, JSON-RPC); they must stay byte-identical across SDK/transport changes.
const (
	goldenMessageHash    = "77f8f7cb068ae0c17a65869c891032d1beffd5140dfac12396086a87485e159b38c459989cacd2f4d17d4a24ca30785826de5a7580fb03b961171246f3bc1fc1"
	goldenRelayerSigs    = "ALqItmoKmigx9KpJrL1JJ1C5GSEfoVuHwCcnFlz143zbsdpnpPR5Jjt6EWxMtqxi6BLZa9hKSDDtFTJ/vN+23gjQSrIydCu0qzoTaL1GFeTm0CJKtxoBa6+FIKMyyXeHNw==AO5+OhgbfBxOtPSaLdZb85bFZZ3eX4iD+eAZ8wMpRDecHDn/dBTljul1HN6tu+PByXzLVKWJoxr3YwkXfhwREwLQSrIydCu0qzoTaL1GFeTm0CJKtxoBa6+FIKMyyXeHNw==AIC1PERsEjpnthIdGYx04oCne2DXbQp277tvhRQacSdHSvwUBrt/cUpmuMQ3WmDWONAnUOiuYxbvgHTMjVE0LgigmqX0emdZgC/5VfjcLSoUpcmdI76X+GQSf/k4NFWk8A==AHIYtsDafGTdPS86K5fQRy7yFr275ByvOfSZ4aW5Gp3mzejFv1oRxuj+xelJ7uwsx3g5GHSdXalUqPrlQuQlkAagmqX0emdZgC/5VfjcLSoUpcmdI76X+GQSf/k4NFWk8A==AFgt7yr20eLatJcECDjQikvqVIJIpo2UaV7meZBaD89FbnO3sqpUyG7UMwHQRfKOnHOpEEohuKab31c4wppiXwwXy3n7K0Eg8rHsZeQZjW4Iso6BP+sB5KQAg5uF4YCAzg==ADEUfcf8/33bjfIqZxWH8jyyC+NRvCBqHYShczG0+jnAYrA+OMvMVryQpvP7Mopivw1YnCNx25aRSzNsc6kJRQ8Xy3n7K0Eg8rHsZeQZjW4Iso6BP+sB5KQAg5uF4YCAzg=="
	goldenTxBytes        = "AAAPAQHhVRPMk9bvv73HhE3xQbMSu2d+5WSlg4t7IqiR+fBcZcf9DgAAAAAAAQEBMsjr9YUxY0cpZM4iaxlK8Fql8uTMR2eJJOvlOKHIhBZqfQYAAAAAAAEAIQGysrKysrKysrKysrKysrKysrKysrKysrKysrKysrKysgAJAQC6HdIFAAAAAAkB+ggAAAAAAAAACJ0EAAAAAAAAAKQCA2C6iLZqCpooMfSqSay9SSdQuRkhH6Fbh8AnJxZc9eN827HaZ6T0eSY7ehFsTLasYugS2WvYSkgw7RUyf7zftt4I0EqyMnQrtKs6E2i9RhXk5tAiSrcaAWuvhSCjMsl3hzdggLU8RGwSOme2Eh0ZjHTigKd7YNdtCnbvu2+FFBpxJ0dK/BQGu39xSma4xDdaYNY40CdQ6K5jFu+AdMyNUTQuCKCapfR6Z1mAL/lV+NwtKhSlyZ0jvpf4ZBJ/+Tg0VaTwYFgt7yr20eLatJcECDjQikvqVIJIpo2UaV7meZBaD89FbnO3sqpUyG7UMwHQRfKOnHOpEEohuKab31c4wppiXwwXy3n7K0Eg8rHsZeQZjW4Iso6BP+sB5KQAg5uF4YCAzgABAAEBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAYBAAAAAAAAAAAAQQKhoaGhoaGhoaGhoaGhoaGhoaGhoaGhoaGhoaGhoaGhocPDw8PDw8PDw8PDw8PDw8PDw8PDw8PDw8PDw8PDw8PDABECL/7N7AgAAAABAAAAAAAAAAARAvkIAAAAAAAA+wgAAAAAAAAACJ0EAAAAAAAAAKQCA2DufjoYG3wcTrT0mi3WW/OWxWWd3l+Ig/ngGfMDKUQ3nBw5/3QU5Y7pdRzerbvjwcl8y1SliaMa92MJF34cERMC0EqyMnQrtKs6E2i9RhXk5tAiSrcaAWuvhSCjMsl3hzdgchi2wNp8ZN09Lzorl9BHLvIWvbvkHK859JnhpbkanebN6MW/WhHG6P7F6Unu7CzHeDkYdJ1dqVSo+uVC5CWQBqCapfR6Z1mAL/lV+NwtKhSlyZ0jvpf4ZBJ/+Tg0VaTwYDEUfcf8/33bjfIqZxWH8jyyC+NRvCBqHYShczG0+jnAYrA+OMvMVryQpvP7Mopivw1YnCNx25aRSzNsc6kJRQ8Xy3n7K0Eg8rHsZeQZjW4Iso6BP+sB5KQAg5uF4YCAzgABAQIAs6aSr2TfV5Q+N2J5mV8hHjwXNNbBw7gzo37QOa5hvAQGYnJpZGdlEGV4ZWN1dGVfdHJhbnNmZXIBB4252dxc1XI+5VclhpYgBz4o+I3fPDYKUS69c8tG8ZA9DGJyaWRnZV90b2tlbgxCUklER0VfVE9LRU4ACgEAAAEBAAECAAEDAAEEAAEFAAEGAAEHAAEBAAEIAACzppKvZN9XlD43YnmZXyEePBc01sHDuDOjftA5rmG8BAZicmlkZ2UQZXhlY3V0ZV90cmFuc2ZlcgEHl8dXH0QGzdepXzAnB1q4DT6ck3wqVnaQ0x4UqxhyzO4DeG1uA1hNTgAKAQAAAQEAAQkAAQoAAQsAAQwAAQ0AAQ4AAQEAAQgAKd+/aIq856tDu45wyuFYrpYRluchRA9RVIL4uhaEOQ8BXzocDostT2p8nhs9X3qcHjtdf5ocPlt9nxo8XnudHzra9Bw/AAAAACA/wCRr0BdKjCgd2kouf1YaJ4QBRo2vihTM1DpTweKIQCnfv2iKvOerQ7uOcMrhWK6WEZbnIUQPUVSC+LoWhDkP6AMAAAAAAADAaHgEAAAAAAA="
	goldenTxSignature    = "ADMoyOdXe5lske+WA0rIX/Cj9JYiIKbjfqk6nLYWfyGg1n7Jx1ZBJXlgKytu8LfaE4m6Na6ZqP+D3r6ma5V0sASKiOPddAnxlf1S2y08ul1yymcJvx2UEhvzdIgBtA9vXA=="
	printGoldenVectorEnv = "SUI_PRINT_GOLDEN"
)

func goldenBatch() *batchProcessor.ArgListsBatch {
	recipient := func(b byte) []byte { return bytes.Repeat([]byte{b}, 32) }
	return &batchProcessor.ArgListsBatch{
		PeerTokens: [][]byte{
			[]byte("0x97c7571f4406cdd7a95f3027075ab80d3e9c937c2a567690d31e14ab1872ccee::xmn::XMN"),
			[]byte("0x8db9d9dc5cd5723ee55725869620073e28f88ddf3c360a512ebd73cb46f1903d::bridge_token::BRIDGE_TOKEN"),
			[]byte("0x97c7571f4406cdd7a95f3027075ab80d3e9c937c2a567690d31e14ab1872ccee::xmn::XMN"),
		},
		Recipients: [][]byte{recipient(0xa1), recipient(0xb2), recipient(0xc3)},
		Amounts:    []*big.Int{big.NewInt(38_332_661_295), big.NewInt(25_000_000_000), big.NewInt(1)},
		Nonces:     []*big.Int{big.NewInt(2297), big.NewInt(2298), big.NewInt(2299)},
		Direction:  batchProcessor.FromMultiversX,
	}
}

func goldenRelayers() []*signer.Signer {
	return []*signer.Signer{
		signer.NewSigner(bytes.Repeat([]byte{0x11}, 32)),
		signer.NewSigner(bytes.Repeat([]byte{0x22}, 32)),
		signer.NewSigner(bytes.Repeat([]byte{0x33}, 32)),
	}
}

func TestSigningCompatibility_GoldenVectors(t *testing.T) {
	args := createMockSuiClientArgs()
	args.PackageId = "0xb3a692af64df57943e376279995f211e3c1734d6c1c3b833a37ed039ae61bc04"
	c, err := NewSuiClient(args)
	require.NoError(t, err)

	const batchID = uint64(1181)
	batch := goldenBatch()
	msgHash, err := c.GenerateMessageHash(batch, batchID)
	require.NoError(t, err)

	serializedSignatures := make([][]byte, 0)
	var allSignatures bytes.Buffer
	for _, relayer := range goldenRelayers() {
		var relayerSigs []byte
		for i := 0; i < len(msgHash)/MessageLength; i++ {
			resp, errSign := relayer.SignPersonalMessageV1(string(msgHash[i*MessageLength : (i+1)*MessageLength]))
			require.NoError(t, errSign)
			relayerSigs = append(relayerSigs, []byte(resp.Signature)...)
		}
		serializedSignatures = append(serializedSignatures, relayerSigs)
		allSignatures.Write(relayerSigs)
	}

	groups, sortedTokenTypes, err := c.groupTransfersByTokenType(batch)
	require.NoError(t, err)
	require.NoError(t, c.processSignaturesOfRelayers(groups, sortedTokenTypes, serializedSignatures))
	calls, err := c.prepareExecuteTransferCallArgs(batchID, groups, sortedTokenTypes)
	require.NoError(t, err)

	gasCoin, err := transaction.NewSuiObjectRef(
		models.SuiAddress("0x5f3a1c0e8b2d4f6a7c9e1b3d5f7a9c1e3b5d7f9a1c3e5b7d9f1a3c5e7b9d1f3a"),
		"1058862298",
		models.ObjectDigest("5Hrcm6c7FodZ3gkyBDzGbUnBSiYUEB2ZsTfeGEhVcYXZ"),
	)
	require.NoError(t, err)
	var sent models.SuiExecuteTransactionBlockRequest
	handler, err := NewTransactionHandler(ArgsTxHandler{
		Proxy: &interactors.SuiProxyStub{
			SuiExecuteTransactionBlockCalled: func(_ context.Context, req models.SuiExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
				sent = req
				return models.SuiTransactionBlockResponse{Digest: "digest", Effects: models.SuiEffects{Status: models.ExecutionStatus{Status: successCodeAfterExecution}}}, nil
			},
		},
		Signer:    c.signer,
		GasPrice:  1000,
		GasBudget: 75_000_000,
	})
	require.NoError(t, err)
	_, err = handler.SendTransactionReturnHash(context.Background(), gasCoin, calls)
	require.NoError(t, err)
	require.Len(t, sent.Signature, 1)

	got := map[string]string{
		"goldenMessageHash": hex.EncodeToString(msgHash),
		"goldenRelayerSigs": allSignatures.String(),
		"goldenTxBytes":     sent.TxBytes,
		"goldenTxSignature": sent.Signature[0],
	}
	if os.Getenv(printGoldenVectorEnv) != "" {
		for _, k := range []string{"goldenMessageHash", "goldenRelayerSigs", "goldenTxBytes", "goldenTxSignature"} {
			t.Logf("%s = %q", k, got[k])
		}
		return
	}
	require.Equal(t, goldenMessageHash, got["goldenMessageHash"])
	require.Equal(t, goldenRelayerSigs, got["goldenRelayerSigs"])
	require.Equal(t, goldenTxBytes, got["goldenTxBytes"])
	require.Equal(t, goldenTxSignature, got["goldenTxSignature"])
}
