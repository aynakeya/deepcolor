package formatter

import "math"

type JsonSchema struct {
	schema map[string]interface{}
}

type JsonArraySchema struct {
	schema interface{}
}

func (s *JsonSchema) format(cache *valueCache, arrayIndexes []int) map[string]interface{} {
	result := make(map[string]interface{})
	valid := false
	for key, value := range s.schema {
		switch v := value.(type) {
		case ValueGetter:
			// re-assign value, so we can identify if it is a nil
			val := cache.getValue(v.WithIndexes(arrayIndexes))
			valid = valid || val != nil
			result[key] = val
		case *JsonSchema:
			val := v.format(cache, arrayIndexes)
			valid = valid || val != nil
			result[key] = val
		case *JsonArraySchema:
			val := v.format(cache, arrayIndexes)
			valid = valid || val != nil
			result[key] = val
		}
	}
	if !valid {
		return nil
	}
	return result
}

func (s *JsonSchema) Format(value string) map[string]interface{} {
	return s.format(newValueCache(value), make([]int, 0))
}

func (s *JsonArraySchema) format(cache *valueCache, arrayIndexes []int) []interface{} {
	switch schema := s.schema.(type) {
	case ValueGetter:
		return cache.getArray(schema.WithIndexes(arrayIndexes))
	case *JsonSchema:
		results := make([]interface{}, 0)
		for i := 0; i < math.MaxInt32; i++ {
			r := schema.format(cache, append(arrayIndexes, i))
			if r == nil {
				break
			}
			results = append(results, schema.format(cache, append(arrayIndexes, i)))
		}
		if len(results) == 0 {
			return nil
		}
		return results
	}
	return nil
}
func (s *JsonArraySchema) Format(value string) []interface{} {
	return s.format(newValueCache(value), make([]int, 0))
}

func NewJsonArraySchema(schema []interface{}) *JsonArraySchema {
	if len(schema) != 1 {
		// invalid schema
		return nil
	}
	switch schema[0].(type) {
	case string:
		return &JsonArraySchema{
			schema: ValueGetterFromStr(schema[0].(string)),
		}
	case map[string]interface{}:
		return &JsonArraySchema{
			schema: NewJsonSchema(schema[0].(map[string]interface{})),
		}
	}
	return nil
}

func NewJsonSchema(schema map[string]interface{}) *JsonSchema {
	result := make(map[string]interface{})
	for key, value := range schema {
		switch value.(type) {
		case string:
			result[key] = ValueGetterFromStr(value.(string))
		case map[string]interface{}:
			v := NewJsonSchema(value.(map[string]interface{}))
			if v == nil {
				return nil
			}
			result[key] = v
		case []interface{}:
			v := NewJsonArraySchema(value.([]interface{}))
			if v == nil {
				return nil
			}
			result[key] = v
		}
	}
	return &JsonSchema{
		schema: result,
	}
}
