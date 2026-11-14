package core

import "testing"

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
