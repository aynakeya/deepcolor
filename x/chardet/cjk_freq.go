package chardet

import "unicode/utf8"

func toRuneSet(v []rune) map[rune]struct{} {
	m := make(map[rune]struct{}, len(v))
	for _, r := range v {
		m[r] = struct{}{}
	}
	return m
}

func scoreCJKFrequency(b []byte, enc Encoding) float64 {
	score := 0.0
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		i += size
		_, inS := frequentSimplifiedSet[r]
		_, inK := frequentKanjiSet[r]
		_, inH := frequentHangulSet[r]
		switch enc {
		case EncodingGBK:
			if inS {
				score += 0.7
			}
			if inK {
				score += 0.2
			}
		case EncodingBig5:
			if inK {
				score += 0.7
			}
			if inS {
				score += 0.2
			}
		case EncodingEUCKR:
			if inH {
				score += 0.8
			}
		}
	}
	return score
}
