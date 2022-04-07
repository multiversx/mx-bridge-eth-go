package batchValidatorManagement

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const minRequestTime = time.Millisecond
const logPath = "BatchValidator"

// ArgsBatchValidator is the DTO used for the creating a new batch validator instance
type ArgsBatchValidator struct {
	SourceChain      clients.Chain
	DestinationChain clients.Chain
	RequestURL       string
	RequestTime      time.Duration
}

type batchValidator struct {
	requestURL  string
	requestTime time.Duration
	log         logger.Logger
	httpClient  HTTPClient
}

// NewBatchValidator returns a new batch validator instance
func NewBatchValidator(args ArgsBatchValidator) (*batchValidator, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	bv := &batchValidator{
		requestURL:  args.RequestURL + "/" + string(args.SourceChain) + "/" + string(args.DestinationChain),
		requestTime: args.RequestTime,
		httpClient:  http.DefaultClient,
	}
	bv.log = logger.GetOrCreate(logPath)
	return bv, nil
}

func checkArgs(args ArgsBatchValidator) error {
	if args.RequestTime < minRequestTime {
		return fmt.Errorf("%w in checkArgs for value RequestTime", clients.ErrInvalidValue)
	}

	return nil
}

func (bv *batchValidator) ValidateBatch(batch string) (bool, error) {
	responseAsBytes, err := bv.doRequest([]byte(batch))
	response := &microserviceResponse{}
	err = json.Unmarshal(responseAsBytes, response)
	if err != nil {
		return false, err
	}
	return response.Valid, nil
}

func (bv *batchValidator) doRequest(batch []byte) ([]byte, error) {
	requestContext, cancel := context.WithTimeout(context.Background(), bv.requestTime)
	defer cancel()
	responseAsBytes, err := bv.doRequestReturningBytes(batch, requestContext)
	if err != nil {
		return nil, err
	}

	return responseAsBytes, nil
}

func (bv *batchValidator) doRequestReturningBytes(batch []byte, ctx context.Context) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, bv.requestURL, bytes.NewBuffer(batch))
	if err != nil {
		return nil, err
	}

	response, err := bv.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bv *batchValidator) IsInterfaceNil() bool {
	return bv == nil
}
