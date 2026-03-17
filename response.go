package deepcolor

import (
	"encoding/json"
	"sync"

	"github.com/aynakeya/deepcolor/internal/httpx"
	"github.com/tidwall/gjson"
)

type Response struct {
	*httpx.Response

	once sync.Once
	json gjson.Result
}

func newResponse(raw *httpx.Response) *Response {
	return &Response{Response: raw}
}

func (r *Response) Raw() *httpx.Response {
	return r.Response
}

func (r *Response) Body() []byte {
	if r == nil || r.Response == nil {
		return nil
	}
	return r.Response.Body()
}

func (r *Response) JSONResult() gjson.Result {
	if r == nil {
		return gjson.Result{}
	}
	r.once.Do(func() {
		r.json = gjson.ParseBytes(r.Body())
	})
	return r.json
}

func (r *Response) JSON(path string) gjson.Result {
	return r.JSONResult().Get(path)
}

func (r *Response) UnmarshalJSON(out any) error {
	return json.Unmarshal(r.Body(), out)
}
