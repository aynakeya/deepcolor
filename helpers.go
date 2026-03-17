package deepcolor

import (
	"net/http"

	"github.com/tidwall/gjson"
)

func GET(url string) *Request {
	return NewRequest(http.MethodGet, url)
}

func POST(url string) *Request {
	return NewRequest(http.MethodPost, url)
}

func PUT(url string) *Request {
	return NewRequest(http.MethodPut, url)
}

func DELETE(url string) *Request {
	return NewRequest(http.MethodDelete, url)
}

func ParseJSON(data []byte) gjson.Result {
	return gjson.ParseBytes(data)
}

func ParseJSONString(data string) gjson.Result {
	return gjson.Parse(data)
}

func ParseResponseJSON(resp *Response) gjson.Result {
	if resp == nil {
		return gjson.Result{}
	}
	return resp.JSONResult()
}
