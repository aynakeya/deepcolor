package chardet

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

func encode(t *testing.T, e encoding.Encoding, s string) []byte {
	t.Helper()
	b, err := e.NewEncoder().Bytes([]byte(s))
	require.NoError(t, err)
	return b
}

func TestUTF8Guess(t *testing.T) {
	d := NewDetector()
	d.Feed([]byte("这是 UTF-8 测试"), true)
	r := d.GuessAssess(nil, true)
	require.Equal(t, EncodingUTF8, r.Encoding)
	require.True(t, r.Confident)
}

func TestUTF8DisallowedFallback(t *testing.T) {
	d := NewDetector()
	d.Feed([]byte("这是 UTF-8 测试"), true)
	r := d.GuessAssess(nil, false)
	require.NotEqual(t, EncodingUTF8, r.Encoding)
}

func TestISO2022JP(t *testing.T) {
	d := NewDetector()
	d.Feed([]byte("\x1B$BF|K\x1B(B"), true)
	require.Equal(t, EncodingISO2022JP, d.Guess(nil, false))
}

func TestShiftJIS(t *testing.T) {
	d := NewDetector()
	d.Feed(encode(t, japanese.ShiftJIS, "これは文字実験です。"), true)
	require.Contains(t, []Encoding{EncodingShiftJIS, EncodingEUCJP}, d.Guess([]byte("jp"), false))
}

func TestEUCJP(t *testing.T) {
	d := NewDetector()
	d.Feed(encode(t, japanese.EUCJP, "これは文字実験です。"), true)
	require.Contains(t, []Encoding{EncodingShiftJIS, EncodingEUCJP}, d.Guess([]byte("jp"), false))
}

func TestGBK(t *testing.T) {
	d := NewDetector()
	d.Feed(encode(t, simplifiedchinese.GBK, "这是一个字符编码测试。"), true)
	require.Contains(t, []Encoding{EncodingGBK, EncodingBig5, EncodingShiftJIS}, d.Guess([]byte("cn"), false))
}

func TestGB18030(t *testing.T) {
	d := NewDetector()
	d.Feed(encode(t, simplifiedchinese.GB18030, "数据库名：c播拨龾龿珳珴𬀩𬀪"), true)
	require.Equal(t, EncodingGB18030, d.Guess([]byte("cn"), false))
}

func TestBig5(t *testing.T) {
	d := NewDetector()
	d.Feed(encode(t, traditionalchinese.Big5, "這是一個字符編碼測試。"), true)
	require.Contains(t, []Encoding{EncodingBig5, EncodingGBK, EncodingShiftJIS}, d.Guess([]byte("tw"), false))
}

func TestEUCKR(t *testing.T) {
	d := NewDetector()
	d.Feed(encode(t, korean.EUCKR, "이것은 문자 인코딩 테스트입니다."), true)
	require.Contains(t, []Encoding{EncodingEUCKR, EncodingShiftJIS}, d.Guess([]byte("kr"), false))
}

func TestWindows1251(t *testing.T) {
	d := NewDetector()
	d.Feed(encode(t, charmap.Windows1251, "Это тест кодировки символов."), true)
	require.Contains(t, []Encoding{EncodingWindows1251, EncodingKOI8U, EncodingIBM866, EncodingISO88595, EncodingBig5, EncodingGBK}, d.Guess([]byte("ru"), false))
}

func TestWindows1256(t *testing.T) {
	d := NewDetector()
	d.Feed(encode(t, charmap.Windows1256, "هذا هو اختبار ترميز الأحرف."), true)
	require.Contains(t, []Encoding{EncodingWindows1256, EncodingISO88596, EncodingShiftJIS}, d.Guess([]byte("sa"), false))
}

func TestStreaming(t *testing.T) {
	d := NewDetector()
	require.False(t, d.Feed([]byte("ASCII only "), false))
	require.False(t, d.Feed([]byte("still ascii"), false))
	require.True(t, d.Feed([]byte{0xE4, 0xB8, 0xAD}, true))
}

func TestTLDMayAffectGuess(t *testing.T) {
	require.True(t, TLDMayAffectGuess([]byte("jp")))
	require.False(t, TLDMayAffectGuess([]byte("com")))
}

func TestInvalidTLDPanics(t *testing.T) {
	require.Panics(t, func() { TLDMayAffectGuess([]byte("CN")) })
	require.Panics(t, func() { TLDMayAffectGuess([]byte("c.n")) })
}

func TestWindows1252OrdinalPatterns(t *testing.T) {
	cases := []string{" © ", " º ", " ª ", ".º ", ".ª ", "Nº1", "Nº", " Mª ", " Dª ", " Sª "}
	for _, s := range cases {
		b, err := charmap.Windows1252.NewEncoder().Bytes([]byte(s))
		require.NoError(t, err)
		d := NewDetector()
		d.Feed(b, true)
		require.Equal(t, EncodingWindows1252, d.Guess(nil, false), s)
	}
}

func TestA0NotIBM866(t *testing.T) {
	b, err := charmap.Windows1252.NewEncoder().Bytes([]byte("\u00a0\u00a0 \u00a0"))
	require.NoError(t, err)
	d := NewDetector()
	d.Feed(b, true)
	require.Equal(t, EncodingWindows1252, d.Guess(nil, false))
}

func TestRussianAndGreekShort(t *testing.T) {
	ru, err := charmap.Windows1251.NewEncoder().Bytes([]byte("Русский"))
	require.NoError(t, err)
	d1 := NewDetector()
	d1.Feed(ru, true)
	require.Contains(t, []Encoding{EncodingWindows1251, EncodingKOI8U, EncodingISO88595, EncodingIBM866}, d1.Guess([]byte("ru"), false))

	el, err := charmap.Windows1253.NewEncoder().Bytes([]byte("Ελληνικά"))
	require.NoError(t, err)
	d2 := NewDetector()
	d2.Feed(el, true)
	require.Contains(t, []Encoding{EncodingWindows1253, EncodingISO88597}, d2.Guess([]byte("gr"), false))
}

func TestHebrewLogicalVisual(t *testing.T) {
	logicalBytes, err := charmap.Windows1255.NewEncoder().Bytes([]byte("עברית"))
	require.NoError(t, err)
	d1 := NewDetector()
	d1.Feed(logicalBytes, true)
	require.Equal(t, EncodingWindows1255, d1.Guess([]byte("il"), false))

	visualBytes, err := charmap.ISO8859_8.NewEncoder().Bytes([]byte(".םיוות דודיק ןחבמ והז"))
	require.NoError(t, err)
	d2 := NewDetector()
	d2.Feed(visualBytes, true)
	require.Equal(t, EncodingISO88598, d2.Guess([]byte("il"), false))
}

func TestCJKSpecialCases(t *testing.T) {
	t.Run("ShiftJISHalfWidthKatakana", func(t *testing.T) {
		b, err := japanese.ShiftJIS.NewEncoder().Bytes([]byte("ﾊｰﾄﾞｳｪｱﾊｰﾄﾞｳｪｱﾊｰﾄﾞｳｪｱ"))
		require.NoError(t, err)
		d := NewDetector()
		d.Feed(b, true)
		require.Equal(t, EncodingShiftJIS, d.Guess([]byte("jp"), false))
	})

	t.Run("GBKSingleByteFF", func(t *testing.T) {
		b := []byte{0xFF}
		for i := 0; i < 80; i++ {
			b = append(b, 0xB5, 0xC4)
		}
		d := NewDetector()
		d.Feed(b, true)
		require.Equal(t, EncodingGBK, d.Guess([]byte("cn"), false))
	})

	t.Run("Big5PUA", func(t *testing.T) {
		b := make([]byte, 0, 100)
		for i := 0; i < 40; i++ {
			b = append(b, 0xA4, 0x40)
		}
		b = append(b, 0x81, 0x40, 0xA4, 0x40)
		d := NewDetector()
		d.Feed(b, true)
		require.Equal(t, EncodingBig5, d.Guess([]byte("tw"), false))
	})
}

func TestDetectAPI(t *testing.T) {
	data := encode(t, simplifiedchinese.GBK, "这是一个字符编码测试。")
	require.Equal(t, EncodingGBK, Detect(data, []byte("cn"), false))
	r := DetectAssess(data, []byte("cn"), false)
	require.Equal(t, EncodingGBK, r.Encoding)
}

func TestDetectAPIConcurrent(t *testing.T) {
	data := encode(t, simplifiedchinese.GB18030, "数据库名：c播拨龾龿珳珴𬀩𬀪")
	const n = 128
	var wg sync.WaitGroup
	wg.Add(n)
	errCh := make(chan string, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if got := Detect(data, []byte("cn"), false); got != EncodingGB18030 {
				errCh <- string(got)
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for got := range errCh {
		require.Fail(t, "unexpected concurrent detect result", got)
	}
}
