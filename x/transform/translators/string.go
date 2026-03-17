package translators

import (
	"fmt"
	"github.com/aynakeya/deepcolor/transform"
	"strings"
)

type StrCase struct {
	transform.BaseTranslator
	Lowercase bool
}

func NewStrCase(lowercase bool) transform.Translator {
	t := &StrCase{transform.BaseTranslator{"StrCase"}, lowercase}
	return t
}

func (c *StrCase) Apply(value interface{}) (interface{}, error) {
	s, ok := value.(string)
	if !ok {
		return s, transform.ErrorWrongSrcType("string")
	}
	if c.Lowercase {
		return strings.ToLower(s), nil
	}
	return strings.ToUpper(s), nil
}

func (c *StrCase) MustApply(value interface{}) interface{} {
	v, _ := c.Apply(value)
	return v
}

type Formatter struct {
	transform.BaseTranslator
	Format string
}

func NewFormatter(format string) transform.Translator {
	t := &Formatter{transform.BaseTranslator{"Formatter"}, format}
	return t
}

func (f *Formatter) Apply(value interface{}) (interface{}, error) {
	return fmt.Sprintf(f.Format, value), nil
}

func (f *Formatter) MustApply(value interface{}) interface{} {
	v, _ := f.Apply(value)
	return v
}
