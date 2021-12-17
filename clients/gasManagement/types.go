package gasManagement

import (
	"encoding/json"
	"fmt"
)

type gasStationResponse struct {
	Fast        int     `json:"fast"`
	Fastest     int     `json:"fastest"`
	SafeLow     int     `json:"safeLow"`
	Average     int     `json:"average"`
	BlockTime   float64 `json:"block_time"`
	BlockNum    int     `json:"blockNum"`
	Speed       float64 `json:"speed"`
	SafeLowWait float64 `json:"safeLowWait"`
	AvgWait     float64 `json:"avgWait"`
	FastWait    float64 `json:"fastWait"`
	FastestWait float64 `json:"fastestWait"`
}

func (gsr *gasStationResponse) String() string {
	data, err := json.Marshal(gsr)
	if err != nil {
		return fmt.Sprintf("gasStationResponse errored with %s", err.Error())
	}

	return string(data)
}
