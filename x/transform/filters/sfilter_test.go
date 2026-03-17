package filters

import (
	"encoding/json"
	"github.com/aynakeya/deepcolor/transform"
	"gotest.tools/assert"
	"regexp"
	"testing"
)

type A struct {
	B string
}

func TestStructFilter(t *testing.T) {
	a := A{B: "【喵萌奶茶屋】★07月新番★[莉可丽丝/Lycoris Recoil][05][1080p][简繁内封][招募翻译校对]"}
	filter := &StructFilter{
		Target: transform.Field("B"),
		Filter: RegExp(regexp.MustCompile("Lycoris Recoil"), true),
	}
	assert.Equal(t, true, filter.Check(a))
	filter1 := &StructFilter{
		Target: transform.Field("B"),
		Filter: RegExp(regexp.MustCompile("Lycoris Recoil"), false),
	}
	assert.Equal(t, false, filter1.Check(a))
	data, err := json.MarshalIndent(filter, "", "  ")
	assert.NilError(t, err)
	var filter2 StructFilter
	err = json.Unmarshal(data, &filter2)
	assert.NilError(t, err)
	assert.Equal(t, filter.Target, filter2.Target)
}
