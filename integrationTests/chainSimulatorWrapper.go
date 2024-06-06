package integrationTests

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
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
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	sdkHttp "github.com/multiversx/mx-sdk-go/core/http"
	"github.com/multiversx/mx-sdk-go/data"
)

var log = logger.GetOrCreate("testscommon/chainsimulator")

const (
	proxyURL                                = "http://127.0.0.1:8085"
	thousandEgld                            = "1000000000000000000000"
	maxAllowedTimeout                       = time.Second
	setMultipleEndpoint                     = "simulator/set-state-overwrite"
	generateBlocksEndpoint                  = "simulator/generate-blocks/%d"
	generateBlocksUntilEpochReachedEndpoint = "simulator/generate-blocks-until-epoch-reached/%d"
	numProbeRetries                         = 10
)

// ArgChainSimulatorWrapper is the DTO used to create a new instance of proxy that relies on a chain simulator
type ArgChainSimulatorWrapper struct {
	ProxyCacherExpirationSeconds uint64
	ProxyMaxNoncesDelta          int
}

type chainSimulatorWrapper struct {
	clientWrapper httpClientWrapper
	proxyInstance multiversx.Proxy
	pkConv        core.PubkeyConverter
}

// CreateChainSimulatorWrapper creates a new instance of the chain simulator wrapper
func CreateChainSimulatorWrapper(args ArgChainSimulatorWrapper) (*chainSimulatorWrapper, error) {
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

	instance := &chainSimulatorWrapper{
		clientWrapper: sdkHttp.NewHttpClientWrapper(nil, proxyURL),
		proxyInstance: proxyInstance,
		pkConv:        pubKeyConverter,
	}

	err = instance.probeURLWithRetries()
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (instance *chainSimulatorWrapper) probeURLWithRetries() error {
	// at this point we should be able to get the network configs

	var err error
	for i := 0; i < numProbeRetries; i++ {
		log.Info("trying to probe the chain simulator", "url", proxyURL, "try", i)

		ctx, done := context.WithTimeout(context.Background(), maxAllowedTimeout)
		_, err = instance.proxyInstance.GetNetworkConfig(ctx)
		done()

		if err == nil {
			log.Info("probe ok, chain simulator instance found", "url", proxyURL)
			return nil
		}

		time.Sleep(maxAllowedTimeout)
	}

	return fmt.Errorf("%w while probing the network config. Please ensure that a chain simulator is running on %s", err, proxyURL)
}

// Proxy returns the managed proxy instance
func (instance *chainSimulatorWrapper) Proxy() multiversx.Proxy {
	return instance.proxyInstance
}

// GetNetworkAddress returns the network address
func (instance *chainSimulatorWrapper) GetNetworkAddress() string {
	return proxyURL
}

// DeploySC will deploy the provided smart contract and return its address
func (instance *chainSimulatorWrapper) DeploySC(ctx context.Context, wasmFilePath string, ownerPK string, ownerSK []byte, parameters []string) (string, error) {
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
func (instance *chainSimulatorWrapper) GetTransactionResult(ctx context.Context, hash string) (*data.TransactionOnNetwork, error) {
	instance.GenerateBlocks(ctx, 10)

	return instance.getTxInfoWithResultsIfTxProcessingFinished(ctx, hash)
}

// GenerateBlocks calls the chain simulator generate block endpoint
func (instance *chainSimulatorWrapper) GenerateBlocks(ctx context.Context, numBlocks int) {
	_, status, err := instance.clientWrapper.PostHTTP(ctx, fmt.Sprintf(generateBlocksEndpoint, numBlocks), nil)
	if err != nil || status != http.StatusOK {
		log.Error("error in chainSimulatorWrapper.GenerateBlocks", "error", err, "status", status)
		return
	}
}

// GenerateBlocksUntilEpochReached will generate blocks until the provided epoch is reached
func (instance *chainSimulatorWrapper) GenerateBlocksUntilEpochReached(ctx context.Context, epoch uint32) {
	_, status, err := instance.clientWrapper.PostHTTP(ctx, fmt.Sprintf(generateBlocksUntilEpochReachedEndpoint, epoch), nil)
	if err != nil || status != http.StatusOK {
		log.Error("error in chainSimulatorWrapper.GenerateBlocksUntilEpochReached", "error", err, "status", status)
		return
	}
}

func (instance *chainSimulatorWrapper) getTxInfoWithResultsIfTxProcessingFinished(ctx context.Context, hash string) (*data.TransactionOnNetwork, error) {
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
func (instance *chainSimulatorWrapper) ScCall(ctx context.Context, senderPK string, senderSK []byte, contract string, value string, function string, parameters []string) (string, error) {
	params := []string{function}
	params = append(params, parameters...)
	txData := strings.Join(params, "@")

	return instance.SendTx(ctx, senderPK, senderSK, contract, value, []byte(txData))
}

// SendTx will build and send a transaction
func (instance *chainSimulatorWrapper) SendTx(ctx context.Context, senderPK string, senderSK []byte, receiver string, value string, dataField []byte) (string, error) {
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
func (instance *chainSimulatorWrapper) FundWallets(ctx context.Context, wallets []string) {
	addressesState := make([]*dtos.AddressState, 0, len(wallets))
	for _, wallet := range wallets {
		addressesState = append(addressesState, &dtos.AddressState{
			Address: wallet,
			Nonce:   new(uint64),
			Balance: thousandEgld,
		})
	}

	buff, err := json.Marshal(addressesState)
	if err != nil {
		log.Error("error in chainSimulatorWrapper.FundWallets", "error", err)
		return
	}

	_, status, err := instance.clientWrapper.PostHTTP(ctx, setMultipleEndpoint, buff)
	if err != nil || status != http.StatusOK {
		log.Error("error in chainSimulatorWrapper.FundWallets - PostHTTP", "error", err, "status", status)
		return
	}
}

// GetESDTBalance returns the balance of the esdt token for the provided address
func (instance *chainSimulatorWrapper) GetESDTBalance(ctx context.Context, address sdkCore.AddressHandler, token string) (string, error) {
	tokenData, err := instance.proxyInstance.GetESDTTokenData(ctx, address, token, apiCore.AccountQueryOptions{
		OnFinalBlock: true,
	})
	if err != nil {
		return "", err
	}

	return tokenData.Balance, nil
}

func (instance *chainSimulatorWrapper) getNonce(ctx context.Context, bech32Address string) (uint64, error) {
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

func (instance *chainSimulatorWrapper) signAndSend(ctx context.Context, senderSK []byte, ftx *transaction.FrontendTransaction) (string, error) {
	sig, err := computeTransactionSignature(senderSK, ftx)
	if err != nil {
		return "", err
	}
	ftx.Signature = hex.EncodeToString(sig)

	hash, err := instance.proxyInstance.SendTransaction(ctx, ftx)
	if err != nil {
		return "", err
	}

	instance.GenerateBlocks(ctx, 1)

	return hash, nil
}

func computeTransactionSignature(senderSk []byte, tx *transaction.FrontendTransaction) ([]byte, error) {
	signer := &singlesig.Ed25519Signer{}
	keyGenerator := signing.NewKeyGenerator(ed25519.NewEd25519())

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
