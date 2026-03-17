package formatter

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewJsonSchemaJson(t *testing.T) {
	scheme := map[string]interface{}{
		"asdf": "json::a.b",
		"values": map[string]interface{}{
			"a": "json::g.x",
			"value": []interface{}{
				map[string]interface{}{
					"a": "json::g.c.#.x",
				},
			},
			"value2": []interface{}{"json::g.c.#.x"},
		},
	}
	testdata, _ := json.Marshal(map[string]interface{}{
		"a": map[string]interface{}{
			"b": "c",
		},
		"g": map[string]interface{}{
			"x": "value1",
			"c": []interface{}{
				map[string]int{
					"x": 1,
				},
				map[string]int{
					"x": 2,
				},
			},
		},
		"hello":     "world1world2",
		"otherdata": "f1f2f3f4",
	})
	schema := NewJsonSchema(scheme)
	require.NotNil(t, schema)
	value := schema.Format(string(testdata))
	_, err := json.Marshal(value)
	//require.Equal(t, `{"asdf":"c","data":"world1","data2":"f1","values":{"a":"value1","value":[{"a":1},{"a":2}],"value2":[1,2],"value3":["f1","f2","f3","f4"]}}`, string(v))
	require.NoError(t, err)
}

func TestNewJsonSchemaRegex(t *testing.T) {
	scheme := map[string]interface{}{
		"data":  "regex::world.",
		"data2": "regex::f.",
		"regex": map[string]interface{}{
			"regexfirst": "regex::f.",
			"regexarray": []interface{}{"regex::f."},
			"regexpmapinarray": []interface{}{
				map[string]interface{}{
					"a": "regex::f.",
					"d": "regex::f.",
					"c": "regex::world.",
					"e": map[string]interface{}{
						"x": "regex::f.",
					},
				},
			},
		},
	}
	testdata := "world1world2f1f2f3uniquef4"
	schema := NewJsonSchema(scheme)
	require.NotNil(t, schema)
	value := schema.Format(testdata)
	_, err := json.Marshal(value)
	require.NoError(t, err)
}

func TestJSONFormatConversion(t *testing.T) {
	// 测试两种JSON格式之间的相互转换

	// 第一种格式：嵌套artists对象数组
	format1JSON := `{
    "result": [
        {
            "title":"1-t",
            "artists": [
                {
                    "name":"1-a-1-n",
                    "pic":"1-a-1-c"
                },
                {
                    "name":"1-a-2-n",
                    "pic":"1-a-2-c"
                }
            ],
            "urls": [
                "1-u-1",
                "1-u-2"
            ],
            "album":"1-al",
            "album-cover":"1-alb"
        },
        {
            "title":"2-t",
            "artists": [
                {
                    "name":"2-a-1-n",
                    "pic":"2-a-1-c"
                }
            ],
            "urls": [
                "2-u-1"
            ],
            "album":"2-al",
            "album-cover":"2-alb"
        }
    ]
}`

	// 第二种格式：扁平化artists数组（保留用于文档说明）
	format2Json := `{
    "songs": [
        {
            "title":"1-t",
            "album": {
                "name":"1-al",
                "pic":"1-alb"
            },
            "artists": ["1-a-1-n","1-a-2-n"],
            "artists-pic": ["1-a-1-c","1-a-2-c"],
            "urls": [
                "1-u-1",
                "1-u-2"
            ]
        },
        {
            "title":"2-t",
            "album": {
                "name":"2-al",
                "pic":"2-alb"
            },
            "artists": ["2-a-1-n"],
            "artists-pic": ["2-a-1-c"],
            "urls": [
                "2-u-1"
            ]
        }
    ]
}`
	var format1Struct map[string]interface{}
	var format2Struct map[string]interface{}
	err := json.Unmarshal([]byte(format1JSON), &format1Struct)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(format2Json), &format2Struct)
	require.NoError(t, err)

	// 测试 Format1 -> Format2 的转换
	t.Run("Format1ToFormat2", func(t *testing.T) {
		// 定义转换schema：从format1转换到format2
		format1ToFormat2Schema := map[string]interface{}{
			"songs": []interface{}{
				map[string]interface{}{
					"title": "json::result.#.title",
					"album": map[string]interface{}{
						"name": "json::result.#.album",
						"pic":  "json::result.#.album-cover",
					},
					"artists":     []interface{}{"json::result.#.artists.#.name"},
					"artists-pic": []interface{}{"json::result.#.artists.#.pic"},
					"urls":        []interface{}{"json::result.#.urls"},
				},
			},
		}

		schema := NewJsonSchema(format1ToFormat2Schema)
		require.NotNil(t, schema)

		result := schema.Format(format1JSON)
		require.NotNil(t, result)

		require.Equal(t, format2Struct, result)

		t.Log("Format1 -> Format2 conversion result:")

		// 验证转换结果
		songs := result["songs"].([]interface{})
		require.Len(t, songs, 2)

		// 验证第一首歌
		song1 := songs[0].(map[string]interface{})
		require.Equal(t, "1-t", song1["title"])

		album1 := song1["album"].(map[string]interface{})
		require.Equal(t, "1-al", album1["name"])
		require.Equal(t, "1-alb", album1["pic"])

		artists1 := song1["artists"].([]interface{})
		require.Contains(t, artists1, "1-a-1-n")
		require.Contains(t, artists1, "1-a-2-n")

		artistsPic1 := song1["artists-pic"].([]interface{})
		require.Contains(t, artistsPic1, "1-a-1-c")
		require.Contains(t, artistsPic1, "1-a-2-c")

		urls1 := song1["urls"].([]interface{})
		require.Contains(t, urls1, "1-u-1")
		require.Contains(t, urls1, "1-u-2")

		// 验证第二首歌
		song2 := songs[1].(map[string]interface{})
		require.Equal(t, "2-t", song2["title"])

		album2 := song2["album"].(map[string]interface{})
		require.Equal(t, "2-al", album2["name"])
		require.Equal(t, "2-alb", album2["pic"])

		artists2 := song2["artists"].([]interface{})
		require.Contains(t, artists2, "2-a-1-n")

		urls2 := song2["urls"].([]interface{})
		require.Contains(t, urls2, "2-u-1")
	})

	t.Run("Format2ToFormat1", func(t *testing.T) {
		format2ToFormat1Schema := map[string]interface{}{
			"result": []interface{}{
				map[string]interface{}{
					"title": "json::songs.#.title",
					"artists": []interface{}{
						map[string]interface{}{
							"name": "json::songs.#.artists.#",
							"pic":  "json::songs.#.artists-pic.#",
						},
					},
					"urls":        []interface{}{"json::songs.#.urls"},
					"album":       "json::songs.#.album.name",
					"album-cover": "json::songs.#.album.pic",
				},
			},
		}

		schema := NewJsonSchema(format2ToFormat1Schema)
		require.NotNil(t, schema)

		result := schema.Format(format2Json)
		require.NotNil(t, result)
		//pp.Println(result)
		require.Equal(t, format1Struct, result)

	})
}

func TestSimpleArrayConversion(t *testing.T) {
	// 测试一个简单的数组转换案例
	simpleJSON := `{
		"result": [
			{"title": "Song 1", "album": "Album 1"},
			{"title": "Song 2", "album": "Album 2"}
		]
	}`

	simpleSchema := map[string]interface{}{
		"songs": []interface{}{
			map[string]interface{}{
				"name":       "json::result.#.title",
				"album_name": "json::result.#.album",
			},
		},
	}

	schema := NewJsonSchema(simpleSchema)
	require.NotNil(t, schema)

	result := schema.Format(simpleJSON)

	songs := result["songs"].([]interface{})
	t.Logf("Number of songs: %d", len(songs))
	require.Len(t, songs, 2)

	song1 := songs[0].(map[string]interface{})
	require.Equal(t, "Song 1", song1["name"])
	require.Equal(t, "Album 1", song1["album_name"])

	song2 := songs[1].(map[string]interface{})
	require.Equal(t, "Song 2", song2["name"])
	require.Equal(t, "Album 2", song2["album_name"])
}
