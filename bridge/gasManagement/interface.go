package gasManagement

import "net/http"

// HTTPClient is the interface we expect to call in order to do the HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
