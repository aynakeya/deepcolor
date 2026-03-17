package pipeline

import "testing"

type userInfo struct {
	Name string
	Meta map[string]any
}

type rootObj struct {
	User userInfo
}

func TestAccessorGetMapAndSlice(t *testing.T) {
	ac, err := CompileAccessor("a.b[1].c")
	if err != nil {
		t.Fatalf("compile accessor failed: %v", err)
	}
	in := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{"c": "x"},
				map[string]any{"c": "ok"},
			},
		},
	}
	v, err := ac.Get(in)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if v != "ok" {
		t.Fatalf("unexpected value: %v", v)
	}
}

func TestAccessorSetStructAndMap(t *testing.T) {
	acName, err := CompileAccessor("User.Name")
	if err != nil {
		t.Fatalf("compile accessor failed: %v", err)
	}
	acMeta, err := CompileAccessor("User.Meta.level")
	if err != nil {
		t.Fatalf("compile accessor failed: %v", err)
	}
	obj := rootObj{User: userInfo{Meta: map[string]any{}}}

	if err = acName.Set(&obj, "A"); err != nil {
		t.Fatalf("set struct field failed: %v", err)
	}
	if err = acMeta.Set(&obj, 3); err != nil {
		t.Fatalf("set map field failed: %v", err)
	}
	if obj.User.Name != "A" {
		t.Fatalf("unexpected name: %s", obj.User.Name)
	}
	if obj.User.Meta["level"] != 3 {
		t.Fatalf("unexpected meta level: %v", obj.User.Meta["level"])
	}
}

func TestAccessorCompileError(t *testing.T) {
	_, err := CompileAccessor("a[")
	if err == nil {
		t.Fatalf("expected compile error for invalid accessor")
	}
}
