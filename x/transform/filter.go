package transform

import (
	"github.com/aynakeya/deepcolor/internal/dynmarshaller"
)

type Filter interface {
	GetType() string
	Check(value interface{}) bool // return true if this value should be keep
}

type BaseFilter struct {
	Type string
}

func (f *BaseFilter) GetType() string {
	return f.Type
}

var filterRegistry = dynmarshaller.NewDynamicUnmarshaller[Filter](make(map[string]Filter))

func RegisterFilter(filters ...Filter) {
	for _, filter := range filters {
		filterRegistry.Register(filter.GetType(), filter)
	}
}

func UnmarshalFilter(data []byte) (Filter, error) {
	return filterRegistry.Unmarshal(data)
}
