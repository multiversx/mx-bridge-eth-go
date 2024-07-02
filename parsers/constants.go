package parsers

const uint32ArgBytes = 4
const uint64ArgBytes = 8

// MissingDataProtocolMarker defines the marker for missing data (simple transfers)
const MissingDataProtocolMarker byte = 0x00

// DataPresentProtocolMarker defines the marker for existing data (transfers with SC calls)
const DataPresentProtocolMarker byte = 0x01
