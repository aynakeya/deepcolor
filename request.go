package deepcolor

import (
	"context"
	"time"
)

type Request struct {
	Method  string
	URL     string
	Query   map[string]any
	Header  map[string]string
	Body    any
	Timeout time.Duration
	Context context.Context
}

func NewRequest(method string, rawURL string) *Request {
	return &Request{
		Method: method,
		URL:    rawURL,
		Query:  make(map[string]any),
		Header: make(map[string]string),
	}
}

func (r *Request) Clone() *Request {
	if r == nil {
		return nil
	}
	cp := &Request{
		Method:  r.Method,
		URL:     r.URL,
		Body:    r.Body,
		Timeout: r.Timeout,
		Context: r.Context,
		Query:   make(map[string]any, len(r.Query)),
		Header:  make(map[string]string, len(r.Header)),
	}
	for k, v := range r.Query {
		cp.Query[k] = v
	}
	for k, v := range r.Header {
		cp.Header[k] = v
	}
	return cp
}

func (r *Request) SetQuery(query map[string]any) *Request {
	if r.Query == nil {
		r.Query = make(map[string]any)
	}
	for k, v := range query {
		r.Query[k] = v
	}
	return r
}

func (r *Request) SetHeader(header map[string]string) *Request {
	if r.Header == nil {
		r.Header = make(map[string]string)
	}
	for k, v := range header {
		r.Header[k] = v
	}
	return r
}

func (r *Request) SetBody(body any) *Request {
	r.Body = body
	return r
}

func (r *Request) SetTimeout(seconds int) *Request {
	r.Timeout = time.Duration(seconds) * time.Second
	return r
}

func (r *Request) SetTimeoutDuration(timeout time.Duration) *Request {
	r.Timeout = timeout
	return r
}

func (r *Request) SetContext(ctx context.Context) *Request {
	r.Context = ctx
	return r
}
