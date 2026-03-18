package chardet

import "strings"

// Encoding is a WHATWG-like encoding label.
type Encoding string

const (
	EncodingUTF8      Encoding = "utf-8"
	EncodingISO2022JP Encoding = "iso-2022-jp"
	EncodingShiftJIS  Encoding = "shift_jis"
	EncodingEUCJP     Encoding = "euc-jp"
	EncodingEUCKR     Encoding = "euc-kr"
	EncodingGBK       Encoding = "gbk"
	EncodingBig5      Encoding = "big5"

	EncodingWindows1252 Encoding = "windows-1252"
	EncodingWindows1251 Encoding = "windows-1251"
	EncodingWindows1250 Encoding = "windows-1250"
	EncodingISO88592    Encoding = "iso-8859-2"
	EncodingWindows1256 Encoding = "windows-1256"
	EncodingWindows1254 Encoding = "windows-1254"
	EncodingWindows874  Encoding = "windows-874"
	EncodingWindows1255 Encoding = "windows-1255"
	EncodingISO88598    Encoding = "iso-8859-8"
	EncodingWindows1253 Encoding = "windows-1253"
	EncodingISO88597    Encoding = "iso-8859-7"
	EncodingWindows1257 Encoding = "windows-1257"
	EncodingISO885913   Encoding = "iso-8859-13"
	EncodingKOI8U       Encoding = "koi8-u"
	EncodingIBM866      Encoding = "ibm866"
	EncodingISO88596    Encoding = "iso-8859-6"
	EncodingWindows1258 Encoding = "windows-1258"
	EncodingISO88594    Encoding = "iso-8859-4"
	EncodingISO88595    Encoding = "iso-8859-5"
)

func normalizeEncodingLabel(s string) Encoding {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "")
	switch s {
	case "utf8":
		return EncodingUTF8
	case "iso2022jp":
		return EncodingISO2022JP
	case "shift-jis", "cp932", "windows-31j":
		return EncodingShiftJIS
	case "eucjp":
		return EncodingEUCJP
	case "euckr", "ksc5601":
		return EncodingEUCKR
	case "windows1252", "cp1252":
		return EncodingWindows1252
	case "windows1251", "cp1251":
		return EncodingWindows1251
	case "windows1250", "cp1250":
		return EncodingWindows1250
	case "windows1256", "cp1256":
		return EncodingWindows1256
	case "windows1254", "cp1254":
		return EncodingWindows1254
	case "windows874", "cp874":
		return EncodingWindows874
	case "windows1255", "cp1255":
		return EncodingWindows1255
	case "windows1253", "cp1253":
		return EncodingWindows1253
	case "windows1257", "cp1257":
		return EncodingWindows1257
	case "windows1258", "cp1258":
		return EncodingWindows1258
	case "iso8859-2":
		return EncodingISO88592
	case "iso8859-4":
		return EncodingISO88594
	case "iso8859-5":
		return EncodingISO88595
	case "iso8859-6":
		return EncodingISO88596
	case "iso8859-7":
		return EncodingISO88597
	case "iso8859-8":
		return EncodingISO88598
	case "iso8859-13":
		return EncodingISO885913
	case "ibm-866", "cp866":
		return EncodingIBM866
	}
	return Encoding(s)
}
