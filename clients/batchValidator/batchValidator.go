package batchValidatorManagement

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/clients/chain"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const minRequestTime = time.Millisecond
const logPath = "BatchValidator"

// ArgsBatchValidator is the DTO used for the creating a new batch validator instance
type ArgsBatchValidator struct {
	SourceChain      chain.Chain
	DestinationChain chain.Chain
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
		requestURL:  fmt.Sprintf("%s/%s/%s", args.RequestURL, args.SourceChain.ToLower(), args.DestinationChain.ToLower()),
		requestTime: args.RequestTime,
		httpClient:  http.DefaultClient,
	}
	bv.log = logger.GetOrCreate(logPath)
	return bv, nil
}

func checkArgs(args ArgsBatchValidator) error {
	switch args.SourceChain {
	case chain.Ethereum, chain.Bsc, chain.MultiversX:
	default:
		return fmt.Errorf("%w: %q", clients.ErrInvalidValue, args.SourceChain)
	}
	switch args.DestinationChain {
	case chain.Ethereum, chain.Bsc, chain.MultiversX:
	default:
		return fmt.Errorf("%w: %q", clients.ErrInvalidValue, args.DestinationChain)
	}
	if args.RequestTime < minRequestTime {
		return fmt.Errorf("%w in checkArgs for value RequestTime", clients.ErrInvalidValue)
	}

	return nil
}

// ValidateBatch checks whether the given batch is the same also on miscroservice side
func (bv *batchValidator) ValidateBatch(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
	body, err := json.Marshal(batch)
	if err != nil {
		return false, fmt.Errorf("%w during request marshal", err)
	}

	responseAsBytes, err := bv.doRequest(ctx, body)
	if err != nil {
		return false, fmt.Errorf("%w while executing request", err)
	}
	if len(responseAsBytes) == 0 {
		return false, errors.New("empty response")
	}

	response := &microserviceResponse{}
	err = json.Unmarshal(responseAsBytes, response)
	if err != nil {
		return false, fmt.Errorf("%w during response unmarshal", err)
	}

	return response.Valid, nil
}

func (bv *batchValidator) doRequest(ctx context.Context, batch []byte) ([]byte, error) {
	requestContext, cancel := context.WithTimeout(ctx, bv.requestTime)
	defer cancel()

	responseAsBytes, err := bv.doRequestReturningBytes(batch, requestContext)
	if err != nil {
		return nil, err
	}

	return responseAsBytes, nil
}

func (bv *batchValidator) doRequestReturningBytes(batch []byte, ctx context.Context) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, bv.requestURL, bytes.NewBuffer(batch))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}

	response, err := bv.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusBadRequest && response.Body != http.NoBody {
		data, _ := ioutil.ReadAll(response.Body)
		badResponse := &microserviceBadRequestBody{}
		err = json.Unmarshal(data, badResponse)
		if err != nil {
			return nil, fmt.Errorf("%w during bad response unmarshal", err)
		}
		return nil, fmt.Errorf("got status %s: %s", response.Status, badResponse.Message)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status %s", response.Status)
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
