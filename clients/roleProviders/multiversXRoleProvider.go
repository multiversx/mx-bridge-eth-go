package roleproviders

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// ArgsMultiversXRoleProvider is the argument for the MultiversX role provider constructor
type ArgsMultiversXRoleProvider struct {
	DataGetter DataGetter
	Log        logger.Logger
}

type multiversXRoleProvider struct {
	dataGetter           DataGetter
	log                  logger.Logger
	whitelistedAddresses map[string]struct{}
	mut                  sync.RWMutex
}

// NewMultiversXRoleProvider creates a new multiversXRoleProvider instance able to fetch the whitelisted addresses
func NewMultiversXRoleProvider(args ArgsMultiversXRoleProvider) (*multiversXRoleProvider, error) {
	err := checkMultiversXRoleProviderSpecificArgs(args)
	if err != nil {
		return nil, err
	}

	erp := &multiversXRoleProvider{
		dataGetter:           args.DataGetter,
		log:                  args.Log,
		whitelistedAddresses: make(map[string]struct{}),
	}

	return erp, nil
}

func checkMultiversXRoleProviderSpecificArgs(args ArgsMultiversXRoleProvider) error {
	if check.IfNil(args.DataGetter) {
		return clients.ErrNilDataGetter
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}

	return nil
}

// Execute will fetch the available relayers and store them in the inner map
func (erp *multiversXRoleProvider) Execute(ctx context.Context) error {
	results, err := erp.dataGetter.GetAllStakedRelayers(ctx)
	if err != nil {
		return err
	}

	return erp.processResults(results)
}

func (erp *multiversXRoleProvider) processResults(results [][]byte) error {
	currentList := make([]string, 0, len(results))
	temporaryMap := make(map[string]struct{})

	for i, result := range results {
		address := data.NewAddressFromBytes(result)
		isValid := address.IsValid()
		if !isValid {
			return fmt.Errorf("%w for index %d, malformed address: %s", ErrInvalidAddressBytes, i, hex.EncodeToString(result))
		}

		bech32Address, err := address.AddressAsBech32String()
		if err != nil {
			return fmt.Errorf("%w for index %d, malformed address: %s", err, i, hex.EncodeToString(result))
		}

		currentList = append(currentList, bech32Address)
		temporaryMap[string(address.AddressBytes())] = struct{}{}
	}

	erp.mut.Lock()
	erp.whitelistedAddresses = temporaryMap
	erp.mut.Unlock()

	erp.log.Debug("fetched whitelisted addresses:\n" + strings.Join(currentList, "\n"))

	return nil
}

// IsWhitelisted returns true if the non-nil address provided is whitelisted or not
func (erp *multiversXRoleProvider) IsWhitelisted(address core.AddressHandler) bool {
	if check.IfNil(address) {
		return false
	}

	erp.mut.RLock()
	defer erp.mut.RUnlock()

	_, exists := erp.whitelistedAddresses[string(address.AddressBytes())]

	return exists
}

// SortedPublicKeys will return all the sorted public keys
func (erp *multiversXRoleProvider) SortedPublicKeys() [][]byte {
	erp.mut.RLock()
	defer erp.mut.RUnlock()

	sortedPublicKeys := make([][]byte, 0, len(erp.whitelistedAddresses))
	for addr := range erp.whitelistedAddresses {
		sortedPublicKeys = append(sortedPublicKeys, []byte(addr))
	}

	sort.Slice(sortedPublicKeys, func(i, j int) bool {
		return bytes.Compare(sortedPublicKeys[i], sortedPublicKeys[j]) < 0
	})
	return sortedPublicKeys
}

// IsInterfaceNil returns true if there is no value under the interface
func (erp *multiversXRoleProvider) IsInterfaceNil() bool {
	return erp == nil
}
