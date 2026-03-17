package translators

import (
	"encoding/json"
	"github.com/aynakeya/deepcolor/transform"
	"gotest.tools/assert"
	"regexp"
	"testing"
)

func TestPipeline_Marshalling(t *testing.T) {
	trans := NewRegExpReplacer(
		regexp.MustCompile(`【喵萌奶茶屋】★07月新番★\[莉可丽丝/Lycoris Recoil]\[(.*)]\[1080p]\[简繁内封]\[招募翻译校对]\[MKV]`),
		"$1",
	)
	pipe := Pipeline{
		Steps: []transform.Translator{trans, trans},
	}
	data, err := marshalIndentUnescape(pipe, "", "  ")
	if err != nil {
		t.Fatalf("Marshlling failed")
	}
	var pipe2 Pipeline
	err = json.Unmarshal([]byte(data), &pipe2)
	if err != nil {
		t.Fatalf("Unmarshlling failed %s", err)
		return
	}
	assert.Equal(t, "05", pipe2.Steps[0].MustApply(testString))
}

func TestPipeline(t *testing.T) {
	trans := NewRegExpReplacer(
		regexp.MustCompile(`【喵萌奶茶屋】★07月新番★\[莉可丽丝/Lycoris Recoil]\[(.*)]\[1080p]\[简繁内封]\[招募翻译校对]\[MKV]`),
		"$1",
	)
	trans1 := NewRegExpReplacer(
		regexp.MustCompile(`\d`),
		"${0}x",
	)
	pipe := Pipeline{
		Steps: []transform.Translator{trans, trans1},
	}
	assert.Equal(t, "0x5x", pipe.MustApply(testString))
}
