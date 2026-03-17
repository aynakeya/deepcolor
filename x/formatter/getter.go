package formatter

import (
	"strings"
)

const N = "#" // indexing character

const (
	ValueTypeBool = iota
	ValueTypeString
	ValueTypeInt
	ValueTypeFloat
	ValueTypeArray
	ValueTypeObject
	ValueTypeNull
)

// IValueGetter returns using specific expression,
// expression format is decided by underlying implementation,
// so it might be various from different value getter
type IValueGetter interface {
	// Value should return a value in following format:
	// bool, float64, int64, string, nil.
	// nil is used when no value found.
	Value(expression string, indexes []int) interface{}
	// Array should return  []interface{}, where interface{} is
	// data type describe in Value, if no result found, should return nil
	Array(expression string, indexes []int) []interface{}
}

// ValueGetter is a struct indicate which value getter should use,
// and its expression
type ValueGetter struct {
	Type       string
	Expression string
	indexes    []int
}

func (value ValueGetter) MissingIndexes() int {
	if value.Type != "json" {
		return 0
	}
	return strings.Count(value.Expression, N)
}

func (value ValueGetter) WithIndexes(indexes []int) ValueGetter {
	return ValueGetter{
		Type:       value.Type,
		Expression: value.Expression,
		indexes:    indexes,
	}
}

func ValueGetterFromStr(value string) ValueGetter {
	values := strings.Split(value, "::")
	if len(values) < 2 {
		return ValueGetter{}
	}
	return ValueGetter{
		Type:       values[0],
		Expression: strings.Join(values[1:], "::"),
	}
}
