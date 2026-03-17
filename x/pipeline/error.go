package pipeline

import (
	"fmt"
)

type ErrorKind string

const (
	ErrKindValidate ErrorKind = "validate"
	ErrKindCompile  ErrorKind = "compile"
	ErrKindRuntime  ErrorKind = "runtime"
	ErrKindRegistry ErrorKind = "registry"
)

type Error struct {
	Kind      ErrorKind
	Path      string
	Op        string
	InputType string
	Message   string
	Cause     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	s := fmt.Sprintf("pipeline %s", e.Kind)
	if e.Path != "" {
		s += " path=" + e.Path
	}
	if e.Op != "" {
		s += " op=" + e.Op
	}
	if e.InputType != "" {
		s += " input=" + e.InputType
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
