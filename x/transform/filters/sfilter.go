package filters

import (
	"encoding/json"
	"github.com/aynakeya/deepcolor/transform"
)

type StructFilter struct {
	transform.BaseFilter
	Target transform.Field
	Filter transform.Filter
}

func (f *StructFilter) Check(value interface{}) bool {
	v := f.Target.GetValue(value)
	return f.Filter.Check(v.Interface())
}

func (f *StructFilter) UnmarshalJSON(data []byte) error {
	type Tmp StructFilter
	aux := &struct {
		*Tmp
		Filter json.RawMessage
	}{
		Tmp: (*Tmp)(f),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	var err error
	f.Filter, err = transform.UnmarshalFilter(aux.Filter)
	return err
}
