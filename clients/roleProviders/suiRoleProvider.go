package roleproviders

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	suiClient "github.com/multiversx/mx-bridge-eth-go/clients/sui"
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

	for i, result := range results {
		isValid := srp.isValidSuiAddress(result)
		if !isValid {
			return fmt.Errorf("%w for index %d, malformed address: %s", ErrInvalidSuiAddress, i, result)
		}
		currentList = append(currentList, string(result))
		temporaryMap[string(result)] = struct{}{}
	}

	srp.mut.Lock()
	srp.whitelistedAddresses = temporaryMap
	srp.mut.Unlock()

	srp.log.Debug("fetched whitelisted addresses:\n" + strings.Join(currentList, "\n"))

	return nil
}

func (srp *suiRoleProvider) isValidSuiAddress(address models.SuiAddress) bool {
	return len(address) == sui.ValidSuiAddressLength && strings.HasPrefix(string(address), "0x")
}

// VerifySignature will verify the provided signature against the message hash. It will also checks if the
// public key is whitelisted or not
func (srp *suiRoleProvider) VerifySignature(signature []byte, messageHash []byte) error {
	srp.log.Debug("Verifying:", "signature", string(signature), "for message hash:", hex.EncodeToString(messageHash))
	if len(signature)%suiClient.EncodedSignatureLength != 0 {
		return fmt.Errorf("%w: expected a multiple of %d, got %d", ErrInvalidSignaturesArray, suiClient.EncodedSignatureLength, len(signature))
	}
	if len(messageHash)%suiClient.MessageLength != 0 {
		return fmt.Errorf("%w: expected a multiple of %d, got %d", ErrInvalidMessagesArray, suiClient.MessageLength, len(messageHash))
	}
	if len(signature)/suiClient.EncodedSignatureLength != len(messageHash)/suiClient.MessageLength {
		return fmt.Errorf("%w: number of signatures (%d), number of message hashes (%d)", ErrInvalidSignaturesCount, len(signature)/suiClient.EncodedSignatureLength, len(messageHash)/suiClient.MessageLength)
	}

	n := len(signature) / suiClient.EncodedSignatureLength
	var relayer string
	pass := true
	for i := 0; i < n; i++ {
		start := i * suiClient.EncodedSignatureLength
		end := start + suiClient.EncodedSignatureLength
		sig := signature[start:end]

		startHash := i * suiClient.MessageLength
		endHash := startHash + suiClient.MessageLength
		msgHash := messageHash[startHash:endHash]

		signer, ok, err := models.VerifyPersonalMessage(string(msgHash), string(sig))
		if err != nil {
			return err
		}
		relayer = signer

		pass = pass && ok
	}

	if !pass {
		return ErrInvalidSignature
	}
	if !srp.isWhitelisted(relayer) {
		return ErrAddressIsNotWhitelisted
	}

	return nil
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
