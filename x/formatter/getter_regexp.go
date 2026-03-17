package formatter

import (
	"regexp"
)

type regexpValueGetter struct {
	data string
}

func (r *regexpValueGetter) Value(expression string, indexes []int) interface{} {
	re, err := regexp.Compile(expression)
	if err != nil {
		return nil
	}
	// Find all matches
	matches := re.FindAllStringSubmatch(r.data, -1)
	lastIndex := 0
	if len(indexes) != 0 {
		lastIndex = indexes[len(indexes)-1]
	}
	if len(matches) <= lastIndex {
		// No match found
		return nil
	}
	if len(matches[lastIndex]) > 1 {
		// If there's a capturing group, return the first one
		return matches[0][1]
	}
	// Otherwise return the entire match
	return matches[lastIndex][0]
}

func (r *regexpValueGetter) Array(expression string, indexes []int) []interface{} {
	re, err := regexp.Compile(expression)
	if err != nil {
		return nil
	}

	// Find all matches
	matches := re.FindAllStringSubmatch(r.data, -1)
	if len(matches) == 0 {
		// No match found
		return nil
	}

	if len(indexes) != 0 {
		return nil
	}

	results := make([]interface{}, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			// If there's a capturing group, return the first one
			results = append(results, m[1])
		} else {
			// Otherwise return the entire match
			results = append(results, m[0])
		}
	}
	return results
}

func NewRegexpValueGetter(data string) IValueGetter {
	return &regexpValueGetter{
		data: data,
	}
}
