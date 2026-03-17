package pipeline

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

func registerBuiltins(r *Registry) {
	r.RegisterExtractor("input", func(_ *Context, st *State, arg string) (any, error) {
		ac, err := CompileAccessor(arg)
		if err != nil {
			return nil, err
		}
		return ac.Get(st.Input)
	})
	r.RegisterExtractor("doc", func(_ *Context, st *State, arg string) (any, error) {
		ac, err := CompileAccessor(arg)
		if err != nil {
			return nil, err
		}
		return ac.Get(st.Doc)
	})
	r.RegisterExtractor("json", func(_ *Context, st *State, arg string) (any, error) {
		var raw string
		switch v := st.Input.(type) {
		case string:
			raw = v
		case []byte:
			raw = string(v)
		default:
			return nil, &Error{Kind: ErrKindRuntime, Op: "json", Message: "json extractor requires string/[]byte input"}
		}
		j := gjson.Parse(raw)
		res := j.Get(arg)
		if !res.Exists() {
			return nil, nil
		}
		return res.Value(), nil
	})
	r.RegisterExtractor("regex", func(_ *Context, st *State, arg string) (any, error) {
		parts := strings.SplitN(arg, "=>", 2)
		if len(parts) == 0 || parts[0] == "" {
			return nil, &Error{Kind: ErrKindRuntime, Op: "regex", Message: "regex arg required"}
		}
		group := 0
		if len(parts) == 2 {
			g, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, err
			}
			group = g
		}
		re, err := regexp.Compile(parts[0])
		if err != nil {
			return nil, err
		}
		var raw string
		switch v := st.Input.(type) {
		case string:
			raw = v
		case []byte:
			raw = string(v)
		default:
			return nil, &Error{Kind: ErrKindRuntime, Op: "regex", Message: "regex extractor requires string/[]byte input"}
		}
		m := re.FindStringSubmatch(raw)
		if len(m) <= group {
			return nil, nil
		}
		return m[group], nil
	})

	r.RegisterMapper("cast_string", EffectPure, func(_ *Context, in any, _ []any) (any, error) { return cast.ToString(in), nil })
	r.RegisterMapper("cast_int", EffectPure, func(_ *Context, in any, _ []any) (any, error) { return cast.ToIntE(in) })
	r.RegisterMapper("lower", EffectPure, func(_ *Context, in any, _ []any) (any, error) { return strings.ToLower(cast.ToString(in)), nil })
	r.RegisterMapper("upper", EffectPure, func(_ *Context, in any, _ []any) (any, error) { return strings.ToUpper(cast.ToString(in)), nil })
	r.RegisterMapper("trim", EffectPure, func(_ *Context, in any, _ []any) (any, error) { return strings.TrimSpace(cast.ToString(in)), nil })
	r.RegisterMapper("format", EffectPure, func(_ *Context, in any, args []any) (any, error) {
		if len(args) == 0 {
			return nil, &Error{Kind: ErrKindRuntime, Op: "format", Message: "missing format string"}
		}
		return fmt.Sprintf(cast.ToString(args[0]), in), nil
	})
	r.RegisterMapper("regexp_replace", EffectPure, func(_ *Context, in any, args []any) (any, error) {
		if len(args) < 2 {
			return nil, &Error{Kind: ErrKindRuntime, Op: "regexp_replace", Message: "need pattern and replacement"}
		}
		re, err := regexp.Compile(cast.ToString(args[0]))
		if err != nil {
			return nil, err
		}
		return re.ReplaceAllString(cast.ToString(in), cast.ToString(args[1])), nil
	})

	r.RegisterPredicate("exists", EffectPure, func(_ *Context, in any, _ []any) (bool, error) { return in != nil, nil })
	r.RegisterPredicate("eq", EffectPure, func(_ *Context, in any, args []any) (bool, error) {
		if len(args) == 0 {
			return false, &Error{Kind: ErrKindRuntime, Op: "eq", Message: "missing expected value"}
		}
		return fmt.Sprint(in) == fmt.Sprint(args[0]), nil
	})
	r.RegisterPredicate("contains", EffectPure, func(_ *Context, in any, args []any) (bool, error) {
		if len(args) == 0 {
			return false, &Error{Kind: ErrKindRuntime, Op: "contains", Message: "missing needle"}
		}
		return strings.Contains(cast.ToString(in), cast.ToString(args[0])), nil
	})
	r.RegisterPredicate("gt", EffectPure, func(_ *Context, in any, args []any) (bool, error) {
		if len(args) == 0 {
			return false, &Error{Kind: ErrKindRuntime, Op: "gt", Message: "missing threshold"}
		}
		return cast.ToFloat64(in) > cast.ToFloat64(args[0]), nil
	})
	r.RegisterPredicate("lt", EffectPure, func(_ *Context, in any, args []any) (bool, error) {
		if len(args) == 0 {
			return false, &Error{Kind: ErrKindRuntime, Op: "lt", Message: "missing threshold"}
		}
		return cast.ToFloat64(in) < cast.ToFloat64(args[0]), nil
	})
}
