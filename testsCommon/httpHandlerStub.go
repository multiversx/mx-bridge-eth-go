package testsCommon

import "net/http"

// HTTPHandlerStub -
type HTTPHandlerStub struct {
	ServeHTTPCalled func(writer http.ResponseWriter, request *http.Request)
}

// ServeHTTP -
func (stub *HTTPHandlerStub) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if stub.ServeHTTPCalled != nil {
		stub.ServeHTTPCalled(writer, request)
	}
}
