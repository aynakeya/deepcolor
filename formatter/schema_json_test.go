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
	v, err := json.Marshal(value)
	//require.Equal(t, `{"asdf":"c","data":"world1","data2":"f1","values":{"a":"value1","value":[{"a":1},{"a":2}],"value2":[1,2],"value3":["f1","f2","f3","f4"]}}`, string(v))
	v, err = json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)
	t.Log(string(v))
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
	v, err := json.Marshal(value)
	v, err = json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)
	t.Log(string(v))
}

func TestComplexNestedSchema(t *testing.T) {
	// 复杂嵌套测试：同时使用JSON和Regex value getter
	scheme := map[string]interface{}{
		"metadata": map[string]interface{}{
			"title":       "json::meta.title",
			"version":     "json::meta.version",
			"tags":        []interface{}{"json::meta.tags.#"},
			"description": "regex::desc:\\s*([^\\n]+)",
		},
		"users": []interface{}{
			map[string]interface{}{
				"id":     "json::users.#.id",
				"name":   "json::users.#.name",
				"email":  "json::users.#.email",
				"status": "regex::status(\\d+)",
				"groups": []interface{}{"json::users.#.groups.#"},
				"profile": map[string]interface{}{
					"age":        "json::users.#.profile.age",
					"department": "json::users.#.profile.department",
					"skills":     []interface{}{"regex::skill:(\\w+)"},
				},
			},
		},
		"summary": map[string]interface{}{
			"total_users":  "json::stats.total",
			"active_count": "regex::active:(\\d+)",
			"department_breakdown": []interface{}{
				map[string]interface{}{
					"dept":  "regex::dept:(\\w+)",
					"count": "regex::count:(\\d+)",
				},
			},
		},
	}

	// 混合数据：JSON + 文本
	jsonData := map[string]interface{}{
		"meta": map[string]interface{}{
			"title":   "User Management System",
			"version": "2.1.0",
			"tags":    []string{"user", "management", "api"},
		},
		"users": []interface{}{
			map[string]interface{}{
				"id":    1,
				"name":  "Alice Johnson",
				"email": "alice@company.com",
				"profile": map[string]interface{}{
					"age":        28,
					"department": "Engineering",
				},
				"groups": []string{"admin", "developer"},
			},
			map[string]interface{}{
				"id":    2,
				"name":  "Bob Smith",
				"email": "bob@company.com",
				"profile": map[string]interface{}{
					"age":        35,
					"department": "Marketing",
				},
				"groups": []string{"editor"},
			},
		},
		"stats": map[string]interface{}{
			"total": 150,
		},
	}

	jsonBytes, _ := json.Marshal(jsonData)

	// 附加文本数据
	textData := "\ndesc: Advanced user management system with role-based access\n" +
		"status1 status2 status3\n" +
		"active:120 inactive:30\n" +
		"skill:golang skill:python skill:javascript skill:docker\n" +
		"dept:Engineering count:80\n" +
		"dept:Marketing count:45\n" +
		"dept:Sales count:25"

	testdata := string(jsonBytes) + textData

	schema := NewJsonSchema(scheme)
	require.NotNil(t, schema)
	value := schema.Format(testdata)

	v, err := json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)
	t.Log("Complex nested result:")
	t.Log(string(v))

	// 验证结构
	require.Contains(t, value, "metadata")
	require.Contains(t, value, "users")
	require.Contains(t, value, "summary")

	metadata := value["metadata"].(map[string]interface{})
	require.Equal(t, "User Management System", metadata["title"])
	require.Equal(t, "2.1.0", metadata["version"])

	users := value["users"].([]interface{})
	require.Len(t, users, 2)

	user1 := users[0].(map[string]interface{})
	require.Equal(t, float64(1), user1["id"])
	require.Equal(t, "Alice Johnson", user1["name"])
	require.Equal(t, "1", user1["status"]) // 来自regex
}

func TestDeepNestedArrays(t *testing.T) {
	// 测试深度嵌套数组的处理
	scheme := map[string]interface{}{
		"categories": []interface{}{
			map[string]interface{}{
				"name": "json::categories.#.name",
				"items": []interface{}{
					map[string]interface{}{
						"title":       "json::categories.#.items.#.title",
						"price":       "json::categories.#.items.#.price",
						"tags":        []interface{}{"json::categories.#.items.#.tags.#"},
						"color_codes": []interface{}{"regex::#([A-Fa-f0-9]{6})"},
					},
				},
			},
		},
		"extracted_colors": []interface{}{"regex::#([A-Fa-f0-9]{6})"},
		"price_range": map[string]interface{}{
			"min": "regex::min:(\\d+\\.\\d+)",
			"max": "regex::max:(\\d+\\.\\d+)",
		},
	}

	jsonData := map[string]interface{}{
		"categories": []interface{}{
			map[string]interface{}{
				"name": "Electronics",
				"items": []interface{}{
					map[string]interface{}{
						"title": "Laptop",
						"price": 999.99,
						"tags":  []string{"computer", "portable", "work"},
					},
					map[string]interface{}{
						"title": "Phone",
						"price": 699.99,
						"tags":  []string{"mobile", "communication"},
					},
				},
			},
			map[string]interface{}{
				"name": "Clothing",
				"items": []interface{}{
					map[string]interface{}{
						"title": "T-Shirt",
						"price": 29.99,
						"tags":  []string{"casual", "cotton"},
					},
				},
			},
		},
	}

	jsonBytes, _ := json.Marshal(jsonData)
	textData := "\nColor palette: #FF5733 #33C3FF #75FF33 #FF33E6\n" +
		"min:29.99 max:999.99\n" +
		"Additional colors: #123456 #ABCDEF"

	testdata := string(jsonBytes) + textData

	schema := NewJsonSchema(scheme)
	require.NotNil(t, schema)
	value := schema.Format(testdata)

	v, err := json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)
	t.Log("Deep nested arrays result:")
	t.Log(string(v))

	// 验证深度嵌套结构
	categories := value["categories"].([]interface{})
	require.Len(t, categories, 2)

	electronics := categories[0].(map[string]interface{})
	require.Equal(t, "Electronics", electronics["name"])

	items := electronics["items"].([]interface{})
	require.Len(t, items, 2)

	// 第一个item应该对应第一个商品
	laptop := items[0].(map[string]interface{})
	laptopTitle := laptop["title"].([]interface{})
	require.Contains(t, laptopTitle, "Laptop")

	laptopPrice := laptop["price"].([]interface{})
	require.Contains(t, laptopPrice, float64(999.99))

	// 验证颜色提取
	extractedColors := value["extracted_colors"].([]interface{})
	require.Contains(t, extractedColors, "FF5733")
	require.Contains(t, extractedColors, "123456")
}

func TestMixedArrayTypes(t *testing.T) {
	// 测试不同类型数组字段的混合使用
	scheme := map[string]interface{}{
		"products": []interface{}{
			map[string]interface{}{
				"id":         "regex::id(\\d+)",
				"name":       "json::products.#.name",
				"categories": []interface{}{"json::products.#.categories.#"},
				"variants":   []interface{}{"regex::variant:(\\w+)"},
				"pricing": map[string]interface{}{
					"base":      "json::products.#.price",
					"discounts": []interface{}{"regex::discount(\\d+)"},
				},
			},
		},
		"global_categories": []interface{}{"json::all_categories.#"},
		"promotion_codes":   []interface{}{"regex::promo:([A-Z0-9]+)"},
	}

	jsonData := map[string]interface{}{
		"products": []interface{}{
			map[string]interface{}{
				"name":       "Gaming Laptop",
				"price":      1299.99,
				"categories": []string{"electronics", "gaming", "computers"},
			},
			map[string]interface{}{
				"name":       "Wireless Mouse",
				"price":      49.99,
				"categories": []string{"electronics", "accessories"},
			},
		},
		"all_categories": []string{"electronics", "gaming", "computers", "accessories", "mobile"},
	}

	jsonBytes, _ := json.Marshal(jsonData)
	textData := "\nProduct variants:\n" +
		"id1 variant:silver variant:black variant:white\n" +
		"id2 variant:red variant:blue\n" +
		"Promotions: promo:SAVE20 promo:NEWUSER promo:WEEKEND50\n" +
		"Discounts: discount10 discount15 discount25"

	testdata := string(jsonBytes) + textData

	schema := NewJsonSchema(scheme)
	require.NotNil(t, schema)
	value := schema.Format(testdata)

	v, err := json.MarshalIndent(value, "", "  ")
	require.NoError(t, err)
	t.Log("Mixed array types result:")
	t.Log(string(v))

	// 验证混合数组处理
	products := value["products"].([]interface{})
	require.Len(t, products, 2)

	product1 := products[0].(map[string]interface{})
	require.Equal(t, "1", product1["id"])
	require.Equal(t, "Gaming Laptop", product1["name"])

	variants := product1["variants"].([]interface{})
	require.Contains(t, variants, "silver")
	require.Contains(t, variants, "black")
	require.Contains(t, variants, "white")

	pricing := product1["pricing"].(map[string]interface{})
	basePrice := pricing["base"].([]interface{})
	require.Contains(t, basePrice, float64(1299.99))

	discounts := pricing["discounts"].([]interface{})
	require.Contains(t, discounts, "10")
	require.Contains(t, discounts, "15")
	require.Contains(t, discounts, "25")
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
