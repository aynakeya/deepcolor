package transform

import (
	"errors"
	"fmt"
)

var errorTranslatorNotFound = errors.New("translator not found")
var errorFailToApply = errors.New("fail to apply translator")

func ErrorWrongSrcType(req string) error {
	return errors.New(fmt.Sprintf("wrong source type, should be req %s", req))
}

func ErrorRegexpInvalidGroup(n int) error {
	return errors.New(fmt.Sprintf("group %d is not valid", n))
}
