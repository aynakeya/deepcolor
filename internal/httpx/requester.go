package httpx

import (
	"net/http"
	"time"
)

type Config struct {
	BaseURL string
	Header  map[string]string
	Cookie  map[string]string
	Timeout time.Duration
}

func NewConfig() *Config {
	return &Config{
		BaseURL: "",
		Header:  make(map[string]string),
		Cookie:  make(map[string]string),
		Timeout: 10 * time.Second,
	}
}

type IRequester interface {
	Config() *Config
	Do(req *Request) (*Response, error)
}

type BaseRequester interface {
	Config() *Config
	Do(req *Request) (*Response, error)
}

func NewRequester(base BaseRequester) IRequester {
	return &requester{base: base}
}

type requester struct {
	base BaseRequester
}

func (r *requester) Config() *Config {
	return r.base.Config()
}

func (r *requester) Do(req *Request) (*Response, error) {
	return r.base.Do(req)
}

// Helper methods remain available on requester wrapper for convenience.
func (r *requester) Get(uri string, headers map[string]string) (*Response, error) {
	u, err := BuildURLE(r.Config().BaseURL, uri)
	if err != nil {
		return nil, err
	}
	return r.base.Do(&Request{
		Method: http.MethodGet,
		Url:    u,
		Header: headers,
	})
}

func (r *requester) Post(uri string, headers map[string]string, body any) (*Response, error) {
	u, err := BuildURLE(r.Config().BaseURL, uri)
	if err != nil {
		return nil, err
	}
	data, err := EncodeBodyData(body)
	if err != nil {
		return nil, err
	}
	return r.base.Do(&Request{
		Method: http.MethodPost,
		Url:    u,
		Header: headers,
		Data:   data,
	})
}

func (r *requester) GetQuery(uri string, query map[string]string, headers map[string]string) (*Response, error) {
	u, err := BuildURLE(r.Config().BaseURL, uri)
	if err != nil {
		return nil, err
	}
	paramVals := u.Query()
	for key, value := range query {
		paramVals.Set(key, value)
	}
	u.RawQuery = paramVals.Encode()
	return r.base.Do(&Request{
		Method: http.MethodGet,
		Url:    u,
		Header: headers,
	})
}
