package deepcolor

import (
	"fmt"
	"reflect"
)

func CheckEqual(resp *Response, path string, expect any) error {
	actual := resp.JSON(path).Value()
	if reflect.DeepEqual(actual, expect) {
		return nil
	}
	return &Error{
		Kind:    ErrKindValidate,
		Op:      "check_equal",
		Code:    path,
		Message: fmt.Sprintf("expect %v, got %v", expect, actual),
	}
}

func MustString(resp *Response, path string) (string, error) {
	v := resp.JSON(path).String()
	if v != "" {
		return v, nil
	}
	return "", &Error{
		Kind:    ErrKindValidate,
		Op:      "must_string",
		Code:    path,
		Message: "empty string",
	}
}
