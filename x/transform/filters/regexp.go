package filters

import (
	"encoding/json"
	"github.com/aynakeya/deepcolor/transform"
	"regexp"
)

type RegExpFilter struct {
	transform.BaseFilter
	Include    bool // true if the filter should include the meta, false if it should exclude it.
	Expression *regexp.Regexp
}

func RegExp(expression *regexp.Regexp, include bool) transform.Filter {
	return &RegExpFilter{
		BaseFilter: transform.BaseFilter{
			Type: "RegExpFilter",
		},
		Expression: expression,
		Include:    include,
	}
}

func (r *RegExpFilter) Check(value interface{}) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}
	return r.Expression.MatchString(s) == r.Include
}

func (r *RegExpFilter) MarshalJSON() ([]byte, error) {
	type Tmp RegExpFilter
	return json.Marshal(&struct {
		*Tmp
		Expression string
	}{
		Tmp:        (*Tmp)(r),
		Expression: r.Expression.String(),
	})
}

func (r *RegExpFilter) UnmarshalJSON(data []byte) error {
	type Tmp RegExpFilter
	aux := &struct {
		*Tmp
		Expression string
	}{
		Tmp: (*Tmp)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Expression = regexp.MustCompile(aux.Expression)
	return nil
}
