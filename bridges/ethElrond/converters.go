package ethElrond

import (
	"encoding/hex"
	"fmt"
)

func convertObjectToString(obj interface{}) string {
	switch objType := obj.(type) {
	case []byte:
		return hex.EncodeToString(objType)
	default:
		return fmt.Sprintf("%v", obj)
	}
}
