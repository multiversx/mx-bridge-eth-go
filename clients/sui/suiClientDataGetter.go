package sui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/mystenbcs"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/transaction"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/sui/dtos"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// ArgsSuiClientDataGetter is the arguments DTO used in the NewSuiClientDataGetter constructor
type ArgsSuiClientDataGetter struct {
	SafeContractAddress   string
	BridgeContractAddress string
	RelayerAddress        string
	Client                sui.ISuiAPI
	Log                   chainCore.Logger
}

type suiClientDataGetter struct {
	safeContractAddress string
	safeObjectIdBytes   models.SuiAddressBytes

	bridgeContractAddress         string
	bridgeObjectIdBytes           models.SuiAddressBytes
	relayerAddress                string
	client                        sui.ISuiAPI
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
	if args.RelayerAddress == "" {
		return nil, fmt.Errorf("%w for the signer address argument", errNilAddress)
	}
	if args.BridgeContractAddress == "" {
		return nil, fmt.Errorf("%w for the BridgePackageId argument", errNilPackageId)
	}
	if args.SafeContractAddress == "" {
		return nil, fmt.Errorf("%w for the SafePackageId argument", errNilPackageId)
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
		relayerAddress:        args.RelayerAddress,
		client:                args.Client,
		log:                   args.Log,
	}, nil
}

// GetBatchByNonce returns the batch of transactions by providing the batch nonce
func (getter *suiClientDataGetter) GetBatchByNonce(ctx context.Context, batchNonce uint64) (dtos.Batch, bool, error) {
	tx := transaction.NewTransaction()

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return dtos.Batch{}, false, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
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
		return dtos.Batch{}, false, fmt.Errorf("failed to get batch: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return dtos.Batch{}, false, fmt.Errorf("failed to get batch: %s", txBlockResp.Effects.Status.Error)
	}

	var batch dtos.Batch
	var isFinalBatch bool
	err = DecodeReturnValues(txBlockResp.Results, &batch, &isFinalBatch)
	if err != nil {
		return dtos.Batch{}, false, fmt.Errorf("failed to decode return value: %w", err)
	}

	return batch, isFinalBatch, nil
}

// GetBatchDeposits returns the transactions of a batch by providing the batch nonce
func (getter *suiClientDataGetter) GetBatchDeposits(ctx context.Context, batchNonce uint64) ([]dtos.Deposit, bool, error) {
	tx := transaction.NewTransaction()

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
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
		return nil, false, fmt.Errorf("failed to get batch deposits: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return nil, false, fmt.Errorf("failed to get batch: %s", txBlockResp.Effects.Status.Error)
	}

	var depositsList []dtos.Deposit
	var areFinalDeposits bool
	err = DecodeReturnValues(txBlockResp.Results, &depositsList, &areFinalDeposits)
	if err != nil {
		return nil, false, fmt.Errorf("failed to decode return value: %w", err)
	}

	return depositsList, areFinalDeposits, nil
}

// GetRelayers returns all whitelisted sui addresses
func (getter *suiClientDataGetter) GetRelayers(ctx context.Context) ([]models.SuiAddress, error) {
	tx := transaction.NewTransaction()

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

	var relayersAddresses []models.SuiAddress
	err = DecodeReturnValues(txBlockResp.Results, &relayersAddresses)
	if err != nil {
		return []models.SuiAddress{}, fmt.Errorf("failed to decode return value: %w", err)
	}

	return relayersAddresses, nil
}

// WasBatchExecuted returns true if the batch was executed
func (getter *suiClientDataGetter) WasBatchExecuted(ctx context.Context, batchNonce uint64) (bool, error) {
	tx := transaction.NewTransaction()

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

	var wasBatchExecuted bool
	err = DecodeReturnValues(txBlockResp.Results, &wasBatchExecuted)
	if err != nil {
		return false, fmt.Errorf("failed to decode return value: %w", err)
	}

	return wasBatchExecuted, nil
}

// IsPaused returns true if the bridge contract is paused
func (getter *suiClientDataGetter) IsPaused(ctx context.Context) (bool, error) {
	tx := transaction.NewTransaction()

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

	var isPaused bool
	err = DecodeReturnValues(txBlockResp.Results, &isPaused)
	if err != nil {
		return false, fmt.Errorf("failed to decode return value: %w", err)
	}

	return isPaused, nil
}

// Quorum returns the current set quorum value
func (getter *suiClientDataGetter) Quorum(ctx context.Context) (uint64, error) {
	tx := transaction.NewTransaction()

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

	var quorum uint64
	err = DecodeReturnValues(txBlockResp.Results, &quorum)
	if err != nil {
		return 0, fmt.Errorf("failed to decode return value: %w", err)
	}

	return quorum, nil
}

// GetStatusesAfterExecution returns the statuses of the last executed transfer
func (getter *suiClientDataGetter) GetStatusesAfterExecution(ctx context.Context, batchNonce uint64) ([]byte, bool, error) {
	tx := transaction.NewTransaction()

	initialSharedVersion, err := getter.getInitialSharedVersionForObject(ctx, getter.bridgeContractAddress)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get initial shared version for object %s: %w", getter.bridgeContractAddress, err)
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
		return nil, false, fmt.Errorf("failed to get quorum: %w", err)
	}

	if txBlockResp.Effects.Status.Status != "success" {
		return nil, false, fmt.Errorf("get quorum transaction failed: %s", txBlockResp.Effects.Status.Error)
	}

	var depositStatuses []byte
	var isFinal bool
	err = DecodeReturnValues(txBlockResp.Results, &depositStatuses, &isFinal)
	if err != nil {
		return nil, false, fmt.Errorf("failed to decode return value: %w", err)
	}

	return depositStatuses, isFinal, nil
}

// GetTotalBalanceFromSafe returns the total balance of the given token
func (getter *suiClientDataGetter) GetTotalBalanceFromSafe(ctx context.Context, coinType string) (uint64, error) {
	tx := transaction.NewTransaction()

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

	var totalBalance uint64
	err = DecodeReturnValues(txBlockResp.Results, &totalBalance)
	if err != nil {
		return 0, fmt.Errorf("failed to decode return value: %w", err)
	}

	return totalBalance, nil
}

// IsTokenWhitelisted returns true if the token is whitelisted
func (getter *suiClientDataGetter) IsTokenWhitelisted(ctx context.Context, coinType string) (bool, error) {
	tx := transaction.NewTransaction()

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

	var isTokenWhitelisted bool
	err = DecodeReturnValues(txBlockResp.Results, &isTokenWhitelisted)
	if err != nil {
		return false, fmt.Errorf("failed to decode return value: %w", err)
	}

	return isTokenWhitelisted, nil
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

// DecodeReturnValues decodes the return value from a transaction block response.
func DecodeReturnValues(data json.RawMessage, out ...interface{}) error {
	var results []dtos.InspectResult
	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("decode dev inspect results: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no dev inspect results")
	}

	returnValues := results[0].ReturnValues
	if len(returnValues) < len(out) {
		return fmt.Errorf("expected at least %d return values, got %d", len(out), len(returnValues))
	}

	for i := range out {
		if _, err := mystenbcs.Unmarshal(returnValues[i].Bytes, out[i]); err != nil {
			return fmt.Errorf("unmarshal return value %d: %w", i, err)
		}
	}

	return nil
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
		Sender:  getter.relayerAddress,
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

func (getter *suiClientDataGetter) IsInterfaceNil() bool {
	return getter == nil
}
