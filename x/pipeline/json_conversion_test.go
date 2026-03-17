package pipeline

import (
	"encoding/json"
	"reflect"
	"testing"
)

func compileAndRunForTest(t *testing.T, plan Plan, input string) any {
	t.Helper()
	reg := NewRegistryWithBuiltins()
	if err := ValidateSchema(plan, reg); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	cp, err := CompilePipeline(plan, reg)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	out, err := cp.Run(DefaultContext(), input)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	return out
}

func mustJSONMap(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("unmarshal expected json failed: %v", err)
	}
	return out
}

func TestMigration_NewJsonSchemaJson(t *testing.T) {
	testdata := `{
		"a":{"b":"c"},
		"g":{"x":"value1","c":[{"x":1},{"x":2}]}
	}`
	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{
				{To: "asdf", Expr: "json:a.b"},
				{To: "values_a", Expr: "json:g.x"},
				{To: "value_0_a", Expr: "json:g.c.0.x"},
				{To: "value_1_a", Expr: "json:g.c.1.x"},
				{To: "value2_0", Expr: "json:g.c.0.x"},
				{To: "value2_1", Expr: "json:g.c.1.x"},
			}},
		},
		Output: ObjectNode{Fields: map[string]Node{
			"asdf": FieldNode{Path: "asdf"},
			"values": ObjectNode{Fields: map[string]Node{
				"a": FieldNode{Path: "values_a"},
				"value": ArrayNode{Items: []Node{
					ObjectNode{Fields: map[string]Node{"a": FieldNode{Path: "value_0_a"}}},
					ObjectNode{Fields: map[string]Node{"a": FieldNode{Path: "value_1_a"}}},
				}},
				"value2": ArrayNode{Items: []Node{
					FieldNode{Path: "value2_0"},
					FieldNode{Path: "value2_1"},
				}},
			}},
		}},
	}
	out := compileAndRunForTest(t, plan, testdata)
	if _, err := json.Marshal(out); err != nil {
		t.Fatalf("marshal output failed: %v", err)
	}
}

func TestMigration_JSONFormatConversion(t *testing.T) {
	format1JSON := `{
		"result": [
			{
				"title":"1-t",
				"artists":[{"name":"1-a-1-n","pic":"1-a-1-c"},{"name":"1-a-2-n","pic":"1-a-2-c"}],
				"urls":["1-u-1","1-u-2"],
				"album":"1-al",
				"album-cover":"1-alb"
			},
			{
				"title":"2-t",
				"artists":[{"name":"2-a-1-n","pic":"2-a-1-c"}],
				"urls":["2-u-1"],
				"album":"2-al",
				"album-cover":"2-alb"
			}
		]
	}`
	format2JSON := `{
		"songs":[
			{
				"title":"1-t",
				"album":{"name":"1-al","pic":"1-alb"},
				"artists":["1-a-1-n","1-a-2-n"],
				"artists-pic":["1-a-1-c","1-a-2-c"],
				"urls":["1-u-1","1-u-2"]
			},
			{
				"title":"2-t",
				"album":{"name":"2-al","pic":"2-alb"},
				"artists":["2-a-1-n"],
				"artists-pic":["2-a-1-c"],
				"urls":["2-u-1"]
			}
		]
	}`

	t.Run("Format1ToFormat2", func(t *testing.T) {
		plan := Plan{
			SchemaVersion: VersionV1,
			Steps: []Step{
				ExtractStep{Assignments: []ExtractAssignment{
					{To: "s0_title", Expr: "json:result.0.title"},
					{To: "s0_album_name", Expr: "json:result.0.album"},
					{To: "s0_album_pic", Expr: "json:result.0.album-cover"},
					{To: "s0_a0_name", Expr: "json:result.0.artists.0.name"},
					{To: "s0_a1_name", Expr: "json:result.0.artists.1.name"},
					{To: "s0_a0_pic", Expr: "json:result.0.artists.0.pic"},
					{To: "s0_a1_pic", Expr: "json:result.0.artists.1.pic"},
					{To: "s0_u0", Expr: "json:result.0.urls.0"},
					{To: "s0_u1", Expr: "json:result.0.urls.1"},

					{To: "s1_title", Expr: "json:result.1.title"},
					{To: "s1_album_name", Expr: "json:result.1.album"},
					{To: "s1_album_pic", Expr: "json:result.1.album-cover"},
					{To: "s1_a0_name", Expr: "json:result.1.artists.0.name"},
					{To: "s1_a0_pic", Expr: "json:result.1.artists.0.pic"},
					{To: "s1_u0", Expr: "json:result.1.urls.0"},
				}},
			},
			Output: ObjectNode{Fields: map[string]Node{
				"songs": ArrayNode{Items: []Node{
					ObjectNode{Fields: map[string]Node{
						"title": FieldNode{Path: "s0_title"},
						"album": ObjectNode{Fields: map[string]Node{
							"name": FieldNode{Path: "s0_album_name"},
							"pic":  FieldNode{Path: "s0_album_pic"},
						}},
						"artists": ArrayNode{Items: []Node{
							FieldNode{Path: "s0_a0_name"},
							FieldNode{Path: "s0_a1_name"},
						}},
						"artists-pic": ArrayNode{Items: []Node{
							FieldNode{Path: "s0_a0_pic"},
							FieldNode{Path: "s0_a1_pic"},
						}},
						"urls": ArrayNode{Items: []Node{
							FieldNode{Path: "s0_u0"},
							FieldNode{Path: "s0_u1"},
						}},
					}},
					ObjectNode{Fields: map[string]Node{
						"title": FieldNode{Path: "s1_title"},
						"album": ObjectNode{Fields: map[string]Node{
							"name": FieldNode{Path: "s1_album_name"},
							"pic":  FieldNode{Path: "s1_album_pic"},
						}},
						"artists": ArrayNode{Items: []Node{
							FieldNode{Path: "s1_a0_name"},
						}},
						"artists-pic": ArrayNode{Items: []Node{
							FieldNode{Path: "s1_a0_pic"},
						}},
						"urls": ArrayNode{Items: []Node{
							FieldNode{Path: "s1_u0"},
						}},
					}},
				}},
			}},
		}
		out := compileAndRunForTest(t, plan, format1JSON)
		want := mustJSONMap(t, format2JSON)
		if !reflect.DeepEqual(want, out) {
			t.Fatalf("format1->format2 mismatch\nwant=%v\ngot=%v", want, out)
		}
	})

	t.Run("Format2ToFormat1", func(t *testing.T) {
		plan := Plan{
			SchemaVersion: VersionV1,
			Steps: []Step{
				ExtractStep{Assignments: []ExtractAssignment{
					{To: "r0_title", Expr: "json:songs.0.title"},
					{To: "r0_album", Expr: "json:songs.0.album.name"},
					{To: "r0_album_cover", Expr: "json:songs.0.album.pic"},
					{To: "r0_a0_name", Expr: "json:songs.0.artists.0"},
					{To: "r0_a1_name", Expr: "json:songs.0.artists.1"},
					{To: "r0_a0_pic", Expr: "json:songs.0.artists-pic.0"},
					{To: "r0_a1_pic", Expr: "json:songs.0.artists-pic.1"},
					{To: "r0_u0", Expr: "json:songs.0.urls.0"},
					{To: "r0_u1", Expr: "json:songs.0.urls.1"},

					{To: "r1_title", Expr: "json:songs.1.title"},
					{To: "r1_album", Expr: "json:songs.1.album.name"},
					{To: "r1_album_cover", Expr: "json:songs.1.album.pic"},
					{To: "r1_a0_name", Expr: "json:songs.1.artists.0"},
					{To: "r1_a0_pic", Expr: "json:songs.1.artists-pic.0"},
					{To: "r1_u0", Expr: "json:songs.1.urls.0"},
				}},
			},
			Output: ObjectNode{Fields: map[string]Node{
				"result": ArrayNode{Items: []Node{
					ObjectNode{Fields: map[string]Node{
						"title": FieldNode{Path: "r0_title"},
						"artists": ArrayNode{Items: []Node{
							ObjectNode{Fields: map[string]Node{
								"name": FieldNode{Path: "r0_a0_name"},
								"pic":  FieldNode{Path: "r0_a0_pic"},
							}},
							ObjectNode{Fields: map[string]Node{
								"name": FieldNode{Path: "r0_a1_name"},
								"pic":  FieldNode{Path: "r0_a1_pic"},
							}},
						}},
						"urls": ArrayNode{Items: []Node{
							FieldNode{Path: "r0_u0"},
							FieldNode{Path: "r0_u1"},
						}},
						"album":       FieldNode{Path: "r0_album"},
						"album-cover": FieldNode{Path: "r0_album_cover"},
					}},
					ObjectNode{Fields: map[string]Node{
						"title": FieldNode{Path: "r1_title"},
						"artists": ArrayNode{Items: []Node{
							ObjectNode{Fields: map[string]Node{
								"name": FieldNode{Path: "r1_a0_name"},
								"pic":  FieldNode{Path: "r1_a0_pic"},
							}},
						}},
						"urls": ArrayNode{Items: []Node{
							FieldNode{Path: "r1_u0"},
						}},
						"album":       FieldNode{Path: "r1_album"},
						"album-cover": FieldNode{Path: "r1_album_cover"},
					}},
				}},
			}},
		}
		out := compileAndRunForTest(t, plan, format2JSON)
		want := mustJSONMap(t, format1JSON)
		if !reflect.DeepEqual(want, out) {
			t.Fatalf("format2->format1 mismatch\nwant=%v\ngot=%v", want, out)
		}
	})
}

func TestMigration_SimpleArrayConversion(t *testing.T) {
	simpleJSON := `{
		"result": [
			{"title": "Song 1", "album": "Album 1"},
			{"title": "Song 2", "album": "Album 2"}
		]
	}`
	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{
				{To: "s0_name", Expr: "json:result.0.title"},
				{To: "s0_album", Expr: "json:result.0.album"},
				{To: "s1_name", Expr: "json:result.1.title"},
				{To: "s1_album", Expr: "json:result.1.album"},
			}},
		},
		Output: ObjectNode{Fields: map[string]Node{
			"songs": ArrayNode{Items: []Node{
				ObjectNode{Fields: map[string]Node{
					"name":       FieldNode{Path: "s0_name"},
					"album_name": FieldNode{Path: "s0_album"},
				}},
				ObjectNode{Fields: map[string]Node{
					"name":       FieldNode{Path: "s1_name"},
					"album_name": FieldNode{Path: "s1_album"},
				}},
			}},
		}},
	}
	out := compileAndRunForTest(t, plan, simpleJSON)
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("output is not map: %T", out)
	}
	songs, ok := m["songs"].([]any)
	if !ok || len(songs) != 2 {
		t.Fatalf("songs invalid: %#v", m["songs"])
	}
	song1 := songs[0].(map[string]any)
	song2 := songs[1].(map[string]any)
	if song1["name"] != "Song 1" || song1["album_name"] != "Album 1" {
		t.Fatalf("song1 mismatch: %#v", song1)
	}
	if song2["name"] != "Song 2" || song2["album_name"] != "Album 2" {
		t.Fatalf("song2 mismatch: %#v", song2)
	}
}

func TestMigration_DynamicArrayExpansionWithArrayMapNode(t *testing.T) {
	format1JSON := `{
		"result":[
			{
				"title":"1-t",
				"artists":[{"name":"1-a-1-n","pic":"1-a-1-c"},{"name":"1-a-2-n","pic":"1-a-2-c"}],
				"urls":["1-u-1","1-u-2"],
				"album":"1-al",
				"album-cover":"1-alb"
			},
			{
				"title":"2-t",
				"artists":[{"name":"2-a-1-n","pic":"2-a-1-c"}],
				"urls":["2-u-1"],
				"album":"2-al",
				"album-cover":"2-alb"
			},
			{
				"title":"3-t",
				"artists":[{"name":"3-a-1-n","pic":"3-a-1-c"},{"name":"3-a-2-n","pic":"3-a-2-c"},{"name":"3-a-3-n","pic":"3-a-3-c"}],
				"urls":["3-u-1","3-u-2","3-u-3"],
				"album":"3-al",
				"album-cover":"3-alb"
			}
		]
	}`
	plan := Plan{
		SchemaVersion: VersionV1,
		Steps: []Step{
			ExtractStep{Assignments: []ExtractAssignment{
				{To: "result", Expr: "json:result"},
			}},
		},
		Output: ObjectNode{Fields: map[string]Node{
			"songs": ArrayMapNode{
				From: "result",
				Item: ObjectNode{Fields: map[string]Node{
					"title": FieldNode{Path: "$item.title"},
					"album": ObjectNode{Fields: map[string]Node{
						"name": FieldNode{Path: "$item.album"},
						"pic":  FieldNode{Path: "$item.album-cover"},
					}},
					"artists": ArrayMapNode{
						From: "$item.artists",
						Item: FieldNode{Path: "$item.name"},
					},
					"artists-pic": ArrayMapNode{
						From: "$item.artists",
						Item: FieldNode{Path: "$item.pic"},
					},
					"urls": FieldNode{Path: "$item.urls"},
				}},
			},
		}},
	}

	out := compileAndRunForTest(t, plan, format1JSON)
	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("output is not map: %T", out)
	}
	songs, ok := m["songs"].([]any)
	if !ok {
		t.Fatalf("songs is not array: %T", m["songs"])
	}
	if len(songs) != 3 {
		t.Fatalf("expected 3 songs, got %d", len(songs))
	}
	song3 := songs[2].(map[string]any)
	if song3["title"] != "3-t" {
		t.Fatalf("unexpected song3 title: %v", song3["title"])
	}
	artists3 := song3["artists"].([]any)
	if !reflect.DeepEqual(artists3, []any{"3-a-1-n", "3-a-2-n", "3-a-3-n"}) {
		t.Fatalf("unexpected song3 artists: %#v", artists3)
	}
}
