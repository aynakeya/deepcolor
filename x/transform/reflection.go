package transform

import (
	"reflect"
	"strings"
)

// todo https://stackoverflow.com/questions/47187680/how-do-i-change-fields-a-slice-of-structs-using-reflect

type Field string

type Value struct {
	reflect.Value
	parent reflect.Value
	field  string
}

func (t Field) GetValue(inst interface{}) Value {
	v := reflect.ValueOf(inst)
	var parent reflect.Value
	var field string
	for _, f := range strings.Split(string(t), ".") {
		parent = v
		field = f
		switch v.Kind() {
		case reflect.Ptr:
			v = v.Elem().FieldByName(f)
		case reflect.Map:
			v = v.MapIndex(reflect.ValueOf(f))
		case reflect.Struct:
			v = v.FieldByName(f)
		default:
			panic("not supported kind" + v.Kind().String())
		}
	}
	return Value{v, parent, field}
}

func (t Field) GetValueE(inst interface{}) (value Value, ok bool) {
	defer func() {
		if p := recover(); p != nil {
			value = Value{}
			ok = false
		}
	}()
	return t.GetValue(value), true
}

func SetFieldValue(src interface{}, dst Value) {
	if dst.Kind() == reflect.Ptr {
		dst.Value = dst.Elem()
	}
	if dst.parent.Kind() == reflect.Map {
		dst.parent.SetMapIndex(reflect.ValueOf(dst.field), reflect.ValueOf(src))
		return
	}
	if dst.Kind() != reflect.Slice {
		dst.Set(reflect.ValueOf(src))
		return
	}
	asrc, ok := src.([]interface{})
	if !ok {
		return
	}
	s := reflect.New(dst.Type()).Elem()
	for i := 0; i < len(asrc); i++ {
		s = reflect.Append(s, reflect.ValueOf(asrc[i]))
	}
	dst.Set(s)
}
