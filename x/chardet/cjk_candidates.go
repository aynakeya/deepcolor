package chardet

const (
	cjkBaseScore             = 41
	cjkSecondaryBaseScore    = 20
	cjkOtherScore            = cjkSecondaryBaseScore / 4
	cjkLatinAdjacencyPenalty = -cjkBaseScore
	cjPunctuationScore       = cjkBaseScore / 2
	shiftJisPUAPenalty       = -(cjkBaseScore * 10)
	shiftJisExtensionPenalty = shiftJisPUAPenalty * 2
	gbkPUAPenalty            = -(cjkBaseScore * 10)
	big5PUAPenalty           = -(cjkBaseScore * 30)
)

type latinCJKState uint8

const (
	latinCJKOther latinCJKState = iota
	latinCJKAscii
	latinCJKCJ
)

func scoreCJKCandidate(buf []byte, enc Encoding) float64 {
	score := 0.0
	switch enc {
	case EncodingShiftJIS:
		score += scoreShiftJISCandidate(buf)
	case EncodingEUCJP:
		score += scoreEUCJPCandidate(buf)
	case EncodingEUCKR:
		score += scoreEUCKRCandidate(buf)
	case EncodingGBK:
		score += scoreGBKCandidate(buf)
	case EncodingBig5:
		score += scoreBig5Candidate(buf)
	}
	return score
}

func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func problematicLead(b byte) bool {
	switch b {
	case 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x9A, 0x8A, 0x9B, 0x8B, 0x9E, 0x8E, 0xB0:
		return true
	default:
		return false
	}
}

func moreProblematicLead(b byte) bool {
	return problematicLead(b) || b == 0x82 || b == 0x84 || b == 0x85 || b == 0xA0
}

func scoreShiftJISCandidate(buf []byte) float64 {
	score := 0.0
	state := latinCJKOther
	halfWidthSeen := false
	halfCount := 0
	cjCount := 0
	badCount := 0
	shiftExclusivePairs := 0
	prevByte := byte(0)
	pending := 0.0
	hasPending := false
	maybePending := func(s float64) float64 {
		if state == latinCJKCJ || !problematicLead(prevByte) {
			return s
		}
		pending = s
		hasPending = true
		return 0
	}
	flushPending := func() {
		if hasPending {
			score += pending
			hasPending = false
			pending = 0
		}
	}
	for i := 0; i < len(buf); {
		b := buf[i]
		if isASCIILetter(b) {
			hasPending = false
			if state == latinCJKCJ {
				score += cjkLatinAdjacencyPenalty
			}
			state = latinCJKAscii
			prevByte = b
			i++
			continue
		}
		if b < 0x80 {
			hasPending = false
			state = latinCJKOther
			prevByte = b
			i++
			continue
		}
		if b >= 0xA1 && b <= 0xDF {
			hasPending = false
			if !halfWidthSeen {
				halfWidthSeen = true
				score -= 75
			}
			score += 1
			halfCount++
			cjCount++
			if state == latinCJKAscii {
				score += cjkLatinAdjacencyPenalty
			}
			state = latinCJKCJ
			prevByte = b
			i++
			continue
		}
		if ((b >= 0x81 && b <= 0x9F) || (b >= 0xE0 && b <= 0xFC)) && i+1 < len(buf) {
			t := buf[i+1]
			if ((t >= 0x40 && t <= 0x7E) || (t >= 0x80 && t <= 0xFC)) && t != 0x7F {
				if b < 0xA1 {
					shiftExclusivePairs++
				}
				flushPending()
				if b < 0x98 || (b == 0x98 && t < 0x73) {
					score += maybePending(cjkBaseScore)
				} else {
					score += maybePending(cjkSecondaryBaseScore)
				}
				cjCount++
				if state == latinCJKAscii {
					score += cjkLatinAdjacencyPenalty
				}
				state = latinCJKCJ
				prevByte = t
				i += 2
				continue
			}
		}
		if ((prevByte >= 0x81 && prevByte <= 0x9F) || (prevByte >= 0xE0 && prevByte <= 0xFC)) &&
			((b >= 0x40 && b <= 0x7E) || (b >= 0x80 && b <= 0xFC)) {
			flushPending()
			score += shiftJisExtensionPenalty
		}
		if b == 0xA0 || b >= 0xFD {
			score -= float64(cjkBaseScore * 20)
		} else {
			score -= float64(cjkBaseScore * 10)
		}
		badCount++
		state = latinCJKOther
		prevByte = b
		i++
	}
	if halfCount >= 6 {
		score += 80
	}
	if shiftExclusivePairs > 0 {
		score += float64(shiftExclusivePairs * 10)
	}
	if cjCount < 2 {
		score -= 220
	}
	score -= float64(badCount * 10)
	return score
}

func scoreEUCJPCandidate(buf []byte) float64 {
	score := 0.0
	state := latinCJKOther
	nonASCIISeen := false
	cjCount := 0
	badCount := 0
	halfLike := 0
	kanaPrefix := 0
	nonEUCLeadLike := 0
	for i := 0; i < len(buf); {
		b := buf[i]
		if isASCIILetter(b) {
			if state == latinCJKCJ {
				score += cjkLatinAdjacencyPenalty
			}
			state = latinCJKAscii
			i++
			continue
		}
		if b < 0x80 {
			state = latinCJKOther
			i++
			continue
		}
		nonASCIISeen = true
		if (b >= 0x81 && b <= 0x9F) || (b >= 0xE0 && b <= 0xFC) {
			nonEUCLeadLike++
		}
		if b >= 0xA1 && b <= 0xDF {
			halfLike++
		}
		if b == 0x8E && i+1 < len(buf) {
			t := buf[i+1]
			if t >= 0xA1 && t <= 0xDF {
				kanaPrefix++
				score += 1
				cjCount++
				if state == latinCJKAscii {
					score += cjkLatinAdjacencyPenalty
				}
				state = latinCJKCJ
				i += 2
				continue
			}
		}
		if b == 0x8F && i+2 < len(buf) {
			b1, b2 := buf[i+1], buf[i+2]
			if b1 >= 0xA1 && b1 <= 0xFE && b2 >= 0xA1 && b2 <= 0xFE {
				score += cjkOtherScore
				cjCount++
				if state == latinCJKAscii {
					score += cjkLatinAdjacencyPenalty
				}
				state = latinCJKCJ
				i += 3
				continue
			}
		}
		if b >= 0xA1 && b <= 0xFE && i+1 < len(buf) {
			t := buf[i+1]
			if t >= 0xA1 && t <= 0xFE {
				if b < 0xD0 {
					score += cjkBaseScore
				} else {
					score += cjkSecondaryBaseScore
				}
				cjCount++
				if state == latinCJKAscii {
					score += cjkLatinAdjacencyPenalty
				}
				state = latinCJKCJ
				i += 2
				continue
			}
		}
		score -= float64(cjkBaseScore * 30)
		badCount++
		state = latinCJKOther
		i++
	}
	if !nonASCIISeen {
		score -= 20
	}
	// In real EUC-JP half-width katakana must follow 0x8E.
	if halfLike > kanaPrefix*3 {
		score -= float64((halfLike - kanaPrefix*3) * 6)
	}
	if nonEUCLeadLike > 0 {
		score -= float64(nonEUCLeadLike * 6)
	}
	if cjCount < 2 {
		score -= 220
	}
	score -= float64(badCount * 10)
	return score
}

func scoreEUCKRCandidate(buf []byte) float64 {
	score := 0.0
	state := latinCJKOther
	cjCount := 0
	badCount := 0
	for i := 0; i < len(buf); {
		b := buf[i]
		if isASCIILetter(b) {
			if state == latinCJKCJ {
				score += cjkLatinAdjacencyPenalty
			}
			state = latinCJKAscii
			i++
			continue
		}
		if b < 0x80 {
			state = latinCJKOther
			i++
			continue
		}
		if b >= 0xA1 && b <= 0xFE && i+1 < len(buf) {
			t := buf[i+1]
			if t >= 0xA1 && t <= 0xFE {
				score += cjkBaseScore + 1
				cjCount++
				if state == latinCJKAscii {
					score += cjkLatinAdjacencyPenalty
				}
				state = latinCJKCJ
				i += 2
				continue
			}
		}
		if b == 0x80 || b == 0xFF || (b >= 0x81 && b <= 0x84) {
			score -= float64(cjkBaseScore * 40)
		} else {
			score -= float64(cjkBaseScore * 10)
		}
		badCount++
		state = latinCJKOther
		i++
	}
	if cjCount < 2 {
		score -= 220
	}
	score -= float64(badCount * 10)
	return score
}

func scoreGBKCandidate(buf []byte) float64 {
	score := 0.0
	state := latinCJKOther
	cjCount := 0
	badCount := 0
	prevByte := byte(0)
	pending := 0.0
	hasPending := false
	maybePending := func(s float64) float64 {
		if state == latinCJKCJ || !moreProblematicLead(prevByte) {
			return s
		}
		pending = s
		hasPending = true
		return 0
	}
	flushPending := func() {
		if hasPending {
			score += pending
			hasPending = false
			pending = 0
		}
	}
	for i := 0; i < len(buf); {
		b := buf[i]
		if isASCIILetter(b) {
			hasPending = false
			if state == latinCJKCJ {
				score += cjkLatinAdjacencyPenalty
			}
			state = latinCJKAscii
			prevByte = b
			i++
			continue
		}
		if b < 0x80 {
			hasPending = false
			state = latinCJKOther
			prevByte = b
			i++
			continue
		}
		if b >= 0x81 && b <= 0xFE && i+1 < len(buf) {
			t := buf[i+1]
			if t >= 0x40 && t <= 0xFE && t != 0x7F {
				flushPending()
				if b >= 0xA1 && b <= 0xD7 {
					score += maybePending(cjkBaseScore)
				} else if b >= 0xD8 {
					score += maybePending(cjkSecondaryBaseScore)
				} else {
					score += maybePending(cjkOtherScore)
				}
				cjCount++
				if state == latinCJKAscii {
					score += cjkLatinAdjacencyPenalty
				}
				state = latinCJKCJ
				prevByte = t
				i += 2
				continue
			}
		}
		if ((prevByte == 0xA0 || prevByte == 0xFE || prevByte == 0xFD) && (b < 0x80 || b == 0xFF)) || b == 0xFF {
			flushPending()
			score += shiftJisExtensionPenalty
		}
		if b == 0xA0 || b == 0xFE || b == 0xFD || b == 0xFF {
			score += gbkPUAPenalty
		} else {
			score -= float64(cjkBaseScore * 10)
		}
		badCount++
		state = latinCJKOther
		prevByte = b
		i++
	}
	score += scoreCJKFrequency(bytesMustDecode(buf, EncodingGBK), EncodingGBK)
	if cjCount < 2 {
		score -= 220
	}
	score -= float64(badCount * 10)
	return score
}

func scoreBig5Candidate(buf []byte) float64 {
	score := 0.0
	state := latinCJKOther
	cjCount := 0
	badCount := 0
	prevByte := byte(0)
	pending := 0.0
	hasPending := false
	maybePending := func(s float64) float64 {
		if state == latinCJKCJ || !problematicLead(prevByte) {
			return s
		}
		pending = s
		hasPending = true
		return 0
	}
	flushPending := func() {
		if hasPending {
			score += pending
			hasPending = false
			pending = 0
		}
	}
	for i := 0; i < len(buf); {
		b := buf[i]
		if isASCIILetter(b) {
			hasPending = false
			if state == latinCJKCJ {
				score += cjkLatinAdjacencyPenalty
			}
			state = latinCJKAscii
			prevByte = b
			i++
			continue
		}
		if b < 0x80 {
			hasPending = false
			state = latinCJKOther
			prevByte = b
			i++
			continue
		}
		if b >= 0x81 && b <= 0xFE && i+1 < len(buf) {
			t := buf[i+1]
			if (t >= 0x40 && t <= 0x7E) || (t >= 0xA1 && t <= 0xFE) {
				flushPending()
				if b >= 0xA4 && b <= 0xC6 {
					score += maybePending(cjkBaseScore)
				} else {
					score += maybePending(cjkSecondaryBaseScore)
				}
				cjCount++
				if state == latinCJKAscii {
					score += cjkLatinAdjacencyPenalty
				}
				state = latinCJKCJ
				prevByte = t
				i += 2
				continue
			}
		}
		if (prevByte >= 0x81 && prevByte <= 0xFE) && ((b >= 0x40 && b <= 0x7E) || (b >= 0xA1 && b <= 0xFE)) {
			flushPending()
			score += big5PUAPenalty
		}
		if b == 0xA0 || b == 0xFD || b == 0xFE || b == 0xFF {
			score += big5PUAPenalty
		} else {
			score -= float64(cjkBaseScore * 10)
		}
		badCount++
		state = latinCJKOther
		prevByte = b
		i++
	}
	score += scoreCJKFrequency(bytesMustDecode(buf, EncodingBig5), EncodingBig5)
	if cjCount < 2 {
		score -= 220
	}
	score -= float64(badCount * 10)
	return score
}

func bytesMustDecode(buf []byte, enc Encoding) []byte {
	impl := encodingImpl(enc)
	if impl == nil {
		return nil
	}
	b, err := impl.NewDecoder().Bytes(buf)
	if err != nil {
		return b
	}
	return b
}
