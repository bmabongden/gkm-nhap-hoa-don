package core

import (
	"strings"
	"testing"
)

// Số dạng chữ từ nhiều phần mềm: dấu . và , có thể là phân cách nghìn hoặc dấu thập phân (Khiêm báo lỗi 14/11/2026).
func TestParseVNNumber(t *testing.T) {
	cases := map[string]float64{
		"250.000": 250000, "250,000": 250000, "1.234.567": 1234567, "1,234,567": 1234567,
		"1.234.567,5": 1234567.5, "1,234,567.5": 1234567.5, "145.000,00": 145000, "145,000.00": 145000,
		"0,35": 0.35, "0.35": 0.35, "0.350": 0.35, "2,00": 2, "2,5": 2.5, "7.5": 7.5, "12": 12,
		"250.000 đ": 250000, "5 %": 5, "-1.500": -1500, "(2.000)": -2000, "1 250 000": 1250000,
	}
	for in, want := range cases {
		got, ok := parseVNNumber(in)
		if !ok || got != want {
			t.Errorf("%q: được %v (ok=%v), cần %v", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", "abc", "1.2.3", "1,23,4", "12a", "NaN", "Inf"} {
		if v, ok := parseVNNumber(bad); ok {
			t.Errorf("%q phải bị từ chối, được %v", bad, v)
		}
	}
}

// Hóa đơn mà mọi đơn giá đều dưới 1.000 đ: app hiểu là nghìn đồng, nhân 1.000 và cảnh báo.
func TestTinyPricesScaledToThousand(t *testing.T) {
	g := newGrid("Sheet1")
	for c, h := range []string{"Mã hàng", "Tên hàng", "Số lượng", "Đơn giá", "Thành tiền"} {
		g.Cells[[2]int{1, c + 1}] = Cell{S: h}
	}
	rows := [][]Cell{
		{{S: "A1"}, {S: "Bóng LED"}, {S: "100", Num: true}, {S: "250", Num: true}, {S: "25000", Num: true}},
		{{S: "A2"}, {S: "Bóng T10"}, {S: "1000", Num: true}, {S: "0.35", Num: true}, {S: "350", Num: true}},
		{{S: "A3"}, {S: "Bóng tặng"}, {S: "5", Num: true}, {S: "0", Num: true}, {S: "0", Num: true}},
	}
	for i, r := range rows {
		for c, v := range r {
			g.Cells[[2]int{i + 2, c + 1}] = v
		}
	}
	g.Cells[[2]int{5, 1}] = Cell{S: "Cộng tiền hàng:"}
	g.Cells[[2]int{5, 5}] = Cell{S: "25350", Num: true}
	g.MaxRow, g.MaxCol = 5, 5
	p, err := Parse(g)
	if err != nil {
		t.Fatal(err)
	}
	out := p.OutputRows()
	if p.Scale != 1000 || out[0].Price != 250000 || out[1].Price != 350 || out[2].Price != 0 {
		t.Fatalf("cần nhân 1.000: scale %v, giá %v %v %v", p.Scale, out[0].Price, out[1].Price, out[2].Price)
	}
	if len(p.Warnings) == 0 || !p.NeedConfirm() {
		t.Fatal("phải có cảnh báo và bắt xác nhận")
	}
}

// Hóa đơn có Đơn giá (gốc), Giảm giá %, Giá bán (đã trừ): nhập giá gốc và ghi % vào cột Giảm giá (%) (Khiêm 14/11/2026).
func TestOriginalPriceWithPercentDiscount(t *testing.T) {
	g := newGrid("ChiTietHoaDon_HD086017_KV141120")
	hdr := []string{"Mã hàng", "Tên hàng", "Số lượng", "Đơn giá", "Giảm giá %", "Giảm giá", "Giá bán", "Thành tiền"}
	for c, h := range hdr {
		g.Cells[[2]int{1, c + 1}] = Cell{S: h}
	}
	num := func(s string) Cell { return Cell{S: s, Num: true} }
	rows := [][]Cell{
		{{S: "HDNVIVAFI"}, {S: "Bộ nhông sên dĩa"}, num("2"), num("317000"), num("10"), num("31700"), num("285300"), num("570600")},
		{{S: "6301ZZ"}, {S: "Bạc đạn"}, num("6"), num("40100"), num("100"), num("40100"), num("0"), num("0")},
		{{S: "X1"}, {S: "Hàng không giảm"}, num("1"), num("50000"), num("0"), num("0"), num("50000"), num("50000")},
	}
	for i, r := range rows {
		for c, v := range r {
			g.Cells[[2]int{i + 2, c + 1}] = v
		}
	}
	g.Cells[[2]int{5, 1}] = Cell{S: "Tổng tiền hàng"}
	g.Cells[[2]int{5, 2}] = num("620600")
	g.MaxRow, g.MaxCol = 5, len(hdr)
	p, err := Parse(g)
	if err != nil {
		t.Fatal(err)
	}
	out := p.OutputRows()
	if out[0].Price != 317000 || out[0].CK.S != "10" {
		t.Fatalf("dòng 1 phải là giá gốc 317000 và giảm 10%%: %v %q", out[0].Price, out[0].CK.S)
	}
	if out[1].Price != 40100 || out[1].CK.S != "100" || out[1].Kind != KindPromo {
		t.Fatalf("dòng giảm 100%%: %v %q %s", out[1].Price, out[1].CK.S, out[1].Kind)
	}
	if out[2].Price != 50000 || !out[2].CK.Empty() {
		t.Fatalf("dòng giảm 0%% phải để trống cột giảm giá: %q", out[2].CK.S)
	}
	if p.Meta.So != "HD086017" {
		t.Fatalf("số hóa đơn từ tên sheet: %q", p.Meta.So)
	}
	for _, c := range p.Checks {
		if c.Level != LevelOK {
			t.Fatalf("đối chiếu %s lệch: %s", c.Name, c.Detail)
		}
	}
}

func discountGrid(amount string) *Grid {
	g := newGrid("Sheet1")
	hdr := []string{"STT", "Mã hàng", "Tên hàng", "Số lượng", "Đơn giá", "Giảm giá %", "Thành tiền"}
	for c, h := range hdr {
		g.Cells[[2]int{1, c + 1}] = Cell{S: h}
	}
	num := func(s string) Cell { return Cell{S: s, Num: true} }
	row := []Cell{num("1"), {S: "A1"}, {S: "Nhông"}, num("2"), num("100000"), num("10"), num(amount)}
	for c, v := range row {
		g.Cells[[2]int{2, c + 1}] = v
	}
	g.Cells[[2]int{3, 1}] = Cell{S: "Tổng tiền hàng"}
	g.Cells[[2]int{3, 7}] = num(amount)
	g.MaxRow, g.MaxCol = 3, len(hdr)
	return g
}

func hasWarning(p *Parsed, sub string) bool {
	for _, w := range p.Warnings {
		if strings.Contains(w, sub) {
			return true
		}
	}
	return false
}

// Thành tiền đã trừ % (2 × 100.000 × 90% = 180.000): chắc chắn đơn giá là giá gốc, không cảnh báo.
// Thành tiền chưa trừ (200.000) và không có cột giá sau giảm: không kiểm chứng được, phải cảnh báo.
func TestDiscountVerification(t *testing.T) {
	p, err := Parse(discountGrid("180000"))
	if err != nil {
		t.Fatal(err)
	}
	if hasWarning(p, "không kiểm chứng được") {
		t.Fatalf("không được cảnh báo khi thành tiền khớp giá sau giảm: %v", p.Warnings)
	}
	p, err = Parse(discountGrid("200000"))
	if err != nil {
		t.Fatal(err)
	}
	if !hasWarning(p, "không kiểm chứng được") || p.OutputRows()[0].CK.S != "10" {
		t.Fatalf("phải cảnh báo khi không kiểm chứng được giá gốc: %v", p.Warnings)
	}
}

// Thả nhầm file nhập hàng đã chuyển đổi: báo lỗi rõ ràng, không coi là hóa đơn.
func TestRejectAlreadyConvertedFile(t *testing.T) {
	g := newGrid("PurchaseOrderTemplate")
	for c, h := range OutHeader {
		g.Cells[[2]int{1, c + 1}] = Cell{S: h}
	}
	g.Cells[[2]int{2, 1}] = Cell{S: "A1"}
	g.Cells[[2]int{2, 4}] = Cell{S: "1000", Num: true}
	g.Cells[[2]int{2, 7}] = Cell{S: "2", Num: true}
	g.MaxRow, g.MaxCol = 2, 7
	if _, err := ParseAll(g, Options{}); err == nil || !strings.Contains(err.Error(), "đã chuyển đổi") {
		t.Fatalf("phải từ chối file đã chuyển đổi: %v", err)
	}
}

// Chiết khấu cả hóa đơn: ghi % thì chia đúng % đó; ghi số tiền thì đổi ra % trên tổng tiền hàng; dòng đã có % riêng thì gộp.
func TestSpreadInvoiceDiscount(t *testing.T) {
	mk := func(label, val string) *Grid {
		g := newGrid("Sheet1")
		hdr := []string{"STT", "Mã hàng", "Tên hàng", "Số lượng", "Đơn giá", "Thành tiền"}
		for c, h := range hdr {
			g.Cells[[2]int{1, c + 1}] = Cell{S: h}
		}
		num := func(s string) Cell { return Cell{S: s, Num: true} }
		rows := [][]Cell{
			{num("1"), {S: "A1"}, {S: "Nhông"}, num("2"), num("100000"), num("200000")},
			{num("2"), {S: "A2"}, {S: "Sên"}, num("1"), num("300000"), num("300000")},
			{num("3"), {S: "A3"}, {S: "Quà tặng"}, num("1"), num("0"), num("0")},
		}
		for i, r := range rows {
			for c, v := range r {
				g.Cells[[2]int{i + 2, c + 1}] = v
			}
		}
		g.Cells[[2]int{5, 2}] = Cell{S: "Tổng tiền hàng:"}
		g.Cells[[2]int{5, 6}] = num("500000")
		g.Cells[[2]int{6, 2}] = Cell{S: label}
		g.Cells[[2]int{6, 3}] = Cell{S: val, Num: !strings.Contains(val, "%")}
		g.MaxRow, g.MaxCol = 6, len(hdr)
		return g
	}
	p, err := Parse(mk("Tỷ lệ CK:", "5 %"))
	if err != nil {
		t.Fatal(err)
	}
	out := p.OutputRows()
	if p.InvoiceCK != 5 || out[0].CK.S != "5" || out[1].CK.S != "5" || !out[2].CK.Empty() {
		t.Fatalf("chia 5%%: %v %q %q %q", p.InvoiceCK, out[0].CK.S, out[1].CK.S, out[2].CK.S)
	}
	p, err = Parse(mk("Chiết khấu:", "25000"))
	if err != nil {
		t.Fatal(err)
	}
	if out = p.OutputRows(); p.InvoiceCK != 5 || out[0].CK.S != "5" || hasWarning(p, "lệch") {
		t.Fatalf("25.000/500.000 phải là 5%% không lệch: %v %q %v", p.InvoiceCK, out[0].CK.S, p.Warnings)
	}
	p, err = Parse(mk("Chiết khấu:", "25123"))
	if err != nil {
		t.Fatal(err)
	}
	if !hasWarning(p, "lệch") {
		t.Fatalf("25.123 đổi ra 5,02%% thì phải cảnh báo lệch làm tròn: %v %v", p.InvoiceCK, p.Warnings)
	}
}
