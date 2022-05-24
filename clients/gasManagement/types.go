package gasManagement

import (
	"encoding/json"
	"fmt"
)

type gasStationResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  struct {
		LastBlock       string `json:"LastBlock"`
		SafeGasPrice    string `json:"SafeGasPrice"`
		ProposeGasPrice string `json:"ProposeGasPrice"`
		FastGasPrice    string `json:"FastGasPrice"`
		SuggestBaseFee  string `json:"suggestBaseFee"`
		GasUsedRatio    string `json:"gasUsedRatio"`
	} `json:"result"`
}

func (gsr *gasStationResponse) String() string {
	data, err := json.Marshal(gsr)
	if err != nil {
		return fmt.Sprintf("gasStationResponse errored with %s", err.Error())
	}

	return string(data)
}
