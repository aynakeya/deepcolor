package chardet

// Detect is a convenience API for one-shot detection.
// It creates an isolated detector per call, so it is safe for concurrent use.
// data is the raw byte content to detect.
// tld is the right-most DNS label in lower-case ASCII (e.g. "cn", "jp", "com"),
// used as a locale hint and can be nil/empty.
// allowUTF8 controls whether UTF-8 is an allowed final result:
// - true: return UTF-8 when input looks like valid UTF-8.
// - false: never return UTF-8; fallback to legacy/non-UTF candidates.
func Detect(data []byte, tld []byte, allowUTF8 bool) Encoding {
	return DetectAssess(data, tld, allowUTF8).Encoding
}

// DetectNoTLD is a convenience API when no TLD hint is available.
// It is equivalent to Detect(data, nil, allowUTF8).
func DetectNoTLD(data []byte, allowUTF8 bool) Encoding {
	return Detect(data, nil, allowUTF8)
}

// DetectAssess is a convenience API for one-shot detection with confidence info.
// It creates an isolated detector per call, so it is safe for concurrent use.
// data/tld/allowUTF8 have the same meaning as Detect.
func DetectAssess(data []byte, tld []byte, allowUTF8 bool) GuessResult {
	d := NewDetector()
	d.Feed(data, true)
	return d.GuessAssess(tld, allowUTF8)
}

// DetectAssessNoTLD is a convenience API when no TLD hint is available.
// It is equivalent to DetectAssess(data, nil, allowUTF8).
func DetectAssessNoTLD(data []byte, allowUTF8 bool) GuessResult {
	return DetectAssess(data, nil, allowUTF8)
}
