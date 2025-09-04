package deepcolor

import (
	"fmt"
	"github.com/aynakeya/deepcolor/dphttp"
	"github.com/spf13/cast"
	"net/http"
)

func NewGetRequestFuncWithSingleQuery(
	uri string,
	query string, headers map[string]string) dphttp.RequestFunc[string] {
	return func(param string) (*dphttp.Request, error) {
		url := dphttp.UrlMustParse(uri)
		paramVals := url.Query()
		paramVals.Set(query, param)
		url.RawQuery = paramVals.Encode()
		return &dphttp.Request{
			Method: http.MethodGet,
			Url:    url,
			Header: headers,
		}, nil
	}
}

func NewGetRequestFuncWithQuery(
	uri string,
	queries []string, headers map[string]string) dphttp.RequestFunc[[]string] {
	return func(params []string) (*dphttp.Request, error) {
		if len(queries) > len(params) {
			return nil, fmt.Errorf("only receive %d parameter, required %d", len(params), len(queries))
		}
		url := dphttp.UrlMustParse(uri)
		paramVals := url.Query()
		for i, _ := range queries {
			paramVals.Set(queries[i], params[i])
		}
		url.RawQuery = paramVals.Encode()
		return &dphttp.Request{
			Method: http.MethodGet,
			Url:    url,
			Header: headers,
		}, nil
	}
}

func NewGetRequestFromUrl(
	uri string,
	headers map[string]string,
	params ...any) *dphttp.Request {
	return &dphttp.Request{
		Method: http.MethodGet,
		Url:    dphttp.UrlMustParse(fmt.Sprintf(uri, params...)),
		Header: headers,
	}
}

func NewGetRequestWithSingleQuery(
	uri string,
	query, value string, headers map[string]string) (*dphttp.Request, error) {
	url := dphttp.UrlMustParse(uri)
	paramVals := url.Query()
	paramVals.Set(query, value)
	url.RawQuery = paramVals.Encode()
	return &dphttp.Request{
		Method: http.MethodGet,
		Url:    url,
		Header: headers,
	}, nil
}

func NewGetRequestWithQuery(
	uri string,
	queries map[string]any, headers map[string]string) (*dphttp.Request, error) {
	url := dphttp.UrlMustParse(uri)
	paramVals := url.Query()
	for key, value := range queries {
		paramVals.Set(key, cast.ToString(value))
	}
	url.RawQuery = paramVals.Encode()
	return &dphttp.Request{
		Method: http.MethodGet,
		Url:    url,
		Header: headers,
	}, nil
}
