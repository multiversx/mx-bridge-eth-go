package sui

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/transaction"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/sui/dtos"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

const suiCoinType = "0x2::sui::SUI"

// ArgsSuiClientDataGetter is the arguments DTO used in the NewSuiClientDataGetter constructor
type ArgsSuiClientDataGetter struct {
	SafeContractAddress   string
	BridgeContractAddress string
	Relayer               *signer.Signer
	Client                sui.ISuiAPI
	Log                   chainCore.Logger
}

type suiClientDataGetter struct {
	safeContractAddress string
	safeObjectIdBytes   models.SuiAddressBytes

	bridgeContractAddress         string
	bridgeObjectIdBytes           models.SuiAddressBytes
	relayer                       *signer.Signer
	client                        *sui.Client
	initialSharedVersionForObject sync.Map // to cache the initial shared version for objects
	log                           chainCore.Logger
	mtx                           sync.Mutex
}

// NewSuiClientDataGetter creates a new instance of type suiClientDataGetter
func NewSuiClientDataGetter(args ArgsSuiClientDataGetter) (*suiClientDataGetter, error) {
	if check.IfNil(args.Log) {
		return nil, clients.ErrNilLogger
	}
	if args.Client == nil {
		return nil, errNilClient
	}
	if args.Relayer == nil {
		return nil, errNilRelayerSigner
	}
	if args.BridgeContractAddress == "" {
		return nil, fmt.Errorf("%w for the BridgeContractAddress argument", errEmptyAddress)
	}
	if args.SafeContractAddress == "" {
		return nil, fmt.Errorf("%w for the SafeContractAddress argument", errEmptyAddress)
	}

	var suiClient, ok = args.Client.(*sui.Client)
	if !ok {
		panic("not sui client")
	}

	safeObjectIdBytes, err := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(args.SafeContractAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to convert address: %w", err)
	}

	bridgeObjectIdBytes, err := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(args.BridgeContractAddress))
	if err != nil {
		return nil, fmt.Errorf("failed to convert address: %w", err)
	}

	return &suiClientDataGetter{
		safeContractAddress:   args.SafeContractAddress,
		safeObjectIdBytes:     *safeObjectIdBytes,
		bridgeContractAddress: args.BridgeContractAddress,
		bridgeObjectIdBytes:   *bridgeObjectIdBytes,
		relayer:               args.Relayer,
		client:                suiClient,
		log:                   args.Log,
	}, nil

}

func (getter *suiClientDataGetter) getGasObject(ctx context.Context) (*transaction.SuiObjectRef, error) {
	gasObjects, err := getter.client.SuiXGetCoins(ctx, models.SuiXGetCoinsRequest{
		Owner:    getter.relayer.Address,
		CoinType: suiCoinType,
		Limit:    1,
	})
	if err != nil {
		return nil, errGetCoinObjectsFailed
	}

	if len(gasObjects.Data) == 0 {
		fmt.Println("NU ai obiecte de gas!")
		return nil, fmt.Errorf("%w for address %s", errNoGasObjects, getter.relayer.Address)
	}

	gasCoin, err := transaction.NewSuiObjectRef(
		models.SuiAddress(gasObjects.Data[0].CoinObjectId),
		gasObjects.Data[0].Version,
		models.ObjectDigest(gasObjects.Data[0].Digest),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gas object ref: %w", err)
	}

	return gasCoin, nil
}

func (getter *suiClientDataGetter) getNewTransaction(ctx context.Context) (*transaction.Transaction, error) {
	gasCoin, err := getter.getGasObject(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas object: %w", err)
	}

	tx := transaction.NewTransaction()
	tx.SetSuiClient(getter.client).
		SetSigner(getter.relayer).
		SetSender(models.SuiAddress(getter.relayer.Address)).
		SetGasPrice(1000).
		SetGasBudget(50000000).
		SetGasPayment([]transaction.SuiObjectRef{*gasCoin}).
		SetGasOwner(models.SuiAddress(getter.relayer.Address))

	return tx, nil
}

func (getter *suiClientDataGetter) getTxBytes(tx *transaction.Transaction) (string, error) {
	bcsEncodedMsg, err := tx.Data.V1.Kind.Marshal()
	if err != nil {
		return "", fmt.Errorf("failed to marshal transaction data: %w", err)
	}
	txBytes := mystenbcs.ToBase64(bcsEncodedMsg)

	return txBytes, nil
}

func (getter *suiClientDataGetter) sendTxGetBlockResponse(ctx context.Context, tx *transaction.Transaction) (models.SuiTransactionBlockResponse, error) {
	txBytes, err := getter.getTxBytes(tx)
	if err != nil {
		return models.SuiTransactionBlockResponse{}, err
	}

	return getter.client.SuiDevInspectTransactionBlock(ctx, models.SuiDevInspectTransactionBlockRequest{
		Sender:  getter.relayer.Address,
		TxBytes: txBytes,
	})
}

func (getter *suiClientDataGetter) getInitialSharedVersionForObject(ctx context.Context, objectId string) (uint64, error) {
	if value, exists := getter.initialSharedVersionForObject.Load(objectId); exists {
		return value.(uint64), nil
	}

	// If not found, make the network call
	rsp, err := getter.client.SuiGetObject(ctx, models.SuiGetObjectRequest{
		ObjectId: objectId,
		Options: models.SuiObjectDataOptions{
			ShowOwner: true,
		},
	})

	if err != nil {
		return 0, fmt.Errorf("failed to get object %s: %w", objectId, err)
	}

	data := rsp.Data
	ownerBytes, err := json.Marshal(data.Owner)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal owner: %w", err)
	}

	var sharedOwner dtos.SharedOwner
	if err = json.Unmarshal(ownerBytes, &sharedOwner); err != nil {
		return 0, fmt.Errorf("failed to unmarshal to SharedOwner: %w", err)
	}

	initialSharedVersion := sharedOwner.Shared.InitialSharedVersion

	if actualValue, loaded := getter.initialSharedVersionForObject.LoadOrStore(objectId, initialSharedVersion); loaded {
		// if another goroutine stored it first, return that value
		return actualValue.(uint64), nil
	}

	return initialSharedVersion, nil
}

// GetBatch returns the batch of transactions by providing the batch nonce
func (getter *suiClientDataGetter) GetBatch(ctx context.Context, batchNonce uint64) (dtos.Batch, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return dtos.Batch{}, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return dtos.Batch{}, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.bridgeContractAddress),
		"bridge",
		"get_batch",
		nil,
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.bridgeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
			tx.Pure(batchNonce),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return dtos.Batch{}, fmt.Errorf("failed to get batch: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return dtos.Batch{}, fmt.Errorf("failed to get batch: %s", txBlockResp.Effects.Status.Error)
	}

	var batch *dtos.Batch
	result := txBlockResp.Results
	returnValues, err := ExtractReturnValues(result)
	if err != nil {
		return dtos.Batch{}, fmt.Errorf("failed to extract return values: %w", err)
	}
	if len(returnValues) > 0 && strings.HasSuffix(returnValues[0].Type, "::Batch") {
		batch, err = DecodeStructFromExtracted(returnValues[0], BatchDecoder)
		if err != nil {
			return dtos.Batch{}, fmt.Errorf("failed to decode batch: %w", err)
		}
	} else {
		return dtos.Batch{}, fmt.Errorf("wrong expected return value, got %s", returnValues[0].Type)
	}

	return *batch, nil
}

// GetBatchDeposits returns the transactions of a batch by providing the batch nonce
func (getter *suiClientDataGetter) GetBatchDeposits(ctx context.Context, batchNonce uint64) ([]dtos.Deposit, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return []dtos.Deposit{}, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.bridgeContractAddress),
		"bridge",
		"get_batch_deposits",
		nil,
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.bridgeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
			tx.Pure(batchNonce),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch deposits: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return nil, fmt.Errorf("failed to get batch: %s", txBlockResp.Effects.Status.Error)
	}

	// TODO: decode result
	return []dtos.Deposit{}, nil
}

// GetRelayers returns all whitelisted sui addresses
func (getter *suiClientDataGetter) GetRelayers(ctx context.Context) ([]models.SuiAddress, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return []models.SuiAddress{}, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.bridgeContractAddress),
		"bridge",
		"get_relayers",
		nil,
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.bridgeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch deposits: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return nil, fmt.Errorf("failed to get batch: %s", txBlockResp.Effects.Status.Error)
	}

	// TODO: decode result
	return []models.SuiAddress{}, nil
}

// WasBatchExecuted returns true if the batch was executed
func (getter *suiClientDataGetter) WasBatchExecuted(ctx context.Context, batchNonce uint64) (bool, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return false, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.bridgeContractAddress),
		"bridge",
		"get_relayers",
		nil,
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.bridgeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
			tx.Pure(batchNonce),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return false, fmt.Errorf("failed to get batch deposits: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return false, fmt.Errorf("failed to get batch: %s", txBlockResp.Effects.Status.Error)
	}

	// TODO: decode result
	return false, nil
}

// IsPaused returns true if the bridge contract is paused
func (getter *suiClientDataGetter) IsPaused(ctx context.Context) (bool, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return false, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.safeContractAddress),
		"bridge",
		"get_pause",
		nil,
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.bridgeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return false, fmt.Errorf("failed to get get_pause: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return false, fmt.Errorf("get_pause transaction failed: %s", txBlockResp.Effects.Status.Error)
	}

	// TODO: decode result
	return false, nil
}

// Quorum returns the current set quorum value
func (getter *suiClientDataGetter) Quorum(ctx context.Context) (uint64, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.bridgeContractAddress),
		"bridge",
		"get_quorum",
		nil,
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.bridgeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf("failed to get quorum: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return 0, fmt.Errorf("get quorum transaction failed: %s", txBlockResp.Effects.Status.Error)
	}

	// TODO: decode result
	return 0, nil
}

// GetStatusesAfterExecution returns the statuses of the last executed transfer
func (getter *suiClientDataGetter) GetStatusesAfterExecution(ctx context.Context, batchNonce uint64) (uint64, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.bridgeContractAddress),
		"bridge",
		"get_statuses_after_execution",
		nil,
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.bridgeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
			tx.Pure(batchNonce),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf("failed to get quorum: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return 0, fmt.Errorf("get quorum transaction failed: %s", txBlockResp.Effects.Status.Error)
	}

	// TODO: decode result
	return 0, nil
}

// TotalBalances returns the total balance of the given token
func (getter *suiClientDataGetter) TotalBalances(ctx context.Context, coinType string) (uint64, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.safeContractAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.safeContractAddress, err)
	}

	coinParts, err := ParseCoinType(coinType)
	if err != nil {
		return 0, fmt.Errorf("failed to parse coin type %s: %w", coinType, err)
	}

	coinIdBytes, err := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(coinParts[0]))
	if err != nil {
		return 0, fmt.Errorf("failed to convert coin type %s: %w", coinType, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.safeContractAddress),
		"safe",
		"get_stored_coin_balance",
		[]transaction.TypeTag{
			{
				Struct: &transaction.StructTag{
					Address: *coinIdBytes,
					Module:  coinParts[1],
					Name:    coinParts[2],
				},
			},
		},
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.safeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return 0, fmt.Errorf("failed to get total balances: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return 0, fmt.Errorf("get total balances transaction failed: %s", txBlockResp.Effects.Status.Error)
	}

	// TODO: decode result
	return 0, nil
}

// WhitelistedTokens returns true if the token is whitelisted
func (getter *suiClientDataGetter) WhitelistedTokens(ctx context.Context, coinType string) (bool, error) {
	tx, err := getter.getNewTransaction(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to create new transaction: %w", err)
	}

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.safeContractAddress)
	if err != nil {
		return false, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.safeContractAddress, err)
	}

	coinParts, err := ParseCoinType(coinType)
	if err != nil {
		return false, fmt.Errorf("failed to parse coin type %s: %w", coinType, err)
	}

	coinIdBytes, err := transaction.ConvertSuiAddressStringToBytes(models.SuiAddress(coinParts[0]))
	if err != nil {
		return false, fmt.Errorf("failed to convert coin type %s: %w", coinType, err)
	}

	tx.MoveCall(
		models.SuiAddress(getter.safeContractAddress),
		"safe",
		"is_token_whitelisted",
		[]transaction.TypeTag{
			{
				Struct: &transaction.StructTag{
					Address: *coinIdBytes,
					Module:  coinParts[1],
					Name:    coinParts[2],
				},
			},
		},
		[]transaction.Argument{
			tx.Object(
				transaction.CallArg{
					Object: &transaction.ObjectArg{
						SharedObject: &transaction.SharedObjectRef{
							ObjectId:             getter.safeObjectIdBytes,
							InitialSharedVersion: initialSharedVersion,
							Mutable:              true,
						},
					},
				},
			),
		},
	)

	txBlockResp, err := getter.sendTxGetBlockResponse(ctx, tx)
	if err != nil {
		return false, fmt.Errorf("failed to get total balances: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return false, fmt.Errorf("get total balances transaction failed: %s", txBlockResp.Effects.Status.Error)
	}

	// TODO: decode result
	return false, nil
}

// GetLatestCheckpoint returns the latest checkpoint sequence number
func (getter *suiClientDataGetter) GetLatestCheckpoint(ctx context.Context) (uint64, error) {
	return getter.client.SuiGetLatestCheckpointSequenceNumber(ctx)
}

// GetBalance returns the sui balance of the given account
func (getter *suiClientDataGetter) GetBalance(ctx context.Context, account string, coinType string) (models.CoinBalanceResponse, error) {
	return getter.client.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{
		Owner:    account,
		CoinType: coinType,
	})
}

// ParseCoinType parses a coin type string in the format "package::module::struct" and returns its components.
func ParseCoinType(coinType string) ([3]string, error) {
	parts := strings.Split(coinType, "::")

	if len(parts) != 3 {
		return [3]string{}, errInvalidCoinType
	}

	packageAddr := strings.TrimSpace(parts[0])
	module := strings.TrimSpace(parts[1])
	structName := strings.TrimSpace(parts[2])

	// Basic validation
	if packageAddr == "" || module == "" || structName == "" {
		return [3]string{}, errors.New("coin type parts cannot be empty")
	}

	return [3]string{packageAddr, module, structName}, nil
}

// ExtractReturnValues extracts and parses return values from json.RawMessage
func ExtractReturnValues(rawResults json.RawMessage) ([]dtos.ReturnValue, error) {
	var returnValues []dtos.ReturnValue

	if len(rawResults) == 0 {
		return returnValues, nil
	}

	var results []dtos.InspectResult
	if err := json.Unmarshal(rawResults, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}

	for _, result := range results {
		for _, rv := range result.ReturnValues {
			if len(rv) >= 2 {
				returnValue := dtos.ReturnValue{
					Value: rv[0],
					Type:  fmt.Sprintf("%v", rv[1]),
				}
				returnValues = append(returnValues, returnValue)
			}
		}
	}

	return returnValues, nil
}

// DecodeStructFromExtracted decodes a struct from the extracted return value.
func DecodeStructFromExtracted[T any](rv dtos.ReturnValue, decoder func([]interface{}) (*T, error)) (*T, error) {
	fields, ok := rv.Value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected struct fields as []interface{}, got %T", rv.Value)
	}
	return decoder(fields)
}

func BatchDecoder(fields []interface{}) (*dtos.Batch, error) {
	if len(fields) != 26 {
		return nil, fmt.Errorf("expected 26 bytes for serialized Sword, got %d", len(fields))
	}

	bytes := make([]byte, 26)
	for i, v := range fields {
		f, ok := v.(float64)
		if !ok {
			return nil, fmt.Errorf("expected float64 byte at index %d, got %T", i, v)
		}
		bytes[i] = byte(f)
	}

	nonce, err := parseU64(bytes[0:8])
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %v", err)
	}

	blockNumber, err := parseU64(bytes[8:16])
	if err != nil {
		return nil, fmt.Errorf("failed to decode block_number: %v", err)
	}

	lastUpdatedBlock, err := parseU64(bytes[16:24])
	if err != nil {
		return nil, fmt.Errorf("failed to decode last_updated_block: %v", err)
	}

	depositsCount, err := parseU16(bytes[24:26])
	if err != nil {
		return nil, fmt.Errorf("failed to decode deposits_count: %v", err)
	}

	return &dtos.Batch{
		Nonce:                  nonce,
		BlockNumber:            blockNumber,
		LastUpdatedBlockNumber: lastUpdatedBlock,
		DepositsCount:          depositsCount,
	}, nil
}

func parseU64(bytes []byte) (uint64, error) {
	if len(bytes) != 8 {
		return 0, fmt.Errorf("expected 8 bytes for u64, got %d", len(bytes))
	}
	return binary.LittleEndian.Uint64(bytes), nil
}

func parseU16(bytes []byte) (uint16, error) {
	if len(bytes) != 2 {
		return 0, fmt.Errorf("expected 2 bytes for u16, got %d", len(bytes))
	}
	return binary.LittleEndian.Uint16(bytes), nil
}
