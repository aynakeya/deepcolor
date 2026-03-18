package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type fieldDef struct {
	Name string
	Typ  string
	Len  int
}

type sbSpec struct {
	EncodingConst string
	LowerField    string
	UpperField    string
	ProbField     string
	ASCIIConst    string
	NonASCIIConst string
}

var (
	reFieldDef  = regexp.MustCompile(`^\s*([a-z0-9_]+): \[(u8|u16); (\d+)\],\s*$`)
	reConst     = regexp.MustCompile(`^\s*(?:pub\s+)?const ([A-Z0-9_]+): usize = (\d+);\s*$`)
	reNum       = regexp.MustCompile(`0x[0-9A-Fa-f]+|\d+`)
	reIndex     = regexp.MustCompile(`^\s*pub const ([A-Z0-9_]+): usize = (\d+);\s*$`)
	reEncoding  = regexp.MustCompile(`encoding:\s*&([A-Z0-9_]+),`)
	reFieldRef  = regexp.MustCompile(`(lower|upper|probabilities):\s*&DETECTOR_DATA\.([a-z0-9_]+),`)
	reAsciiRef  = regexp.MustCompile(`ascii:\s*([A-Z0-9_]+),`)
	reNAsciiRef = regexp.MustCompile(`non_ascii:\s*([A-Z0-9_]+),`)
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	in := filepath.Join(root, "chardetng", "src", "data.rs")
	out := filepath.Join(root, "x", "chardet", "singlebyte_data_gen.go")
	raw, err := os.ReadFile(in)
	if err != nil {
		panic(err)
	}
	txt := string(raw)

	fields := parseFieldDefs(txt)
	arrays := parseDetectorDataArrays(txt, fields)
	consts := parseConsts(txt)
	specs := parseSingleByteSpecs(txt)
	indices := parseIndexConsts(txt)

	code := render(fields, arrays, consts, specs, indices)
	if err := os.WriteFile(out, []byte(code), 0o644); err != nil {
		panic(err)
	}
	fmt.Println("generated", out)
}

func parseFieldDefs(txt string) []fieldDef {
	start := strings.Index(txt, "pub struct DetectorData {")
	end := strings.Index(txt[start:], "}\n\n#[rustfmt::skip]")
	if start < 0 || end < 0 {
		panic("DetectorData struct not found")
	}
	blk := txt[start : start+end]
	lines := strings.Split(blk, "\n")
	out := make([]fieldDef, 0)
	for _, ln := range lines {
		m := reFieldDef.FindStringSubmatch(ln)
		if len(m) == 4 {
			n, _ := strconv.Atoi(m[3])
			out = append(out, fieldDef{Name: m[1], Typ: m[2], Len: n})
		}
	}
	if len(out) == 0 {
		panic("no fields parsed")
	}
	return out
}

func parseDetectorDataArrays(txt string, fields []fieldDef) map[string][]string {
	start := strings.Index(txt, "pub static DETECTOR_DATA: DetectorData = DetectorData {")
	if start < 0 {
		panic("DETECTOR_DATA init not found")
	}
	body := txt[start:]
	body = body[strings.Index(body, "{")+1:]
	end := strings.Index(body, "};")
	if end < 0 {
		panic("DETECTOR_DATA end not found")
	}
	body = body[:end]

	known := map[string]fieldDef{}
	for _, f := range fields {
		known[f.Name] = f
	}

	res := map[string][]string{}
	for i := 0; i < len(body); {
		for i < len(body) && (body[i] == ' ' || body[i] == '\n' || body[i] == '\t' || body[i] == ',') {
			i++
		}
		if i >= len(body) {
			break
		}
		j := i
		for j < len(body) && ((body[j] >= 'a' && body[j] <= 'z') || body[j] == '_' || (body[j] >= '0' && body[j] <= '9')) {
			j++
		}
		if j == i {
			i++
			continue
		}
		name := body[i:j]
		if _, ok := known[name]; !ok {
			i = j
			continue
		}
		k := strings.Index(body[j:], "[")
		if k < 0 {
			panic("array start not found for " + name)
		}
		k = j + k
		depth := 0
		l := k
		for ; l < len(body); l++ {
			if body[l] == '[' {
				depth++
			} else if body[l] == ']' {
				depth--
				if depth == 0 {
					break
				}
			}
		}
		if l >= len(body) {
			panic("array end not found for " + name)
		}
		arr := body[k+1 : l]
		vals := reNum.FindAllString(arr, -1)
		res[name] = vals
		i = l + 1
	}

	for _, f := range fields {
		vals, ok := res[f.Name]
		if !ok {
			panic("missing data for field " + f.Name)
		}
		if len(vals) != f.Len {
			panic(fmt.Sprintf("field %s len mismatch: got %d want %d", f.Name, len(vals), f.Len))
		}
	}
	return res
}

func parseConsts(txt string) map[string]string {
	lines := strings.Split(txt, "\n")
	out := map[string]string{}
	for _, ln := range lines {
		m := reConst.FindStringSubmatch(ln)
		if len(m) == 3 {
			out[m[1]] = m[2]
		}
	}
	return out
}

func parseSingleByteSpecs(txt string) []sbSpec {
	start := strings.Index(txt, "pub static SINGLE_BYTE_DATA: [SingleByteData; 20] = [")
	if start < 0 {
		panic("SINGLE_BYTE_DATA not found")
	}
	body := txt[start:]
	body = body[strings.Index(body, "[")+1:]
	end := strings.Index(body, "];\n\npub const WINDOWS_1258_INDEX")
	if end < 0 {
		panic("SINGLE_BYTE_DATA end not found")
	}
	body = body[:end]

	parts := strings.Split(body, "SingleByteData {")
	out := make([]sbSpec, 0, 20)
	for _, p := range parts[1:] {
		p = p[:strings.Index(p, "},")]
		var s sbSpec
		if m := reEncoding.FindStringSubmatch(p); len(m) == 2 {
			s.EncodingConst = m[1]
		}
		for _, m := range reFieldRef.FindAllStringSubmatch(p, -1) {
			switch m[1] {
			case "lower":
				s.LowerField = m[2]
			case "upper":
				s.UpperField = m[2]
			case "probabilities":
				s.ProbField = m[2]
			}
		}
		if m := reAsciiRef.FindStringSubmatch(p); len(m) == 2 {
			s.ASCIIConst = m[1]
		}
		if m := reNAsciiRef.FindStringSubmatch(p); len(m) == 2 {
			s.NonASCIIConst = m[1]
		}
		if s.EncodingConst == "" || s.LowerField == "" || s.UpperField == "" || s.ProbField == "" || s.ASCIIConst == "" || s.NonASCIIConst == "" {
			panic("failed to parse SingleByteData entry")
		}
		out = append(out, s)
	}
	if len(out) != 20 {
		panic(fmt.Sprintf("single byte entries mismatch: %d", len(out)))
	}
	return out
}

func parseIndexConsts(txt string) map[string]string {
	out := map[string]string{}
	for _, ln := range strings.Split(txt, "\n") {
		m := reIndex.FindStringSubmatch(ln)
		if len(m) == 3 {
			if !strings.HasSuffix(m[1], "_INDEX") {
				continue
			}
			out[m[1]] = m[2]
		}
	}
	return out
}

func render(fields []fieldDef, arrays map[string][]string, consts map[string]string, specs []sbSpec, indices map[string]string) string {
	var b strings.Builder
	b.WriteString("// Code generated by x/chardet/tools/gen_singlebyte_data.go; DO NOT EDIT.\n")
	b.WriteString("package chardet\n\n")

	b.WriteString("const implausibilityPenalty int64 = -220\n")
	keys := make([]string, 0, len(consts))
	for k := range consts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	writtenConsts := map[string]struct{}{}
	for _, k := range keys {
		if strings.HasSuffix(k, "_ASCII") || strings.HasSuffix(k, "_NON_ASCII") || k == "ASCII_DIGIT" || k == "WINDOWS_1256_ZWNJ" {
			name := toGoConst(k)
			if _, ok := writtenConsts[name]; ok {
				continue
			}
			writtenConsts[name] = struct{}{}
			b.WriteString("const " + name + " = " + consts[k] + "\n")
		}
	}
	b.WriteString("\n")

	for _, f := range fields {
		goName := toGoField(f.Name)
		goTyp := "uint8"
		if f.Typ == "u16" {
			goTyp = "uint16"
		}
		b.WriteString("var " + goName + " = [" + strconv.Itoa(f.Len) + "]" + goTyp + "{\n")
		vals := arrays[f.Name]
		for i, v := range vals {
			if i%16 == 0 {
				b.WriteString("\t")
			}
			b.WriteString(v)
			b.WriteString(", ")
			if i%16 == 15 {
				b.WriteString("\n")
			}
		}
		if len(vals)%16 != 0 {
			b.WriteString("\n")
		}
		b.WriteString("}\n\n")
	}

	b.WriteString("type singleByteData struct {\n")
	b.WriteString("\tencoding     Encoding\n")
	b.WriteString("\tlower        *[128]uint8\n")
	b.WriteString("\tupper        *[128]uint8\n")
	b.WriteString("\tprobabilities []uint8\n")
	b.WriteString("\tascii         int\n")
	b.WriteString("\tnonASCII      int\n")
	b.WriteString("}\n\n")

	b.WriteString("func (s *singleByteData) classify(byt byte) uint8 {\n")
	b.WriteString("\tif byt < 0x80 {\n\t\treturn s.lower[int(byt)]\n\t}\n")
	b.WriteString("\treturn s.upper[int(byt-0x80)]\n")
	b.WriteString("}\n\n")

	b.WriteString("func (s *singleByteData) isLatinAlphabetic(c uint8) bool {\n")
	b.WriteString("\tcu := int(c)\n\treturn cu > 0 && cu < (s.ascii+s.nonASCII)\n")
	b.WriteString("}\n\n")

	b.WriteString("func (s *singleByteData) isNonLatinAlphabetic(c uint8, isWindows1256 bool) bool {\n")
	b.WriteString("\tlowerBound := 1\n\tif isWindows1256 { lowerBound = windows1256ZWNJ }\n")
	b.WriteString("\tcu := int(c)\n\treturn cu > lowerBound && cu < (s.ascii+s.nonASCII)\n")
	b.WriteString("}\n\n")

	b.WriteString("func computeIndex(x, y, asciiClasses, nonASCIIClasses int) (int, bool) {\n")
	b.WriteString("\tif x == 0 && y == 0 { return 0, false }\n")
	b.WriteString("\tif x < asciiClasses && y < asciiClasses { return 0, false }\n")
	b.WriteString("\tif y >= asciiClasses {\n")
	b.WriteString("\t\treturn (asciiClasses*nonASCIIClasses) + (asciiClasses+nonASCIIClasses)*(y-asciiClasses) + x, true\n\t}\n")
	b.WriteString("\treturn y*nonASCIIClasses + x - asciiClasses, true\n")
	b.WriteString("}\n\n")

	b.WriteString("func (s *singleByteData) score(currentClass, previousClass uint8, isWindows1256 bool) int64 {\n")
	b.WriteString("\tcurrent := int(currentClass)\n\tprevious := int(previousClass)\n\tstoredBoundary := s.ascii + s.nonASCII\n")
	b.WriteString("\tif current < storedBoundary {\n")
	b.WriteString("\t\tif previous < storedBoundary {\n")
	b.WriteString("\t\t\tif idx, ok := computeIndex(previous, current, s.ascii, s.nonASCII); ok {\n")
	b.WriteString("\t\t\t\tb := s.probabilities[idx]\n\t\t\t\tif b == 255 { return implausibilityPenalty }\n\t\t\t\treturn int64(b)\n\t\t\t}\n\t\t\treturn 0\n\t\t}\n")
	b.WriteString("\t\tif current == 0 || current == asciiDigit || (isWindows1256 && current == windows1256ZWNJ) { return 0 }\n")
	b.WriteString("\t\tpreviousUnstored := previous - storedBoundary\n")
	b.WriteString("\t\tswitch previousUnstored {\n")
	b.WriteString("\t\tcase 0: return 0\n\t\tcase 1, 2: return implausibilityPenalty\n\t\tcase 3: return 0\n")
	b.WriteString("\t\tcase 4: if current < s.ascii { return implausibilityPenalty }; return 0\n")
	b.WriteString("\t\tcase 5: if current < s.ascii { return 0 }; return implausibilityPenalty\n")
	b.WriteString("\t\tdefault: return 0\n\t\t}\n\t}\n")
	b.WriteString("\tif previous < storedBoundary {\n")
	b.WriteString("\t\tif previous == 0 || previous == asciiDigit || (isWindows1256 && previous == windows1256ZWNJ) { return 0 }\n")
	b.WriteString("\t\tcurrentUnstored := current - storedBoundary\n")
	b.WriteString("\t\tswitch currentUnstored {\n")
	b.WriteString("\t\tcase 0: return 0\n\t\tcase 1, 3: return implausibilityPenalty\n\t\tcase 2: return 0\n")
	b.WriteString("\t\tcase 4: if previous < s.ascii { return implausibilityPenalty }; return 0\n")
	b.WriteString("\t\tcase 5: if previous < s.ascii { return 0 }; return implausibilityPenalty\n")
	b.WriteString("\t\tdefault: return 0\n\t\t}\n\t}\n")
	b.WriteString("\tif current == asciiDigit || previous == asciiDigit { return 0 }\n")
	b.WriteString("\treturn implausibilityPenalty\n")
	b.WriteString("}\n\n")

	b.WriteString("var singleByteDataTable = [...]singleByteData{\n")
	for _, sp := range specs {
		b.WriteString("\t{\n")
		b.WriteString("\t\tencoding: " + mapEncoding(sp.EncodingConst) + ",\n")
		b.WriteString("\t\tlower: &" + toGoField(sp.LowerField) + ",\n")
		b.WriteString("\t\tupper: &" + toGoField(sp.UpperField) + ",\n")
		b.WriteString("\t\tprobabilities: " + toGoFieldSlice(sp.ProbField) + ",\n")
		b.WriteString("\t\tascii: " + toGoConst(sp.ASCIIConst) + ",\n")
		b.WriteString("\t\tnonASCII: " + toGoConst(sp.NonASCIIConst) + ",\n")
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n\n")

	iKeys := make([]string, 0, len(indices))
	for k := range indices {
		iKeys = append(iKeys, k)
	}
	sort.Strings(iKeys)
	for _, k := range iKeys {
		b.WriteString("const " + toGoConst(k) + " = " + indices[k] + "\n")
	}
	return b.String()
}

func toGoField(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if i == 0 {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

func toGoFieldSlice(s string) string {
	f := toGoField(s)
	return f + "[:]"
}

func toGoConst(s string) string {
	parts := strings.Split(strings.ToLower(s), "_")
	if len(parts) == 0 {
		return s
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		p := parts[i]
		switch p {
		case "ascii":
			p = "ASCII"
		case "zwnj":
			p = "ZWNJ"
		case "ibm":
			p = "IBM"
		case "koi8":
			p = "KOI8"
		case "iso":
			p = "ISO"
		default:
			p = strings.ToUpper(p[:1]) + p[1:]
		}
		out += p
	}
	return out
}

func mapEncoding(c string) string {
	switch c {
	case "WINDOWS_1258_INIT":
		return "EncodingWindows1258"
	case "WINDOWS_1250_INIT":
		return "EncodingWindows1250"
	case "ISO_8859_2_INIT":
		return "EncodingISO88592"
	case "WINDOWS_1251_INIT":
		return "EncodingWindows1251"
	case "KOI8_U_INIT":
		return "EncodingKOI8U"
	case "ISO_8859_5_INIT":
		return "EncodingISO88595"
	case "IBM866_INIT":
		return "EncodingIBM866"
	case "WINDOWS_1252_INIT":
		return "EncodingWindows1252"
	case "WINDOWS_1253_INIT":
		return "EncodingWindows1253"
	case "ISO_8859_7_INIT":
		return "EncodingISO88597"
	case "WINDOWS_1254_INIT":
		return "EncodingWindows1254"
	case "WINDOWS_1255_INIT":
		return "EncodingWindows1255"
	case "ISO_8859_8_INIT":
		return "EncodingISO88598"
	case "WINDOWS_1256_INIT":
		return "EncodingWindows1256"
	case "ISO_8859_6_INIT":
		return "EncodingISO88596"
	case "WINDOWS_1257_INIT":
		return "EncodingWindows1257"
	case "ISO_8859_13_INIT":
		return "EncodingISO885913"
	case "ISO_8859_4_INIT":
		return "EncodingISO88594"
	case "WINDOWS_874_INIT":
		return "EncodingWindows874"
	default:
		panic("unmapped encoding const: " + c)
	}
}
