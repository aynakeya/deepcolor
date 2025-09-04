package formatter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXmlValueGetter(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<library>
	<book id="1">
		<title>Go Programming</title>
		<author>John Doe</author>
		<year>2023</year>
		<tags>
			<tag>programming</tag>
			<tag>golang</tag>
			<tag>software</tag>
		</tags>
	</book>
	<book id="2">
		<title>Advanced Go</title>
		<author>Jane Smith</author>
		<year>2024</year>
		<tags>
			<tag>advanced</tag>
			<tag>golang</tag>
		</tags>
	</book>
</library>`

	getter := NewXmlValueGetter(xmlData)
	require.NotNil(t, getter)

	// Test single value extraction
	title := getter.Value("//book[1]/title", nil)
	require.Equal(t, "Go Programming", title)

	author := getter.Value("//book[2]/author", nil)
	require.Equal(t, "Jane Smith", author)

	// Test attribute access
	bookId := getter.Value("//book[1]/@id", nil)
	require.Equal(t, "1", bookId)

	// Test non-existent path
	notFound := getter.Value("//book/nonexistent", nil)
	require.Nil(t, notFound)

	// Test array extraction
	allTitles := getter.Array("//book/title", nil)
	require.Len(t, allTitles, 2)
	require.Contains(t, allTitles, "Go Programming")
	require.Contains(t, allTitles, "Advanced Go")

	// Test nested array extraction
	firstBookTags := getter.Array("//book[1]/tags/tag", nil)
	require.Len(t, firstBookTags, 3)
	require.Contains(t, firstBookTags, "programming")
	require.Contains(t, firstBookTags, "golang")
	require.Contains(t, firstBookTags, "software")

	// Test all tags from all books
	allTags := getter.Array("//book/tags/tag", nil)
	require.Len(t, allTags, 5)
	require.Contains(t, allTags, "programming")
	require.Contains(t, allTags, "advanced")
}

func TestXmlValueGetterWithIndexes(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<bookstore>
	<book>
		<title>Book 1</title>
		<authors>
			<author>Author 1A</author>
			<author>Author 1B</author>
		</authors>
	</book>
	<book>
		<title>Book 2</title>
		<authors>
			<author>Author 2A</author>
		</authors>
	</book>
</bookstore>`

	getter := NewXmlValueGetter(xmlData)
	require.NotNil(t, getter)

	// Test indexed access with # notation
	firstBookTitle := getter.Value("//book[#]/title", []int{0})
	require.Equal(t, "Book 1", firstBookTitle)

	secondBookTitle := getter.Value("//book[#]/title", []int{1})
	require.Equal(t, "Book 2", secondBookTitle)

	// Test nested indexed access
	firstBookFirstAuthor := getter.Value("//book[#]/authors/author[#]", []int{0, 0})
	require.Equal(t, "Author 1A", firstBookFirstAuthor)

	firstBookSecondAuthor := getter.Value("//book[#]/authors/author[#]", []int{0, 1})
	require.Equal(t, "Author 1B", firstBookSecondAuthor)

	secondBookFirstAuthor := getter.Value("//book[#]/authors/author[#]", []int{1, 0})
	require.Equal(t, "Author 2A", secondBookFirstAuthor)

	// Test array access with indexes
	firstBookAuthors := getter.Array("//book[#]/authors/author", []int{0})
	require.Len(t, firstBookAuthors, 2)
	require.Contains(t, firstBookAuthors, "Author 1A")
	require.Contains(t, firstBookAuthors, "Author 1B")
}

func TestXmlValueGetterInvalidData(t *testing.T) {
	// Test with invalid XML
	invalidXml := `<invalid><unclosed>`
	getter := NewXmlValueGetter(invalidXml)
	require.NotNil(t, getter)

	// Should return nil for any query on invalid XML
	result := getter.Value("//anything", nil)
	require.Nil(t, result)

	arrayResult := getter.Array("//anything", nil)
	require.Nil(t, arrayResult)
}

func TestXmlSchemaIntegration(t *testing.T) {
	// Test XML getter integration with schema
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<products>
	<product>
		<name>Laptop</name>
		<price>999.99</price>
		<category>Electronics</category>
		<features>
			<feature>Fast</feature>
			<feature>Lightweight</feature>
		</features>
	</product>
	<product>
		<name>Mouse</name>
		<price>29.99</price>
		<category>Electronics</category>
		<features>
			<feature>Wireless</feature>
		</features>
	</product>
</products>`

	schema := map[string]interface{}{
		"product_list": []interface{}{
			map[string]interface{}{
				"product_name": "xml:://product[#]/name",
				"cost":         "xml:://product[#]/price",
				"type":         "xml:://product[#]/category",
				"features":     []interface{}{"xml:://product[#]/features/feature"},
			},
		},
		"all_features": []interface{}{"xml:://product/features/feature"},
	}

	jsonSchema := NewJsonSchema(schema)
	require.NotNil(t, jsonSchema)

	result := jsonSchema.Format(xmlData)
	require.NotNil(t, result)

	// Verify product list
	productList := result["product_list"].([]interface{})
	require.Len(t, productList, 2)

	product1 := productList[0].(map[string]interface{})
	require.Equal(t, "Laptop", product1["product_name"])
	require.Equal(t, "999.99", product1["cost"])
	require.Equal(t, "Electronics", product1["type"])

	features1 := product1["features"].([]interface{})
	require.Len(t, features1, 2)
	require.Contains(t, features1, "Fast")
	require.Contains(t, features1, "Lightweight")

	product2 := productList[1].(map[string]interface{})
	require.Equal(t, "Mouse", product2["product_name"])
	require.Equal(t, "29.99", product2["cost"])

	// Verify all features
	allFeatures := result["all_features"].([]interface{})
	require.Len(t, allFeatures, 3)
	require.Contains(t, allFeatures, "Fast")
	require.Contains(t, allFeatures, "Lightweight")
	require.Contains(t, allFeatures, "Wireless")

	// Note: XPath count() function testing removed for simplicity
}

func TestXmlToJsonSchemaConversion(t *testing.T) {
	// 复杂的XML数据 - 电商产品目录
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<catalog>
	<metadata>
		<name>Spring Electronics Catalog</name>
		<version>2024.1</version>
		<categories>Electronics,Computers,Mobile</categories>
		<last_updated>2024-03-15</last_updated>
	</metadata>
	<products>
		<product id="P001" featured="true">
			<name>Gaming Laptop Pro</name>
			<price currency="USD">1299.99</price>
			<category>Computers</category>
			<description>High-performance gaming laptop with RTX graphics</description>
			<specifications>
				<cpu>Intel i7-12700H</cpu>
				<ram>32GB DDR5</ram>
				<storage>1TB NVMe SSD</storage>
				<gpu>RTX 4070</gpu>
			</specifications>
			<tags>
				<tag>gaming</tag>
				<tag>laptop</tag>
				<tag>high-performance</tag>
			</tags>
			<reviews>
				<review rating="5">Excellent performance!</review>
				<review rating="4">Great value for money</review>
				<review rating="5">Perfect for gaming</review>
			</reviews>
		</product>
		<product id="P002" featured="false">
			<name>Wireless Mouse</name>
			<price currency="USD">49.99</price>
			<category>Electronics</category>
			<description>Ergonomic wireless mouse with precision tracking</description>
			<specifications>
				<type>Optical</type>
				<battery>Rechargeable</battery>
				<connectivity>Bluetooth 5.0</connectivity>
			</specifications>
			<tags>
				<tag>wireless</tag>
				<tag>mouse</tag>
				<tag>ergonomic</tag>
			</tags>
			<reviews>
				<review rating="4">Very comfortable</review>
				<review rating="5">Great build quality</review>
			</reviews>
		</product>
		<product id="P003" featured="true">
			<name>Smartphone X</name>
			<price currency="USD">799.99</price>
			<category>Mobile</category>
			<description>Latest smartphone with advanced camera system</description>
			<specifications>
				<camera>108MP Triple Camera</camera>
				<display>6.7" OLED</display>
				<battery>5000mAh</battery>
				<storage>256GB</storage>
			</specifications>
			<tags>
				<tag>smartphone</tag>
				<tag>camera</tag>
				<tag>mobile</tag>
			</tags>
			<reviews>
				<review rating="5">Amazing camera quality</review>
				<review rating="4">Great battery life</review>
				<review rating="5">Beautiful display</review>
			</reviews>
		</product>
	</products>
</catalog>`

	// 目标JSON Schema - 电商API响应格式
	targetSchema := map[string]interface{}{
		"catalog_info": map[string]interface{}{
			"catalog_name":    "xml:://metadata/name",
			"version":         "xml:://metadata/version",
			"last_updated":    "xml:://metadata/last_updated",
			"available_categories": "xml:://metadata/categories",
		},
		"featured_products": []interface{}{
			map[string]interface{}{
				"product_id":   "xml:://product[@featured='true'][#]/@id",
				"name":         "xml:://product[@featured='true'][#]/name",
				"price":        "xml:://product[@featured='true'][#]/price",
				"currency":     "xml:://product[@featured='true'][#]/price/@currency",
				"category":     "xml:://product[@featured='true'][#]/category",
				"description":  "xml:://product[@featured='true'][#]/description",
				"tags":         []interface{}{"xml:://product[@featured='true'][#]/tags/tag"},
			},
		},
		"all_products": []interface{}{
			map[string]interface{}{
				"id":           "xml:://product[#]/@id",
				"name":         "xml:://product[#]/name",
				"price_info": map[string]interface{}{
					"amount":   "xml:://product[#]/price",
					"currency": "xml:://product[#]/price/@currency",
				},
				"category":     "xml:://product[#]/category",
				"is_featured":  "xml:://product[#]/@featured",
				"specifications": map[string]interface{}{
					"details": []interface{}{"xml:://product[#]/specifications/*"},
				},
				"review_summary": map[string]interface{}{
					"reviews": []interface{}{"xml:://product[#]/reviews/review"},
				},
			},
		},
		"all_categories": []interface{}{"xml:://product/category"},
	}

	// 执行转换
	schema := NewJsonSchema(targetSchema)
	require.NotNil(t, schema)

	result := schema.Format(xmlData)
	require.NotNil(t, result)

	// 验证转换结果
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	t.Log("XML to JSON conversion result:")
	t.Log(string(resultJSON))

	// 验证目录信息
	catalogInfo := result["catalog_info"].(map[string]interface{})
	require.Equal(t, "Spring Electronics Catalog", catalogInfo["catalog_name"])
	require.Equal(t, "2024.1", catalogInfo["version"])
	require.Equal(t, "2024-03-15", catalogInfo["last_updated"])
	require.Equal(t, "Electronics,Computers,Mobile", catalogInfo["available_categories"])

	// 验证特色产品
	featuredProducts := result["featured_products"].([]interface{})
	require.Len(t, featuredProducts, 2) // P001 and P003 are featured

	featured1 := featuredProducts[0].(map[string]interface{})
	require.Equal(t, "P001", featured1["product_id"])
	require.Equal(t, "Gaming Laptop Pro", featured1["name"])
	require.Equal(t, "1299.99", featured1["price"])
	require.Equal(t, "USD", featured1["currency"])
	require.Equal(t, "Computers", featured1["category"])

	featuredTags1 := featured1["tags"].([]interface{})
	require.Len(t, featuredTags1, 3)
	require.Contains(t, featuredTags1, "gaming")
	require.Contains(t, featuredTags1, "laptop")
	require.Contains(t, featuredTags1, "high-performance")

	// 验证所有产品
	allProducts := result["all_products"].([]interface{})
	require.Len(t, allProducts, 3)

	product1 := allProducts[0].(map[string]interface{})
	require.Equal(t, "P001", product1["id"])
	require.Equal(t, "Gaming Laptop Pro", product1["name"])
	require.Equal(t, "true", product1["is_featured"])

	priceInfo1 := product1["price_info"].(map[string]interface{})
	require.Equal(t, "1299.99", priceInfo1["amount"])
	require.Equal(t, "USD", priceInfo1["currency"])

	specs1 := product1["specifications"].(map[string]interface{})
	specDetails1 := specs1["details"].([]interface{})
	require.Len(t, specDetails1, 4) // cpu, ram, storage, gpu
	require.Contains(t, specDetails1, "Intel i7-12700H")
	require.Contains(t, specDetails1, "32GB DDR5")

	reviewSummary1 := product1["review_summary"].(map[string]interface{})
	reviews1 := reviewSummary1["reviews"].([]interface{})
	require.Len(t, reviews1, 3)
	require.Contains(t, reviews1, "Excellent performance!")

	// 验证第二个产品（非特色产品）
	product2 := allProducts[1].(map[string]interface{})
	require.Equal(t, "P002", product2["id"])
	require.Equal(t, "Wireless Mouse", product2["name"])
	require.Equal(t, "false", product2["is_featured"])
	require.Equal(t, "Electronics", product2["category"])

	// 验证第三个产品（特色产品）
	product3 := allProducts[2].(map[string]interface{})
	require.Equal(t, "P003", product3["id"])
	require.Equal(t, "Smartphone X", product3["name"])
	require.Equal(t, "true", product3["is_featured"])
	require.Equal(t, "Mobile", product3["category"])

	priceInfo3 := product3["price_info"].(map[string]interface{})
	require.Equal(t, "799.99", priceInfo3["amount"])
	
	// 验证所有分类
	allCategories := result["all_categories"].([]interface{})
	require.Len(t, allCategories, 3) // Computers, Electronics, Mobile
	require.Contains(t, allCategories, "Computers")
	require.Contains(t, allCategories, "Electronics")
	require.Contains(t, allCategories, "Mobile")
}

func TestXmlToJsonApiResponse(t *testing.T) {
	// 模拟RSS/XML API响应转换为现代JSON API格式
	xmlApiResponse := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Tech News Feed</title>
		<description>Latest technology news and updates</description>
		<lastBuildDate>2024-03-15T10:30:00Z</lastBuildDate>
		<item>
			<title>AI Breakthrough in Healthcare</title>
			<description>New AI system shows promising results in medical diagnosis</description>
			<pubDate>2024-03-15T09:00:00Z</pubDate>
			<category>AI</category>
			<category>Healthcare</category>
			<author>john.doe@techblog.com</author>
			<link>https://techblog.com/ai-healthcare-breakthrough</link>
		</item>
		<item>
			<title>Quantum Computing Milestone</title>
			<description>Researchers achieve new quantum supremacy benchmark</description>
			<pubDate>2024-03-14T15:30:00Z</pubDate>
			<category>Quantum</category>
			<category>Research</category>
			<author>jane.smith@techblog.com</author>
			<link>https://techblog.com/quantum-milestone</link>
		</item>
		<item>
			<title>5G Network Expansion</title>
			<description>Major cities see significant 5G coverage improvements</description>
			<pubDate>2024-03-13T12:00:00Z</pubDate>
			<category>Networking</category>
			<category>Infrastructure</category>
			<author>mike.wilson@techblog.com</author>
			<link>https://techblog.com/5g-expansion</link>
		</item>
	</channel>
</rss>`

	// 转换为现代JSON API响应格式
	apiSchema := map[string]interface{}{
		"feed_metadata": map[string]interface{}{
			"title":        "xml:://channel/title",
			"description":  "xml:://channel/description",
			"last_updated": "xml:://channel/lastBuildDate",
			"version":      "xml:://rss/@version",
		},
		"articles": []interface{}{
			map[string]interface{}{
				"id":          "xml:://item[#]/link", // 使用链接作为ID
				"headline":    "xml:://item[#]/title",
				"summary":     "xml:://item[#]/description",
				"published_at": "xml:://item[#]/pubDate",
				"author_email": "xml:://item[#]/author",
				"url":         "xml:://item[#]/link",
				"categories":  []interface{}{"xml:://item[#]/category"},
			},
		},
		"all_categories": []interface{}{"xml:://item/category"},
		"authors": []interface{}{"xml:://item/author"},
	}

	schema := NewJsonSchema(apiSchema)
	require.NotNil(t, schema)

	result := schema.Format(xmlApiResponse)
	require.NotNil(t, result)

	// 输出转换结果
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	t.Log("RSS to JSON API conversion:")
	t.Log(string(resultJSON))

	// 验证feed元数据
	feedMeta := result["feed_metadata"].(map[string]interface{})
	require.Equal(t, "Tech News Feed", feedMeta["title"])
	require.Equal(t, "Latest technology news and updates", feedMeta["description"])
	require.Equal(t, "2.0", feedMeta["version"])

	// 验证文章数据
	articles := result["articles"].([]interface{})
	require.Len(t, articles, 3)

	article1 := articles[0].(map[string]interface{})
	require.Equal(t, "AI Breakthrough in Healthcare", article1["headline"])
	require.Equal(t, "https://techblog.com/ai-healthcare-breakthrough", article1["id"])
	require.Equal(t, "https://techblog.com/ai-healthcare-breakthrough", article1["url"])
	require.Equal(t, "john.doe@techblog.com", article1["author_email"])

	categories1 := article1["categories"].([]interface{})
	require.Len(t, categories1, 2)
	require.Contains(t, categories1, "AI")
	require.Contains(t, categories1, "Healthcare")

	// 验证聚合数据
	allCategories := result["all_categories"].([]interface{})
	require.Len(t, allCategories, 6) // AI, Healthcare, Quantum, Research, Networking, Infrastructure
	require.Contains(t, allCategories, "AI")
	require.Contains(t, allCategories, "Quantum")
	require.Contains(t, allCategories, "Networking")

	authors := result["authors"].([]interface{})
	require.Len(t, authors, 3)
	require.Contains(t, authors, "john.doe@techblog.com")
	require.Contains(t, authors, "jane.smith@techblog.com")
	require.Contains(t, authors, "mike.wilson@techblog.com")
}
