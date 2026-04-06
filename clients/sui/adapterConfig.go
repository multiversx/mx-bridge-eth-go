package sui

import "github.com/block-vision/sui-go-sdk/models"

// ParsedAdapterObject is a pre-parsed shared object reference for an adapter PTB call.
type ParsedAdapterObject struct {
	objectIdBytes        models.SuiAddressBytes
	initialSharedVersion uint64
	mutable              bool
}

// NewParsedAdapterObject creates a ParsedAdapterObject.
func NewParsedAdapterObject(objectIdBytes models.SuiAddressBytes, initialSharedVersion uint64, mutable bool) ParsedAdapterObject {
	return ParsedAdapterObject{
		objectIdBytes:        objectIdBytes,
		initialSharedVersion: initialSharedVersion,
		mutable:              mutable,
	}
}

// ParsedAdapterConfig holds the routing info for a mint-burn adapter token type.
// When a batch contains the associated coin type, the relayer calls
// <package>::<adapterModule>::execute_transfer instead of bridge::execute_transfer,
// appending adapterObjects after the is_batch_complete argument.
type ParsedAdapterConfig struct {
	adapterModule  string
	adapterObjects []ParsedAdapterObject
}

// NewParsedAdapterConfig creates a ParsedAdapterConfig.
func NewParsedAdapterConfig(adapterModule string, adapterObjects []ParsedAdapterObject) ParsedAdapterConfig {
	return ParsedAdapterConfig{
		adapterModule:  adapterModule,
		adapterObjects: adapterObjects,
	}
}
