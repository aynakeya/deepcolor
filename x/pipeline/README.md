# x/pipeline

`x/pipeline` is the unified advanced processing engine in Deepcolor v2.

It provides a typed, compiled pipeline model:

- `Extract -> Map -> Filter -> Project`
- instance-based `Registry` (no global registry)
- compile-time checks (`ValidateSchema`, `CompilePipeline`)
- runtime execution policy (`Executor`)

## Quick Start

```go
reg := pipeline.NewRegistryWithBuiltins()

plan := pipeline.Plan{
	SchemaVersion: pipeline.VersionV1,
	Source:        pipeline.SourceSpec{Name: "demo"},
	Steps: []pipeline.Step{
		pipeline.ExtractStep{Assignments: []pipeline.ExtractAssignment{
			{To: "title", Expr: "json:data.title"},
			{To: "artist", Expr: "json:data.artist"},
		}},
		pipeline.MapStep{Transforms: []pipeline.MapTransform{
			{Path: "title", Ops: []pipeline.OpSpec{{Name: "trim"}}},
			{Path: "artist", Ops: []pipeline.OpSpec{{Name: "lower"}}},
		}},
		pipeline.FilterStep{Mode: pipeline.FilterAll, Conditions: []pipeline.Condition{
			{Path: "title", Op: pipeline.OpSpec{Name: "exists"}},
		}},
	},
	Output: pipeline.ObjectNode{Fields: map[string]pipeline.Node{
		"name":   pipeline.FieldNode{Path: "title"},
		"artist": pipeline.FieldNode{Path: "artist"},
	}},
}

if err := pipeline.ValidateSchema(plan, reg); err != nil {
	panic(err)
}
cp, err := pipeline.CompilePipeline(plan, reg)
if err != nil {
	panic(err)
}
out, err := cp.Run(pipeline.DefaultContext(), []byte(`{"data":{"title":"  Song ","artist":"REOL"}}`))
_ = out
_ = err
```

## Plan Model

- `Plan.SchemaVersion`: currently `v1`
- `Plan.Steps`: execution steps
- `Plan.Output`: final output node tree

### Step Types

- `ExtractStep`: extract values to pipeline document fields
- `MapStep`: transform document fields with mapper ops
- `FilterStep`: keep/drop by predicate conditions
- `ProjectStep`: project output (optional if `Plan.Output` is set)

## Node Types

- `FieldNode{Path: "a.b"}`
- `ValueNode{Value: ...}`
- `ObjectNode{Fields: ...}`
- `ArrayNode{Items: ...}`
- `ArrayMapNode{From: "result", Item: ...}` for unknown-length expansion
- `OpNode{Input: ..., Op: ...}`

`ArrayMapNode` supports scoped values in item nodes:

- `$item`
- `$item.xxx`
- `$index`

## Expression Format

`ExtractAssignment.Expr` uses `kind:arg`:

- `json:data.title`
- `regex:ep([0-9]+)=>1`
- `xml://book[1]/title`
- `xml_all://book/title`
- `input:user.name`
- `doc:meta.id`

## Builtins

### Extractors

- `input`
- `doc`
- `json`
- `regex`
- `regex_all`
- `xml`
- `xml_all`

### Mappers

- `cast_string`, `cast_int`
- `lower`, `upper`, `trim`, `format`
- `regexp_replace`, `regex_find`
- `default`, `split`, `join`
- `switch_first`

### Predicates

- `exists`, `eq`, `contains`, `gt`, `lt`

## Registry Extension

```go
reg := pipeline.NewRegistryWithBuiltins()
reg.RegisterMapper("num_to_int", pipeline.EffectPure, func(ctx *pipeline.Context, in any, args []any) (any, error) {
	// custom transform
	return in, nil
})
```

## Executor Policy

```go
cp.WithExecutor(pipeline.Executor{
	ErrorMode:        pipeline.ErrorModeFailFast, // or continue / collect
	AllowSideEffects: false,
})
```

## Error Model

All pipeline errors use `pipeline.Error` with structured fields:

- `Kind`: validate / compile / runtime / registry
- `Path`, `Op`, `InputType`, `Message`, `Cause`
