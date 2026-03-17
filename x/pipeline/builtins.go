package pipeline

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
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
	r.RegisterExtractor("regex_all", func(_ *Context, st *State, arg string) (any, error) {
		parts := strings.SplitN(arg, "=>", 2)
		if len(parts) == 0 || parts[0] == "" {
			return nil, &Error{Kind: ErrKindRuntime, Op: "regex_all", Message: "regex arg required"}
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
			return nil, &Error{Kind: ErrKindRuntime, Op: "regex_all", Message: "regex_all extractor requires string/[]byte input"}
		}
		matches := re.FindAllStringSubmatch(raw, -1)
		if len(matches) == 0 {
			return []any{}, nil
		}
		out := make([]any, 0, len(matches))
		for _, m := range matches {
			if len(m) <= group {
				continue
			}
			out = append(out, m[group])
		}
		return out, nil
	})
	r.RegisterExtractor("xml", func(_ *Context, st *State, arg string) (any, error) {
		doc, err := parseXMLInput(st.Input)
		if err != nil {
			return nil, err
		}
		expr := applyXMLIndexPlaceholder(arg, st.Scope)
		node := xmlquery.FindOne(doc, expr)
		if node == nil {
			return nil, nil
		}
		return node.InnerText(), nil
	})
	r.RegisterExtractor("xml_all", func(_ *Context, st *State, arg string) (any, error) {
		doc, err := parseXMLInput(st.Input)
		if err != nil {
			return nil, err
		}
		expr := applyXMLIndexPlaceholder(arg, st.Scope)
		nodes := xmlquery.Find(doc, expr)
		if len(nodes) == 0 {
			return []any{}, nil
		}
		out := make([]any, 0, len(nodes))
		for _, n := range nodes {
			out = append(out, n.InnerText())
		}
		return out, nil
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
	r.RegisterMapper("regex_find", EffectPure, func(_ *Context, in any, args []any) (any, error) {
		if len(args) == 0 {
			return nil, &Error{Kind: ErrKindRuntime, Op: "regex_find", Message: "need pattern"}
		}
		group := 0
		if len(args) > 1 {
			group = cast.ToInt(args[1])
		}
		re, err := regexp.Compile(cast.ToString(args[0]))
		if err != nil {
			return nil, err
		}
		m := re.FindStringSubmatch(cast.ToString(in))
		if len(m) <= group {
			return nil, nil
		}
		return m[group], nil
	})
	r.RegisterMapper("default", EffectPure, func(_ *Context, in any, args []any) (any, error) {
		if len(args) == 0 {
			return in, nil
		}
		if in == nil || cast.ToString(in) == "" {
			return args[0], nil
		}
		return in, nil
	})
	r.RegisterMapper("split", EffectPure, func(_ *Context, in any, args []any) (any, error) {
		sep := ","
		if len(args) > 0 {
			sep = cast.ToString(args[0])
		}
		parts := strings.Split(cast.ToString(in), sep)
		out := make([]any, 0, len(parts))
		for _, p := range parts {
			out = append(out, p)
		}
		return out, nil
	})
	r.RegisterMapper("join", EffectPure, func(_ *Context, in any, args []any) (any, error) {
		sep := ","
		if len(args) > 0 {
			sep = cast.ToString(args[0])
		}
		arr, ok := in.([]any)
		if !ok {
			return nil, &Error{Kind: ErrKindRuntime, Op: "join", Message: "join expects []any"}
		}
		parts := make([]string, 0, len(arr))
		for _, item := range arr {
			parts = append(parts, cast.ToString(item))
		}
		return strings.Join(parts, sep), nil
	})
	r.RegisterMapper("switch_first", EffectPure, func(ctx *Context, in any, args []any) (any, error) {
		var lastErr error
		for _, candidate := range args {
			op, err := parseOpCandidate(candidate)
			if err != nil {
				lastErr = err
				continue
			}
			entry, err := r.getMapper(op.Name)
			if err != nil {
				lastErr = err
				continue
			}
			out, err := entry.fn(ctx, in, op.Args)
			if err == nil {
				if out != nil {
					return out, nil
				}
			} else {
				lastErr = err
			}
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return in, nil
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

func parseOpCandidate(v any) (OpSpec, error) {
	switch t := v.(type) {
	case OpSpec:
		if t.Name == "" {
			return OpSpec{}, &Error{Kind: ErrKindRuntime, Op: "switch_first", Message: "candidate op name is empty"}
		}
		return t, nil
	case map[string]any:
		name := cast.ToString(t["name"])
		if name == "" {
			return OpSpec{}, &Error{Kind: ErrKindRuntime, Op: "switch_first", Message: "candidate map.name is empty"}
		}
		rawArgs, ok := t["args"]
		if !ok {
			return OpSpec{Name: name}, nil
		}
		args, ok := rawArgs.([]any)
		if !ok {
			return OpSpec{}, &Error{Kind: ErrKindRuntime, Op: "switch_first", Message: "candidate map.args must be []any"}
		}
		return OpSpec{Name: name, Args: args}, nil
	case string:
		if t == "" {
			return OpSpec{}, &Error{Kind: ErrKindRuntime, Op: "switch_first", Message: "candidate string is empty"}
		}
		return OpSpec{Name: t}, nil
	default:
		return OpSpec{}, &Error{Kind: ErrKindRuntime, Op: "switch_first", Message: "unsupported candidate type"}
	}
}

func parseXMLInput(input any) (*xmlquery.Node, error) {
	switch v := input.(type) {
	case string:
		return xmlquery.Parse(strings.NewReader(v))
	case []byte:
		return xmlquery.Parse(strings.NewReader(string(v)))
	default:
		return nil, &Error{Kind: ErrKindRuntime, Op: "xml", Message: "xml extractor requires string/[]byte input"}
	}
}

func applyXMLIndexPlaceholder(expr string, scope map[string]any) string {
	if !strings.Contains(expr, "#") {
		return expr
	}
	if scope == nil {
		return expr
	}
	rawIdx, ok := scope["$index"]
	if !ok {
		return expr
	}
	idx := cast.ToInt(rawIdx) + 1
	return strings.ReplaceAll(expr, "#", strconv.Itoa(idx))
}
