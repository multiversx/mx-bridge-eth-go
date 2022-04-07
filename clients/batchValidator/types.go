package batchValidatorManagement

import (
	"encoding/json"
	"fmt"
)

type microserviceResponse struct {
	Valid bool `json:"valid"`
}

func (msr *microserviceResponse) String() string {
	data, err := json.Marshal(msr)
	if err != nil {
		return fmt.Sprintf("microserviceResponse errored with %s", err.Error())
	}

	return string(data)
}
