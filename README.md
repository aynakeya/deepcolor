# Deepcolor v2

Deepcolor v2 is a lightweight HTTP workflow toolkit with two first-class styles:

- `Raw`: full manual control (`Client.Do(*Request)`)
- `Flow`: chain style with generic output (`Flow[T]`)

Template reuse is also unified into Flow through `Template[P,R]`.

## Install

```bash
go get github.com/aynakeya/deepcolor
```

## Quick Start

```go
dc := deepcolor.New(
	deepcolor.WithHeader(map[string]string{"User-Agent": "deepcolor-v2"}),
	deepcolor.WithTimeout(8),
)

resp, err := dc.GET("https://httpbin.org/get").
	Query(map[string]any{"q": "hello"}).
	JSON().
	Response(dc)
if err != nil {
	// handle
}
_ = resp
```

## Typed Flow

```go
type Info struct {
	Name string `json:"name"`
}

info, err := deepcolor.Typed[Info](
	dc.GET("https://example.com/api/info"),
).JSON().IntoJSON().Result(dc)
_ = info
_ = err
```

## Raw Style

```go
req := deepcolor.GET("https://example.com/api").
	SetQuery(map[string]any{"id": 1}).
	SetHeader(map[string]string{"X-Trace": "abc"})

resp, err := dc.Do(req)
if err != nil {
	// handle
}
```

## Reusable Flow Template

```go
infoFlow := deepcolor.Template(func(id string) *deepcolor.Flow[string] {
	return deepcolor.Typed[string](
		dc.GET("https://example.com/api/info"),
	).Query(map[string]any{"id": id}).
		Decode(func(resp *deepcolor.Response, out *string) error {
			*out = resp.JSON("data.name").String()
			return nil
		})
})

name, err := infoFlow.Call(dc, "1001")
_ = name
_ = err
```

## Bind Then Run

```go
flow := infoFlow.Bind("1001")
resp, err := flow.Response(dc)
_ = resp
_ = err
```

## Notes

- v2 is a breaking redesign and does not preserve v1 root-level APIs.
- Internal transport/runtime code is moved under `internal/` and is not public API.
