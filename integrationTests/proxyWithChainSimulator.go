package integrationTests

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	apiCore "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-simulator-go/config"
	"github.com/multiversx/mx-chain-simulator-go/pkg/factory"
	"github.com/multiversx/mx-chain-simulator-go/pkg/process"
	"github.com/multiversx/mx-chain-simulator-go/pkg/proxy"
	"github.com/multiversx/mx-chain-simulator-go/pkg/proxy/configs"
	"github.com/multiversx/mx-chain-simulator-go/pkg/proxy/creator"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

var log = logger.GetOrCreate("testscommon/chainsimulator")

const (
	proxyURL     = "http://127.0.0.1:8085"
	thousandEgld = "1000000000000000000000"
)

// ArgProxyWithChainSimulator is the dto used to create a new instance of proxy that relies on a chain simulator
type ArgProxyWithChainSimulator struct {
	BypassTxsSignature           bool
	WorkingDir                   string
	RoundDurationInMs            uint64
	RoundsPerEpoch               uint64
	NodeConfigs                  string
	ProxyConfigs                 string
	NumOfShards                  uint32
	BlockTimeInMs                uint64
	ServerPort                   int
	ProxyCacherExpirationSeconds uint64
	ProxyMaxNoncesDelta          int
}

type proxyWithChainSimulator struct {
	simulatorProxyInstance proxy.ProxyHandler
	blocksGenerator        process.BlocksGenerator
	proxyInstance          multiversx.Proxy
	simulator              chainSimulatorHandler
	pkConv                 core.PubkeyConverter
	roundDuration          time.Duration
}

// CreateProxyWithChainSimulator creates a new instance of chain simulator with proxy
func CreateProxyWithChainSimulator(args ArgProxyWithChainSimulator) (*proxyWithChainSimulator, error) {
	roundDurationInMillis := args.RoundDurationInMs
	rounds := core.OptionalUint64{
		HasValue: true,
		Value:    args.RoundsPerEpoch,
	}

	apiConfigurator := api.NewFreePortAPIConfigurator("localhost")
	argsChainSimulator := chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: args.BypassTxsSignature,
		TempDir:                args.WorkingDir,
		PathToInitialConfig:    args.NodeConfigs,
		NumOfShards:            args.NumOfShards,
		GenesisTimestamp:       time.Now().Unix(),
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         rounds,
		ApiInterface:           apiConfigurator,
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
		InitialRound:           0,
		InitialNonce:           0,
		InitialEpoch:           0,
	}
	simulator, err := chainSimulator.NewChainSimulator(argsChainSimulator)
	if err != nil {
		return nil, err
	}

	log.Info("simulators were initialized")

	err = simulator.GenerateBlocks(1)
	if err != nil {
		return nil, err
	}

	generator, err := factory.CreateBlocksGenerator(simulator, config.BlocksGeneratorConfig{
		AutoGenerateBlocks: true,
		BlockTimeInMs:      args.BlockTimeInMs,
	})
	if err != nil {
		return nil, err
	}

	metaNode := simulator.GetNodeHandler(core.MetachainShardId)
	restApiInterfaces := simulator.GetRestAPIInterfaces()
	outputProxyConfigs, err := configs.CreateProxyConfigs(configs.ArgsProxyConfigs{
		TemDir:            args.WorkingDir,
		PathToProxyConfig: args.ProxyConfigs,
		ServerPort:        args.ServerPort,
		RestApiInterfaces: restApiInterfaces,
		InitialWallets:    simulator.GetInitialWalletKeys().ShardWallets,
	})
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second)

	simulatorProxyInstance, err := creator.CreateProxy(creator.ArgsProxy{
		Config:        outputProxyConfigs.Config,
		NodeHandler:   metaNode,
		PathToConfig:  outputProxyConfigs.PathToTempConfig,
		PathToPemFile: outputProxyConfigs.PathToPemFile,
	})
	if err != nil {
		return nil, err
	}

	simulatorProxyInstance.Start()

	log.Info("chain simulator proxy was started")

	argsProxy := blockchain.ArgsProxy{
		ProxyURL:            proxyURL,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		AllowedDeltaToFinal: args.ProxyMaxNoncesDelta,
		CacheExpirationTime: time.Second * time.Duration(args.ProxyCacherExpirationSeconds),
		EntityType:          sdkCore.Proxy,
	}
	proxyInstance, err := blockchain.NewProxy(argsProxy)
	if err != nil {
		return nil, err
	}

	pubKeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	if err != nil {
		return nil, err
	}

	return &proxyWithChainSimulator{
		simulatorProxyInstance: simulatorProxyInstance,
		blocksGenerator:        generator,
		proxyInstance:          proxyInstance,
		simulator:              simulator,
		pkConv:                 pubKeyConverter,
		roundDuration:          time.Duration(args.RoundDurationInMs) * time.Millisecond,
	}, nil
}

// Proxy returns the managed proxy instance
func (instance *proxyWithChainSimulator) Proxy() multiversx.Proxy {
	return instance.proxyInstance
}

// GetNetworkAddress returns the network address
func (instance *proxyWithChainSimulator) GetNetworkAddress() string {
	return proxyURL
}

// DeploySC will deploy the provided smart contract and return its address
func (instance *proxyWithChainSimulator) DeploySC(ctx context.Context, wasmFilePath string, ownerPK string, ownerSK []byte, parameters []string) (string, error) {
	networkConfig, err := instance.proxyInstance.GetNetworkConfig(ctx)
	if err != nil {
		return "", err
	}

	nonce, err := instance.getNonce(ctx, ownerPK)
	if err != nil {
		return "", err
	}

	emptyAddress, err := instance.pkConv.Encode(make([]byte, 32))
	if err != nil {
		return "", err
	}

	scCode := wasm.GetSCCode(wasmFilePath)
	params := []string{scCode, wasm.VMTypeHex, wasm.DummyCodeMetadataHex}
	params = append(params, parameters...)
	txData := strings.Join(params, "@")

	ftx := &transaction.FrontendTransaction{
		Nonce:    nonce,
		Value:    "0",
		Receiver: emptyAddress,
		Sender:   ownerPK,
		GasPrice: networkConfig.MinGasPrice,
		GasLimit: 600000000,
		Data:     []byte(txData),
		ChainID:  networkConfig.ChainID,
		Version:  1,
	}

	hash, err := instance.signAndSend(ctx, ownerSK, ftx)
	if err != nil {
		return "", err
	}

	log.Info("contract deployed", "hash", hash)

	txResult, errGet := instance.GetTransactionResult(ctx, hash)
	if errGet != nil {
		return "", errGet
	}

	return txResult.Logs.Events[0].Address, nil
}

// GetTransactionResult tries to get a transaction result. It may wait a few blocks
func (instance *proxyWithChainSimulator) GetTransactionResult(ctx context.Context, hash string) (data.TransactionOnNetwork, error) {
	txResult, err := instance.getTxInfoWithResultsIfTxProcessingFinished(ctx, hash)
	if err == nil && txResult != nil {
		return *txResult, nil
	}

	// wait for tx to be done, in order to get the contract address
	timeoutTimer := time.NewTimer(instance.roundDuration * 20)
	for {
		select {
		case <-time.After(instance.roundDuration):
			txResult, err = instance.getTxInfoWithResultsIfTxProcessingFinished(ctx, hash)
			if err == nil && txResult != nil {
				return *txResult, nil
			}
			if err != nil {
				return data.TransactionOnNetwork{}, err
			}
		case <-timeoutTimer.C:
			return data.TransactionOnNetwork{}, errors.New("timeout")
		}
	}
}

func (instance *proxyWithChainSimulator) getTxInfoWithResultsIfTxProcessingFinished(ctx context.Context, hash string) (*data.TransactionOnNetwork, error) {
	txStatus, err := instance.proxyInstance.ProcessTransactionStatus(ctx, hash)
	if err != nil {
		return nil, err
	}

	if txStatus == transaction.TxStatusPending {
		return nil, nil
	}

	if txStatus != transaction.TxStatusSuccess {
		log.Warn("something went wrong with the transaction", "hash", hash, "status", txStatus)
	}

	txResult, errGet := instance.proxyInstance.GetTransactionInfoWithResults(ctx, hash)
	if errGet != nil {
		return nil, err
	}

	return &txResult.Data.Transaction, nil

}

// ScCall will make the provided sc call
func (instance *proxyWithChainSimulator) ScCall(ctx context.Context, senderPK string, senderSK []byte, contract string, value string, function string, parameters []string) (string, error) {
	params := []string{function}
	params = append(params, parameters...)
	txData := strings.Join(params, "@")

	return instance.SendTx(ctx, senderPK, senderSK, contract, value, []byte(txData))
}

// SendTx will build and send a transaction
func (instance *proxyWithChainSimulator) SendTx(ctx context.Context, senderPK string, senderSK []byte, receiver string, value string, dataField []byte) (string, error) {
	networkConfig, err := instance.proxyInstance.GetNetworkConfig(ctx)
	if err != nil {
		return "", err
	}

	nonce, err := instance.getNonce(ctx, senderPK)
	if err != nil {
		return "", err
	}

	ftx := &transaction.FrontendTransaction{
		Nonce:    nonce,
		Value:    value,
		Receiver: receiver,
		Sender:   senderPK,
		GasPrice: networkConfig.MinGasPrice,
		GasLimit: 600000000,
		Data:     dataField,
		ChainID:  networkConfig.ChainID,
		Version:  1,
	}

	return instance.signAndSend(ctx, senderSK, ftx)
}

// FundWallets sends funds to the provided addresses
func (instance *proxyWithChainSimulator) FundWallets(wallets []string) {
	addressesState := make([]*dtos.AddressState, 0, len(wallets))
	for _, wallet := range wallets {
		addressesState = append(addressesState, &dtos.AddressState{
			Address: wallet,
			Nonce:   new(uint64),
			Balance: thousandEgld,
		})
	}
	err := instance.simulator.SetStateMultiple(addressesState)
	log.LogIfError(err)
}

// GetESDTBalance returns the balance of the esdt token for the provided address
func (instance *proxyWithChainSimulator) GetESDTBalance(ctx context.Context, address sdkCore.AddressHandler, token string) (string, error) {
	tokenData, err := instance.proxyInstance.GetESDTTokenData(ctx, address, token, apiCore.AccountQueryOptions{
		OnFinalBlock: true,
	})
	if err != nil {
		return "", err
	}

	return tokenData.Balance, nil
}

func (instance *proxyWithChainSimulator) getNonce(ctx context.Context, bech32Address string) (uint64, error) {
	address, err := data.NewAddressFromBech32String(bech32Address)
	if err != nil {
		return 0, err
	}

	account, err := instance.proxyInstance.GetAccount(ctx, address)
	if err != nil {
		return 0, err
	}

	return account.Nonce, nil
}

func (instance *proxyWithChainSimulator) signAndSend(ctx context.Context, senderSK []byte, ftx *transaction.FrontendTransaction) (string, error) {
	sig, err := computeTransactionSignature(senderSK, ftx)
	if err != nil {
		return "", err
	}
	ftx.Signature = hex.EncodeToString(sig)

	return instance.proxyInstance.SendTransaction(ctx, ftx)
}

func computeTransactionSignature(senderSk []byte, tx *transaction.FrontendTransaction) ([]byte, error) {
	signer := &singlesig.Ed25519Signer{}
	keyGenerator := signing.NewKeyGenerator(ed25519.NewEd25519())

	senderSk, err := hex.DecodeString(string(senderSk))
	if err != nil {
		return nil, err
	}

	privateKey, err := keyGenerator.PrivateKeyFromByteArray(senderSk)
	if err != nil {
		return nil, err
	}

	dataToSign, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	return signer.Sign(privateKey, dataToSign)
}

// Close closes the internal components
func (instance *proxyWithChainSimulator) Close() {
	instance.blocksGenerator.Close()
	instance.simulatorProxyInstance.Close()
}
