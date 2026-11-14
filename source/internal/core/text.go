package core

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	reDate      = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})$`)
	reDateWords = regexp.MustCompile(`ngày\s+(\d+)\s+tháng\s+(\d+)\s+năm\s+(\d{4})`)
)

// normText: bỏ khoảng trắng thừa, chữ thường, chuẩn hóa NFC (để so nhãn cột/chuỗi đánh dấu).
func normText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(norm.NFC.String(s))), " ")
}

// stripAccents bỏ dấu tiếng Việt, dùng cho tên file.
func stripAccents(s string) string {
	s = strings.NewReplacer("đ", "d", "Đ", "D").Replace(s)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return out
}

// parseVNNumber đọc số dạng chữ do nhiều phần mềm xuất ra, không phụ thuộc dấu . hay , làm phân cách:
// "250.000" và "250,000" -> 250000; "1.234.567,5" và "1,234,567.5" -> 1234567.5; "0,35" và "0.35" -> 0.35.
// Quy ước: có cả hai dấu thì dấu đứng sau là dấu thập phân; chỉ một loại dấu mà mọi nhóm sau nó đúng 3 chữ số
// (và phần đầu khác 0) thì là phân cách nghìn; còn lại là dấu thập phân.
func parseVNNumber(s string) (float64, bool) {
	t := strings.TrimSpace(s)
	t = strings.NewReplacer(" ", "", "\u00a0", "", "%", "", "đ", "", "Đ", "").Replace(t)
	neg := false
	if strings.HasPrefix(t, "(") && strings.HasSuffix(t, ")") {
		neg, t = true, t[1:len(t)-1]
	}
	if strings.HasPrefix(t, "-") {
		neg, t = !neg, t[1:]
	}
	if t == "" {
		return 0, false
	}
	dot, comma := strings.LastIndex(t, "."), strings.LastIndex(t, ",")
	switch {
	case dot >= 0 && comma >= 0:
		dec, th := ".", ","
		if comma > dot {
			dec, th = ",", "."
		}
		t = strings.ReplaceAll(t, th, "")
		t = strings.Replace(t, dec, ".", 1)
	case dot >= 0 || comma >= 0:
		sep := "."
		if comma >= 0 {
			sep = ","
		}
		parts := strings.Split(t, sep)
		thousands := len(parts[0]) >= 1 && len(parts[0]) <= 3 && parts[0] != "0"
		for _, g := range parts[1:] {
			if len(g) != 3 {
				thousands = false
			}
		}
		if thousands {
			t = strings.ReplaceAll(t, sep, "")
		} else if len(parts) == 2 {
			t = parts[0] + "." + parts[1]
		} else {
			return 0, false
		}
	}
	for _, r := range t {
		if (r < '0' || r > '9') && r != '.' && r != 'e' && r != 'E' && r != '+' && r != '-' {
			return 0, false
		}
	}
	f, err := strconv.ParseFloat(t, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	if neg {
		f = -f
	}
	return f, true
}

// numFromText đọc số từ chuỗi (số lượng, thành tiền, tổng): xem parseVNNumber.
func numFromText(s string) (float64, bool) { return parseVNNumber(s) }

// parseMoney đọc đơn giá: ô số giữ nguyên, ô chữ đọc bằng parseVNNumber. ok=false nếu trống/không phải số.
func parseMoney(c Cell) (float64, bool) {
	if c.Empty() {
		return 0, false
	}
	if c.Num {
		return c.Float()
	}
	return parseVNNumber(c.S)
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// vnNum định dạng số kiểu Việt Nam: 71.299.000 hoặc 47.132.326,5
func vnNum(f float64) string {
	neg := f < 0
	if neg {
		f = -f
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if i := strings.IndexByte(s, '.'); i >= 0 && len(s)-i-1 > 4 {
		s = strconv.FormatFloat(math.Round(f*10000)/10000, 'f', -1, 64)
	}
	intPart, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i+1:]
	}
	var b strings.Builder
	for i, ch := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(ch)
	}
	out := b.String()
	if frac != "" {
		out += "," + frac
	}
	if neg {
		out = "-" + out
	}
	return out
}

// dateParts trả ngày/tháng/năm từ "dd/mm/yyyy".
func dateParts(s string) (d, m, y string, ok bool) {
	g := reDate.FindStringSubmatch(strings.TrimSpace(s))
	if g == nil {
		return "", "", "", false
	}
	di, _ := strconv.Atoi(g[1])
	mi, _ := strconv.Atoi(g[2])
	return fmt.Sprintf("%02d", di), fmt.Sprintf("%02d", mi), g[3], true
}
