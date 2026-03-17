package httpx

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
)

var absoluteUrlRegex = regexp.MustCompile(`([a-z][a-z\d+\-.]*:)?//`)
var baseUrlRegex = regexp.MustCompile(`/+$`)
var refUrlRegex = regexp.MustCompile(`^/+`)

func EncodeBodyData(data any) ([]byte, error) {
	switch data.(type) {
	case nil:
		return nil, nil
	case string:
		return []byte(data.(string)), nil
	case []byte:
		return data.([]byte), nil
	}
	rs, err := json.Marshal(data)
	if err != nil {
		return nil, &Error{
			Kind:    ErrKindEncodeBody,
			Op:      "encode_body",
			Message: "failed to marshal request body",
			Cause:   err,
		}
	}
	return rs, nil
}

// FormatBodyData is kept for compatibility and returns nil on encoding errors.
func FormatBodyData(data any) []byte {
	rs, _ := EncodeBodyData(data)
	return rs
}

func UrlMustParse(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

// UrlJoin adapted from axios core
func BuildUrl(base string, ref string) *url.URL {
	u, _ := BuildURLE(base, ref)
	return u
}

func BuildURLE(base string, ref string) (*url.URL, error) {
	if base == "" || absoluteUrlRegex.MatchString(ref) {
		u, err := url.Parse(ref)
		if err != nil {
			return nil, &Error{
				Kind:    ErrKindInvalidURL,
				Op:      "build_url",
				URL:     ref,
				Message: "invalid reference url",
				Cause:   err,
			}
		}
		return u, nil
	}
	if ref == "" {
		return nil, errors.New("empty reference url")
	}
	joined := baseUrlRegex.ReplaceAllString(base, "") + "/" + refUrlRegex.ReplaceAllString(ref, "")
	u, err := url.Parse(joined)
	if err != nil {
		return nil, &Error{
			Kind:    ErrKindInvalidURL,
			Op:      "build_url",
			URL:     joined,
			Message: "invalid joined url",
			Cause:   err,
		}
	}
	return u, nil
}
