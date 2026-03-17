package translators

import (
	"encoding/json"
	"github.com/aynakeya/deepcolor/transform"
	"regexp"
)

type RegExpReplacer struct {
	transform.BaseTranslator
	Expression *regexp.Regexp
	Repl       string
}

func (r *RegExpReplacer) MarshalJSON() ([]byte, error) {
	type Tmp RegExpReplacer
	return json.Marshal(&struct {
		*Tmp
		Expression string
	}{
		Tmp:        (*Tmp)(r),
		Expression: r.Expression.String(),
	})
}

func (r *RegExpReplacer) UnmarshalJSON(data []byte) error {
	type Tmp RegExpReplacer
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

func NewRegExpReplacer(expression *regexp.Regexp, repl string) transform.Translator {
	t := &RegExpReplacer{
		BaseTranslator: transform.BaseTranslator{
			Type: "RegExpReplacer",
		},
		Expression: expression,
		Repl:       repl,
	}
	return t
}

func (r *RegExpReplacer) Apply(value interface{}) (interface{}, error) {
	s, ok := value.(string)
	if !ok {
		return "", transform.ErrorWrongSrcType("string")
	}
	return r.Expression.ReplaceAllString(s, r.Repl), nil
}

func (r *RegExpReplacer) MustApply(value interface{}) interface{} {
	v, _ := r.Apply(value)
	return v
}

type RegExpFind struct {
	transform.BaseTranslator
	Expression *regexp.Regexp
	GroupNum   int
	All        bool
}

func NewRegExpFindFirst(expression *regexp.Regexp, groupNum int) transform.Translator {
	t := &RegExpFind{
		BaseTranslator: transform.BaseTranslator{
			Type: "RegExpFind",
		},
		Expression: expression,
		GroupNum:   groupNum,
		All:        false,
	}
	return t
}

func NewRegExpFindAll(expression *regexp.Regexp, groupNum int) transform.Translator {
	t := &RegExpFind{
		BaseTranslator: transform.BaseTranslator{
			Type: "RegExpFind",
		},
		Expression: expression,
		GroupNum:   groupNum,
		All:        true,
	}
	return t
}

func (r *RegExpFind) MarshalJSON() ([]byte, error) {
	type Tmp RegExpFind
	return json.Marshal(&struct {
		*Tmp
		Expression string
	}{
		Tmp:        (*Tmp)(r),
		Expression: r.Expression.String(),
	})
}

func (r *RegExpFind) UnmarshalJSON(data []byte) error {
	type Tmp RegExpFind
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

func (t *RegExpFind) Apply(value interface{}) (interface{}, error) {
	if t.All {
		return t.applyAll(value)
	}
	return t.applySingle(value)
}

func (t *RegExpFind) applySingle(value interface{}) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", transform.ErrorWrongSrcType("string")
	}
	rs := t.Expression.FindStringSubmatch(s)
	if t.GroupNum < len(rs) {
		return rs[t.GroupNum], nil
	}
	return "", transform.ErrorRegexpInvalidGroup(t.GroupNum)
}

func (t *RegExpFind) applyAll(value interface{}) ([]string, error) {
	s, ok := value.(string)
	if !ok {
		return []string{}, transform.ErrorWrongSrcType("string")
	}
	vs := []string{}
	rs := t.Expression.FindAllStringSubmatch(s, -1)
	if len(rs) == 0 {
		return vs, transform.ErrorRegexpInvalidGroup(t.GroupNum)
	}
	if t.GroupNum >= len(rs[0]) {
		return vs, transform.ErrorRegexpInvalidGroup(t.GroupNum)
	}
	for _, tmp := range rs {
		vs = append(vs, tmp[t.GroupNum])
	}
	return vs, nil
}

func (t *RegExpFind) MustApply(value interface{}) interface{} {
	v, _ := t.Apply(value)
	return v
}
