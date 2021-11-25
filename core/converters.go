package core

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"strings"
)

var log = logger.GetOrCreate("core")

type addressConverter struct {
	converter core.PubkeyConverter
}

// NewAddressConverter will create an address converter instance
func NewAddressConverter() *addressConverter {
	var err error
	ac := &addressConverter{}
	ac.converter, err = pubkeyConverter.NewBech32PubkeyConverter(erdgoCore.AddressLen, log)
	if err != nil {
		log.Error("error while creating and addressConverter", "error", err)
		return nil
	}

	return ac
}

// ToHexString will convert the addressBytes to the hex representation
func (ac *addressConverter) ToHexString(addressBytes []byte) string {
	return hex.EncodeToString(addressBytes)
}

// ToBech32String will convert the addressBytes to the bech32 representation
func (ac *addressConverter) ToBech32String(addressBytes []byte) string {
	return ac.converter.Encode(addressBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *addressConverter) IsInterfaceNil() bool {
	return ac == nil
}

// TODO - move this as a method in AddressHandler

// ConvertFromByteSliceToArray will convert the provided buffer to its [32]byte representation
func ConvertFromByteSliceToArray(buff []byte) [32]byte {
	var result [32]byte
	copy(result[:], buff)

	return result
}

// TrimWhiteSpaceCharacters will remove the white spaces from the input string
func TrimWhiteSpaceCharacters(input string) string {
	cutset := "\n\t "

	return strings.Trim(input, cutset)
}
