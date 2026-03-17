package translators

import "github.com/aynakeya/deepcolor/transform"

type FilterTranslator struct {
	transform.BaseTranslator
	Filter transform.Filter
}

func NewFilterTranslator(filter transform.Filter) transform.Translator {
	t := &FilterTranslator{
		BaseTranslator: transform.BaseTranslator{
			Type: "FilterTranslator",
		},
		Filter: filter,
	}
	return t
}

func (f *FilterTranslator) Apply(value interface{}) (interface{}, error) {
	v, ok := value.([]interface{})
	retval := make([]interface{}, 0)
	if !ok {
		return value, transform.ErrorWrongSrcType("[]interface")
	}
	for index, _ := range v {
		if f.Filter.Check(v[index]) {
			retval = append(retval, v[index])
		}
	}
	return retval, nil
}

func (f *FilterTranslator) MustApply(value interface{}) interface{} {
	v, _ := f.Apply(value)
	return v
}
