package roleProviders

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// ArgsElrondRoleProvider is the argument for the elrond role provider constructor
type ArgsElrondRoleProvider struct {
	DataGetter DataGetter
	Log        logger.Logger
}

type elrondRoleProvider struct {
	dataGetter           DataGetter
	log                  logger.Logger
	whitelistedAddresses map[string]struct{}
	mut                  sync.RWMutex
}

// NewElrondRoleProvider creates a new elrond role provider instance able to fetch the whitelisted addresses
func NewElrondRoleProvider(args ArgsElrondRoleProvider) (*elrondRoleProvider, error) {
	err := checkElrondRoleProviderSpecificArgs(args)
	if err != nil {
		return nil, err
	}

	erp := &elrondRoleProvider{
		dataGetter:           args.DataGetter,
		log:                  args.Log,
		whitelistedAddresses: make(map[string]struct{}),
	}

	return erp, nil
}

func checkElrondRoleProviderSpecificArgs(args ArgsElrondRoleProvider) error {
	if check.IfNil(args.DataGetter) {
		return clients.ErrNilDataGetter
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}

	return nil
}

// Execute will fetch the available relayers and store them in the inner map
func (erp *elrondRoleProvider) Execute(ctx context.Context) error {
	results, err := erp.dataGetter.GetAllStakedRelayers(ctx)
	if err != nil {
		return err
	}

	return erp.processResults(results)
}

func (erp *elrondRoleProvider) processResults(results [][]byte) error {
	currentList := make([]string, 0, len(results))
	temporaryMap := make(map[string]struct{})

	for i, result := range results {
		address := data.NewAddressFromBytes(result)
		isValid := address.IsValid()
		if !isValid {
			return fmt.Errorf("%w for index %d, malformed address: %s", ErrInvalidAddressBytes, i, hex.EncodeToString(result))
		}

		currentList = append(currentList, address.AddressAsBech32String())
		temporaryMap[string(address.AddressBytes())] = struct{}{}
	}

	erp.mut.Lock()
	erp.whitelistedAddresses = temporaryMap
	erp.mut.Unlock()

	erp.log.Debug("fetched whitelisted addresses:\n" + strings.Join(currentList, "\n"))

	return nil
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
