package formatter

import (
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
)

type xmlValueGetter struct {
	doc *xmlquery.Node
}

func (x *xmlValueGetter) Value(expression string, indexes []int) interface{} {
	if x.doc == nil {
		return nil
	}
	
	// Apply indexing to the XPath expression if needed
	indexedExpression := x.applyIndexesToExpression(expression, indexes)
	
	node := xmlquery.FindOne(x.doc, indexedExpression)
	if node == nil {
		return nil
	}
	
	// Return the inner text of the node
	return node.InnerText()
}

func (x *xmlValueGetter) Array(expression string, indexes []int) []interface{} {
	if x.doc == nil {
		return nil
	}
	
	// Apply indexing to the XPath expression if needed
	indexedExpression := x.applyIndexesToExpression(expression, indexes)
	
	nodes := xmlquery.Find(x.doc, indexedExpression)
	if len(nodes) == 0 {
		return nil
	}
	
	results := make([]interface{}, 0, len(nodes))
	for _, node := range nodes {
		results = append(results, node.InnerText())
	}
	return results
}

// applyIndexesToExpression applies array indexes to XPath expressions
// For XML, we can use position() predicate to select specific elements
func (x *xmlValueGetter) applyIndexesToExpression(expression string, indexes []int) string {
	if len(indexes) == 0 {
		return expression
	}
	
	result := expression
	indexPos := 0
	
	// Replace each occurrence of # with the corresponding index
	for indexPos < len(indexes) {
		hashPos := strings.Index(result, "#")
		if hashPos == -1 {
			break
		}
		
		// Convert 0-based index to 1-based for XPath
		position := indexes[indexPos] + 1
		replacement := strconv.Itoa(position)
		
		result = result[:hashPos] + replacement + result[hashPos+1:]
		indexPos++
	}
	
	return result
}

func NewXmlValueGetter(data string) IValueGetter {
	doc, err := xmlquery.Parse(strings.NewReader(data))
	if err != nil {
		// Return a getter that always returns nil if parsing fails
		return &xmlValueGetter{doc: nil}
	}
	
	return &xmlValueGetter{
		doc: doc,
	}
}
