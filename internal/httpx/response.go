package httpx

import (
	"net/http"
	"strings"
)

// Response struct
// adapt from resty response.go
type Response struct {
	Request     *Request
	RawResponse *http.Response
	RawBody     []byte
	Size        int64
}

func (r *Response) Body() []byte {
	if r.RawResponse == nil {
		return []byte{}
	}
	return r.RawBody
}

func (r *Response) StatusCode() int {
	if r.RawResponse == nil {
		return 0
	}
	return r.RawResponse.StatusCode
}

func (r *Response) Header() http.Header {
	if r.RawResponse == nil {
		return http.Header{}
	}
	return r.RawResponse.Header
}

func (r *Response) String() string {
	if r.RawBody == nil {
		return ""
	}
	return strings.TrimSpace(string(r.RawBody))
}
