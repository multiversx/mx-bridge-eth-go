package roleProvider

import (
	"context"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

const getAllStakedRelayersFunctionName = "getAllStakedRelayers"

// ArgsElrondRoleProvider is the argument for the elrond role provider constructor
type ArgsElrondRoleProvider struct {
	ElrondChainInteractor ElrondChainInteractor
	Log                   logger.Logger
}

type elrondRoleProvider struct {
	elrondChainInteractor ElrondChainInteractor
	log                   logger.Logger
	whitelistedAddresses  map[string]struct{}
	mut                   sync.RWMutex
}

// NewElrondRoleProvider creates a new elrond role provider instance able to fetch the whitelisted addresses
func NewElrondRoleProvider(args ArgsElrondRoleProvider) (*elrondRoleProvider, error) {
	err := checkElrondRoleProviderSpecificArgs(args)
	if err != nil {
		return nil, err
	}

	erp := &elrondRoleProvider{
		elrondChainInteractor: args.ElrondChainInteractor,
		log:                   args.Log,
		whitelistedAddresses:  make(map[string]struct{}),
	}

	return erp, nil
}

func checkElrondRoleProviderSpecificArgs(args ArgsElrondRoleProvider) error {
	if check.IfNil(args.ElrondChainInteractor) {
		return ErrNilElrondChainInteractor
	}
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}

	return nil
}

// Execute will fetch the available relayers and store them in the inner map
func (erp *elrondRoleProvider) Execute(_ context.Context) error {
	results, err := erp.elrondChainInteractor.ExecuteVmQueryOnBridgeContract(getAllStakedRelayersFunctionName)
	if err != nil {
		return err
	}

	erp.processResults(results)

	return nil
}

func (erp *elrondRoleProvider) processResults(results [][]byte) {
	currentList := make([]string, 0, len(results))

	erp.mut.Lock()
	erp.whitelistedAddresses = make(map[string]struct{})

	for _, result := range results {
		address := data.NewAddressFromBytes(result)
		currentList = append(currentList, address.AddressAsBech32String())

		erp.whitelistedAddresses[string(address.AddressBytes())] = struct{}{}
	}
	erp.mut.Unlock()

	erp.log.Debug("fetched whitelisted addresses:\n" + strings.Join(currentList, "\n"))
}

// IsWhitelisted returns true if the non-nil address provided is whitelisted or not
func (erp *elrondRoleProvider) IsWhitelisted(address core.AddressHandler) bool {
	if check.IfNil(address) {
		return false
	}

	erp.mut.RLock()
	defer erp.mut.RUnlock()

	_, exists := erp.whitelistedAddresses[string(address.AddressBytes())]

	return exists
}

// IsInterfaceNil returns true if there is no value under the interface
func (erp *elrondRoleProvider) IsInterfaceNil() bool {
	return erp == nil
}
