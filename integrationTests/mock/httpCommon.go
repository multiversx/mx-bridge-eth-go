package mock

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

func getBodyAsByteSlice(req *http.Request) []byte {
	defer func() {
		if req.Body != nil {
			return
		}

		_ = req.Body.Close()
	}()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("getBodyAsByteSlice", "error", err)
		return nil
	}

	return body
}

func writeResponse(rw http.ResponseWriter, httpCode int, responseName string, data interface{}, err error) {
	var wrappedData interface{}
	code := shared.ReturnCodeInternalError
	if httpCode == http.StatusOK {
		wrappedData = gin.H{responseName: data}
		code = shared.ReturnCodeSuccess
	}

	errString := ""
	if err != nil {
		errString = err.Error()
	}

	response := shared.GenericAPIResponse{
		Data:  wrappedData,
		Error: errString,
		Code:  code,
	}

	buff, _ := json.Marshal(response)
	rw.WriteHeader(httpCode)
	_, _ = rw.Write(buff)
}
