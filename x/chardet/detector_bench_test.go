package chardet

import (
	"sync"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

type benchItem struct {
	tld       []byte
	allowUTF8 bool
	data      []byte
}

var (
	benchItemsOnce sync.Once
	benchItemsData []benchItem
)

func buildBenchItems() []benchItem {
	tryEncode := func(enc encoding.Encoding, s string) ([]byte, bool) {
		if enc == nil {
			return []byte(s), true
		}
		b, err := enc.NewEncoder().Bytes([]byte(s))
		if err != nil {
			return nil, false
		}
		return b, true
	}
	items := []benchItem{
		{tld: []byte("com"), allowUTF8: true, data: []byte("这是 UTF-8 测试 mixed English 123")},
	}
	add := func(tld string, allow bool, enc encoding.Encoding, text string) {
		b, ok := tryEncode(enc, text)
		if !ok {
			return
		}
		items = append(items, benchItem{tld: []byte(tld), allowUTF8: allow, data: b})
	}
	add("jp", false, japanese.ISO2022JP, "日本語のエンコーディング検出テストです。")
	add("jp", false, japanese.ShiftJIS, "これは文字実験です。")
	add("jp", false, japanese.EUCJP, "日本語のテキストです。")
	add("kr", false, korean.EUCKR, "이것은 문자 인코딩 테스트입니다.")
	add("cn", false, simplifiedchinese.GBK, "这是一个字符编码测试。")
	add("tw", false, traditionalchinese.Big5, "這是一個字符編碼測試。")
	add("ru", false, charmap.Windows1251, "Это тест кодировки символов.")
	add("ru", false, charmap.KOI8U, "Це тест на кодування символів.")
	add("ru", false, charmap.CodePage866, "Это тест кодировки символов.")
	add("ru", false, charmap.ISO8859_5, "Это тест кодировки символов.")
	add("gr", false, charmap.Windows1253, "Πρόκειται για δοκιμή κωδικοποίησης χαρακτήρων")
	add("gr", false, charmap.ISO8859_7, "Πρόκειται για δοκιμή κωδικοποίησης χαρακτήρων")
	add("sa", false, charmap.Windows1256, "هذا هو اختبار ترميز الأحرف.")
	add("sa", false, charmap.ISO8859_6, "هذا هو اختبار ترميز الأحرف.")
	add("il", false, charmap.Windows1255, "עברית")
	add("il", false, charmap.ISO8859_8, ".םיוות דודיק ןחבמ והז")
	add("tr", false, charmap.Windows1254, "Turkce test isguc")
	add("vn", false, charmap.Windows1258, "Tieng Viet ma hoa ky tu")
	add("pl", false, charmap.Windows1250, "To jest test kodowania znakow.")
	add("pl", false, charmap.ISO8859_2, "To jest test kodowania znakow.")
	add("lv", false, charmap.Windows1257, "Sis ir rakstzimju kodesanas tests.")
	add("lv", false, charmap.ISO8859_13, "Sis ir rakstzimju kodesanas tests.")
	add("lv", false, charmap.ISO8859_4, "Sis ir rakstzimju kodesanas tests.")
	add("th", false, charmap.Windows874, "ทดสอบการเข้ารหัส")
	return items
}

func getBenchItems() []benchItem {
	benchItemsOnce.Do(func() {
		benchItemsData = buildBenchItems()
	})
	return benchItemsData
}

func BenchmarkDetectorGuessMixed(b *testing.B) {
	items := getBenchItems()
	d := NewDetector()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := items[i%len(items)]
		d.Reset()
		d.Feed(it.data, true)
		_ = d.Guess(it.tld, it.allowUTF8)
	}
}

func BenchmarkScoreDecodedByEncoding(b *testing.B) {
	items := getBenchItems()
	if len(items) == 0 {
		b.Fatal("no bench items")
	}
	src := items[0].data
	for _, it := range items {
		if len(it.data) > len(src) {
			src = it.data
		}
	}
	encs := []Encoding{
		EncodingShiftJIS, EncodingEUCJP, EncodingEUCKR, EncodingGBK, EncodingBig5,
		EncodingWindows1252, EncodingWindows1251, EncodingWindows1250, EncodingISO88592,
		EncodingWindows1256, EncodingISO88596, EncodingWindows1255, EncodingISO88598,
		EncodingWindows1253, EncodingISO88597, EncodingWindows1254, EncodingWindows1258,
		EncodingWindows1257, EncodingISO885913, EncodingISO88594, EncodingWindows874,
		EncodingKOI8U, EncodingIBM866, EncodingISO88595,
	}
	for _, enc := range encs {
		b.Run(string(enc), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = scoreDecoded(src, enc)
			}
		})
	}
}
