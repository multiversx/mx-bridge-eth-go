package framework

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/stretchr/testify/require"
)

const (
	setBridgeProxyContractAddressFunction  = "setBridgeProxyContractAddress"
	setWrappingContractAddressFunction     = "setWrappingContractAddress"
	setBridgedTokensWrapperAddressFunction = "setBridgedTokensWrapperAddress"
	setMultiTransferAddressFunction        = "setMultiTransferAddress"
	setEsdtSafeAddressFunction             = "setEsdtSafeAddress"
	setEsdtSafeOnMultiTransferFunction     = "setEsdtSafeOnMultiTransfer"
)

// DeployAndSetContractsForV3p0 will deploy all required contracts (v3.0) on MultiversX side and do the proper wiring
func (handler *MultiversxHandler) DeployAndSetContractsForV3p0(ctx context.Context) {
	handler.deployContractsV3p0(ctx)

	handler.wireMultiTransfer(ctx)
	handler.wireSCProxy(ctx)
	handler.wireSafe(ctx)

	handler.changeOwners(ctx)
	handler.finishSettings(ctx)
}

func (handler *MultiversxHandler) deployContractsV3p0(ctx context.Context) {
	// deploy aggregator
	stakeValue, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	aggregatorDeployParams := []string{
		hex.EncodeToString([]byte("EGLD")),
		hex.EncodeToString(stakeValue.Bytes()),
		"01",
		"02",
		"03",
	}

	for _, oracleKey := range handler.OraclesKeys {
		aggregatorDeployParams = append(aggregatorDeployParams, oracleKey.MvxAddress.Hex())
	}

	hash := ""
	handler.AggregatorAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		normalizePathToRelayersTests(aggregatorContractPath),
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		aggregatorDeployParams,
	)
	require.NotEqual(handler, emptyAddress, handler.AggregatorAddress)
	log.Info("Deploy: aggregator contract", "address", handler.AggregatorAddress, "transaction hash", hash, "num oracles", len(handler.OraclesKeys))

	// deploy wrapper
	handler.WrapperAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		normalizePathToRelayersTests(fmt.Sprintf(wrapperContractPathTemplate, ContractsVersion3p0)),
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.WrapperAddress)
	log.Info("Deploy: wrapper contract", "address", handler.WrapperAddress, "transaction hash", hash)

	// deploy multi-transfer
	handler.MultiTransferAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		normalizePathToRelayersTests(fmt.Sprintf(multiTransferContractPathTemplate, ContractsVersion3p0)),
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.MultiTransferAddress)
	log.Info("Deploy: multi-transfer contract", "address", handler.MultiTransferAddress, "transaction hash", hash)

	// deploy safe
	handler.SafeAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		normalizePathToRelayersTests(fmt.Sprintf(safeContractPathTemplate, ContractsVersion3p0)),
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{
			handler.AggregatorAddress.Hex(),
			handler.MultiTransferAddress.Hex(),
			"01",
		},
	)
	require.NotEqual(handler, emptyAddress, handler.SafeAddress)
	log.Info("Deploy: safe contract", "address", handler.SafeAddress, "transaction hash", hash)

	// deploy bridge proxy
	handler.ScProxyAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		normalizePathToRelayersTests(fmt.Sprintf(bridgeProxyContractPathTemplate, ContractsVersion3p0)),
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{
			handler.MultiTransferAddress.Hex(),
		},
	)
	require.NotEqual(handler, emptyAddress, handler.ScProxyAddress)
	log.Info("Deploy: SC proxy contract", "address", handler.ScProxyAddress, "transaction hash", hash)

	// deploy multisig
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	params := []string{
		handler.SafeAddress.Hex(),
		handler.MultiTransferAddress.Hex(),
		handler.ScProxyAddress.Hex(),
		minRelayerStakeHex,
		slashAmount,
		handler.Quorum}
	for _, relayerKeys := range handler.RelayersKeys {
		params = append(params, relayerKeys.MvxAddress.Hex())
	}
	handler.MultisigAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		normalizePathToRelayersTests(fmt.Sprintf(multisigContractPathTemplate, ContractsVersion3p0)),
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		params,
	)
	require.NotEqual(handler, emptyAddress, handler.MultisigAddress)
	log.Info("Deploy: multisig contract", "address", handler.MultisigAddress, "transaction hash", hash)

	// deploy test-caller
	handler.CalleeScAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		normalizePathToRelayersTests(testCallerContractPath),
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.CalleeScAddress)
	log.Info("Deploy: test-caller contract", "address", handler.CalleeScAddress, "transaction hash", hash)
}

func (handler *MultiversxHandler) wireMultiTransfer(ctx context.Context) {
	// setBridgeProxyContractAddress
	hash, txResult, _ := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgeProxyContractAddressFunction,
		[]string{
			handler.ScProxyAddress.Hex(),
		},
	)
	log.Info("Set in multi-transfer contract the SC proxy contract", "transaction hash", hash, "status", txResult.Status)

	// setWrappingContractAddress
	hash, txResult, _ = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		setWrappingContractAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in multi-transfer contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) wireSCProxy(ctx context.Context) {
	// setBridgedTokensWrapper in SC bridge proxy
	hash, txResult, _ := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgedTokensWrapperAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)

	// setMultiTransferAddress in SC bridge proxy
	hash, txResult, _ = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setMultiTransferAddressFunction,
		[]string{
			handler.MultiTransferAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the multi-transfer contract", "transaction hash", hash, "status", txResult.Status)

	// setEsdtSafeAddress on bridge proxy
	hash, txResult, _ = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setEsdtSafeAddressFunction,
		[]string{
			handler.SafeAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the safe contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) wireSafe(ctx context.Context) {
	// setBridgedTokensWrapperAddress
	hash, txResult, _ := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgedTokensWrapperAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in safe contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)

	//setBridgeProxyContractAddress
	hash, txResult, _ = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgeProxyContractAddressFunction,
		[]string{
			handler.ScProxyAddress.Hex(),
		},
	)
	log.Info("Set in safe contract the SC proxy contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) changeOwners(ctx context.Context) {
	// ChangeOwnerAddress for safe
	hash, txResult, _ := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for safe contract", "transaction hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for multi-transfer
	hash, txResult, _ = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for multi-transfer contract", "transaction hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for bridge proxy
	hash, txResult, _ = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for SC proxy contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) finishSettings(ctx context.Context) {
	// unpause sc proxy
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseProxyFunction)
	log.Info("Un-paused SC proxy contract", "transaction hash", hash, "status", txResult.Status)

	// setEsdtSafeOnMultiTransfer
	hash, txResult, _ = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		setEsdtSafeOnMultiTransferFunction,
		[]string{},
	)
	log.Info("Set in multisig contract the safe contract (automatically)", "transaction hash", hash, "status", txResult.Status)

	// stake relayers on multisig
	handler.stakeAddressesOnContract(ctx, handler.MultisigAddress, handler.RelayersKeys)

	// stake relayers on price aggregator
	handler.stakeAddressesOnContract(ctx, handler.AggregatorAddress, handler.OraclesKeys)

	// unpause multisig
	hash, txResult = handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseFunction)
	log.Info("Un-paused multisig contract", "transaction hash", hash, "status", txResult.Status)

	handler.UnPauseContractsAfterTokenChanges(ctx)
}
