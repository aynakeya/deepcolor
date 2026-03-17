package pipeline

import "testing"

func TestBuiltinsExtractors_All(t *testing.T) {
	reg := NewRegistryWithBuiltins()
	ctx := &Context{}

	// input
	exInput, err := reg.getExtractor("input")
	if err != nil {
		t.Fatalf("get input extractor failed: %v", err)
	}
	stInput := &State{
		Input: map[string]any{"a": map[string]any{"b": "x"}},
		Doc:   map[string]any{},
		Scope: map[string]any{},
	}
	v, err := exInput(ctx, stInput, "a.b")
	if err != nil || v != "x" {
		t.Fatalf("input extractor mismatch: v=%v err=%v", v, err)
	}

	// doc
	exDoc, err := reg.getExtractor("doc")
	if err != nil {
		t.Fatalf("get doc extractor failed: %v", err)
	}
	stDoc := &State{
		Input: nil,
		Doc:   map[string]any{"x": map[string]any{"y": 3}},
		Scope: map[string]any{},
	}
	v, err = exDoc(ctx, stDoc, "x.y")
	if err != nil || v != 3 {
		t.Fatalf("doc extractor mismatch: v=%v err=%v", v, err)
	}

	// json
	exJSON, err := reg.getExtractor("json")
	if err != nil {
		t.Fatalf("get json extractor failed: %v", err)
	}
	stJSON := &State{Input: []byte(`{"data":{"title":"hello"}}`), Doc: map[string]any{}, Scope: map[string]any{}}
	v, err = exJSON(ctx, stJSON, "data.title")
	if err != nil || v != "hello" {
		t.Fatalf("json extractor mismatch: v=%v err=%v", v, err)
	}

	// regex
	exRegex, err := reg.getExtractor("regex")
	if err != nil {
		t.Fatalf("get regex extractor failed: %v", err)
	}
	stRegex := &State{Input: "abc123xyz", Doc: map[string]any{}, Scope: map[string]any{}}
	v, err = exRegex(ctx, stRegex, `([0-9]+)=>1`)
	if err != nil || v != "123" {
		t.Fatalf("regex extractor mismatch: v=%v err=%v", v, err)
	}

	// regex_all
	exRegexAll, err := reg.getExtractor("regex_all")
	if err != nil {
		t.Fatalf("get regex_all extractor failed: %v", err)
	}
	v, err = exRegexAll(ctx, &State{Input: "f1f2f3", Doc: map[string]any{}, Scope: map[string]any{}}, `f([0-9])=>1`)
	if err != nil {
		t.Fatalf("regex_all extractor failed: %v", err)
	}
	arr, ok := v.([]any)
	if !ok || len(arr) != 3 || arr[0] != "1" || arr[2] != "3" {
		t.Fatalf("regex_all extractor mismatch: %#v", v)
	}

	// xml
	exXML, err := reg.getExtractor("xml")
	if err != nil {
		t.Fatalf("get xml extractor failed: %v", err)
	}
	xmlInput := `<?xml version="1.0"?><library><book><title>A</title></book><book><title>B</title></book></library>`
	v, err = exXML(ctx, &State{Input: xmlInput, Doc: map[string]any{}, Scope: map[string]any{}}, "//book[1]/title")
	if err != nil || v != "A" {
		t.Fatalf("xml extractor mismatch: v=%v err=%v", v, err)
	}
	// xml index placeholder replacement (# + $index)
	v, err = exXML(ctx, &State{Input: xmlInput, Doc: map[string]any{}, Scope: map[string]any{"$index": 1}}, "//book[#]/title")
	if err != nil || v != "B" {
		t.Fatalf("xml extractor index mismatch: v=%v err=%v", v, err)
	}

	// xml_all
	exXMLAll, err := reg.getExtractor("xml_all")
	if err != nil {
		t.Fatalf("get xml_all extractor failed: %v", err)
	}
	v, err = exXMLAll(ctx, &State{Input: xmlInput, Doc: map[string]any{}, Scope: map[string]any{}}, "//book/title")
	if err != nil {
		t.Fatalf("xml_all extractor failed: %v", err)
	}
	arr, ok = v.([]any)
	if !ok || len(arr) != 2 || arr[0] != "A" || arr[1] != "B" {
		t.Fatalf("xml_all extractor mismatch: %#v", v)
	}
}

func TestBuiltinsMappers_All(t *testing.T) {
	reg := NewRegistryWithBuiltins()
	ctx := &Context{}

	tests := []struct {
		name   string
		op     string
		in     any
		args   []any
		assert func(any, error) bool
	}{
		{name: "cast_string", op: "cast_string", in: 12, assert: func(v any, err error) bool { return err == nil && v == "12" }},
		{name: "cast_int", op: "cast_int", in: "42", assert: func(v any, err error) bool { return err == nil && v == 42 }},
		{name: "lower", op: "lower", in: "REOL", assert: func(v any, err error) bool { return err == nil && v == "reol" }},
		{name: "upper", op: "upper", in: "reol", assert: func(v any, err error) bool { return err == nil && v == "REOL" }},
		{name: "trim", op: "trim", in: "  x  ", assert: func(v any, err error) bool { return err == nil && v == "x" }},
		{name: "format", op: "format", in: "x", args: []any{"[%s]"}, assert: func(v any, err error) bool { return err == nil && v == "[x]" }},
		{name: "regexp_replace", op: "regexp_replace", in: "abc123", args: []any{"[0-9]+", "X"}, assert: func(v any, err error) bool { return err == nil && v == "abcX" }},
		{name: "regex_find", op: "regex_find", in: "ep12v2", args: []any{`ep([0-9]+)`, 1}, assert: func(v any, err error) bool { return err == nil && v == "12" }},
		{name: "default", op: "default", in: "", args: []any{"fallback"}, assert: func(v any, err error) bool { return err == nil && v == "fallback" }},
		{name: "split", op: "split", in: "a,b,c", args: []any{","}, assert: func(v any, err error) bool {
			arr, ok := v.([]any)
			return err == nil && ok && len(arr) == 3 && arr[0] == "a" && arr[2] == "c"
		}},
		{name: "join", op: "join", in: []any{"a", "b"}, args: []any{"|"}, assert: func(v any, err error) bool { return err == nil && v == "a|b" }},
		{name: "switch_first", op: "switch_first", in: "REOL", args: []any{
			OpSpec{Name: "cast_int"}, // fail
			OpSpec{Name: "lower"},    // success
		}, assert: func(v any, err error) bool { return err == nil && v == "reol" }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := reg.getMapper(tc.op)
			if err != nil {
				t.Fatalf("get mapper failed: %v", err)
			}
			v, e := entry.fn(ctx, tc.in, tc.args)
			if !tc.assert(v, e) {
				t.Fatalf("unexpected mapper result for %s: v=%#v err=%v", tc.op, v, e)
			}
		})
	}
}

func TestBuiltinsPredicates_All(t *testing.T) {
	reg := NewRegistryWithBuiltins()
	ctx := &Context{}

	tests := []struct {
		name string
		op   string
		in   any
		args []any
		want bool
	}{
		{name: "exists_true", op: "exists", in: 1, want: true},
		{name: "exists_false", op: "exists", in: nil, want: false},
		{name: "eq_true", op: "eq", in: "12", args: []any{12}, want: true},
		{name: "contains_true", op: "contains", in: "hello world", args: []any{"world"}, want: true},
		{name: "gt_true", op: "gt", in: 10, args: []any{3}, want: true},
		{name: "lt_true", op: "lt", in: 2, args: []any{3}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := reg.getPredicate(tc.op)
			if err != nil {
				t.Fatalf("get predicate failed: %v", err)
			}
			got, e := entry.fn(ctx, tc.in, tc.args)
			if e != nil {
				t.Fatalf("predicate returned err: %v", e)
			}
			if got != tc.want {
				t.Fatalf("predicate mismatch: want=%v got=%v", tc.want, got)
			}
		})
	}
}
