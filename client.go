package deepcolor

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aynakeya/deepcolor/internal/httpx"
	"github.com/aynakeya/deepcolor/internal/httpx/requesters"
)

type Option func(*Client)

type Client struct {
	requester      httpx.IRequester
	baseURL        string
	defaultHeader  map[string]string
	defaultTimeout time.Duration
}

func New(opts ...Option) *Client {
	c := &Client{
		requester:      requesters.NewRestyRequester(),
		defaultHeader:  map[string]string{},
		defaultTimeout: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithRequester(requester httpx.IRequester) Option {
	return func(c *Client) {
		c.requester = requester
	}
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithHeader(header map[string]string) Option {
	return func(c *Client) {
		for k, v := range header {
			c.defaultHeader[k] = v
		}
	}
}

func WithTimeout(seconds int) Option {
	return func(c *Client) {
		c.defaultTimeout = time.Duration(seconds) * time.Second
	}
}

func WithTimeoutDuration(timeout time.Duration) Option {
	return func(c *Client) {
		c.defaultTimeout = timeout
	}
}

func (c *Client) Do(req *Request) (*Response, error) {
	if c == nil || c.requester == nil {
		return nil, &Error{Kind: ErrKindTransport, Op: "do", Message: "nil client/requester"}
	}
	if req == nil {
		return nil, &Error{Kind: ErrKindValidate, Op: "do", Message: "nil request"}
	}
	httpReq, err := c.buildRequest(req)
	if err != nil {
		return nil, err
	}
	rawResp, err := c.requester.Do(httpReq)
	if err != nil {
		var httpErr *httpx.Error
		if errors.As(err, &httpErr) {
			kind := ErrKindTransport
			switch httpErr.Kind {
			case httpx.ErrKindHTTPStatus:
				kind = ErrKindHTTP
			case httpx.ErrKindEncodeBody, httpx.ErrKindInvalidURL, httpx.ErrKindMethod:
				kind = ErrKindValidate
			}
			return newResponse(rawResp), &Error{
				Kind:    kind,
				Op:      httpErr.Op,
				URL:     httpErr.URL,
				Status:  httpErr.Status,
				Message: httpErr.Message,
				Cause:   err,
			}
		}
		return newResponse(rawResp), &Error{
			Kind:  ErrKindTransport,
			Op:    "http",
			URL:   httpReq.Url.String(),
			Cause: err,
		}
	}
	if rawResp != nil && rawResp.StatusCode() >= http.StatusBadRequest {
		return newResponse(rawResp), &Error{
			Kind:    ErrKindHTTP,
			Op:      "http_status",
			URL:     httpReq.Url.String(),
			Status:  rawResp.StatusCode(),
			Message: "unexpected http status",
		}
	}
	return newResponse(rawResp), nil
}

func (c *Client) GET(url string) *Flow[any] {
	return newFlowAny(NewRequest(http.MethodGet, url))
}

func (c *Client) POST(url string) *Flow[any] {
	return newFlowAny(NewRequest(http.MethodPost, url))
}

func (c *Client) PUT(url string) *Flow[any] {
	return newFlowAny(NewRequest(http.MethodPut, url))
}

func (c *Client) DELETE(url string) *Flow[any] {
	return newFlowAny(NewRequest(http.MethodDelete, url))
}

func (c *Client) buildRequest(req *Request) (*httpx.Request, error) {
	targetURL, err := httpx.BuildURLE(c.baseURL, req.URL)
	if err != nil {
		return nil, &Error{Kind: ErrKindValidate, Op: "build_request", Message: err.Error(), Cause: err}
	}
	query := targetURL.Query()
	for k, v := range req.Query {
		query.Set(k, fmt.Sprint(v))
	}
	targetURL.RawQuery = query.Encode()

	header := map[string]string{}
	for k, v := range c.defaultHeader {
		header[k] = v
	}
	for k, v := range req.Header {
		header[k] = v
	}

	timeout := req.Timeout
	if timeout == 0 {
		timeout = c.defaultTimeout
	}

	body, err := httpx.EncodeBodyData(req.Body)
	if err != nil {
		return nil, &Error{Kind: ErrKindValidate, Op: "encode_body", Message: err.Error(), Cause: err}
	}

	return &httpx.Request{
		Method:  req.Method,
		Url:     targetURL,
		Header:  header,
		Data:    body,
		Timeout: timeout,
		Context: req.Context,
	}, nil
}
