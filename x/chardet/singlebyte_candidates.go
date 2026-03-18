package chardet

const (
	implausibleLatinCaseTransitionPenalty int64 = -180
	ordinalBonus                          int64 = 300
	copyrightBonus                        int64 = 222
)

type latinCaseState uint8

const (
	latinCaseSpace latinCaseState = iota
	latinCaseUpper
	latinCaseLower
	latinCaseAllCaps
)

type ordinalState uint8

const (
	ordinalOther ordinalState = iota
	ordinalSpace
	ordinalPeriodAfterN
	ordinalExpectingSpace
	ordinalExpectingSpaceUndoImplausibility
	ordinalExpectingSpaceOrDigit
	ordinalExpectingSpaceOrDigitUndoImplausibility
	ordinalUpperN
	ordinalLowerN
	ordinalFeminineAbbreviationStartLetter
	ordinalDigit
	ordinalRoman
	ordinalCopyright
)

type latinCandidate struct {
	data         *singleByteData
	prev         uint8
	caseState    latinCaseState
	prevNonASCII uint32
	ordinalState ordinalState
	windows1252  bool
}

func newLatinCandidate(data *singleByteData, windows1252 bool) *latinCandidate {
	return &latinCandidate{
		data:         data,
		caseState:    latinCaseSpace,
		ordinalState: ordinalSpace,
		windows1252:  windows1252,
	}
}

func (c *latinCandidate) feed(buffer []byte) (int64, bool) {
	score := int64(0)
	for _, b := range buffer {
		class := c.data.classify(b)
		if class == 255 {
			return 0, false
		}
		caselessClass := class & 0x7F
		ascii := b < 0x80
		asciiPair := c.prevNonASCII == 0 && ascii

		nonASCIIPenalty := int64(0)
		switch c.prevNonASCII {
		case 0, 1, 2:
			nonASCIIPenalty = 0
		case 3:
			nonASCIIPenalty = -5
		case 4:
			nonASCIIPenalty = -20
		default:
			nonASCIIPenalty = -200
		}
		score += nonASCIIPenalty

		if !c.data.isLatinAlphabetic(caselessClass) {
			c.caseState = latinCaseSpace
		} else if (class >> 7) == 0 {
			if c.caseState == latinCaseAllCaps && !asciiPair {
				score += implausibleLatinCaseTransitionPenalty
			}
			c.caseState = latinCaseLower
		} else {
			switch c.caseState {
			case latinCaseSpace:
				c.caseState = latinCaseUpper
			case latinCaseUpper, latinCaseAllCaps:
				c.caseState = latinCaseAllCaps
			case latinCaseLower:
				if !asciiPair {
					score += implausibleLatinCaseTransitionPenalty
				}
				c.caseState = latinCaseUpper
			}
		}

		asciiIshPair := asciiPair || (ascii && c.prev == 0) || (caselessClass == 0 && c.prevNonASCII == 0)
		if !asciiIshPair {
			score += c.data.score(caselessClass, c.prev, false)
		}

		if c.windows1252 {
			score += c.feedOrdinal(b, caselessClass)
		}

		if ascii {
			c.prevNonASCII = 0
		} else {
			c.prevNonASCII++
		}
		c.prev = caselessClass
	}
	return score, true
}

func (c *latinCandidate) feedOrdinal(b byte, caselessClass uint8) int64 {
	score := int64(0)
	switch c.ordinalState {
	case ordinalOther:
		if caselessClass == 0 {
			c.ordinalState = ordinalSpace
		}
	case ordinalSpace:
		switch {
		case caselessClass == 0:
		case b == 0xAA || b == 0xBA:
			c.ordinalState = ordinalExpectingSpace
		case b == 'M' || b == 'D' || b == 'S':
			c.ordinalState = ordinalFeminineAbbreviationStartLetter
		case b == 'N':
			c.ordinalState = ordinalUpperN
		case b == 'n':
			c.ordinalState = ordinalLowerN
		case caselessClass == uint8(asciiDigit):
			c.ordinalState = ordinalDigit
		case caselessClass == 9 || caselessClass == 22 || caselessClass == 24:
			c.ordinalState = ordinalRoman
		case b == 0xA9:
			c.ordinalState = ordinalCopyright
		default:
			c.ordinalState = ordinalOther
		}
	case ordinalExpectingSpace:
		if caselessClass == 0 {
			score += ordinalBonus
			c.ordinalState = ordinalSpace
		} else {
			c.ordinalState = ordinalOther
		}
	case ordinalExpectingSpaceUndoImplausibility:
		if caselessClass == 0 {
			score += ordinalBonus - implausibilityPenalty
			c.ordinalState = ordinalSpace
		} else {
			c.ordinalState = ordinalOther
		}
	case ordinalExpectingSpaceOrDigit:
		if caselessClass == 0 || caselessClass == uint8(asciiDigit) {
			score += ordinalBonus
			if caselessClass == 0 {
				c.ordinalState = ordinalSpace
			} else {
				c.ordinalState = ordinalOther
			}
		} else {
			c.ordinalState = ordinalOther
		}
	case ordinalExpectingSpaceOrDigitUndoImplausibility:
		if caselessClass == 0 || caselessClass == uint8(asciiDigit) {
			score += ordinalBonus - implausibilityPenalty
			if caselessClass == 0 {
				c.ordinalState = ordinalSpace
			} else {
				c.ordinalState = ordinalOther
			}
		} else {
			c.ordinalState = ordinalOther
		}
	case ordinalUpperN:
		switch {
		case b == 0xAA:
			c.ordinalState = ordinalExpectingSpaceUndoImplausibility
		case b == 0xBA:
			c.ordinalState = ordinalExpectingSpaceOrDigitUndoImplausibility
		case b == '.':
			c.ordinalState = ordinalPeriodAfterN
		case caselessClass == 0:
			c.ordinalState = ordinalSpace
		default:
			c.ordinalState = ordinalOther
		}
	case ordinalLowerN:
		switch {
		case b == 0xBA:
			c.ordinalState = ordinalExpectingSpaceOrDigitUndoImplausibility
		case b == '.':
			c.ordinalState = ordinalPeriodAfterN
		case caselessClass == 0:
			c.ordinalState = ordinalSpace
		default:
			c.ordinalState = ordinalOther
		}
	case ordinalFeminineAbbreviationStartLetter:
		if b == 0xAA {
			c.ordinalState = ordinalExpectingSpaceUndoImplausibility
		} else if caselessClass == 0 {
			c.ordinalState = ordinalSpace
		} else {
			c.ordinalState = ordinalOther
		}
	case ordinalDigit:
		if b == 0xAA || b == 0xBA {
			c.ordinalState = ordinalExpectingSpace
		} else if caselessClass == 0 {
			c.ordinalState = ordinalSpace
		} else if caselessClass != uint8(asciiDigit) {
			c.ordinalState = ordinalOther
		}
	case ordinalRoman:
		if b == 0xAA || b == 0xBA {
			c.ordinalState = ordinalExpectingSpaceUndoImplausibility
		} else if caselessClass == 0 {
			c.ordinalState = ordinalSpace
		} else if caselessClass != 9 && caselessClass != 22 && caselessClass != 24 {
			c.ordinalState = ordinalOther
		}
	case ordinalPeriodAfterN:
		if b == 0xBA {
			c.ordinalState = ordinalExpectingSpaceOrDigit
		} else if caselessClass == 0 {
			c.ordinalState = ordinalSpace
		} else {
			c.ordinalState = ordinalOther
		}
	case ordinalCopyright:
		if caselessClass == 0 {
			score += copyrightBonus
			c.ordinalState = ordinalSpace
		} else {
			c.ordinalState = ordinalOther
		}
	}
	return score
}

const (
	latinLetter                 uint8 = 1
	latinAdjacencyPenalty       int64 = -50
	nonLatinCapitalizationBonus int64 = 40
	nonLatinAllCapsPenalty      int64 = -40
	nonLatinMixedCasePenalty    int64 = -20
)

type nonLatinCaseState uint8

const (
	nonLatinSpace nonLatinCaseState = iota
	nonLatinUpper
	nonLatinLower
	nonLatinUpperLower
	nonLatinAllCaps
	nonLatinMix
)

type nonLatinCasedCandidate struct {
	data           *singleByteData
	prev           uint8
	caseState      nonLatinCaseState
	prevASCII      bool
	currentWordLen uint64
	longestWord    uint64
	ibm866         bool
	koi8u          bool
	prevWasA0      bool
}

func newNonLatinCasedCandidate(data *singleByteData, ibm866, koi8u bool) *nonLatinCasedCandidate {
	return &nonLatinCasedCandidate{data: data, caseState: nonLatinSpace, prevASCII: true, ibm866: ibm866, koi8u: koi8u}
}

func (c *nonLatinCasedCandidate) feed(buffer []byte) (int64, bool, uint64) {
	score := int64(0)
	for _, b := range buffer {
		class := c.data.classify(b)
		if class == 255 {
			return 0, false, 0
		}
		caseless := class & 0x7F
		ascii := b < 0x80
		asciiPair := c.prevASCII && ascii
		nonASCIIAlpha := c.data.isNonLatinAlphabetic(caseless, false)

		if caseless == latinLetter {
			c.caseState = nonLatinMix
		} else if !nonASCIIAlpha {
			switch c.caseState {
			case nonLatinUpperLower:
				score += nonLatinCapitalizationBonus
			case nonLatinAllCaps:
				if c.koi8u {
					score += nonLatinAllCapsPenalty
				}
			case nonLatinMix:
				score += nonLatinMixedCasePenalty * int64(c.currentWordLen)
			}
			c.caseState = nonLatinSpace
		} else if (class >> 7) == 0 {
			switch c.caseState {
			case nonLatinSpace:
				c.caseState = nonLatinLower
			case nonLatinUpper:
				c.caseState = nonLatinUpperLower
			case nonLatinAllCaps:
				c.caseState = nonLatinMix
			}
		} else {
			switch c.caseState {
			case nonLatinSpace:
				c.caseState = nonLatinUpper
			case nonLatinUpper:
				c.caseState = nonLatinAllCaps
			case nonLatinLower, nonLatinUpperLower:
				c.caseState = nonLatinMix
			}
		}

		if nonASCIIAlpha {
			c.currentWordLen++
		} else {
			if c.currentWordLen > c.longestWord {
				c.longestWord = c.currentWordLen
			}
			c.currentWordLen = 0
		}

		isA0 := b == 0xA0
		if !asciiPair {
			ignore := c.ibm866 && ((isA0 && (c.prevWasA0 || c.prev == 0)) || (caseless == 0 && c.prevWasA0))
			if !ignore {
				score += c.data.score(caseless, c.prev, false)
			}
			if c.prev == latinLetter && nonASCIIAlpha {
				score += latinAdjacencyPenalty
			} else if caseless == latinLetter && c.data.isNonLatinAlphabetic(c.prev, false) {
				score += latinAdjacencyPenalty
			}
		}
		c.prevASCII = ascii
		c.prev = caseless
		c.prevWasA0 = isA0
	}
	if c.currentWordLen > c.longestWord {
		c.longestWord = c.currentWordLen
	}
	return score, true, c.longestWord
}

type caselessCandidate struct {
	data           *singleByteData
	prev           uint8
	prevASCII      bool
	currentWordLen uint64
	longestWord    uint64
}

func newCaselessCandidate(data *singleByteData) *caselessCandidate {
	return &caselessCandidate{data: data, prevASCII: true}
}

func (c *caselessCandidate) feed(buffer []byte, isWindows1256 bool) (int64, bool, uint64) {
	score := int64(0)
	for _, b := range buffer {
		class := c.data.classify(b)
		if class == 255 {
			return 0, false, 0
		}
		caseless := class & 0x7F
		ascii := b < 0x80
		asciiPair := c.prevASCII && ascii
		nonASCIIAlpha := c.data.isNonLatinAlphabetic(caseless, false)
		if nonASCIIAlpha {
			c.currentWordLen++
		} else {
			if c.currentWordLen > c.longestWord {
				c.longestWord = c.currentWordLen
			}
			c.currentWordLen = 0
		}
		if !asciiPair {
			score += c.data.score(caseless, c.prev, isWindows1256)
			if c.prev == latinLetter && nonASCIIAlpha {
				score += latinAdjacencyPenalty
			} else if caseless == latinLetter && c.data.isNonLatinAlphabetic(c.prev, isWindows1256) {
				score += latinAdjacencyPenalty
			}
		}
		c.prevASCII = ascii
		c.prev = caseless
	}
	if c.currentWordLen > c.longestWord {
		c.longestWord = c.currentWordLen
	}
	return score, true, c.longestWord
}

type arabicFrenchCandidate struct {
	data           *singleByteData
	prev           uint8
	caseState      latinCaseState
	prevASCII      bool
	currentWordLen uint64
	longestWord    uint64
}

func newArabicFrenchCandidate(data *singleByteData) *arabicFrenchCandidate {
	return &arabicFrenchCandidate{data: data, caseState: latinCaseSpace, prevASCII: true}
}

func (c *arabicFrenchCandidate) feed(buffer []byte) (int64, bool, uint64) {
	score := int64(0)
	for _, b := range buffer {
		class := c.data.classify(b)
		if class == 255 {
			return 0, false, 0
		}
		caseless := class & 0x7F
		ascii := b < 0x80
		asciiPair := c.prevASCII && ascii
		if caseless != latinLetter {
			c.caseState = latinCaseSpace
		} else if (class >> 7) == 0 {
			if c.caseState == latinCaseAllCaps && !asciiPair {
				score += implausibleLatinCaseTransitionPenalty
			}
			c.caseState = latinCaseLower
		} else {
			switch c.caseState {
			case latinCaseSpace:
				c.caseState = latinCaseUpper
			case latinCaseUpper, latinCaseAllCaps:
				c.caseState = latinCaseAllCaps
			case latinCaseLower:
				if !asciiPair {
					score += implausibleLatinCaseTransitionPenalty
				}
				c.caseState = latinCaseUpper
			}
		}
		nonASCIIAlpha := c.data.isNonLatinAlphabetic(caseless, true)
		if nonASCIIAlpha {
			c.currentWordLen++
		} else {
			if c.currentWordLen > c.longestWord {
				c.longestWord = c.currentWordLen
			}
			c.currentWordLen = 0
		}
		if !asciiPair {
			score += c.data.score(caseless, c.prev, true)
			if c.prev == latinLetter && nonASCIIAlpha {
				score += latinAdjacencyPenalty
			} else if caseless == latinLetter && c.data.isNonLatinAlphabetic(c.prev, true) {
				score += latinAdjacencyPenalty
			}
		}
		c.prevASCII = ascii
		c.prev = caseless
	}
	if c.currentWordLen > c.longestWord {
		c.longestWord = c.currentWordLen
	}
	return score, true, c.longestWord
}

func isASCIIPunctuation(b byte) bool {
	switch b {
	case '.', ',', ':', ';', '?', '!':
		return true
	default:
		return false
	}
}

type logicalCandidate struct {
	data                 *singleByteData
	prev                 uint8
	prevASCII            bool
	plausiblePunctuation uint64
	currentWordLen       uint64
	longestWord          uint64
}

func newLogicalCandidate(data *singleByteData) *logicalCandidate {
	return &logicalCandidate{data: data, prevASCII: true}
}

func (c *logicalCandidate) feed(buffer []byte) (int64, bool, uint64, uint64) {
	score := int64(0)
	for _, b := range buffer {
		class := c.data.classify(b)
		if class == 255 {
			return 0, false, 0, 0
		}
		caseless := class & 0x7F
		ascii := b < 0x80
		asciiPair := c.prevASCII && ascii
		nonASCIIAlpha := c.data.isNonLatinAlphabetic(caseless, false)
		if nonASCIIAlpha {
			c.currentWordLen++
		} else {
			if c.currentWordLen > c.longestWord {
				c.longestWord = c.currentWordLen
			}
			c.currentWordLen = 0
		}
		if !asciiPair {
			score += c.data.score(caseless, c.prev, false)
			prevNonASCIIAlpha := c.data.isNonLatinAlphabetic(c.prev, false)
			if caseless == 0 && prevNonASCIIAlpha && isASCIIPunctuation(b) {
				c.plausiblePunctuation++
			}
			if c.prev == latinLetter && nonASCIIAlpha {
				score += latinAdjacencyPenalty
			} else if caseless == latinLetter && prevNonASCIIAlpha {
				score += latinAdjacencyPenalty
			}
		}
		c.prevASCII = ascii
		c.prev = caseless
	}
	if c.currentWordLen > c.longestWord {
		c.longestWord = c.currentWordLen
	}
	return score, true, c.longestWord, c.plausiblePunctuation
}

type visualCandidate struct {
	data                 *singleByteData
	prev                 uint8
	prevASCII            bool
	prevPunctuation      bool
	plausiblePunctuation uint64
	currentWordLen       uint64
	longestWord          uint64
}

func newVisualCandidate(data *singleByteData) *visualCandidate {
	return &visualCandidate{data: data, prevASCII: true}
}

func (c *visualCandidate) feed(buffer []byte) (int64, bool, uint64, uint64) {
	score := int64(0)
	for _, b := range buffer {
		class := c.data.classify(b)
		if class == 255 {
			return 0, false, 0, 0
		}
		caseless := class & 0x7F
		ascii := b < 0x80
		asciiPair := c.prevASCII && ascii
		nonASCIIAlpha := c.data.isNonLatinAlphabetic(caseless, false)
		if nonASCIIAlpha {
			c.currentWordLen++
		} else {
			if c.currentWordLen > c.longestWord {
				c.longestWord = c.currentWordLen
			}
			c.currentWordLen = 0
		}
		if !asciiPair {
			score += c.data.score(caseless, c.prev, false)
			if nonASCIIAlpha && c.prevPunctuation {
				c.plausiblePunctuation++
			}
			if c.prev == latinLetter && nonASCIIAlpha {
				score += latinAdjacencyPenalty
			} else if caseless == latinLetter && c.data.isNonLatinAlphabetic(c.prev, false) {
				score += latinAdjacencyPenalty
			}
		}
		c.prevASCII = ascii
		c.prev = caseless
		c.prevPunctuation = caseless == 0 && isASCIIPunctuation(b)
	}
	if c.currentWordLen > c.longestWord {
		c.longestWord = c.currentWordLen
	}
	return score, true, c.longestWord, c.plausiblePunctuation
}
