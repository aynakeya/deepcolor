package deepcolor

import (
	"errors"
	"fmt"
)

const (
	ErrKindTransport = "transport"
	ErrKindHTTP      = "http_status"
	ErrKindDecode    = "decode"
	ErrKindValidate  = "validate"
	ErrKindUpstream  = "upstream"
)

type Error struct {
	Kind    string
	Op      string
	URL     string
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	msg := fmt.Sprintf("deepcolor %s", e.Kind)
	if e.Op != "" {
		msg += " op=" + e.Op
	}
	if e.URL != "" {
		msg += " url=" + e.URL
	}
	if e.Status != 0 {
		msg += fmt.Sprintf(" status=%d", e.Status)
	}
	if e.Code != "" {
		msg += " code=" + e.Code
	}
	if e.Message != "" {
		msg += " msg=" + e.Message
	}
	if e.Cause != nil {
		msg += ": " + e.Cause.Error()
	}
	return msg
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func IsKind(err error, kind string) bool {
	var dcErr *Error
	if !errors.As(err, &dcErr) {
		return false
	}
	return dcErr.Kind == kind
}
