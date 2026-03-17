package filters

import (
	"encoding/json"
	"github.com/aynakeya/deepcolor/transform"
	"github.com/stretchr/testify/assert"
	"reflect"
	"regexp"
	"testing"
)

func TestRegExpFilter(t *testing.T) {
	title := "【喵萌奶茶屋】★07月新番★[莉可丽丝/Lycoris Recoil][05][1080p][简繁内封][招募翻译校对]"
	var include_filter = RegExp(
		regexp.MustCompile("Lycoris Recoil"),
		true,
	)
	assert.Equal(t, "RegExpFilter", include_filter.GetType())
	assert.True(t, include_filter.Check(title))
	exclude_filter := RegExp(
		regexp.MustCompile("Lycoris Recoil"),
		false,
	)
	assert.False(t, exclude_filter.Check(title))
	filter3 := RegExp(
		regexp.MustCompile(`【喵萌奶茶屋】★07月新番★\[莉可丽丝/Lycoris Recoil]\[\d+]\[1080p]\[简繁内封]\[招募翻译校对]`),
		true,
	)
	assert.True(t, filter3.Check(title))
}

func TestRegExpFilter_Marshalling(t *testing.T) {
	var filter1 transform.Filter = RegExp(regexp.MustCompile("Lycoris Recoil [0-9]*"), true)
	data, err := json.MarshalIndent(filter1, "", "  ")
	assert.Nil(t, err)
	var regF RegExpFilter
	err = json.Unmarshal(data, &regF)
	assert.Nil(t, err)
	assert.Equal(t, regF.Expression.String(), filter1.(*RegExpFilter).Expression.String(), "unmarshlling field not match")
	assert.Equal(t, regF.Expression.FindString("b9834hgbsaLycoris Recoil 33H%$43"), "Lycoris Recoil 33", "unmarshlling field not match")
	transform.RegisterFilter(&RegExpFilter{})
	f1, err := transform.UnmarshalFilter([]byte(data))
	assert.Nilf(t, err, "unmarshlling using UnmarshalFilter failed %s", err)
	assert.Equal(t, reflect.TypeOf(f1).String(), reflect.TypeOf(filter1).String())
}
