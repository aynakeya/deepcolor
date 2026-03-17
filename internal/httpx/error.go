package httpx

import "fmt"

const (
	ErrKindTransport   = "transport"
	ErrKindHTTPStatus  = "http_status"
	ErrKindMethod      = "method_not_supported"
	ErrKindEncodeBody  = "encode_body"
	ErrKindInvalidURL  = "invalid_url"
	ErrKindNilRequester = "nil_requester"
)

type Error struct {
	Kind    string
	Op      string
	URL     string
	Method  string
	Status  int
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	s := fmt.Sprintf("httpx %s", e.Kind)
	if e.Op != "" {
		s += " op=" + e.Op
	}
	if e.Method != "" {
		s += " method=" + e.Method
	}
	if e.URL != "" {
		s += " url=" + e.URL
	}
	if e.Status != 0 {
		s += fmt.Sprintf(" status=%d", e.Status)
	}
	if e.Message != "" {
		s += " msg=" + e.Message
	}
	if e.Cause != nil {
		s += ": " + e.Cause.Error()
	}
	return s
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}
