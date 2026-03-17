package pipeline

import (
	"testing"
)

func TestBuiltinXMLExtractorBasic(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<library>
	<book id="1">
		<title>Go Programming</title>
		<author>John Doe</author>
		<tags><tag>programming</tag><tag>golang</tag></tags>
	</book>
	<book id="2">
		<title>Advanced Go</title>
		<author>Jane Smith</author>
		<tags><tag>advanced</tag><tag>golang</tag></tags>
	</book>
</library>`

	reg := NewRegistryWithBuiltins()
	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{
				{To: "title1", Expr: "xml://book[1]/title"},
				{To: "author2", Expr: "xml://book[2]/author"},
				{To: "book_id", Expr: "xml://book[1]/@id"},
				{To: "all_titles", Expr: "xml_all://book/title"},
				{To: "all_tags", Expr: "xml_all://book/tags/tag"},
			}},
		},
		Output: ObjectNode{Fields: map[string]Node{
			"title1":     FieldNode{Path: "title1"},
			"author2":    FieldNode{Path: "author2"},
			"book_id":    FieldNode{Path: "book_id"},
			"all_titles": FieldNode{Path: "all_titles"},
			"all_tags":   FieldNode{Path: "all_tags"},
		}},
	}

	cp, err := CompilePipeline(plan, reg)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	out, err := cp.Run(DefaultContext(), xmlData)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	m := out.(map[string]any)
	if m["title1"] != "Go Programming" {
		t.Fatalf("unexpected title1: %v", m["title1"])
	}
	if m["author2"] != "Jane Smith" {
		t.Fatalf("unexpected author2: %v", m["author2"])
	}
	if m["book_id"] != "1" {
		t.Fatalf("unexpected book_id: %v", m["book_id"])
	}
	if titles, ok := m["all_titles"].([]any); !ok || len(titles) != 2 {
		t.Fatalf("unexpected all_titles: %#v", m["all_titles"])
	}
	if tags, ok := m["all_tags"].([]any); !ok || len(tags) != 4 {
		t.Fatalf("unexpected all_tags: %#v", m["all_tags"])
	}
}

func TestBuiltinXMLExtractorWithArrayMapIndex(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<bookstore>
	<book><title>Book 1</title></book>
	<book><title>Book 2</title></book>
	<book><title>Book 3</title></book>
</bookstore>`

	reg := NewRegistryWithBuiltins()
	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{
				{To: "titles_raw", Expr: "xml_all://book/title"},
			}},
		},
		Output: ObjectNode{Fields: map[string]Node{
			"titles": ArrayMapNode{
				From: "titles_raw",
				Item: FieldNode{Path: "$item"},
			},
		}},
	}

	cp, err := CompilePipeline(plan, reg)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	out, err := cp.Run(DefaultContext(), xmlData)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	m := out.(map[string]any)
	titles, ok := m["titles"].([]any)
	if !ok || len(titles) != 3 {
		t.Fatalf("unexpected titles: %#v", m["titles"])
	}
}

func TestBuiltinRegexExtractors(t *testing.T) {
	input := "world1world2f1f2f3uniquef4"
	reg := NewRegistryWithBuiltins()
	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{
				{To: "first_world", Expr: "regex:world.=>0"},
				{To: "first_f", Expr: "regex:f.=>0"},
				{To: "all_f", Expr: "regex_all:f.=>0"},
			}},
		},
		Output: ObjectNode{Fields: map[string]Node{
			"first_world": FieldNode{Path: "first_world"},
			"first_f":     FieldNode{Path: "first_f"},
			"all_f":       FieldNode{Path: "all_f"},
		}},
	}
	cp, err := CompilePipeline(plan, reg)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	out, err := cp.Run(DefaultContext(), input)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	m := out.(map[string]any)
	if m["first_world"] != "world1" {
		t.Fatalf("unexpected first_world: %v", m["first_world"])
	}
	if m["first_f"] != "f1" {
		t.Fatalf("unexpected first_f: %v", m["first_f"])
	}
	allF, ok := m["all_f"].([]any)
	if !ok || len(allF) != 4 {
		t.Fatalf("unexpected all_f: %#v", m["all_f"])
	}
}
