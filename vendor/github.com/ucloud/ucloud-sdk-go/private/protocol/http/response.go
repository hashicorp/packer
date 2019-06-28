package http

import (
	"net/http"
)

// HttpResponse is a simple wrapper of "net/http" response
type HttpResponse struct {
	body               []byte
	statusCode         int
	originHttpResponse *http.Response // origin "net/http" response
}

// NewHttpResponse will create a new response of http request
func NewHttpResponse() *HttpResponse {
	return &HttpResponse{}
}

// GetBody will get body from from sdk http request
func (h *HttpResponse) GetBody() []byte {
	return h.body
}

func (h *HttpResponse) GetHeaders() http.Header {
	if h.originHttpResponse == nil {
		return http.Header{}
	}
	return h.originHttpResponse.Header
}

// SetBody will set body into http response
// it usually used for restore the body already read from an stream
// it will also cause extra memory usage
func (h *HttpResponse) SetBody(body []byte) error {
	h.body = body
	return nil
}

// GetStatusCode will return status code of origin http response
func (h *HttpResponse) GetStatusCode() int {
	return h.statusCode
}

// SetStatusCode will return status code of origin http response
func (h *HttpResponse) SetStatusCode(code int) {
	h.statusCode = code
}

func (h *HttpResponse) setHttpResponse(resp *http.Response) {
	h.statusCode = resp.StatusCode
	h.originHttpResponse = resp
}
