package converters

import (
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
)

var log = logger.GetOrCreate("core")

const hexPrefix = "0x"

type addressConverter struct {
	converter core.PubkeyConverter
}

// NewAddressConverter will create an address converter instance
func NewAddressConverter() (*addressConverter, error) {
	var err error
	ac := &addressConverter{}
	ac.converter, err = pubkeyConverter.NewBech32PubkeyConverter(erdgoCore.AddressBytesLen, log)
	if err != nil {
		log.Error("error while creating and addressConverter", "error", err)
		return nil, err
	}

	return ac, nil
}

// ToHexString will convert the addressBytes to the hex representation
func (ac *addressConverter) ToHexString(addressBytes []byte) string {
	return hex.EncodeToString(addressBytes)
}

// ToHexStringWithPrefix will convert the addressBytes to the hex representation adding the hex prefix
func (ac *addressConverter) ToHexStringWithPrefix(addressBytes []byte) string {
	return hexPrefix + hex.EncodeToString(addressBytes)
}

// ToBech32String will convert the addressBytes to the bech32 representation
func (ac *addressConverter) ToBech32String(addressBytes []byte) string {
	return ac.converter.Encode(addressBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *addressConverter) IsInterfaceNil() bool {
	return ac == nil
}

// TrimWhiteSpaceCharacters will remove the white spaces from the input string
func TrimWhiteSpaceCharacters(input string) string {
	cutset := "\n\t "

	return strings.Trim(input, cutset)
}
