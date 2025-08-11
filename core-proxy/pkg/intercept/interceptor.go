package intercept

import "net/http"

// Interceptor is the interface for an HTTP/HTTPS traffic interceptor.
// It allows for modifying requests and responses.
type Interceptor interface {
	// InterceptRequest is called before a request is sent to the target server.
	// It can be used to view or modify the request.
	InterceptRequest(*http.Request) error

	// InterceptResponse is called after a response is received from the target server.
	// It can be used to view or modify the response.
	InterceptResponse(*http.Response) error
}
