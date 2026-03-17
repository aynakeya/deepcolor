package requesters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aynakeya/deepcolor/internal/httpx"
	"github.com/go-resty/resty/v2"
)

type restyRequester struct {
	config *httpx.Config
	client *resty.Client
}

func NewRestyRequester() httpx.IRequester {
	return httpx.NewRequester(&restyRequester{
		client: resty.New(),
		config: httpx.NewConfig(),
	})
}

func (r *restyRequester) Config() *httpx.Config {
	return r.config
}

func (r *restyRequester) Do(req *httpx.Request) (*httpx.Response, error) {
	var resp *resty.Response
	var err error

	if req == nil {
		return nil, &httpx.Error{
			Kind:    httpx.ErrKindTransport,
			Op:      "request",
			Message: "nil request",
		}
	}
	if req.Url == nil {
		return &httpx.Response{Request: req}, &httpx.Error{
			Kind:    httpx.ErrKindInvalidURL,
			Op:      "request",
			Method:  req.Method,
			Message: "nil request url",
		}
	}

	req.Header = mergeHeader(req.Header, r.config.Header)
	if req.Timeout <= 0 {
		req.Timeout = r.config.Timeout
	}
	if req.Context == nil {
		req.Context = context.Background()
	}
	ctx, cancel := context.WithTimeout(req.Context, req.Timeout)
	defer cancel()

	switch req.Method {
	case http.MethodGet:
		resp, err = r.client.R().
			SetContext(ctx).
			SetHeaders(req.Header).
			Get(req.Url.String())
	case http.MethodPost:
		resp, err = r.client.R().
			SetContext(ctx).
			SetHeaders(req.Header).
			SetBody(req.Data).
			Post(req.Url.String())
	case http.MethodHead:
		resp, err = r.client.R().
			SetContext(ctx).
			SetHeaders(req.Header).
			Head(req.Url.String())
	default:
		return &httpx.Response{Request: req}, &httpx.Error{
			Kind:    httpx.ErrKindMethod,
			Op:      "request",
			Method:  req.Method,
			URL:     req.Url.String(),
			Message: fmt.Sprintf("unsupported method %s", req.Method),
		}
	}
	if err != nil {
		return &httpx.Response{
			Request: req,
		}, &httpx.Error{
			Kind:   httpx.ErrKindTransport,
			Op:     "request",
			Method: req.Method,
			URL:    req.Url.String(),
			Cause:  err,
		}
	}
	return &httpx.Response{
		Request:     req,
		RawResponse: resp.RawResponse,
		RawBody:     resp.Body(),
		Size:        resp.Size(),
	}, nil
}

func mergeHeader(src map[string]string, updated map[string]string) map[string]string {
	header := make(map[string]string)
	for k, v := range src {
		header[k] = v
	}
	for k, v := range updated {
		header[k] = v
	}
	return header
}
