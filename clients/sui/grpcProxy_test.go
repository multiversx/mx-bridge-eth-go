package sui

import (
	"context"
	"encoding/base64"
	"net"
	"testing"

	"github.com/block-vision/sui-go-sdk/models"
	v2 "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	"github.com/block-vision/sui-go-sdk/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const testSuiAddress = "0x29dfbf688abce7ab43bb8e70cae158ae961196e721440f515482f8ba1684390f"

type executionServerStub struct {
	v2.UnimplementedTransactionExecutionServiceServer
	simulateRequest *v2.SimulateTransactionRequest
	executeRequest  *v2.ExecuteTransactionRequest
}

func (stub *executionServerStub) SimulateTransaction(_ context.Context, req *v2.SimulateTransactionRequest) (*v2.SimulateTransactionResponse, error) {
	stub.simulateRequest = req
	return &v2.SimulateTransactionResponse{
		Transaction: &v2.ExecutedTransaction{Effects: &v2.TransactionEffects{Status: &v2.ExecutionStatus{Success: proto.Bool(true)}}},
		CommandOutputs: []*v2.CommandResult{{ReturnValues: []*v2.CommandOutput{
			{Value: &v2.Bcs{Value: []byte{3, 0, 0, 0, 0, 0, 0, 0}}},
		}}},
	}, nil
}

func (stub *executionServerStub) ExecuteTransaction(_ context.Context, req *v2.ExecuteTransactionRequest) (*v2.ExecuteTransactionResponse, error) {
	stub.executeRequest = req
	return &v2.ExecuteTransactionResponse{Transaction: &v2.ExecutedTransaction{
		Digest:  proto.String("digest"),
		Effects: &v2.TransactionEffects{Status: &v2.ExecutionStatus{Success: proto.Bool(true)}},
	}}, nil
}

func startExecutionServer(t *testing.T) (*grpcProxy, *executionServerStub) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	stub := &executionServerStub{}
	v2.RegisterTransactionExecutionServiceServer(server, stub)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	proxy, err := NewGrpcProxy("http://" + listener.Addr().String())
	require.NoError(t, err)
	return proxy, stub
}

func TestParseGrpcTarget(t *testing.T) {
	t.Parallel()

	target, useTLS, err := parseGrpcTarget("https://fullnode.mainnet.sui.io:443")
	require.NoError(t, err)
	assert.Equal(t, "fullnode.mainnet.sui.io:443", target)
	assert.True(t, useTLS)

	target, useTLS, err = parseGrpcTarget("https://fullnode.mainnet.sui.io")
	require.NoError(t, err)
	assert.Equal(t, "fullnode.mainnet.sui.io:443", target)
	assert.True(t, useTLS)

	target, useTLS, err = parseGrpcTarget("http://localhost:9000")
	require.NoError(t, err)
	assert.Equal(t, "localhost:9000", target)
	assert.False(t, useTLS)

	_, _, err = parseGrpcTarget("fullnode.mainnet.sui.io:443")
	assert.Error(t, err)
}

func TestTransactionDataForInspection_MatchesSdkEncoding(t *testing.T) {
	t.Parallel()

	sender, err := transaction.ConvertSuiAddressStringToBytes(testSuiAddress)
	require.NoError(t, err)

	tx := transaction.NewTransaction()
	tx.MoveCall("0xb3a692af64df57943e376279995f211e3c1734d6c1c3b833a37ed039ae61bc04", "bridge", "get_quorum", nil,
		[]transaction.Argument{tx.Pure(uint64(7))})
	kind, err := tx.Data.V1.Kind.Marshal()
	require.NoError(t, err)

	tx.SetSender(testSuiAddress).
		SetGasOwner(testSuiAddress).
		SetGasPayment([]transaction.SuiObjectRef{}).
		SetGasPrice(inspectionGasPrice).
		SetGasBudget(inspectionGasBudget)
	expected, err := tx.Data.Marshal()
	require.NoError(t, err)

	assert.Equal(t, expected, transactionDataForInspection(kind, *sender))
}

func TestConvertSimulation_DecodesThroughExistingDecoder(t *testing.T) {
	t.Parallel()

	response := &v2.SimulateTransactionResponse{
		Transaction: &v2.ExecutedTransaction{Effects: &v2.TransactionEffects{Status: &v2.ExecutionStatus{Success: proto.Bool(true)}}},
		CommandOutputs: []*v2.CommandResult{{ReturnValues: []*v2.CommandOutput{
			{Value: &v2.Bcs{Value: []byte{3, 0, 0, 0, 0, 0, 0, 0}}},
			{Value: &v2.Bcs{Value: []byte{1}}},
		}}},
	}
	converted, err := convertSimulation(response)
	require.NoError(t, err)
	assert.Equal(t, successCodeAfterExecution, converted.Effects.Status.Status)

	dg, err := NewSuiClientDataGetter(createMockArgsSuiClientDataGetter())
	require.NoError(t, err)
	var quorum uint64
	var flag bool
	require.NoError(t, dg.decodeReturnValues(converted.Results, &quorum, &flag))
	assert.Equal(t, uint64(3), quorum)
	assert.True(t, flag)
}

func TestConvertExecutionStatus(t *testing.T) {
	t.Parallel()

	assert.Equal(t, models.ExecutionStatus{Status: successCodeAfterExecution}, convertExecutionStatus(&v2.ExecutionStatus{Success: proto.Bool(true)}))
	assert.Equal(t, models.ExecutionStatus{Status: failureStatus, Error: "MoveAbort code 7"}, convertExecutionStatus(&v2.ExecutionStatus{
		Success: proto.Bool(false),
		Error:   &v2.ExecutionError{Description: proto.String("MoveAbort code 7")},
	}))
	assert.Equal(t, models.ExecutionStatus{Status: failureStatus, Error: "INSUFFICIENT_GAS"}, convertExecutionStatus(&v2.ExecutionStatus{
		Success: proto.Bool(false),
		Error:   &v2.ExecutionError{Kind: v2.ExecutionError_INSUFFICIENT_GAS.Enum()},
	}))
	assert.Equal(t, failureStatus, convertExecutionStatus(nil).Status)
}

func TestConvertOwnedCoins(t *testing.T) {
	t.Parallel()

	coins := convertOwnedCoins(&v2.ListOwnedObjectsResponse{
		Objects: []*v2.Object{{
			ObjectId: proto.String("0x63eb"),
			Version:  proto.Uint64(1055631385),
			Digest:   proto.String("BfNgQDhhzDvUNbLGtZkXWDJQontDMKwWqXfMWFuxZTSR"),
			Balance:  proto.Uint64(685106282),
		}},
		NextPageToken: []byte("next"),
	}, suiCoinType)

	assert.Equal(t, []models.CoinData{{
		CoinType:     suiCoinType,
		CoinObjectId: "0x63eb",
		Version:      "1055631385",
		Digest:       "BfNgQDhhzDvUNbLGtZkXWDJQontDMKwWqXfMWFuxZTSR",
		Balance:      "685106282",
	}}, coins.Data)
	assert.True(t, coins.HasNextPage)
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("next")), coins.NextCursor)
}

func TestGrpcProxy_DevInspectAndExecuteRequests(t *testing.T) {
	t.Parallel()

	proxy, server := startExecutionServer(t)
	kind := []byte{0, 1, 2}

	inspected, err := proxy.SuiDevInspectTransactionBlock(context.Background(), models.SuiDevInspectTransactionBlockRequest{
		Sender:  testSuiAddress,
		TxBytes: base64.StdEncoding.EncodeToString(kind),
	})
	require.NoError(t, err)
	assert.Equal(t, successCodeAfterExecution, inspected.Effects.Status.Status)
	sender, _ := transaction.ConvertSuiAddressStringToBytes(testSuiAddress)
	assert.Equal(t, transactionDataForInspection(kind, *sender), server.simulateRequest.GetTransaction().GetBcs().GetValue())
	assert.Equal(t, v2.SimulateTransactionRequest_DISABLED, server.simulateRequest.GetChecks())
	assert.False(t, server.simulateRequest.GetDoGasSelection())

	executed, err := proxy.SuiExecuteTransactionBlock(context.Background(), models.SuiExecuteTransactionBlockRequest{
		TxBytes:   base64.StdEncoding.EncodeToString([]byte{0xaa}),
		Signature: []string{base64.StdEncoding.EncodeToString([]byte{0x00, 0x01})},
	})
	require.NoError(t, err)
	assert.Equal(t, "digest", executed.Digest)
	assert.Equal(t, successCodeAfterExecution, executed.Effects.Status.Status)
	assert.Equal(t, []byte{0xaa}, server.executeRequest.GetTransaction().GetBcs().GetValue())
	require.Len(t, server.executeRequest.GetSignatures(), 1)
	assert.Equal(t, []byte{0x00, 0x01}, server.executeRequest.GetSignatures()[0].GetBcs().GetValue())
}
