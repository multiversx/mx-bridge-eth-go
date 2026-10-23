package sui

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/block-vision/sui-go-sdk/models"
	v2 "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	"github.com/block-vision/sui-go-sdk/transaction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	grpcRequestTimeout  = 5 * time.Second
	inspectionGasPrice  = 1000
	inspectionGasBudget = 50_000_000
	failureStatus       = "failure"
)

var errUnsupportedByGrpcProxy = errors.New("not supported by the Sui gRPC proxy")

type grpcProxy struct {
	ledger    v2.LedgerServiceClient
	state     v2.StateServiceClient
	execution v2.TransactionExecutionServiceClient
}

// NewGrpcProxy creates a Proxy backed by the Sui full node gRPC API. networkAddress is https://host[:port] or http://host[:port]
func NewGrpcProxy(networkAddress string) (*grpcProxy, error) {
	target, useTLS, err := parseGrpcTarget(networkAddress)
	if err != nil {
		return nil, err
	}

	transportCredentials := insecure.NewCredentials()
	if useTLS {
		transportCredentials = credentials.NewTLS(nil)
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("create Sui gRPC client for %s: %w", target, err)
	}

	return &grpcProxy{
		ledger:    v2.NewLedgerServiceClient(conn),
		state:     v2.NewStateServiceClient(conn),
		execution: v2.NewTransactionExecutionServiceClient(conn),
	}, nil
}

func parseGrpcTarget(networkAddress string) (string, bool, error) {
	parsed, err := url.Parse(networkAddress)
	if err != nil {
		return "", false, fmt.Errorf("invalid Sui network address %q: %w", networkAddress, err)
	}

	port := parsed.Port()
	switch parsed.Scheme {
	case "https":
		if port == "" {
			port = "443"
		}
		return net.JoinHostPort(parsed.Hostname(), port), true, nil
	case "http":
		if port == "" {
			port = "80"
		}
		return net.JoinHostPort(parsed.Hostname(), port), false, nil
	default:
		return "", false, fmt.Errorf("invalid Sui network address %q: expected an http:// or https:// URL", networkAddress)
	}
}

// SuiGetLatestCheckpointSequenceNumber returns the latest checkpoint known by the node
func (proxy *grpcProxy) SuiGetLatestCheckpointSequenceNumber(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcRequestTimeout)
	defer cancel()

	info, err := proxy.ledger.GetServiceInfo(ctx, &v2.GetServiceInfoRequest{})
	if err != nil {
		return 0, err
	}

	return info.GetCheckpointHeight(), nil
}

// SuiXGetCoins returns one page of the owner's coins of the requested type
func (proxy *grpcProxy) SuiXGetCoins(ctx context.Context, req models.SuiXGetCoinsRequest) (models.PaginatedCoinsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, grpcRequestTimeout)
	defer cancel()

	coinType := req.CoinType
	if coinType == "" {
		coinType = suiCoinType
	}
	objectType := fmt.Sprintf("0x2::coin::Coin<%s>", coinType)
	request := &v2.ListOwnedObjectsRequest{
		Owner:      &req.Owner,
		ObjectType: &objectType,
		ReadMask:   &fieldmaskpb.FieldMask{Paths: []string{"object_id", "version", "digest", "balance"}},
	}
	if req.Limit > 0 {
		pageSize := uint32(req.Limit)
		request.PageSize = &pageSize
	}
	if cursor, ok := req.Cursor.(string); ok && cursor != "" {
		pageToken, err := base64.StdEncoding.DecodeString(cursor)
		if err != nil {
			return models.PaginatedCoinsResponse{}, fmt.Errorf("invalid coins cursor: %w", err)
		}
		request.PageToken = pageToken
	}

	response, err := proxy.state.ListOwnedObjects(ctx, request)
	if err != nil {
		return models.PaginatedCoinsResponse{}, err
	}

	return convertOwnedCoins(response, coinType), nil
}

func convertOwnedCoins(response *v2.ListOwnedObjectsResponse, coinType string) models.PaginatedCoinsResponse {
	coins := make([]models.CoinData, 0, len(response.GetObjects()))
	for _, object := range response.GetObjects() {
		coins = append(coins, models.CoinData{
			CoinType:     coinType,
			CoinObjectId: object.GetObjectId(),
			Version:      strconv.FormatUint(object.GetVersion(), 10),
			Digest:       object.GetDigest(),
			Balance:      strconv.FormatUint(object.GetBalance(), 10),
		})
	}

	return models.PaginatedCoinsResponse{
		Data:        coins,
		NextCursor:  base64.StdEncoding.EncodeToString(response.GetNextPageToken()),
		HasNextPage: len(response.GetNextPageToken()) > 0,
	}
}

// SuiDevInspectTransactionBlock runs the transaction kind as a simulation with checks disabled, like the JSON-RPC devInspect
func (proxy *grpcProxy) SuiDevInspectTransactionBlock(ctx context.Context, req models.SuiDevInspectTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	kind, err := base64.StdEncoding.DecodeString(req.TxBytes)
	if err != nil {
		return models.SuiTransactionBlockResponse{}, fmt.Errorf("invalid transaction kind bytes: %w", err)
	}
	sender, err := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(req.Sender))
	if err != nil {
		return models.SuiTransactionBlockResponse{}, fmt.Errorf("invalid sender %s: %w", req.Sender, err)
	}

	ctx, cancel := context.WithTimeout(ctx, grpcRequestTimeout)
	defer cancel()

	checks := v2.SimulateTransactionRequest_DISABLED
	doGasSelection := false
	response, err := proxy.execution.SimulateTransaction(ctx, &v2.SimulateTransactionRequest{
		Transaction:    &v2.Transaction{Bcs: &v2.Bcs{Value: transactionDataForInspection(kind, *sender)}},
		ReadMask:       &fieldmaskpb.FieldMask{Paths: []string{"transaction.effects.status", "command_outputs"}},
		Checks:         &checks,
		DoGasSelection: &doGasSelection,
	})
	if err != nil {
		return models.SuiTransactionBlockResponse{}, err
	}

	return convertSimulation(response)
}

// gRPC simulation takes a full TransactionData while devInspect takes only the TransactionKind
func transactionDataForInspection(kind []byte, sender models.SuiAddressBytes) []byte {
	var data bytes.Buffer
	data.WriteByte(0) // TransactionData::V1
	data.Write(kind)
	data.Write(sender[:])
	data.WriteByte(0) // empty gas payment
	data.Write(sender[:])
	_ = binary.Write(&data, binary.LittleEndian, uint64(inspectionGasPrice))
	_ = binary.Write(&data, binary.LittleEndian, uint64(inspectionGasBudget))
	data.WriteByte(0) // TransactionExpiration::None

	return data.Bytes()
}

func convertSimulation(response *v2.SimulateTransactionResponse) (models.SuiTransactionBlockResponse, error) {
	type inspectResult struct {
		ReturnValues [][]interface{} `json:"returnValues"`
	}

	results := make([]inspectResult, 0, len(response.GetCommandOutputs()))
	for _, commandOutput := range response.GetCommandOutputs() {
		returnValues := make([][]interface{}, 0, len(commandOutput.GetReturnValues()))
		for _, returnValue := range commandOutput.GetReturnValues() {
			valueBytes := returnValue.GetValue().GetValue()
			valueAsNumbers := make([]int, len(valueBytes))
			for i, b := range valueBytes {
				valueAsNumbers[i] = int(b)
			}
			returnValues = append(returnValues, []interface{}{valueAsNumbers, returnValue.GetValue().GetName()})
		}
		results = append(results, inspectResult{ReturnValues: returnValues})
	}

	encodedResults, err := json.Marshal(results)
	if err != nil {
		return models.SuiTransactionBlockResponse{}, err
	}

	return models.SuiTransactionBlockResponse{
		Effects: models.SuiEffects{Status: convertExecutionStatus(response.GetTransaction().GetEffects().GetStatus())},
		Results: encodedResults,
	}, nil
}

// SuiExecuteTransactionBlock submits a signed transaction
func (proxy *grpcProxy) SuiExecuteTransactionBlock(ctx context.Context, req models.SuiExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	txData, err := base64.StdEncoding.DecodeString(req.TxBytes)
	if err != nil {
		return models.SuiTransactionBlockResponse{}, fmt.Errorf("invalid transaction bytes: %w", err)
	}
	signatures := make([]*v2.UserSignature, 0, len(req.Signature))
	for _, encodedSignature := range req.Signature {
		signature, errDecode := base64.StdEncoding.DecodeString(encodedSignature)
		if errDecode != nil {
			return models.SuiTransactionBlockResponse{}, fmt.Errorf("invalid transaction signature: %w", errDecode)
		}
		signatures = append(signatures, &v2.UserSignature{Bcs: &v2.Bcs{Value: signature}})
	}

	ctx, cancel := context.WithTimeout(ctx, grpcRequestTimeout)
	defer cancel()

	response, err := proxy.execution.ExecuteTransaction(ctx, &v2.ExecuteTransactionRequest{
		Transaction: &v2.Transaction{Bcs: &v2.Bcs{Value: txData}},
		Signatures:  signatures,
		ReadMask:    &fieldmaskpb.FieldMask{Paths: []string{"digest", "effects.status"}},
	})
	if err != nil {
		return models.SuiTransactionBlockResponse{}, err
	}

	return models.SuiTransactionBlockResponse{
		Digest:  response.GetTransaction().GetDigest(),
		Effects: models.SuiEffects{Status: convertExecutionStatus(response.GetTransaction().GetEffects().GetStatus())},
	}, nil
}

func convertExecutionStatus(status *v2.ExecutionStatus) models.ExecutionStatus {
	if status.GetSuccess() {
		return models.ExecutionStatus{Status: successCodeAfterExecution}
	}
	if status == nil {
		return models.ExecutionStatus{Status: failureStatus, Error: "missing execution status"}
	}

	description := status.GetError().GetDescription()
	if description == "" {
		description = status.GetError().GetKind().String()
	}

	return models.ExecutionStatus{Status: failureStatus, Error: description}
}

// Pay is not used by the relayer
func (proxy *grpcProxy) Pay(_ context.Context, _ models.PayRequest) (models.TxnMetaData, error) {
	return models.TxnMetaData{}, fmt.Errorf("Pay: %w", errUnsupportedByGrpcProxy)
}

// Publish is not used by the relayer
func (proxy *grpcProxy) Publish(_ context.Context, _ models.PublishRequest) (models.TxnMetaData, error) {
	return models.TxnMetaData{}, fmt.Errorf("Publish: %w", errUnsupportedByGrpcProxy)
}

// SuiXGetBalance is not used by the relayer
func (proxy *grpcProxy) SuiXGetBalance(_ context.Context, _ models.SuiXGetBalanceRequest) (models.CoinBalanceResponse, error) {
	return models.CoinBalanceResponse{}, fmt.Errorf("SuiXGetBalance: %w", errUnsupportedByGrpcProxy)
}

// SuiXGetOwnedObjects is not used by the relayer
func (proxy *grpcProxy) SuiXGetOwnedObjects(_ context.Context, _ models.SuiXGetOwnedObjectsRequest) (models.PaginatedObjectsResponse, error) {
	return models.PaginatedObjectsResponse{}, fmt.Errorf("SuiXGetOwnedObjects: %w", errUnsupportedByGrpcProxy)
}

// SplitCoin is not used by the relayer
func (proxy *grpcProxy) SplitCoin(_ context.Context, _ models.SplitCoinRequest) (models.TxnMetaData, error) {
	return models.TxnMetaData{}, fmt.Errorf("SplitCoin: %w", errUnsupportedByGrpcProxy)
}

// MoveCall is not used by the relayer
func (proxy *grpcProxy) MoveCall(_ context.Context, _ models.MoveCallRequest) (models.TxnMetaData, error) {
	return models.TxnMetaData{}, fmt.Errorf("MoveCall: %w", errUnsupportedByGrpcProxy)
}

// SignAndExecuteTransactionBlock is not used by the relayer
func (proxy *grpcProxy) SignAndExecuteTransactionBlock(_ context.Context, _ models.SignAndExecuteTransactionBlockRequest) (models.SuiTransactionBlockResponse, error) {
	return models.SuiTransactionBlockResponse{}, fmt.Errorf("SignAndExecuteTransactionBlock: %w", errUnsupportedByGrpcProxy)
}

// TransferObject is not used by the relayer
func (proxy *grpcProxy) TransferObject(_ context.Context, _ models.TransferObjectRequest) (models.TxnMetaData, error) {
	return models.TxnMetaData{}, fmt.Errorf("TransferObject: %w", errUnsupportedByGrpcProxy)
}
