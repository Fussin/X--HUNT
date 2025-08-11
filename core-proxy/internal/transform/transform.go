package transform

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sentinelx/core-proxy/pkg/intercept"
)

// Transformer applies a list of interceptors.
type Transformer struct {
	interceptors []intercept.Interceptor
}

// NewTransformer creates a new Transformer.
func NewTransformer(interceptors ...intercept.Interceptor) *Transformer {
	return &Transformer{interceptors: interceptors}
}

// InterceptRequest applies all request interceptors.
func (t *Transformer) InterceptRequest(req *http.Request) error {
	for _, i := range t.interceptors {
		if err := i.InterceptRequest(req); err != nil {
			return err
		}
	}
	return nil
}

// InterceptResponse applies all response interceptors.
func (t *Transformer) InterceptResponse(res *http.Response) error {
	for _, i := range t.interceptors {
		if err := i.InterceptResponse(res); err != nil {
			return err
		}
	}
	return nil
}

// MatchReplaceInterceptor is a simple interceptor that does a string replacement.
type MatchReplaceInterceptor struct {
	Match   []byte
	Replace []byte
}

// InterceptRequest is a no-op for this interceptor.
func (i *MatchReplaceInterceptor) InterceptRequest(req *http.Request) error {
	return nil
}

// InterceptResponse replaces the match string in the response body.
func (i *MatchReplaceInterceptor) InterceptResponse(res *http.Response) error {
	if res.Body == nil {
		return nil
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	newBody := bytes.ReplaceAll(body, i.Match, i.Replace)
	res.Body = io.NopCloser(bytes.NewReader(newBody))
	res.ContentLength = int64(len(newBody))
	return nil
}

// JSONModifyFunc is a function that can modify a parsed JSON body.
type JSONModifyFunc func(data map[string]interface{}) (map[string]interface{}, error)

// JSONInterceptor is an interceptor that can modify JSON response bodies.
type JSONInterceptor struct {
	ModifyFunc JSONModifyFunc
}

// InterceptRequest is a no-op for this interceptor.
func (i *JSONInterceptor) InterceptRequest(req *http.Request) error {
	return nil
}

// InterceptResponse parses and modifies a JSON response body.
func (i *JSONInterceptor) InterceptResponse(res *http.Response) error {
	if res.Body == nil {
		return nil
	}
	// TODO: Check Content-Type header to make sure it's JSON.

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// Not a JSON body, so we don't do anything.
		// We need to put the original body back.
		res.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	}

	newData, err := i.ModifyFunc(data)
	if err != nil {
		return err
	}

	newBody, err := json.Marshal(newData)
	if err != nil {
		return err
	}

	res.Body = io.NopCloser(bytes.NewReader(newBody))
	res.ContentLength = int64(len(newBody))
	return nil
}
