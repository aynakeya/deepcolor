package chardet

import "bytes"

func classifyTLD(tld []byte) tldClass {
	switch len(tld) {
	case 2:
		if c, ok := twoLetterTLDMap[string(tld)]; ok {
			return c
		}
		return tldWestern
	case 3:
		if bytes.Equal(tld, []byte("edu")) || bytes.Equal(tld, []byte("gov")) || bytes.Equal(tld, []byte("mil")) {
			return tldWestern
		}
		return tldGeneric
	default:
		if len(tld) >= 8 && bytes.HasPrefix(tld, []byte("xn--")) {
			if c, ok := punycodeTLDMap[string(bytes.ToLower(tld[4:]))]; ok {
				return c
			}
		}
		return tldGeneric
	}
}
