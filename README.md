# Deepcolor v2

Deepcolor v2 provides two layers:

- `deepcolor` root package: lightweight HTTP request/flow API
- `x/pipeline`: unified advanced data pipeline (`Extract -> Map -> Filter -> Project`)

## Install

```bash
go get github.com/aynakeya/deepcolor
```

## Root HTTP Flow

```go
dc := deepcolor.New(deepcolor.WithTimeout(8))

resp, err := dc.GET("https://httpbin.org/get").
	Query(map[string]any{"q": "hello"}).
	JSON().
	Response(dc)
_ = resp
_ = err
```

## Unified Pipeline (`x/pipeline`)

```go
reg := pipeline.NewRegistryWithBuiltins()

plan := pipeline.Plan{
	SchemaVersion: pipeline.VersionV1,
	Source:        pipeline.SourceSpec{Name: "example"},
	Steps: []pipeline.Step{
		pipeline.ExtractStep{Assignments: []pipeline.ExtractAssignment{
			{To: "title", Expr: "json:data.title"},
			{To: "artist", Expr: "json:data.artist"},
		}},
		pipeline.MapStep{Transforms: []pipeline.MapTransform{
			{Path: "title", Ops: []pipeline.OpSpec{{Name: "trim"}}},
			{Path: "artist", Ops: []pipeline.OpSpec{{Name: "lower"}}},
		}},
		pipeline.FilterStep{
			Mode: pipeline.FilterAll,
			Conditions: []pipeline.Condition{
				{Path: "title", Op: pipeline.OpSpec{Name: "exists"}},
			},
		},
	},
	Output: pipeline.ObjectNode{Fields: map[string]pipeline.Node{
		"name":   pipeline.FieldNode{Path: "title"},
		"artist": pipeline.FieldNode{Path: "artist"},
	}},
}

if err := pipeline.ValidateSchema(plan, reg); err != nil {
	panic(err)
}

compiled, err := pipeline.CompilePipeline(plan, reg)
if err != nil {
	panic(err)
}

out, err := compiled.WithExecutor(pipeline.Executor{
	ErrorMode:        pipeline.ErrorModeFailFast,
	AllowSideEffects: false,
}).Run(pipeline.DefaultContext(), []byte(`{"data":{"title":"  Song ","artist":"REOL"}}`))
_ = out
_ = err
```

## Design Guarantees in `x/pipeline`

- Typed schema nodes: `FieldNode` / `ArrayNode` / `ObjectNode` / `OpNode`
- No global registry: use instance `Registry`
- Unified error model with kind/path/op/input-type
- Path accessor is compiled once and reused at runtime
- Executor policy controls error strategy and side-effect allowance
- Compile-time APIs: `ValidateSchema` and `CompilePipeline`

## Breaking Changes

- Old `x/formatter` and `x/transform` are removed.
- Advanced transformations now live in `x/pipeline` only.
