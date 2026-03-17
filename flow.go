package deepcolor

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Ctx struct {
	Request *Request
}

type QueryFunc func(*Ctx) (map[string]any, error)
type HeaderFunc func(*Ctx) (map[string]string, error)
type BodyFunc func(*Ctx) (any, error)
type CheckFunc func(*Response) error
type DecodeFunc[T any] func(*Response, *T) error

type Flow[T any] struct {
	req *Request

	queryFns  []QueryFunc
	headerFns []HeaderFunc
	bodyFns   []BodyFunc
	checks    []CheckFunc
	decoder   DecodeFunc[T]
}

func newFlowAny(req *Request) *Flow[any] {
	return &Flow[any]{
		req: req,
	}
}

// Typed converts a default Flow[any] into Flow[T] without losing chained state.
func Typed[T any](f *Flow[any]) *Flow[T] {
	if f == nil {
		return &Flow[T]{}
	}
	tf := &Flow[T]{
		req:       f.req,
		queryFns:  f.queryFns,
		headerFns: f.headerFns,
		bodyFns:   f.bodyFns,
		checks:    f.checks,
	}
	return tf
}

func (f *Flow[T]) Query(query map[string]any) *Flow[T] {
	f.req.SetQuery(query)
	return f
}

func (f *Flow[T]) QueryFn(fn QueryFunc) *Flow[T] {
	f.queryFns = append(f.queryFns, fn)
	return f
}

func (f *Flow[T]) Header(header map[string]string) *Flow[T] {
	f.req.SetHeader(header)
	return f
}

func (f *Flow[T]) HeaderFn(fn HeaderFunc) *Flow[T] {
	f.headerFns = append(f.headerFns, fn)
	return f
}

func (f *Flow[T]) Body(body any) *Flow[T] {
	f.req.SetBody(body)
	return f
}

func (f *Flow[T]) BodyFn(fn BodyFunc) *Flow[T] {
	f.bodyFns = append(f.bodyFns, fn)
	return f
}

func (f *Flow[T]) Timeout(seconds int) *Flow[T] {
	f.req.SetTimeout(seconds)
	return f
}

func (f *Flow[T]) Check(path string, expect any) *Flow[T] {
	f.checks = append(f.checks, func(resp *Response) error {
		actual := resp.JSON(path).Value()
		if !reflect.DeepEqual(actual, expect) {
			return &Error{
				Kind:    ErrKindValidate,
				Op:      "check",
				Code:    path,
				Message: fmt.Sprintf("expect %v, got %v", expect, actual),
			}
		}
		return nil
	})
	return f
}

func (f *Flow[T]) CheckFn(fn CheckFunc) *Flow[T] {
	f.checks = append(f.checks, fn)
	return f
}

func (f *Flow[T]) JSON() *Flow[T] {
	f.checks = append(f.checks, func(resp *Response) error {
		if !resp.JSONResult().Exists() && len(resp.Body()) > 0 {
			return &Error{Kind: ErrKindDecode, Op: "json", Message: "invalid json"}
		}
		return nil
	})
	return f
}

func (f *Flow[T]) Decode(fn DecodeFunc[T]) *Flow[T] {
	f.decoder = fn
	return f
}

// IntoJSON uses encoding/json to decode response body into T.
func (f *Flow[T]) IntoJSON() *Flow[T] {
	f.decoder = func(resp *Response, out *T) error {
		if out == nil {
			return &Error{Kind: ErrKindDecode, Op: "into_json", Message: "nil output pointer"}
		}
		if err := json.Unmarshal(resp.Body(), out); err != nil {
			return &Error{Kind: ErrKindDecode, Op: "into_json", Cause: err}
		}
		return nil
	}
	return f
}

func (f *Flow[T]) BuildRequest() (*Request, error) {
	if f == nil {
		return nil, &Error{Kind: ErrKindValidate, Op: "build_request", Message: "nil flow"}
	}
	req := f.req.Clone()
	ctx := &Ctx{Request: req}

	for _, fn := range f.queryFns {
		query, err := fn(ctx)
		if err != nil {
			return nil, err
		}
		req.SetQuery(query)
	}
	for _, fn := range f.headerFns {
		header, err := fn(ctx)
		if err != nil {
			return nil, err
		}
		req.SetHeader(header)
	}
	for _, fn := range f.bodyFns {
		body, err := fn(ctx)
		if err != nil {
			return nil, err
		}
		req.SetBody(body)
	}
	return req, nil
}

func (f *Flow[T]) Result(client *Client) (T, error) {
	var out T

	if client == nil {
		return out, &Error{Kind: ErrKindValidate, Op: "run", Message: "nil client"}
	}
	req, err := f.BuildRequest()
	if err != nil {
		return out, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return out, err
	}
	for _, check := range f.checks {
		if e := check(resp); e != nil {
			return out, e
		}
	}
	if f.decoder != nil {
		if e := f.decoder(resp, &out); e != nil {
			return out, e
		}
	}
	return out, nil
}

func (f *Flow[T]) Response(client *Client) (*Response, error) {
	if client == nil {
		return nil, &Error{Kind: ErrKindValidate, Op: "run_response", Message: "nil client"}
	}
	req, err := f.BuildRequest()
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	for _, check := range f.checks {
		if e := check(resp); e != nil {
			return resp, e
		}
	}
	return resp, nil
}

// FlowTemplate is a reusable constructor for typed Flow instances.
// It composes with Flow instead of introducing a separate execution model.
type FlowTemplate[P any, R any] struct {
	build func(P) *Flow[R]
}

func Template[P any, R any](build func(P) *Flow[R]) FlowTemplate[P, R] {
	return FlowTemplate[P, R]{build: build}
}

func (t FlowTemplate[P, R]) Bind(param P) *Flow[R] {
	if t.build == nil {
		return &Flow[R]{}
	}
	return t.build(param)
}

func (t FlowTemplate[P, R]) Call(client *Client, param P) (R, error) {
	return t.Bind(param).Result(client)
}
