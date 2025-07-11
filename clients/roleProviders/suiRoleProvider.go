package roleproviders

import (
	"context"
	"strings"
	"sync"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ArgsSuiRoleProvider is the argument for the Sui role provider constructor
type ArgsSuiRoleProvider struct {
	DataGetter SuiDataGetter
	Log        logger.Logger
}

type suiRoleProvider struct {
	dataGetter           SuiDataGetter
	log                  logger.Logger
	whitelistedAddresses map[string]struct{}
	mut                  sync.RWMutex
}

func NewSuiRoleProvider(args ArgsSuiRoleProvider) (*suiRoleProvider, error) {
	err := checkSuiRoleProviderSpecificArgs(args)
	if err != nil {
		return nil, err
	}

	srp := &suiRoleProvider{
		dataGetter:           args.DataGetter,
		log:                  args.Log,
		whitelistedAddresses: make(map[string]struct{}),
	}

	return srp, nil
}

func checkSuiRoleProviderSpecificArgs(args ArgsSuiRoleProvider) error {
	if check.IfNil(args.DataGetter) {
		return clients.ErrNilDataGetter
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}

	return nil
}

// Execute will fetch the available relayers and store them in the inner map
func (srp *suiRoleProvider) Execute(ctx context.Context) error {
	addresses, err := srp.dataGetter.GetRelayers(ctx)
	if err != nil {
		return err
	}

	return srp.processResults(addresses)
}

func (srp *suiRoleProvider) processResults(results []models.SuiAddress) error {
	currentList := make([]string, 0, len(results))
	temporaryMap := make(map[string]struct{})

	for _, result := range results {
		currentList = append(currentList, string(result))
		temporaryMap[string(result)] = struct{}{}
	}

	srp.mut.Lock()
	srp.whitelistedAddresses = temporaryMap
	srp.mut.Unlock()

	srp.log.Debug("fetched whitelisted addresses:\n" + strings.Join(currentList, "\n"))

	return nil
}

// VerifySignature will verify the provided signature against the message hash. It will also checks if the
// public key is whitelisted or not
func (srp *suiRoleProvider) VerifySignature(signature []byte, messageHash []byte) error {
	panic("not implemented yet") // TODO: Implement this method
}

func (srp *suiRoleProvider) isWhitelisted(address string) bool {
	srp.mut.RLock()
	defer srp.mut.RUnlock()

	_, exists := srp.whitelistedAddresses[address]

	return exists
}

// IsInterfaceNil returns true if there is no value under the interface
func (srp *suiRoleProvider) IsInterfaceNil() bool {
	return srp == nil
}
