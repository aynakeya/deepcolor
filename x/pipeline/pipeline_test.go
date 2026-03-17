package pipeline

import "testing"

func TestPipelineCompileAndRunHappyPath(t *testing.T) {
	reg := NewRegistryWithBuiltins()
	plan := Plan{
		SchemaVersion: VersionV1,
		Source:        SourceSpec{Name: "unit"},
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{
				{To: "title", Expr: "json:data.title"},
				{To: "artist", Expr: "json:data.artist"},
			}},
			MapStep{Transforms: []MapTransform{
				{Path: "title", Ops: []OpSpec{{Name: "trim"}}},
				{Path: "artist", Ops: []OpSpec{{Name: "lower"}}},
			}},
			FilterStep{
				Mode: FilterAll,
				Conditions: []Condition{
					{Path: "title", Op: OpSpec{Name: "exists"}},
				},
			},
		},
		Output: ObjectNode{Fields: map[string]Node{
			"name":   FieldNode{Path: "title"},
			"artist": FieldNode{Path: "artist"},
		}},
	}

	cp, err := CompilePipeline(plan, reg)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	out, err := cp.Run(DefaultContext(), []byte(`{"data":{"title":"  Song  ","artist":"REOL"}}`))
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("output is not map: %T", out)
	}
	if m["name"] != "Song" {
		t.Fatalf("unexpected name: %v", m["name"])
	}
	if m["artist"] != "reol" {
		t.Fatalf("unexpected artist: %v", m["artist"])
	}
}

func TestValidateSchemaUnknownMapper(t *testing.T) {
	reg := NewRegistryWithBuiltins()
	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			MapStep{Transforms: []MapTransform{
				{Path: "x", Ops: []OpSpec{{Name: "not_exists"}}},
			}},
		},
		Output: FieldNode{Path: "x"},
	}
	err := ValidateSchema(plan, reg)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestErrorModeContinue(t *testing.T) {
	reg := NewRegistryWithBuiltins()
	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{
				{To: "title", Expr: "json:data.title"},
				{To: "artist", Expr: "json:data.artist"},
			}},
			MapStep{Transforms: []MapTransform{
				{Path: "title", Ops: []OpSpec{{Name: "regexp_replace", Args: []any{"(", "x"}}}},
				{Path: "artist", Ops: []OpSpec{{Name: "lower"}}},
			}},
		},
		Output: ObjectNode{Fields: map[string]Node{
			"title":  FieldNode{Path: "title"},
			"artist": FieldNode{Path: "artist"},
		}},
	}
	cp, err := CompilePipeline(plan, reg)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	cp.WithExecutor(Executor{ErrorMode: ErrorModeContinue, AllowSideEffects: false})
	out, err := cp.Run(DefaultContext(), []byte(`{"data":{"title":"Song","artist":"REOL"}}`))
	if err != nil {
		t.Fatalf("run failed under continue mode: %v", err)
	}
	m := out.(map[string]any)
	if m["artist"] != "reol" {
		t.Fatalf("continue mode should keep subsequent transforms, got: %v", m["artist"])
	}
}

func TestSideEffectPolicy(t *testing.T) {
	reg := NewRegistryWithBuiltins()
	reg.RegisterMapper("effect_mark", EffectSideEffect, func(_ *Context, in any, _ []any) (any, error) {
		return "marked:" + in.(string), nil
	})

	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{{To: "title", Expr: "json:data.title"}}},
			MapStep{Transforms: []MapTransform{
				{Path: "title", Ops: []OpSpec{{Name: "effect_mark"}}},
			}},
		},
		Output: FieldNode{Path: "title"},
	}
	cp, err := CompilePipeline(plan, reg)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	_, err = cp.Run(DefaultContext(), []byte(`{"data":{"title":"song"}}`))
	if err == nil {
		t.Fatalf("expected side effect disabled error")
	}

	cp.WithExecutor(Executor{ErrorMode: ErrorModeFailFast, AllowSideEffects: true})
	out, err := cp.Run(DefaultContext(), []byte(`{"data":{"title":"song"}}`))
	if err != nil {
		t.Fatalf("run failed with side effects allowed: %v", err)
	}
	if out.(string) != "marked:song" {
		t.Fatalf("unexpected side effect output: %v", out)
	}
}

func TestRegistryIsolation(t *testing.T) {
	reg1 := NewRegistryWithBuiltins()
	reg2 := NewRegistry()

	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			MapStep{Transforms: []MapTransform{
				{Path: "x", Ops: []OpSpec{{Name: "lower"}}},
			}},
		},
		Output: FieldNode{Path: "x"},
	}

	if err := ValidateSchema(plan, reg1); err != nil {
		t.Fatalf("reg1 should pass: %v", err)
	}
	if err := ValidateSchema(plan, reg2); err == nil {
		t.Fatalf("reg2 should fail due to missing builtins")
	}
}
