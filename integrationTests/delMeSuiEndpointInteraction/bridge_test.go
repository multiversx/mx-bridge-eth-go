package delMeSuiEndpointInteraction

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/transaction"
	"github.com/block-vision/sui-go-sdk/utils"
	"github.com/stretchr/testify/require"

	"github.com/block-vision/sui-go-sdk/constant"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
)

const (
	// Dummy mnemonic—only for tests
	testMnemonic = "mail cage popular decrease illegal gravity soda flush cancel claw summer apology"
	// 1 SUI = 1e9 MIST
	minBalance = 1_000_000_000
	// Path to your compiled .mv file
	modulePath = "./contracts/bridgeMock/build/bridgeMock/bytecode_modules/bridgemock.mv"
)

type Batch struct {
	Nonce       uint64
	BlockNumber uint64
}

type ReturnValue struct {
	Bytes []byte
	Type  string
}

// UnmarshalJSON - the returnValues items are split into two arrays - unmarshal each into a flat struct
func (rv *ReturnValue) UnmarshalJSON(data []byte) error {
	var parts []json.RawMessage
	err := json.Unmarshal(data, &parts)
	if err != nil {
		return fmt.Errorf("split into parts: %w", err)
	}
	if len(parts) != 2 {
		return fmt.Errorf("expected 2 elements, got %d", len(parts))
	}

	err = json.Unmarshal(parts[0], &rv.Bytes)
	if err != nil {
		return fmt.Errorf("decode byte array: %w", err)
	}

	err = json.Unmarshal(parts[1], &rv.Type)
	if err != nil {
		return fmt.Errorf("decode type string: %w", err)
	}

	return nil
}

type MyInspect struct {
	ReturnValues []ReturnValue `json:"returnValues"`
}

func readModuleBytes(t *testing.T) []byte {
	b, err := ioutil.ReadFile(modulePath)
	require.NoError(t, err)
	return b
}

func initTest(t *testing.T) (context.Context, sui.ISuiAPI, *signer.Signer) {
	ctx := context.Background()

	acc, err := signer.NewSignertWithMnemonic(testMnemonic)
	require.NoError(t, err)
	t.Logf("Test address: %s", acc.Address)

	cli := sui.NewSuiClient(constant.SuiTestnetEndpoint)

	balResp, err := cli.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{Owner: acc.Address, CoinType: "0x2::sui::SUI"})
	require.NoError(t, err)
	t.Logf("Balance before faucet: %s MIST", balResp.TotalBalance)

	total, err := strconv.ParseInt(balResp.TotalBalance, 10, 64)
	if err != nil {
		t.Fatalf("parse balance %q: %v", balResp.TotalBalance, err)
	}

	require.True(t, total > 0)

	return ctx, cli, acc
}

// TestDeploy deploys a mock bridge contract
func TestDeploy(t *testing.T) {
	ctx, cli, acc := initTest(t)

	mv := readModuleBytes(t)
	encoded := base64.StdEncoding.EncodeToString(mv)
	pubReq := models.PublishRequest{
		Sender:          acc.Address,
		CompiledModules: []string{encoded},
		Dependencies: []string{
			"0x1", // Sui Framework
			"0x2", // Move Standard Library
		},
		GasBudget: "100000000",
	}
	txMeta, err := cli.Publish(ctx, pubReq)
	require.NoError(t, err)

	exec, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: txMeta,
		PriKey:      acc.PriKey,
		Options:     models.SuiTransactionBlockOptions{ShowEffects: true},
		RequestType: "WaitForLocalExecution",
	})

	require.NoError(t, err)
	utils.PrettyPrint(exec)
}

// TestAddBatch puts another batch in our mock contract
func TestAddBatch(t *testing.T) {
	ctx, cli, acc := initTest(t)

	// Grab the new package object ID
	pkgID := "0x89ce33d1d1d85be4b7d46fb563a66c910b960fee962f09fca17d966c59fc8119"
	objID := "0x6486586aab2fe590283d25148e248c2583ffa03d650a6fdf586cc20ee768cb3c"

	// 7. Invoke `add_batch` on our Safe
	callMeta, err := cli.MoveCall(ctx, models.MoveCallRequest{
		Signer:          acc.Address,
		PackageObjectId: pkgID,
		Module:          "bridgemock",
		Function:        "add_batch",
		TypeArguments:   []interface{}{},
		Arguments:       []interface{}{objID, "1", "1"},
		GasBudget:       "100000000",
	})
	require.NoError(t, err)

	callExec, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: callMeta,
		PriKey:      acc.PriKey,
		Options:     models.SuiTransactionBlockOptions{ShowEffects: true},
		RequestType: "WaitForLocalExecution",
	})

	require.NoError(t, err)
	utils.PrettyPrint(callExec)
}

// TestGetBatch fetches a batch and decodes it into a predefined structure
func TestGetBatch(t *testing.T) {
	ctx, cli, acc := initTest(t)

	pkgID := "0x89ce33d1d1d85be4b7d46fb563a66c910b960fee962f09fca17d966c59fc8119"
	objID := "0x6486586aab2fe590283d25148e248c2583ffa03d650a6fdf586cc20ee768cb3c"

	objIDBytes, _ := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(objID))
	txb := transaction.NewTransaction()
	txb.MoveCall(
		models.SuiAddress(pkgID),
		"bridgemock",
		"get_batch",
		[]transaction.TypeTag{},
		// TODO: this should be generalised so the consumer code stays clean - check if the library can do it, probably not
		[]transaction.Argument{
			txb.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             *objIDBytes,
							InitialSharedVersion: 349179941, // I actually hardcoded the version
							Mutable:              true,
						},
					},
				},
			),
			// BatchID
			txb.Pure(uint64(1)),
		},
	)

	bcsEncodedMsg, err := txb.Data.V1.Kind.Marshal()
	txBytes := mystenbcs.ToBase64(bcsEncodedMsg)
	require.NoError(t, err)

	callExec, err := cli.SuiDevInspectTransactionBlock(ctx, models.SuiDevInspectTransactionBlockRequest{
		TxBytes: txBytes,
		Sender:  acc.Address,
	})
	require.NoError(t, err)

	var result []MyInspect
	err = json.Unmarshal(callExec.Results, &result)
	require.NoError(t, err)

	var batch Batch
	// TODO: we should also generalise the parsing of the results and return values. Return values obviously depend
	//  of the contract function implementation and it's straight fw - but when do we have more than one result?
	//  (check with different contract and return types/values/tuples)
	//
	// As a general note, we only pass the byte array from the return value (we don't need the type)
	// Values of the structure need to be in the same order as the ones in the contract - we don't
	//  need annotations, just keep the same ordering
	_, err = mystenbcs.Unmarshal(result[0].ReturnValues[0].Bytes, &batch)

	require.NoError(t, err)
	utils.PrettyPrint(batch)
}
