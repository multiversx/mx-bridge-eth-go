package ethElrond

import "time"

// InvalidActionID represents an invalid id for an action on Elrond
const InvalidActionID = uint64(0)

const durationLimit = time.Duration(time.Second)

type ClientStatus int

const (
	Available ClientStatus = 0
	Unavailable
)

func (cs ClientStatus) String() string {
	return []string{"Available", "Unavailable"}[cs]
}
