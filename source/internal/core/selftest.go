package core

import (
	"fmt"
	"sort"
	"strings"
)

// SampleCase là kỳ vọng của một cặp mẫu theo QUY TẮC HIỆN HÀNH (Khiêm xác nhận 02/10/2026):
// mã hàng giữ nguyên, điền tên hàng, không giảm giá thì để trống; hàng đổi/bảo hành chỉ có số lượng thì vẫn nhập với
// đơn giá 0 (Khiêm 04/10/2026, thay quy tắc cũ "hàng đổi không nhập"); dòng không có số lượng thì bỏ.
//
//	Mẫu 1: 44 dòng, khác kết quả cũ đúng 1 ô (A44: PTXD0164 thay vì o).
//	Mẫu 2: 80 dòng, khác 80 ô Tên hàng (cột B) vì nay điền tên.
//	Mẫu 3: 55 dòng đầu khớp hoàn toàn kết quả mẫu (kết quả mẫu cũ chưa có 9 dòng hàng đổi); thêm 9 dòng hàng đổi đơn giá 0 ở cuối, tổng 64 dòng.
type SampleCase struct {
	N                             int
	Input, Expected               string
	Profile                       string
	Rows, Normal, Promo, Exch, Ex int
	ValueDiffs                    int
	DiffCol                       string // cột duy nhất được phép khác ("" nếu không ô nào khác)
	FirstDiff                     string // mô tả ô khác đầu tiên (nếu có yêu cầu cụ thể)
	Name                          string
	Alt                           []string // cùng hóa đơn lưu ở định dạng khác (.xls...), phải cho đúng cùng kết quả
}

var SampleCases = []SampleCase{
	{1, "Mau 1.xlsx", "Ket qua mau 1.xlsx", ProfileA, 44, 43, 1, 0, 0, 1, "A", "A44|PTXD0164|o", "NhapHang_PBH002331_23072026.xlsx", []string{"dinh-dang-khac/Mau1.xls"}},
	{2, "Mau 2.xlsx", "Ket qua mau 2.xlsx", ProfileB, 80, 80, 0, 0, 0, 80, "B", "", "NhapHang_BinhMinh_25072026.xlsx", []string{"dinh-dang-khac/Mau2.xls"}},
	{3, "Mau 3.xlsx", "Ket qua mau 3.xlsx", ProfileC, 64, 37, 18, 9, 0, 0, "", "", "NhapHang_08-26-03_01082026.xlsx", []string{"dinh-dang-khac/Mau3.xls", "dinh-dang-khac/Mau3_xml2003.xls"}},
}

// SelfTestResult là kết quả kiểm tra một mẫu.
type SelfTestResult struct {
	N        int
	OK       bool
	Summary  string
	Problems []string
}

// RunSelfTest chạy cả 3 mẫu qua đúng đường xử lý của app và so từng ô với kết quả mẫu.
func RunSelfTest(read func(name string) ([]byte, error), template []byte) []SelfTestResult {
	var out []SelfTestResult
	for _, c := range SampleCases {
		for i, input := range append([]string{c.Input}, c.Alt...) {
			out = append(out, runCase(read, template, c, input, i > 0))
		}
	}
	out = append(out, runExtra(read, template)...)
	return out
}

func runCase(read func(name string) ([]byte, error), template []byte, c SampleCase, input string, alt bool) SelfTestResult {
	r := SelfTestResult{N: c.N, OK: true}
	fail := func(f string, a ...any) { r.OK = false; r.Problems = append(r.Problems, fmt.Sprintf(f, a...)) }
	in, err := read(input)
	if err != nil {
		fail("không đọc được %s: %v", input, err)
		return r
	}
	exp, err := read(c.Expected)
	if err != nil {
		fail("không đọc được %s: %v", c.Expected, err)
		return r
	}
	g, _, err := ReadGridAny(in)
	if err != nil {
		fail("%v", err)
		return r
	}
	p, err := Parse(g)
	if err != nil {
		fail("%v", err)
		return r
	}
	if p.Profile != c.Profile || !p.Known {
		fail("nhận dạng mẫu %s, cần %s", p.Profile, c.Profile)
	}
	nm, pr, dx, ex := p.Counts()
	rows := p.OutputRows()
	if len(rows) != c.Rows || nm != c.Normal || pr != c.Promo || dx != c.Exch || ex != c.Ex {
		fail("số dòng %d (thường %d, khuyến mại %d, hàng đổi %d, bỏ %d), cần %d (%d, %d, %d, %d)", len(rows), nm, pr, dx, ex, c.Rows, c.Normal, c.Promo, c.Exch, c.Ex)
	}
	// hàng đổi chỉ có số lượng: vẫn nhập, đơn giá 0, số lượng dương
	var cmpRows []OutRow
	for _, row := range rows {
		if row.Kind == KindExchange {
			if q, ok := row.Qty.Float(); row.Price != 0 || !ok || q <= 0 {
				fail("dòng hàng đổi %s phải có đơn giá 0 và số lượng dương", row.Code)
			}
			continue
		}
		cmpRows = append(cmpRows, row)
	}
	if len(p.Errors) > 0 {
		fail("có lỗi: %s", strings.Join(p.Errors, "; "))
	}
	if p.SuggestedName() != c.Name {
		fail("tên file %s, cần %s", p.SuggestedName(), c.Name)
	}
	bytesOut, err := BuildWorkbook(template, rows)
	if err != nil {
		fail("%v", err)
		return r
	}
	if err := VerifyExport(bytesOut, rows); err != nil {
		fail("mở lại file xuất: %v", err)
	}
	cmpOut := bytesOut
	if len(cmpRows) != len(rows) { // kết quả mẫu cũ chưa có dòng hàng đổi: so 55 dòng còn lại
		if cmpOut, err = BuildWorkbook(template, cmpRows); err != nil {
			fail("%v", err)
			return r
		}
	}
	cmp, err := Compare(cmpOut, exp)
	if err != nil {
		fail("%v", err)
		return r
	}
	if cmp.RowsGot != cmp.RowsWant {
		fail("số dòng %d khác kết quả mẫu %d", cmp.RowsGot, cmp.RowsWant)
	}
	if len(cmp.ValueDiffs) != c.ValueDiffs {
		fail("%d ô khác giá trị, cần đúng %d", len(cmp.ValueDiffs), c.ValueDiffs)
	}
	if cmp.TypeDiffs != 0 {
		fail("%d ô khác kiểu dữ liệu (số/văn bản)", cmp.TypeDiffs)
	}
	var cols []string
	for _, d := range cmp.ValueDiffs {
		if c.DiffCol == "" || d.Cell[:1] != c.DiffCol {
			cols = append(cols, d.Cell)
		}
	}
	if len(cols) > 0 {
		sort.Strings(cols)
		fail("ô khác ngoài dự kiến: %s", strings.Join(cols, ", "))
	}
	if c.FirstDiff != "" && len(cmp.ValueDiffs) > 0 {
		d := cmp.ValueDiffs[0]
		if got := d.Cell + "|" + d.Got + "|" + d.Want; got != c.FirstDiff {
			fail("ô khác không như dự kiến: %s", got)
		}
	}
	label := fmt.Sprintf("Mẫu %d", c.N)
	if alt {
		label = fmt.Sprintf("Mẫu %d (bản %s)", c.N, input)
	}
	r.Summary = fmt.Sprintf("%s: %d dòng (%d thường, %d khuyến mại, %d hàng đổi giá 0, %d bỏ vì không có số lượng), %d ô khác kết quả mẫu, %d ô khác kiểu", label, len(rows), nm, pr, dx, ex, len(cmp.ValueDiffs), cmp.TypeDiffs)
	return r
}

// runExtra kiểm tra hai hóa đơn nhà cung cấp mới (chưa có kết quả mẫu để so từng ô): chỉ kiểm tra đọc đúng, đối chiếu tổng khớp và xuất file đạt.
func runExtra(read func(name string) ([]byte, error), template []byte) []SelfTestResult {
	type extra struct {
		n       int
		file    string
		profile string
		rows    int
		name    string
		price0  float64
	}
	cases := []extra{
		{4, "Mau 4 - Toa hang Tan Hoa Loi.xlsx", ProfileD, 27, "NhapHang_TanHoaLoi_01042026.xlsx", 7500},
		{4, "dinh-dang-khac/Mau4.xls", ProfileD, 27, "NhapHang_TanHoaLoi_01042026.xlsx", 7500},
		{5, "Mau 5 - gk 28-2.xls", ProfileE, 281, "NhapHang_DB26HA-00427_27022026.xlsx", 63892.8},
		{6, "Mau 6 - Phieu xuat kho SL lay.xlsx", ProfileF, 35, "NhapHang_PhieuXuatKho_20102026.xlsx", 249381.72},
	}
	var out []SelfTestResult
	for _, c := range cases {
		r := SelfTestResult{N: c.n, OK: true}
		fail := func(f string, a ...any) { r.OK = false; r.Problems = append(r.Problems, fmt.Sprintf(f, a...)) }
		data, err := read(c.file)
		if err != nil {
			fail("không đọc được %s: %v", c.file, err)
			out = append(out, r)
			continue
		}
		g, kind, err := ReadGridAny(data)
		var p *Parsed
		if err == nil {
			p, err = Parse(g)
		}
		if err != nil {
			fail("%v", err)
			out = append(out, r)
			continue
		}
		rows := p.OutputRows()
		if p.Profile != c.profile || len(rows) != c.rows || len(p.Errors) != 0 {
			fail("mẫu %s, %d dòng, lỗi %v; cần mẫu %s, %d dòng, không lỗi", p.Profile, len(rows), p.Errors, c.profile, c.rows)
		}
		if p.SuggestedName() != c.name {
			fail("tên file %s, cần %s", p.SuggestedName(), c.name)
		}
		if len(rows) > 0 && rows[0].Price != c.price0 {
			fail("đơn giá dòng 1 là %v, cần %v", rows[0].Price, c.price0)
		}
		for _, ck := range p.Checks {
			if ck.Level != LevelOK {
				fail("đối chiếu %q không khớp: %s", ck.Name, ck.Detail)
			}
		}
		if b, err := BuildWorkbook(template, rows); err != nil {
			fail("%v", err)
		} else if err := VerifyExport(b, rows); err != nil {
			fail("mở lại file xuất: %v", err)
		}
		r.Summary = fmt.Sprintf("Mẫu %d (%s, %s): %d dòng, đối chiếu tổng khớp, %d cảnh báo cần xác nhận khi dùng thật (chưa có kết quả mẫu để so từng ô)", c.n, c.file, kind, len(rows), len(p.Warnings))
		out = append(out, r)
	}
	return out
}
