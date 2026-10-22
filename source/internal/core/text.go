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
	reThousandComma = regexp.MustCompile(`^\d{1,3}(,\d{3})+$`)
	reMoneyComma    = regexp.MustCompile(`^\d{1,3}(,\d{3})+(\.\d+)?$`)
	reMoneyDot      = regexp.MustCompile(`^\d{1,3}(\.\d{3})+(,\d+)?$`)
	reDate          = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})$`)
	reDateWords     = regexp.MustCompile(`ngày\s+(\d+)\s+tháng\s+(\d+)\s+năm\s+(\d{4})`)
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

// numFromText đọc số từ chuỗi để đối chiếu: "492,000" -> 492000, "2,00" -> 2, "5 %" -> 5.
func numFromText(s string) (float64, bool) {
	t := strings.ReplaceAll(strings.TrimSpace(s), " ", "")
	t = strings.ReplaceAll(t, "%", "")
	if t == "" {
		return 0, false
	}
	if reThousandComma.MatchString(t) {
		t = strings.ReplaceAll(t, ",", "")
	} else {
		t = strings.ReplaceAll(t, ",", ".")
	}
	f, err := strconv.ParseFloat(t, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
}

// parseMoney đọc đơn giá: số giữ nguyên, chuỗi "492,000" đổi thành 492000. ok=false nếu trống/không phải số.
func parseMoney(c Cell) (float64, bool) {
	if c.Empty() {
		return 0, false
	}
	if c.Num {
		return c.Float()
	}
	t := strings.ReplaceAll(strings.TrimSpace(c.S), " ", "")
	switch {
	case reMoneyComma.MatchString(t):
		t = strings.ReplaceAll(t, ",", "")
	case reMoneyDot.MatchString(t):
		t = strings.ReplaceAll(t, ".", "")
		t = strings.ReplaceAll(t, ",", ".")
	}
	f, err := strconv.ParseFloat(t, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
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
