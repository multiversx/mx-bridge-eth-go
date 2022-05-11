package ethElrond

import "time"

// InvalidActionID represents an invalid id for an action on Elrond
const InvalidActionID = uint64(0)

const durationLimit = time.Duration(time.Second)

// ClientStatus represents the possible statuses of a client
type ClientStatus int

const (
	Available   ClientStatus = 0
	Unavailable ClientStatus = 1
)

//String will return status as string based on the int value
func (cs ClientStatus) String() string {
	switch cs {
	case Available:
		return "Available"
	case Unavailable:
		return "Unavailable"
	default:
		return "Invalid status"
	}
}
