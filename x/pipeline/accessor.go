package pipeline

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type pathSegment struct {
	field   string
	index   int
	isIndex bool
}

type Accessor struct {
	raw      string
	segments []pathSegment
}

func CompileAccessor(path string) (*Accessor, error) {
	if path == "" {
		return nil, &Error{Kind: ErrKindCompile, Op: "compile_accessor", Path: path, Message: "empty path"}
	}
	segments := make([]pathSegment, 0, 8)
	for _, token := range strings.Split(path, ".") {
		if token == "" {
			continue
		}
		for len(token) > 0 {
			lb := strings.IndexByte(token, '[')
			if lb == -1 {
				if idx, err := strconv.Atoi(token); err == nil {
					segments = append(segments, pathSegment{isIndex: true, index: idx})
				} else {
					segments = append(segments, pathSegment{field: token})
				}
				break
			}
			if lb > 0 {
				segments = append(segments, pathSegment{field: token[:lb]})
			}
			rb := strings.IndexByte(token[lb:], ']')
			if rb == -1 {
				return nil, &Error{Kind: ErrKindCompile, Op: "compile_accessor", Path: path, Message: "missing ]"}
			}
			rawIdx := token[lb+1 : lb+rb]
			idx, err := strconv.Atoi(rawIdx)
			if err != nil {
				return nil, &Error{Kind: ErrKindCompile, Op: "compile_accessor", Path: path, Message: "invalid index"}
			}
			segments = append(segments, pathSegment{isIndex: true, index: idx})
			token = token[lb+rb+1:]
		}
	}
	if len(segments) == 0 {
		return nil, &Error{Kind: ErrKindCompile, Op: "compile_accessor", Path: path, Message: "no path segment"}
	}
	return &Accessor{raw: path, segments: segments}, nil
}

func (a *Accessor) Get(root any) (any, error) {
	cur := reflect.ValueOf(root)
	for _, seg := range a.segments {
		for cur.IsValid() && (cur.Kind() == reflect.Ptr || cur.Kind() == reflect.Interface) {
			if cur.Kind() == reflect.Interface {
				if cur.IsNil() {
					return nil, nil
				}
				cur = cur.Elem()
				continue
			}
			if cur.IsNil() {
				return nil, &Error{Kind: ErrKindRuntime, Op: "get", Path: a.raw, Message: "nil pointer"}
			}
			cur = cur.Elem()
		}
		if !cur.IsValid() {
			return nil, nil
		}
		if seg.isIndex {
			if cur.Kind() != reflect.Slice && cur.Kind() != reflect.Array {
				return nil, &Error{Kind: ErrKindRuntime, Op: "get", Path: a.raw, InputType: cur.Kind().String(), Message: "not indexable"}
			}
			if seg.index < 0 || seg.index >= cur.Len() {
				return nil, nil
			}
			cur = cur.Index(seg.index)
			continue
		}
		switch cur.Kind() {
		case reflect.Map:
			mv := cur.MapIndex(reflect.ValueOf(seg.field))
			if !mv.IsValid() {
				return nil, nil
			}
			cur = mv
		case reflect.Struct:
			cur = cur.FieldByName(seg.field)
			if !cur.IsValid() {
				return nil, nil
			}
		default:
			return nil, &Error{Kind: ErrKindRuntime, Op: "get", Path: a.raw, InputType: cur.Kind().String(), Message: "not traversable"}
		}
	}
	if !cur.IsValid() {
		return nil, nil
	}
	return cur.Interface(), nil
}

func (a *Accessor) Set(root any, value any) error {
	if len(a.segments) == 0 {
		return &Error{Kind: ErrKindRuntime, Op: "set", Path: a.raw, Message: "empty accessor"}
	}
	cur := reflect.ValueOf(root)
	if cur.Kind() != reflect.Ptr || cur.IsNil() {
		return &Error{Kind: ErrKindRuntime, Op: "set", Path: a.raw, Message: "root must be non-nil pointer"}
	}
	cur = cur.Elem()

	for i, seg := range a.segments {
		isLast := i == len(a.segments)-1
		for cur.IsValid() && (cur.Kind() == reflect.Ptr || cur.Kind() == reflect.Interface) {
			if cur.Kind() == reflect.Interface {
				if cur.IsNil() {
					return &Error{Kind: ErrKindRuntime, Op: "set", Path: a.raw, Message: "nil interface target"}
				}
				cur = cur.Elem()
				continue
			}
			if cur.IsNil() {
				cur.Set(reflect.New(cur.Type().Elem()))
			}
			cur = cur.Elem()
		}
		if seg.isIndex {
			if cur.Kind() != reflect.Slice {
				return &Error{Kind: ErrKindRuntime, Op: "set", Path: a.raw, InputType: cur.Kind().String(), Message: "not slice"}
			}
			if seg.index < 0 || seg.index >= cur.Len() {
				return &Error{Kind: ErrKindRuntime, Op: "set", Path: a.raw, Message: "index out of range"}
			}
			if isLast {
				return assignValue(cur.Index(seg.index), value, a.raw)
			}
			cur = cur.Index(seg.index)
			continue
		}

		switch cur.Kind() {
		case reflect.Map:
			if cur.IsNil() {
				cur.Set(reflect.MakeMap(cur.Type()))
			}
			key := reflect.ValueOf(seg.field)
			if isLast {
				cur.SetMapIndex(key, reflect.ValueOf(value))
				return nil
			}
			next := cur.MapIndex(key)
			if !next.IsValid() || (next.Kind() == reflect.Map && next.IsNil()) {
				nm := map[string]any{}
				cur.SetMapIndex(key, reflect.ValueOf(nm))
				next = cur.MapIndex(key)
			}
			cur = next
		case reflect.Struct:
			field := cur.FieldByName(seg.field)
			if !field.IsValid() {
				return &Error{Kind: ErrKindRuntime, Op: "set", Path: a.raw, Message: fmt.Sprintf("field %s not found", seg.field)}
			}
			if isLast {
				return assignValue(field, value, a.raw)
			}
			cur = field
		default:
			return &Error{Kind: ErrKindRuntime, Op: "set", Path: a.raw, InputType: cur.Kind().String(), Message: "not traversable"}
		}
	}
	return nil
}

func assignValue(dst reflect.Value, src any, path string) error {
	for dst.Kind() == reflect.Ptr {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}
	if !dst.CanSet() {
		return &Error{Kind: ErrKindRuntime, Op: "set", Path: path, Message: "destination not settable"}
	}
	sv := reflect.ValueOf(src)
	if !sv.IsValid() {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	if sv.Type().AssignableTo(dst.Type()) {
		dst.Set(sv)
		return nil
	}
	if sv.Type().ConvertibleTo(dst.Type()) {
		dst.Set(sv.Convert(dst.Type()))
		return nil
	}
	return &Error{Kind: ErrKindRuntime, Op: "set", Path: path, Message: "type mismatch"}
}
