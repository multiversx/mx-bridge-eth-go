package gasManagement

import "errors"

// ErrLatestGasPricesWereNotFetched signals that the latest gas price values couldn't have been fetched
var ErrLatestGasPricesWereNotFetched = errors.New("latest gas price values couldn't have been fetched")

// ErrInvalidGasPriceSelector signals that an invalid gas price selector has been provided
var ErrInvalidGasPriceSelector = errors.New("invalid gas price selector")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

// ErrGasPriceIsHigherThanTheMaximumSet signals that the fetched gas price is higher than the maximum set
var ErrGasPriceIsHigherThanTheMaximumSet = errors.New("fetched gas price is higher than the maximum set")
