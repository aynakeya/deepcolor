package translators

import "github.com/aynakeya/deepcolor/transform"

type Value struct {
	transform.BaseTranslator
	Value interface{}
}

func NewSetter(value interface{}) transform.Translator {
	t := &Value{
		BaseTranslator: transform.BaseTranslator{
			Type: "Value",
		},
		Value: value,
	}
	return t
}

func NewValue(value interface{}) transform.Translator {
	return NewSetter(value)
}

func (f *Value) Apply(value interface{}) (interface{}, error) {
	return f.Value, nil
}

func (f *Value) MustApply(value interface{}) interface{} {
	v, _ := f.Apply(value)
	return v
}
