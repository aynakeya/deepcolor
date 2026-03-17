package formatter

import (
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
)

type jsonValueGetter struct {
	result gjson.Result
}

func (j *jsonValueGetter) Value(expression string, indexes []int) interface{} {
	for _, index := range indexes {
		expression = strings.Replace(expression, N, strconv.Itoa(index), 1)
	}
	result := j.result.Get(expression)
	if result.Type == gjson.Null || result.Type == gjson.JSON {
		return nil
	}
	return result.Value()
}

func (j *jsonValueGetter) Array(expression string, indexes []int) []interface{} {
	for _, index := range indexes {
		expression = strings.Replace(expression, N, strconv.Itoa(index), 1)
	}
	val := make([]interface{}, 0)
	if !j.result.Get(expression).Exists() {
		return nil
	}
	j.result.Get(expression).ForEach(func(key, value gjson.Result) bool {
		val = append(val, value.Value())
		return true
	})
	if len(val) == 0 {
		return nil
	}
	return val
}

func NewJsonValueGetter(data string) IValueGetter {
	return &jsonValueGetter{
		result: gjson.Parse(data),
	}
}
