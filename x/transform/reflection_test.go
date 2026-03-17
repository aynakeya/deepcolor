package transform

import (
	"gotest.tools/assert"
	"reflect"
	"testing"
)

type A struct {
	B string
	Y map[string]interface{}
	X interface{}
}

type C struct {
	D *A
}

type E struct {
	F C
}

func TestField_GetValue(t *testing.T) {
	s := &E{
		F: C{
			D: &A{
				B: "final",
				Y: map[string]interface{}{
					"key": "asdff",
				},
				X: "x",
			},
		},
	}
	assert.Equal(t, "final", Field("F.D.B").GetValue(s).Interface())
	assert.Equal(t, "final", Field("F.D.B").GetValue(*s).Interface())
	assert.Equal(t, "asdff", Field("F.D.Y.key").GetValue(*s).Interface())
	Field("F.D.B").GetValue(s).Set(reflect.ValueOf("3333"))
	SetFieldValue("4444", Field("F.D.X").GetValue(s))
	SetFieldValue("4445", Field("F.D.Y.key").GetValue(s))
	assert.Equal(t, "3333", Field("F.D.B").GetValue(s).Interface())
	assert.Equal(t, "4444", s.F.D.X)
	assert.Equal(t, "4445", s.F.D.Y["key"])
	_, ok := Field("F.D.H.key").GetValueE(s)
	assert.Equal(t, false, ok)
}
