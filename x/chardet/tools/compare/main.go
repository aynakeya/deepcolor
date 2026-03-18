package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aynakeya/deepcolor/x/chardet"
	saintchardet "github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

type sampleDef struct {
	Encoding  string
	TLD       string
	AllowUTF8 bool
	Texts     []string
}

type datasetItem struct {
	ID        string
	Expected  string
	TLD       string
	TLDBytes  []byte
	AllowUTF8 bool
	Bytes     []byte
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: compare <generate|compare>")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "generate":
		must(runGenerate(os.Args[2:]))
	case "compare":
		must(runCompare(os.Args[2:]))
	default:
		fmt.Fprintln(os.Stderr, "usage: compare <generate|compare>")
		os.Exit(2)
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	out := fs.String("out", "dataset.tsv", "output tsv")
	repeat := fs.Int("repeat", 20, "repeat multiplier")
	if err := fs.Parse(args); err != nil {
		return err
	}

	items := make([]datasetItem, 0)
	id := 0
	for _, d := range defs() {
		enc := getEncoding(d.Encoding)
		if enc == nil && d.Encoding != "utf-8" {
			return fmt.Errorf("unknown encoding %s", d.Encoding)
		}
		for r := 0; r < *repeat; r++ {
			for _, t := range d.Texts {
				var b []byte
				var err error
				if d.Encoding == "utf-8" {
					b = []byte(t)
				} else {
					b, err = enc.NewEncoder().Bytes([]byte(t))
					if err != nil {
						continue
					}
				}
				id++
				items = append(items, datasetItem{ID: fmt.Sprintf("%06d", id), Expected: d.Encoding, TLD: d.TLD, AllowUTF8: d.AllowUTF8, Bytes: b})
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(*out)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	fmt.Fprintln(w, "id\texpected\ttld\tallow_utf8\thex")
	for _, it := range items {
		allow := "0"
		if it.AllowUTF8 {
			allow = "1"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", it.ID, it.Expected, it.TLD, allow, hex.EncodeToString(it.Bytes))
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Printf("generated %d samples -> %s\n", len(items), *out)
	return nil
}

func runCompare(args []string) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	data := fs.String("data", "dataset.tsv", "dataset tsv")
	rustBin := fs.String("rust-bin", "tools/rust_detect_tsv/target/release/detect_tsv", "rust binary")
	tmpOut := fs.String("rust-out", "rust_out.tsv", "rust output")
	report := fs.String("report", "last_report.txt", "report output file")
	history := fs.String("history", "history_report.log", "append-only history report")
	if err := fs.Parse(args); err != nil {
		return err
	}

	items, err := readDataset(*data)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return errors.New("empty dataset")
	}

	startGo := time.Now()
	goPred := make(map[string]string, len(items))
	det := chardet.NewDetector()
	for _, it := range items {
		det.Reset()
		det.Feed(it.Bytes, true)
		goPred[it.ID] = string(det.Guess(it.TLDBytes, it.AllowUTF8))
	}
	goDur := time.Since(startGo)

	startSaint := time.Now()
	saintPred := make(map[string]string, len(items))
	saintDetector := saintchardet.NewTextDetector()
	for _, it := range items {
		r, err := saintDetector.DetectBest(it.Bytes)
		if err != nil || r == nil {
			saintPred[it.ID] = ""
			continue
		}
		saintPred[it.ID] = normalizeSaintfish(r.Charset)
	}
	saintDur := time.Since(startSaint)

	if _, err := os.Stat(*rustBin); err != nil {
		return fmt.Errorf("rust binary not found: %s", *rustBin)
	}
	startRust := time.Now()
	cmd := exec.Command(*rustBin, *data, *tmpOut)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	rustDur := time.Since(startRust)

	rustPred, err := readRustOut(*tmpOut)
	if err != nil {
		return err
	}

	goAcc, goFam := accuracy(items, goPred)
	rustAcc, rustFam := accuracy(items, rustPred)
	saintAcc, saintFam := accuracy(items, saintPred)

	var out strings.Builder
	fmt.Fprintf(&out, "samples: %d\n", len(items))
	fmt.Fprintf(&out, "go     exact=%.4f family=%.4f time=%s throughput=%.2f/s\n", goAcc, goFam, goDur, float64(len(items))/goDur.Seconds())
	fmt.Fprintf(&out, "rust   exact=%.4f family=%.4f time=%s throughput=%.2f/s\n", rustAcc, rustFam, rustDur, float64(len(items))/rustDur.Seconds())
	fmt.Fprintf(&out, "saint  exact=%.4f family=%.4f time=%s throughput=%.2f/s\n", saintAcc, saintFam, saintDur, float64(len(items))/saintDur.Seconds())
	out.WriteString("top mismatches (go):\n")
	for _, ln := range topMismatchLines(items, goPred, 10) {
		out.WriteString("  " + ln + "\n")
	}
	out.WriteString("top mismatches (rust):\n")
	for _, ln := range topMismatchLines(items, rustPred, 10) {
		out.WriteString("  " + ln + "\n")
	}
	out.WriteString("top mismatches (saint):\n")
	for _, ln := range topMismatchLines(items, saintPred, 10) {
		out.WriteString("  " + ln + "\n")
	}

	reportText := out.String()
	fmt.Print(reportText)
	if err := os.WriteFile(*report, []byte(reportText), 0o644); err != nil {
		return err
	}
	if err := appendHistory(*history, reportText); err != nil {
		return err
	}
	fmt.Printf("report written: %s\n", *report)
	fmt.Printf("history appended: %s\n", *history)
	return nil
}

func appendHistory(path, report string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	ts := time.Now().Format(time.RFC3339)
	if _, err := f.WriteString("=== " + ts + " ===\n"); err != nil {
		return err
	}
	if _, err := f.WriteString(report); err != nil {
		return err
	}
	_, err = f.WriteString("\n")
	return err
}

func defs() []sampleDef {
	return []sampleDef{
		{Encoding: "utf-8", TLD: "com", AllowUTF8: true, Texts: []string{"这是 UTF-8 测试 mixed English 123", "emoji 😀 UTF-8 only 文本"}},
		{Encoding: "iso-2022-jp", TLD: "jp", AllowUTF8: false, Texts: []string{"日本語のエンコーディング検出テストです。"}},
		{Encoding: "shift_jis", TLD: "jp", AllowUTF8: false, Texts: []string{"これは文字実験です。", "ﾊｰﾄﾞｳｪｱﾊｰﾄﾞｳｪｱ"}},
		{Encoding: "euc-jp", TLD: "jp", AllowUTF8: false, Texts: []string{"これは文字実験です。", "日本語のテキストです。"}},
		{Encoding: "euc-kr", TLD: "kr", AllowUTF8: false, Texts: []string{"이것은 문자 인코딩 테스트입니다.", "한글 데이터 분석"}},
		{Encoding: "gbk", TLD: "cn", AllowUTF8: false, Texts: []string{"这是一个字符编码测试。", "中文内容用于检测。"}},
		{Encoding: "gb18030", TLD: "cn", AllowUTF8: false, Texts: []string{"数据库名：c播拨龾龿珳珴𬀩𬀪", "扩展字符：𠀀𠀁𠀂"}},
		{Encoding: "big5", TLD: "tw", AllowUTF8: false, Texts: []string{"這是一個字符編碼測試。", "繁體中文資料檢測"}},
		{Encoding: "windows-1251", TLD: "ru", AllowUTF8: false, Texts: []string{"Это тест кодировки символов.", "Русский текст для проверки."}},
		{Encoding: "koi8-u", TLD: "ru", AllowUTF8: false, Texts: []string{"Це тест на кодування символів."}},
		{Encoding: "ibm866", TLD: "ru", AllowUTF8: false, Texts: []string{"Это тест кодировки символов."}},
		{Encoding: "iso-8859-5", TLD: "ru", AllowUTF8: false, Texts: []string{"Это тест кодировки символов."}},
		{Encoding: "windows-1253", TLD: "gr", AllowUTF8: false, Texts: []string{"Πρόκειται για δοκιμή κωδικοποίησης χαρακτήρων"}},
		{Encoding: "iso-8859-7", TLD: "gr", AllowUTF8: false, Texts: []string{"Πρόκειται για δοκιμή κωδικοποίησης χαρακτήρων"}},
		{Encoding: "windows-1256", TLD: "sa", AllowUTF8: false, Texts: []string{"هذا هو اختبار ترميز الأحرف."}},
		{Encoding: "iso-8859-6", TLD: "sa", AllowUTF8: false, Texts: []string{"هذا هو اختبار ترميز الأحرف."}},
		{Encoding: "windows-1255", TLD: "il", AllowUTF8: false, Texts: []string{"עברית"}},
		{Encoding: "iso-8859-8", TLD: "il", AllowUTF8: false, Texts: []string{".םיוות דודיק ןחבמ והז"}},
		{Encoding: "windows-1254", TLD: "tr", AllowUTF8: false, Texts: []string{"Bu bir karakter kodlama testidir. Türkçe: ışğüöç"}},
		{Encoding: "windows-1258", TLD: "vn", AllowUTF8: false, Texts: []string{"Đây là một thử nghiệm mã hóa ký tự."}},
		{Encoding: "windows-1250", TLD: "pl", AllowUTF8: false, Texts: []string{"To jest test kodowania znaków."}},
		{Encoding: "iso-8859-2", TLD: "pl", AllowUTF8: false, Texts: []string{"To jest test kodowania znaków."}},
		{Encoding: "windows-1257", TLD: "lv", AllowUTF8: false, Texts: []string{"Šis ir rakstzīmju kodēšanas tests."}},
		{Encoding: "iso-8859-13", TLD: "lv", AllowUTF8: false, Texts: []string{"Šis ir rakstzīmju kodēšanas tests."}},
		{Encoding: "iso-8859-4", TLD: "lv", AllowUTF8: false, Texts: []string{"Šis ir rakstzīmju kodēšanas tests."}},
		{Encoding: "windows-1252", TLD: "com", AllowUTF8: false, Texts: []string{"Rock ’n Roll © Nº1", "Codificació de caràcters"}},
		{Encoding: "windows-874", TLD: "th", AllowUTF8: false, Texts: []string{"นี่คือการทดสอบการเข้ารหัสอักขระ"}},
	}
}

func getEncoding(name string) encoding.Encoding {
	switch name {
	case "shift_jis":
		return japanese.ShiftJIS
	case "euc-jp":
		return japanese.EUCJP
	case "iso-2022-jp":
		return japanese.ISO2022JP
	case "euc-kr":
		return korean.EUCKR
	case "gbk":
		return simplifiedchinese.GBK
	case "gb18030":
		return simplifiedchinese.GB18030
	case "big5":
		return traditionalchinese.Big5
	case "windows-1252":
		return charmap.Windows1252
	case "windows-1251":
		return charmap.Windows1251
	case "windows-1250":
		return charmap.Windows1250
	case "iso-8859-2":
		return charmap.ISO8859_2
	case "windows-1256":
		return charmap.Windows1256
	case "windows-1254":
		return charmap.Windows1254
	case "windows-874":
		return charmap.Windows874
	case "windows-1255":
		return charmap.Windows1255
	case "iso-8859-8":
		return charmap.ISO8859_8
	case "windows-1253":
		return charmap.Windows1253
	case "iso-8859-7":
		return charmap.ISO8859_7
	case "windows-1257":
		return charmap.Windows1257
	case "iso-8859-13":
		return charmap.ISO8859_13
	case "koi8-u":
		return charmap.KOI8U
	case "ibm866":
		return charmap.CodePage866
	case "iso-8859-6":
		return charmap.ISO8859_6
	case "windows-1258":
		return charmap.Windows1258
	case "iso-8859-4":
		return charmap.ISO8859_4
	case "iso-8859-5":
		return charmap.ISO8859_5
	default:
		return nil
	}
}

func readDataset(path string) ([]datasetItem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	out := make([]datasetItem, 0)
	first := true
	for s.Scan() {
		line := s.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if first {
			first = false
			if strings.HasPrefix(line, "id\t") {
				continue
			}
		}
		p := strings.Split(line, "\t")
		if len(p) != 5 {
			continue
		}
		b, err := hex.DecodeString(p[4])
		if err != nil {
			continue
		}
		out = append(out, datasetItem{
			ID:        p[0],
			Expected:  normalize(p[1]),
			TLD:       p[2],
			TLDBytes:  []byte(p[2]),
			AllowUTF8: p[3] == "1",
			Bytes:     b,
		})
	}
	return out, s.Err()
}

func readRustOut(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	out := make(map[string]string)
	for s.Scan() {
		p := strings.Split(s.Text(), "\t")
		if len(p) != 2 {
			continue
		}
		out[p[0]] = normalize(p[1])
	}
	return out, s.Err()
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "-")
	if s == "gb-18030" {
		return "gb18030"
	}
	if s == "x-user-defined" {
		return "windows-1252"
	}
	return s
}

func normalizeSaintfish(s string) string {
	n := normalize(s)
	switch n {
	case "windows-31j", "cp932":
		return "shift-jis"
	case "gb-18030", "gb18030":
		return "gb18030"
	case "ibm-866", "cp866":
		return "ibm866"
	case "iso-8859-8-i":
		return "windows-1255"
	}
	return n
}

func family(enc string) string {
	switch enc {
	case "shift-jis", "euc-jp", "iso-2022-jp":
		return "jp"
	case "gbk", "gb18030", "big5":
		return "zh"
	case "euc-kr":
		return "kr"
	case "windows-1251", "koi8-u", "ibm866", "iso-8859-5":
		return "cyr"
	case "windows-1253", "iso-8859-7":
		return "gr"
	case "windows-1256", "iso-8859-6":
		return "ar"
	case "windows-1255", "iso-8859-8":
		return "he"
	case "windows-1250", "iso-8859-2":
		return "central"
	case "windows-1257", "iso-8859-13", "iso-8859-4":
		return "baltic"
	case "windows-1254":
		return "tr"
	case "windows-1258":
		return "vi"
	case "windows-874":
		return "th"
	case "windows-1252":
		return "west"
	case "utf-8":
		return "utf8"
	default:
		return enc
	}
}

func accuracy(items []datasetItem, pred map[string]string) (exact, fam float64) {
	if len(items) == 0 {
		return 0, 0
	}
	var ex, fa int
	for _, it := range items {
		p := normalize(pred[it.ID])
		e := normalize(it.Expected)
		if p == e {
			ex++
		}
		if family(p) == family(e) {
			fa++
		}
	}
	return float64(ex) / float64(len(items)), float64(fa) / float64(len(items))
}

func topMismatchLines(items []datasetItem, pred map[string]string, limit int) []string {
	m := map[string]int{}
	for _, it := range items {
		p := normalize(pred[it.ID])
		e := normalize(it.Expected)
		if p != e {
			m[e+" -> "+p]++
		}
	}
	type kv struct {
		K string
		V int
	}
	arr := make([]kv, 0, len(m))
	for k, v := range m {
		arr = append(arr, kv{k, v})
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].V == arr[j].V {
			return arr[i].K < arr[j].K
		}
		return arr[i].V > arr[j].V
	})
	if len(arr) < limit {
		limit = len(arr)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, fmt.Sprintf("%s : %d", arr[i].K, arr[i].V))
	}
	if len(out) == 0 {
		out = append(out, "(none)")
	}
	return out
}
