package chardet

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

const maxBufferSize = 1 << 20

type Detector struct {
	buf           []byte
	nonASCIICnt   int
	escSeen       bool
	closed        bool
	utf8Candidate bool
}

type GuessResult struct {
	Encoding  Encoding
	Confident bool
}

const (
	encWindows1252 = iota
	encWindows1251
	encWindows1250
	encISO88592
	encWindows1256
	encISO88596
	encWindows1255
	encISO88598
	encWindows1253
	encISO88597
	encWindows1254
	encWindows1258
	encWindows1257
	encISO885913
	encISO88594
	encWindows874
	encKOI8U
	encIBM866
	encISO88595
	encGBK
	encBig5
	encEUCKR
	encEUCJP
	encShiftJIS
	encISO2022JP
	encUTF8
	encCount
)

type scoreArray [encCount]float64

var encByIndex = [encCount]Encoding{
	encWindows1252: EncodingWindows1252,
	encWindows1251: EncodingWindows1251,
	encWindows1250: EncodingWindows1250,
	encISO88592:    EncodingISO88592,
	encWindows1256: EncodingWindows1256,
	encISO88596:    EncodingISO88596,
	encWindows1255: EncodingWindows1255,
	encISO88598:    EncodingISO88598,
	encWindows1253: EncodingWindows1253,
	encISO88597:    EncodingISO88597,
	encWindows1254: EncodingWindows1254,
	encWindows1258: EncodingWindows1258,
	encWindows1257: EncodingWindows1257,
	encISO885913:   EncodingISO885913,
	encISO88594:    EncodingISO88594,
	encWindows874:  EncodingWindows874,
	encKOI8U:       EncodingKOI8U,
	encIBM866:      EncodingIBM866,
	encISO88595:    EncodingISO88595,
	encGBK:         EncodingGBK,
	encBig5:        EncodingBig5,
	encEUCKR:       EncodingEUCKR,
	encEUCJP:       EncodingEUCJP,
	encShiftJIS:    EncodingShiftJIS,
	encISO2022JP:   EncodingISO2022JP,
	encUTF8:        EncodingUTF8,
}

var rankingIndexOrder = [...]int{
	encWindows1252, encWindows1251, encWindows1250, encISO88592,
	encWindows1256, encISO88596, encWindows1255, encISO88598,
	encWindows1253, encISO88597, encWindows1254, encWindows1258,
	encWindows1257, encISO885913, encISO88594, encWindows874,
	encKOI8U, encIBM866, encISO88595,
	encGBK, encBig5, encEUCKR, encEUCJP, encShiftJIS,
	encISO2022JP, encUTF8,
}

func NewDetector() *Detector {
	return &Detector{}
}

func (d *Detector) Reset() {
	d.buf = d.buf[:0]
	d.nonASCIICnt = 0
	d.escSeen = false
	d.closed = false
	d.utf8Candidate = false
}

func (d *Detector) Feed(chunk []byte, last bool) bool {
	if d.closed {
		panic("chardet: feed after closed stream")
	}
	if last {
		d.closed = true
	}
	for _, b := range chunk {
		if b >= 0x80 {
			d.nonASCIICnt++
		}
		if b == 0x1B {
			d.escSeen = true
		}
	}
	if len(d.buf)+len(chunk) <= maxBufferSize {
		d.buf = append(d.buf, chunk...)
	} else {
		remain := maxBufferSize - len(d.buf)
		if remain > 0 {
			d.buf = append(d.buf, chunk[:remain]...)
		}
	}
	if d.nonASCIICnt > 0 && utf8.Valid(d.buf) {
		d.utf8Candidate = true
	}
	return d.nonASCIICnt > 0
}

func (d *Detector) Guess(tld []byte, allowUTF8 bool) Encoding {
	return d.GuessAssess(tld, allowUTF8).Encoding
}

func (d *Detector) GuessAssess(tld []byte, allowUTF8 bool) GuessResult {
	class := tldGeneric
	if len(tld) > 0 {
		if containsUpperCasePeriodOrNonASCII(tld) {
			panic("chardet: invalid tld label")
		}
		class = classifyTLD(tld)
	}

	if d.nonASCIICnt == 0 {
		if d.escSeen && hasISO2022JPEscape(d.buf) {
			return GuessResult{Encoding: EncodingISO2022JP, Confident: true}
		}
		return GuessResult{Encoding: defaultEncodingForTLD(class), Confident: false}
	}

	if d.utf8Candidate && allowUTF8 {
		return GuessResult{Encoding: EncodingUTF8, Confident: true}
	}
	if d.utf8Candidate && !allowUTF8 {
		return GuessResult{Encoding: defaultEncodingForTLD(class), Confident: true}
	}

	scores := scoreAll(d.buf)
	applyTLDBias(&scores, class)
	if !allowUTF8 {
		scores[encUTF8] = -1e18
	}

	bestEnc := EncodingWindows1252
	best := -1e18
	second := -1e18
	for _, idx := range rankingIndexOrder {
		s := scores[idx]
		if s > best {
			second = best
			best = s
			bestEnc = encByIndex[idx]
		} else if s > second {
			second = s
		}
	}
	// Rust-like special case for Hebrew visual vs logical ordering.
	visualScore, visualPunct, visualOK, logicalPunct := hebrewVisualAssessment(d.buf)
	if visualOK && (visualScore > best || bestEnc == EncodingWindows1255) && visualPunct > logicalPunct {
		bestEnc = EncodingISO88598
		best = visualScore
	}
	return GuessResult{Encoding: bestEnc, Confident: best-second >= 10}
}

func TLDMayAffectGuess(tld []byte) bool {
	if len(tld) == 0 {
		return false
	}
	if containsUpperCasePeriodOrNonASCII(tld) {
		panic("chardet: invalid tld label")
	}
	return classifyTLD(tld) != tldGeneric
}

func containsUpperCasePeriodOrNonASCII(label []byte) bool {
	for _, b := range label {
		if b >= 0x80 || b == '.' || (b >= 'A' && b <= 'Z') {
			return true
		}
	}
	return false
}

func hasISO2022JPEscape(b []byte) bool {
	return bytes.Contains(b, []byte("\x1B$B")) || bytes.Contains(b, []byte("\x1B(J")) || bytes.Contains(b, []byte("\x1B(B"))
}

func encodingImpl(enc Encoding) encoding.Encoding {
	switch enc {
	case EncodingShiftJIS:
		return japanese.ShiftJIS
	case EncodingEUCJP:
		return japanese.EUCJP
	case EncodingEUCKR:
		return korean.EUCKR
	case EncodingGBK:
		return simplifiedchinese.GBK
	case EncodingBig5:
		return traditionalchinese.Big5
	case EncodingWindows1252:
		return charmap.Windows1252
	case EncodingWindows1251:
		return charmap.Windows1251
	case EncodingWindows1250:
		return charmap.Windows1250
	case EncodingISO88592:
		return charmap.ISO8859_2
	case EncodingWindows1256:
		return charmap.Windows1256
	case EncodingWindows1254:
		return charmap.Windows1254
	case EncodingWindows874:
		return charmap.Windows874
	case EncodingWindows1255:
		return charmap.Windows1255
	case EncodingISO88598:
		return charmap.ISO8859_8
	case EncodingWindows1253:
		return charmap.Windows1253
	case EncodingISO88597:
		return charmap.ISO8859_7
	case EncodingWindows1257:
		return charmap.Windows1257
	case EncodingISO885913:
		return charmap.ISO8859_13
	case EncodingKOI8U:
		return charmap.KOI8U
	case EncodingIBM866:
		return charmap.CodePage866
	case EncodingISO88596:
		return charmap.ISO8859_6
	case EncodingWindows1258:
		return charmap.Windows1258
	case EncodingISO88594:
		return charmap.ISO8859_4
	case EncodingISO88595:
		return charmap.ISO8859_5
	default:
		return nil
	}
}

func scoreAll(buf []byte) scoreArray {
	var s scoreArray
	for i := range s {
		s[i] = -1e18
	}
	s[encUTF8] = scoreUTF8(buf)
	s[encISO2022JP] = scoreISO2022JP(buf)
	s[encShiftJIS] = scoreDecoded(buf, EncodingShiftJIS)
	s[encEUCJP] = scoreDecoded(buf, EncodingEUCJP)
	s[encEUCKR] = scoreDecoded(buf, EncodingEUCKR)
	s[encGBK] = scoreDecoded(buf, EncodingGBK)
	s[encBig5] = scoreDecoded(buf, EncodingBig5)
	s[encWindows1252] = scoreDecoded(buf, EncodingWindows1252)
	s[encWindows1251] = scoreDecoded(buf, EncodingWindows1251)
	s[encWindows1250] = scoreDecoded(buf, EncodingWindows1250)
	s[encISO88592] = scoreDecoded(buf, EncodingISO88592)
	s[encWindows1256] = scoreDecoded(buf, EncodingWindows1256)
	s[encWindows1254] = scoreDecoded(buf, EncodingWindows1254)
	s[encWindows874] = scoreDecoded(buf, EncodingWindows874)
	s[encWindows1255] = scoreDecoded(buf, EncodingWindows1255)
	s[encISO88598] = scoreDecoded(buf, EncodingISO88598)
	s[encWindows1253] = scoreDecoded(buf, EncodingWindows1253)
	s[encISO88597] = scoreDecoded(buf, EncodingISO88597)
	s[encWindows1257] = scoreDecoded(buf, EncodingWindows1257)
	s[encISO885913] = scoreDecoded(buf, EncodingISO885913)
	s[encKOI8U] = scoreDecoded(buf, EncodingKOI8U)
	s[encIBM866] = scoreDecoded(buf, EncodingIBM866)
	s[encISO88596] = scoreDecoded(buf, EncodingISO88596)
	s[encWindows1258] = scoreDecoded(buf, EncodingWindows1258)
	s[encISO88594] = scoreDecoded(buf, EncodingISO88594)
	s[encISO88595] = scoreDecoded(buf, EncodingISO88595)
	s[encWindows1252] += windows1252OrdinalHint(buf)
	sjBoost, ejPenalty := shiftJISHalfWidthSignal(buf)
	s[encShiftJIS] += sjBoost
	s[encEUCJP] -= ejPenalty
	ejBoost, sjPenalty := eucJPPrefixSignal(buf)
	s[encEUCJP] += ejBoost
	s[encShiftJIS] -= sjPenalty
	return s
}

func shiftJISHalfWidthSignal(buf []byte) (sjBoost float64, ejPenalty float64) {
	half := 0
	prefix8E := 0
	for i := 0; i < len(buf); i++ {
		b := buf[i]
		if b >= 0xA1 && b <= 0xDF {
			half++
		}
		if b == 0x8E && i+1 < len(buf) {
			t := buf[i+1]
			if t >= 0xA1 && t <= 0xDF {
				prefix8E++
			}
		}
	}
	if half >= 6 && prefix8E == 0 {
		return float64(half * 8), float64(half * 6)
	}
	return 0, 0
}

func eucJPPrefixSignal(buf []byte) (ejBoost float64, sjPenalty float64) {
	p8e := 0
	p8f := 0
	valid8e := 0
	valid8f := 0
	for i := 0; i < len(buf); i++ {
		b := buf[i]
		if b == 0x8E {
			p8e++
			if i+1 < len(buf) {
				t := buf[i+1]
				if t >= 0xA1 && t <= 0xDF {
					valid8e++
				}
			}
		}
		if b == 0x8F {
			p8f++
			if i+2 < len(buf) {
				b1, b2 := buf[i+1], buf[i+2]
				if b1 >= 0xA1 && b1 <= 0xFE && b2 >= 0xA1 && b2 <= 0xFE {
					valid8f++
				}
			}
		}
	}
	if valid8e+valid8f == 0 {
		return 0, 0
	}
	boost := float64(valid8e*18 + valid8f*26)
	noise := (p8e - valid8e) + (p8f - valid8f)
	if noise > 0 {
		boost -= float64(noise * 8)
	}
	if boost < 0 {
		boost = 0
	}
	penalty := boost * 0.7
	return boost, penalty
}

func windows1252OrdinalHint(buf []byte) float64 {
	score := 0.0
	if bytes.Contains(buf, []byte{' ', 0xA9, ' '}) {
		score += 120
	}
	if bytes.Contains(buf, []byte{' ', 0xBA, ' '}) || bytes.Contains(buf, []byte{' ', 0xAA, ' '}) {
		score += 220
	}
	if bytes.Contains(buf, []byte{'.', 0xBA, ' '}) || bytes.Contains(buf, []byte{'.', 0xAA, ' '}) {
		score += 220
	}
	if bytes.Contains(buf, []byte{'N', 0xBA}) || bytes.Contains(buf, []byte{'n', 0xBA}) {
		score += 220
	}
	if bytes.Contains(buf, []byte{'M', 0xAA}) || bytes.Contains(buf, []byte{'D', 0xAA}) || bytes.Contains(buf, []byte{'S', 0xAA}) {
		score += 180
	}
	return score
}

func scoreUTF8(buf []byte) float64 {
	if utf8.Valid(buf) {
		return 1000 + float64(countNonASCII(buf))
	}
	return -1000
}

func scoreISO2022JP(buf []byte) float64 {
	if hasISO2022JPEscape(buf) {
		return 1200
	}
	return -500
}

func scoreDecoded(buf []byte, enc Encoding) float64 {
	impl := encodingImpl(enc)
	if impl == nil {
		return -1e9
	}
	if enc == EncodingShiftJIS || enc == EncodingEUCJP || enc == EncodingEUCKR || enc == EncodingGBK || enc == EncodingBig5 {
		return scoreCJKCandidate(buf, enc)
	}
	if idx, ok := singleByteIndexForEncoding(enc); ok {
		base := scoreSingleByteEncoding(buf, enc, idx)
		switch enc {
		case EncodingWindows1254:
			if likelyTurkishBytes(buf) {
				decoded, _ := impl.NewDecoder().Bytes(buf)
				base += scoreTurkishHintFast(decoded) * 10
			}
		case EncodingWindows1258:
			if likelyVietnameseBytes(buf) {
				decoded, _ := impl.NewDecoder().Bytes(buf)
				base += scoreVietnameseHintFast(decoded) * 3
			}
		case EncodingWindows1252:
			decoded, _ := impl.NewDecoder().Bytes(buf)
			base += scoreWesternHintFast(decoded) * 2
		case EncodingWindows1250, EncodingWindows1257:
			decoded, _ := impl.NewDecoder().Bytes(buf)
			base += scoreCentralBalticHintFast(decoded) * 2
		}
		return base
	}
	decoded, err := impl.NewDecoder().Bytes(buf)
	score := 0.0
	if err != nil {
		score -= 120
	}
	var repl, cyr, greek, arabic, hebrew, thai, han, hira, kata, hangul, latin, digit, ctrl int
	for i := 0; i < len(decoded); {
		r, size := utf8.DecodeRune(decoded[i:])
		i += size
		switch {
		case r == unicode.ReplacementChar:
			repl++
		case unicode.Is(unicode.Cyrillic, r):
			cyr++
		case unicode.Is(unicode.Greek, r):
			greek++
		case unicode.Is(unicode.Arabic, r):
			arabic++
		case unicode.Is(unicode.Hebrew, r):
			hebrew++
		case unicode.Is(unicode.Thai, r):
			thai++
		case unicode.Is(unicode.Han, r):
			han++
		case unicode.Is(unicode.Hiragana, r):
			hira++
		case unicode.Is(unicode.Katakana, r):
			kata++
		case unicode.Is(unicode.Hangul, r):
			hangul++
		case unicode.IsLetter(r):
			latin++
		case unicode.IsDigit(r):
			digit++
		case r < 0x20 && r != '\n' && r != '\r' && r != '\t':
			ctrl++
		}
	}

	score += float64(digit) * 0.2
	score += float64(latin) * 0.2
	score -= float64(repl) * 8
	score -= float64(ctrl) * 10

	switch enc {
	case EncodingShiftJIS, EncodingEUCJP:
		score += float64(han)*1.3 + float64(hira)*2.0 + float64(kata)*1.8
	case EncodingGBK, EncodingBig5:
		score += float64(han) * 1.6
	case EncodingEUCKR:
		score += float64(hangul)*2.2 + float64(han)*0.2
	case EncodingWindows1251, EncodingISO88595, EncodingKOI8U, EncodingIBM866:
		score += float64(cyr) * 2
	case EncodingWindows1253, EncodingISO88597:
		score += float64(greek) * 2
	case EncodingWindows1255, EncodingISO88598:
		score += float64(hebrew) * 2
	case EncodingWindows1256, EncodingISO88596:
		score += float64(arabic) * 2
	case EncodingWindows874:
		score += float64(thai) * 2.3
	case EncodingWindows1258:
		score += scoreVietnameseHintFast(decoded)
	case EncodingWindows1254:
		score += scoreTurkishHintFast(decoded)
	case EncodingWindows1250, EncodingISO88592, EncodingWindows1257, EncodingISO885913, EncodingISO88594:
		score += scoreCentralBalticHintFast(decoded)
	case EncodingWindows1252:
		score += scoreWesternHintFast(decoded)
	}
	if enc == EncodingShiftJIS {
		score += scoreKanaFromBytes(buf)
	}
	score += scoreMultibyteStructure(buf, enc)
	if enc == EncodingBig5 || enc == EncodingGBK {
		score += scoreCJKFrequency(decoded, enc)
	}
	if isSingleByteEncoding(enc) {
		// Prevent single-byte families from dominating CJK-heavy buffers.
		score -= float64(countNonASCII(buf)) * 0.35
	}
	return score
}

func singleByteIndexForEncoding(enc Encoding) (int, bool) {
	switch enc {
	case EncodingWindows1258:
		return windows1258Index, true
	case EncodingWindows1250:
		return windows1250Index, true
	case EncodingISO88592:
		return iso88592Index, true
	case EncodingWindows1251:
		return windows1251Index, true
	case EncodingKOI8U:
		return koi8UIndex, true
	case EncodingISO88595:
		return iso88595Index, true
	case EncodingIBM866:
		return ibm866Index, true
	case EncodingWindows1252:
		return windows1252Index, true
	case EncodingWindows1253:
		return windows1253Index, true
	case EncodingISO88597:
		return iso88597Index, true
	case EncodingWindows1254:
		return windows1254Index, true
	case EncodingWindows1255:
		return windows1255Index, true
	case EncodingISO88598:
		return iso88598Index, true
	case EncodingWindows1256:
		return windows1256Index, true
	case EncodingISO88596:
		return iso88596Index, true
	case EncodingWindows1257:
		return windows1257Index, true
	case EncodingISO885913:
		return iso885913Index, true
	case EncodingISO88594:
		return iso88594Index, true
	case EncodingWindows874:
		return windows874Index, true
	default:
		return 0, false
	}
}

func scoreSingleByte(buf []byte, data *singleByteData, isWindows1256 bool) int64 {
	var prev uint8
	var score int64
	for _, b := range buf {
		class := data.classify(b)
		if class == 255 {
			return -1 << 20
		}
		caseless := class & 0x7F
		score += data.score(caseless, prev, isWindows1256)
		prev = caseless
	}
	return score
}

func scoreSingleByteEncoding(buf []byte, enc Encoding, idx int) float64 {
	isLatinFamily := enc == EncodingWindows1252 || enc == EncodingWindows1250 || enc == EncodingISO88592 ||
		enc == EncodingWindows1254 || enc == EncodingWindows1258 || enc == EncodingWindows1257 ||
		enc == EncodingISO885913 || enc == EncodingISO88594
	if isLatinFamily {
		c := newLatinCandidate(&singleByteDataTable[idx], enc == EncodingWindows1252)
		s, ok := c.feed(buf)
		if !ok {
			return -1e9
		}
		// windows-1252 has a second candidate with icelandic profile in chardetng.
		if enc == EncodingWindows1252 {
			ci := newLatinCandidate(&singleByteDataTable[windows1252IcelandicIndex], true)
			s2, ok2 := ci.feed(buf)
			if ok2 && s2 > s {
				s = s2
			}
		}
		return float64(s)
	}
	switch enc {
	case EncodingWindows1251, EncodingKOI8U, EncodingISO88595, EncodingIBM866, EncodingWindows1253, EncodingISO88597:
		c := newNonLatinCasedCandidate(&singleByteDataTable[idx], enc == EncodingIBM866, enc == EncodingKOI8U)
		s, ok, longest := c.feed(buf)
		if !ok || longest < 2 {
			return -1e9
		}
		return float64(s)
	case EncodingWindows1256:
		c := newArabicFrenchCandidate(&singleByteDataTable[idx])
		s, ok, longest := c.feed(buf)
		if !ok || longest < 2 {
			return -1e9
		}
		return float64(s)
	case EncodingISO88596, EncodingWindows874:
		c := newCaselessCandidate(&singleByteDataTable[idx])
		s, ok, longest := c.feed(buf, enc == EncodingISO88596)
		if !ok || longest < 2 {
			return -1e9
		}
		return float64(s)
	case EncodingWindows1255, EncodingISO88598:
		if enc == EncodingWindows1255 {
			c := newLogicalCandidate(&singleByteDataTable[idx])
			s, ok, longest, _ := c.feed(buf)
			if !ok || longest < 2 {
				return -1e9
			}
			return float64(s)
		}
		c := newVisualCandidate(&singleByteDataTable[idx])
		s, ok, longest, _ := c.feed(buf)
		if !ok || longest < 2 {
			return -1e9
		}
		return float64(s)
	}
	return float64(scoreSingleByte(buf, &singleByteDataTable[idx], enc == EncodingWindows1256))
}

func likelyTurkishBytes(buf []byte) bool {
	for _, b := range buf {
		switch b {
		case 0xD0, 0xDD, 0xDE, 0xF0, 0xFD, 0xFE:
			return true
		}
	}
	return false
}

func likelyVietnameseBytes(buf []byte) bool {
	for _, b := range buf {
		// windows-1258 Vietnamese text usually contains at least one high byte diacritic.
		if b >= 0xC0 {
			return true
		}
	}
	return false
}

func hebrewVisualAssessment(buf []byte) (visualScore float64, visualPunctuation uint64, ok bool, logicalPunctuation uint64) {
	li := windows1255Index
	vi := iso88598Index
	l := newLogicalCandidate(&singleByteDataTable[li])
	ls, lok, ll, lp := l.feed(buf)
	if !lok || ll < 2 {
		return -1e9, 0, false, 0
	}
	v := newVisualCandidate(&singleByteDataTable[vi])
	vs, vok, vl, vp := v.feed(buf)
	if !vok || vl < 2 {
		return -1e9, 0, false, lp
	}
	_ = ls
	return float64(vs), vp, true, lp
}

func isSingleByteEncoding(enc Encoding) bool {
	switch enc {
	case EncodingWindows1252, EncodingWindows1251, EncodingWindows1250, EncodingISO88592,
		EncodingWindows1256, EncodingWindows1254, EncodingWindows874, EncodingWindows1255,
		EncodingISO88598, EncodingWindows1253, EncodingISO88597, EncodingWindows1257,
		EncodingISO885913, EncodingKOI8U, EncodingIBM866, EncodingISO88596, EncodingWindows1258,
		EncodingISO88594, EncodingISO88595:
		return true
	default:
		return false
	}
}

var westernHintSet = buildRuneSet("éèêàáâäïîôöùúüçñßœæøå")
var turkishHintSet = buildRuneSet("ğĞıİşŞçÇöÖüÜ")
var vietnameseHintSet = buildRuneSet("ăâđêôơưĂÂĐÊÔƠƯắằẳẵặấầẩẫậếềểễệốồổỗộớờởỡợứừửữự")
var centralBalticHintSet = buildRuneSet("ąćęłńóśźżčďěňřšťůžāēīķļņšūž")

func buildRuneSet(chars string) map[rune]struct{} {
	m := make(map[rune]struct{}, len(chars))
	for _, r := range chars {
		m[r] = struct{}{}
	}
	return m
}

func countRunesInSetBytes(b []byte, set map[rune]struct{}) int {
	n := 0
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		i += size
		if _, ok := set[r]; ok {
			n++
		}
	}
	return n
}

func scoreWesternHintFast(b []byte) float64 {
	return float64(countRunesInSetBytes(b, westernHintSet)) * 1.2
}

func scoreTurkishHintFast(b []byte) float64 {
	return float64(countRunesInSetBytes(b, turkishHintSet)) * 2
}

func scoreVietnameseHintFast(b []byte) float64 {
	return float64(countRunesInSetBytes(b, vietnameseHintSet)) * 2
}

func scoreCentralBalticHintFast(b []byte) float64 {
	return float64(countRunesInSetBytes(b, centralBalticHintSet)) * 1.6
}

func countNonASCII(b []byte) int {
	n := 0
	for _, c := range b {
		if c >= 0x80 {
			n++
		}
	}
	return n
}

func scoreKanaFromBytes(b []byte) float64 {
	n := 0
	for _, c := range b {
		if c >= 0xA1 && c <= 0xDF {
			n++
		}
	}
	return float64(n) * 1.5
}

func scoreMultibyteStructure(b []byte, enc Encoding) float64 {
	var ok, bad, soft int
	i := 0
	for i < len(b) {
		c := b[i]
		if c < 0x80 {
			i++
			continue
		}
		consumed, valid := consumeMultibyte(b[i:], enc)
		if consumed == 0 {
			bad++
			i++
			continue
		}
		if valid {
			if enc == EncodingShiftJIS && consumed == 1 {
				soft++
			} else {
				ok++
			}
		} else {
			bad++
		}
		i += consumed
	}
	if ok == 0 && bad == 0 {
		return 0
	}
	return float64(ok*20 + soft - bad*25)
}

func consumeMultibyte(b []byte, enc Encoding) (int, bool) {
	if len(b) == 0 {
		return 0, false
	}
	c := b[0]
	switch enc {
	case EncodingShiftJIS:
		if c >= 0xA1 && c <= 0xDF {
			return 1, true
		}
		if (c >= 0x81 && c <= 0x9F) || (c >= 0xE0 && c <= 0xFC) {
			if len(b) < 2 {
				return 1, false
			}
			t := b[1]
			if (t >= 0x40 && t <= 0x7E) || (t >= 0x80 && t <= 0xFC) {
				if t != 0x7F {
					return 2, true
				}
			}
			return 2, false
		}
		return 1, false
	case EncodingEUCJP:
		if c == 0x8E {
			if len(b) < 2 {
				return 1, false
			}
			return 2, b[1] >= 0xA1 && b[1] <= 0xDF
		}
		if c == 0x8F {
			if len(b) < 3 {
				return 1, false
			}
			return 3, b[1] >= 0xA1 && b[1] <= 0xFE && b[2] >= 0xA1 && b[2] <= 0xFE
		}
		if c >= 0xA1 && c <= 0xFE {
			if len(b) < 2 {
				return 1, false
			}
			return 2, b[1] >= 0xA1 && b[1] <= 0xFE
		}
		return 1, false
	case EncodingEUCKR:
		if c >= 0xA1 && c <= 0xFE {
			if len(b) < 2 {
				return 1, false
			}
			return 2, b[1] >= 0xA1 && b[1] <= 0xFE
		}
		return 1, false
	case EncodingGBK:
		if c >= 0x81 && c <= 0xFE {
			if len(b) < 2 {
				return 1, false
			}
			t := b[1]
			return 2, t >= 0x40 && t <= 0xFE && t != 0x7F
		}
		return 1, false
	case EncodingBig5:
		if c >= 0x81 && c <= 0xFE {
			if len(b) < 2 {
				return 1, false
			}
			t := b[1]
			return 2, (t >= 0x40 && t <= 0x7E) || (t >= 0xA1 && t <= 0xFE)
		}
		return 1, false
	default:
		return 0, false
	}
}

func applyTLDBias(scores *scoreArray, class tldClass) {
	boost := func(idx int, v float64) {
		scores[idx] += v
	}
	switch class {
	case tldGeneric:
		boost(encWindows1252, 12)
	case tldJapanese:
		boost(encShiftJIS, 320)
		boost(encEUCJP, 300)
		boost(encEUCKR, -140)
		boost(encGBK, -120)
		boost(encBig5, -120)
	case tldKorean:
		boost(encEUCKR, 140)
	case tldTraditional:
		boost(encBig5, 140)
	case tldSimplified:
		boost(encGBK, 140)
	case tldCyrillic, tldWesternCyrillic, tldCentralCyrillic:
		boost(encWindows1251, 90)
		boost(encKOI8U, 40)
		boost(encIBM866, 40)
	case tldGreek:
		boost(encWindows1253, 120)
		boost(encISO88597, 120)
	case tldArabic, tldWesternArabic:
		boost(encWindows1256, 90)
		boost(encISO88596, 90)
	case tldHebrew:
		boost(encWindows1255, 120)
		boost(encISO88598, 80)
	case tldThai:
		boost(encWindows874, 120)
	case tldVietnamese:
		boost(encWindows1258, 90)
	case tldTurkishAzeri:
		boost(encWindows1254, 200)
	case tldCentralWindows:
		boost(encWindows1250, 70)
	case tldCentralIso:
		boost(encISO88592, 70)
	case tldBaltic:
		boost(encWindows1257, 70)
		boost(encISO885913, 60)
		boost(encISO88594, 50)
	}
}

func defaultEncodingForTLD(class tldClass) Encoding {
	switch class {
	case tldJapanese:
		return EncodingShiftJIS
	case tldKorean:
		return EncodingEUCKR
	case tldTraditional, tldTraditionalSimplified:
		return EncodingBig5
	case tldSimplified, tldSimplifiedTraditional:
		return EncodingGBK
	case tldCyrillic, tldWesternCyrillic, tldCentralCyrillic:
		return EncodingWindows1251
	case tldGreek:
		return EncodingWindows1253
	case tldArabic, tldWesternArabic:
		return EncodingWindows1256
	case tldHebrew:
		return EncodingWindows1255
	case tldThai:
		return EncodingWindows874
	case tldVietnamese:
		return EncodingWindows1258
	case tldTurkishAzeri:
		return EncodingWindows1254
	case tldCentralWindows:
		return EncodingWindows1250
	case tldCentralIso:
		return EncodingISO88592
	case tldBaltic:
		return EncodingWindows1257
	default:
		return EncodingWindows1252
	}
}
