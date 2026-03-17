package httpx

import (
	"context"
	"net/url"
	"time"
)

type Request struct {
	Method  string            `json:"method"`
	Url     *url.URL          `json:"url"`
	Header  map[string]string `json:"header"`
	Data    []byte            `json:"data"`
	Timeout time.Duration     `json:"timeout"`
	Context context.Context   `json:"-"`
}

// todo
func (r *Request) SetCookie(key string, value string) {
}
func (r *Request) SetCookies(map[string]string) {
}
