package translators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aynakeya/deepcolor/transform"
	"gotest.tools/assert"
	"regexp"
	"testing"
)

func marshalIndentUnescape(v interface{}, prefix, indent string) (string, error) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	err := jsonEncoder.Encode(v)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, bf.Bytes(), prefix, indent)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

var testString = "【喵萌奶茶屋】★07月新番★[莉可丽丝/Lycoris Recoil][05][1080p][简繁内封][招募翻译校对][MKV]"

func TestRegExpTranslator(t *testing.T) {
	trans := NewRegExpReplacer(
		regexp.MustCompile(`【喵萌奶茶屋】★07月新番★\[莉可丽丝/Lycoris Recoil]\[(.*)]\[1080p]\[简繁内封]\[招募翻译校对]\[MKV]`),
		"$1",
	)
	assert.Equal(t, "05", trans.MustApply(testString))
}

func TestRegExpTranslator_Marshalling(t *testing.T) {
	trans := NewRegExpReplacer(
		regexp.MustCompile(`【喵萌奶茶屋】★07月新番★\[莉可丽丝/Lycoris Recoil]\[(.*)]\[1080p]\[简繁内封]\[招募翻译校对].*`),
		"$1",
	)
	data, err := marshalIndentUnescape(trans, "", "  ")
	if err != nil {
		t.Fatalf("Marshlling failed")
	}
	var trans1 RegExpReplacer
	fmt.Println(data)
	err = json.Unmarshal([]byte(data), &trans1)
	if err != nil {
		t.Fatalf("Unmarshlling failed")
		return
	}
	assert.Equal(t, "05", trans1.MustApply(testString))
	trans2, err := transform.UnmarshalTranslator([]byte(data))
	if err != nil {
		t.Fatalf("fail to unmarshal using UnmarshalTranslator, %s", err)
	}
	assert.Equal(t, "05", trans2.MustApply(testString))
}
