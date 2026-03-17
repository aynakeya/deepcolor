package translators

import "github.com/aynakeya/deepcolor/transform"

type Foreach struct {
	transform.BaseTranslator
	InternTrans transform.Translator
}

func NewForeach(translator transform.Translator) transform.Translator {
	t := &Foreach{
		BaseTranslator: transform.BaseTranslator{
			Type: "Foreach",
		},
		InternTrans: translator,
	}
	return t
}

func (f *Foreach) Apply(value interface{}) (interface{}, error) {
	v, ok := value.([]interface{})
	if !ok {
		return value, transform.ErrorWrongSrcType("[]interface")
	}
	for index, _ := range v {
		v[index] = f.InternTrans.MustApply(v[index])
	}
	return v, nil
}

func (f *Foreach) MustApply(value interface{}) interface{} {
	v, _ := f.Apply(value)
	return v
}
