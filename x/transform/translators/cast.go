package translators

import (
	"errors"
	"fmt"
	"github.com/aynakeya/deepcolor/transform"
	"github.com/spf13/cast"
)

type Cast struct {
	transform.BaseTranslator
	ToType string
}

func NewCast(destType string) transform.Translator {
	t := &Cast{
		BaseTranslator: transform.BaseTranslator{
			Type: "Cast",
		},
		ToType: destType,
	}
	return t
}

func (c *Cast) Apply(value interface{}) (interface{}, error) {
	switch c.ToType {
	case "string":
		return cast.ToStringE(value)
	case "bool":
		return cast.ToBoolE(value)
	case "int":
		return cast.ToIntE(value)
	}
	return value, errors.New(fmt.Sprintf("%s not support in casting", c.ToType))
}

func (c *Cast) MustApply(value interface{}) interface{} {
	v, _ := c.Apply(value)
	return v
}
