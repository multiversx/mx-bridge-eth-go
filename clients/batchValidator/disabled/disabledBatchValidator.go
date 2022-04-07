package disabled

type DisabledBatchValidator struct{}

// ValidateBatch returns true,nil and will result in skipping batch validation
func (dbv *DisabledBatchValidator) ValidateBatch(_ string) (bool, error) {
	return true, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dbv *DisabledBatchValidator) IsInterfaceNil() bool {
	return dbv == nil
}
