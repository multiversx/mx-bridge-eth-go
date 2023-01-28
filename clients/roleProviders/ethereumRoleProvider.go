package roleproviders

import (
	"context"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const ethSignatureSize = 64

// ArgsEthereumRoleProvider is the argument for the ethereum role provider constructor
type ArgsEthereumRoleProvider struct {
	EthereumChainInteractor EthereumChainInteractor
	Log                     logger.Logger
}

type ethereumRoleProvider struct {
	ethereumChainInteractor EthereumChainInteractor
	log                     logger.Logger
	whitelistedAddresses    map[common.Address]struct{}
	mut                     sync.RWMutex
}

// NewEthereumRoleProvider creates a new ethereum role provider instance able to fetch the
// whitelisted addresses and able to check ethereum signatures
func NewEthereumRoleProvider(args ArgsEthereumRoleProvider) (*ethereumRoleProvider, error) {
	err := checkEthereumRoleProviderSpecificArgs(args)
	if err != nil {
		return nil, err
	}

	erp := &ethereumRoleProvider{
		whitelistedAddresses:    make(map[common.Address]struct{}),
		ethereumChainInteractor: args.EthereumChainInteractor,
		log:                     args.Log,
	}

	return erp, nil
}

func checkEthereumRoleProviderSpecificArgs(args ArgsEthereumRoleProvider) error {
	if check.IfNil(args.EthereumChainInteractor) {
		return ErrNilEthereumChainInteractor
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}

	return nil
}

// Execute will fetch the available whitelisted addresses and store them in the inner map
func (erp *ethereumRoleProvider) Execute(ctx context.Context) error {
	addresses, err := erp.ethereumChainInteractor.GetRelayers(ctx)
	if err != nil {
		return err
	}

	erp.processResults(addresses)

	return nil
}

func (erp *ethereumRoleProvider) processResults(results []common.Address) {
	currentList := make([]string, 0, len(results))

	erp.mut.Lock()
	erp.whitelistedAddresses = make(map[common.Address]struct{})

	for _, addr := range results {
		erp.whitelistedAddresses[addr] = struct{}{}
		currentList = append(currentList, addr.String())
	}
	erp.mut.Unlock()

	erp.log.Debug("fetched Ethereum whitelisted addresses:\n" + strings.Join(currentList, "\n"))
}

// VerifyEthSignature will verify the provided signature against the message hash. It will also checks if the
// resulting public key is whitelisted or not
func (erp *ethereumRoleProvider) VerifyEthSignature(signature []byte, messageHash []byte) error {
	pkBytes, err := crypto.Ecrecover(messageHash, signature)
	if err != nil {
		return err
	}

	pk, err := crypto.UnmarshalPubkey(pkBytes)
	if err != nil {
		return err
	}

	address := crypto.PubkeyToAddress(*pk)
	if !erp.isWhitelisted(address) {
		return ErrAddressIsNotWhitelisted
	}

	if len(signature) > ethSignatureSize {
		// signatures might contain the recovery byte
		signature = signature[:ethSignatureSize]
	}

	sigOk := crypto.VerifySignature(pkBytes, messageHash, signature)
	if !sigOk {
		return ErrInvalidSignature
	}

	return nil
}

func (erp *ethereumRoleProvider) isWhitelisted(address common.Address) bool {
	erp.mut.RLock()
	defer erp.mut.RUnlock()

	_, exists := erp.whitelistedAddresses[address]

	return exists
}

// IsInterfaceNil returns true if there is no value under the interface
func (erp *ethereumRoleProvider) IsInterfaceNil() bool {
	return erp == nil
}
