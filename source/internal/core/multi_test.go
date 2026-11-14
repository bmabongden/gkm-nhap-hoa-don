package core

import (
	"strconv"
	"testing"
)

// toaGrid dựng một toa kiểu Tân Hòa Lợi bắt đầu từ dòng r0: tên NCC, ngày, tiêu đề, các dòng hàng, dòng tổng.
func toaGrid(g *Grid, r0 int, ncc, ngay string, rows [][3]any) int {
	set := func(r, c int, v any) {
		switch x := v.(type) {
		case string:
			g.Cells[[2]int{r, c}] = Cell{S: x}
		case float64:
			g.Cells[[2]int{r, c}] = Cell{S: strconv.FormatFloat(x, 'f', -1, 64), Num: true}
		}
		if r > g.MaxRow {
			g.MaxRow = r
		}
		if c > g.MaxCol {
			g.MaxCol = c
		}
	}
	set(r0, 3, ncc)
	set(r0+1, 1, "TOA HÀNG")
	set(r0+2, 6, ngay)
	for c, h := range []string{"STT", "MÃ HÀNG", "HIỆU", "SL", "ĐƠN GIÁ", "THÀNH TIỀN"} {
		set(r0+3, c+1, h)
	}
	r, sum := r0+4, 0.0
	for i, it := range rows {
		set(r, 1, float64(i+1))
		set(r, 2, it[0].(string))
		set(r, 3, "KOYO")
		set(r, 4, it[1].(float64))
		set(r, 5, it[2].(float64))
		set(r, 6, it[1].(float64)*it[2].(float64))
		sum += it[1].(float64) * it[2].(float64)
		r++
	}
	set(r, 1, "TỔNG TOA:")
	set(r, 6, sum)
	return r
}

func newGrid(name string) *Grid { return &Grid{Sheet: name, Cells: map[[2]int]Cell{}} }

// Hai toa ở hai sheet: mỗi toa một kết quả, đơn giá nhân 1.000, tên file theo nhà cung cấp.
func TestParseAllTwoSheets(t *testing.T) {
	g := newGrid("THL")
	toaGrid(g, 1, "TAN  HOA  LOI  CO. , LTD", "11/11/2026", [][3]any{{"6004-ZZ", 20.0, 42.0}, {"6201-ZZ", 30.0, 30.0}})
	h := newGrid("HNP")
	toaGrid(h, 1, "CÔNG TY TNHH HỒNG NHẤT PHÁT", "Ngày: 12/11/2026", [][3]any{{"TC-30X52X10", 10.0, 3.9}})
	g.Extra = []*Grid{h}
	parts, err := ParseAll(g, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 {
		t.Fatalf("cần 2 toa, được %d", len(parts))
	}
	if n := parts[0].SuggestedName(); n != "NhapHang_TanHoaLoi_11112026.xlsx" {
		t.Fatalf("tên toa 1: %s", n)
	}
	if n := parts[1].SuggestedName(); n != "NhapHang_HongNhatPhat_12112026.xlsx" {
		t.Fatalf("tên toa 2: %s", n)
	}
	if r := parts[1].OutputRows(); len(r) != 1 || r[0].Price != 3900 {
		t.Fatalf("toa 2 sai: %+v", r)
	}
	for _, p := range parts {
		for _, c := range p.Checks {
			if c.Level != LevelOK {
				t.Fatalf("%s: đối chiếu %s lệch: %s", p.Part, c.Name, c.Detail)
			}
		}
	}
}

// Hai toa xếp liền nhau trong một sheet (cách nhau bằng dòng tổng): vẫn tách thành 2, không lẫn dòng, trùng tên thì thêm số toa.
func TestParseAllStackedInOneSheet(t *testing.T) {
	g := newGrid("Sheet1")
	end := toaGrid(g, 1, "TAN HOA LOI", "11/11/2026", [][3]any{{"6004-ZZ", 20.0, 42.0}, {"6201-ZZ", 30.0, 30.0}})
	toaGrid(g, end+3, "TAN HOA LOI", "11/11/2026", [][3]any{{"6205-2RS", 50.0, 17.0}})
	parts, err := ParseAll(g, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 || len(parts[0].OutputRows()) != 2 || len(parts[1].OutputRows()) != 1 {
		t.Fatalf("tách sai: %d toa", len(parts))
	}
	if parts[0].SuggestedName() == parts[1].SuggestedName() {
		t.Fatalf("hai toa trùng tên file: %s", parts[0].SuggestedName())
	}
	if parts[1].Meta.Ngay != "11/11/2026" {
		t.Fatalf("toa 2 mất ngày: %q", parts[1].Meta.Ngay)
	}
}

// File một toa: giữ nguyên như trước (1 kết quả, không ghi chú nhiều toa).
func TestParseAllSingle(t *testing.T) {
	g := newGrid("Sheet1")
	toaGrid(g, 1, "TAN HOA LOI", "11/11/2026", [][3]any{{"6004-ZZ", 20.0, 42.0}})
	g.Extra = []*Grid{{Sheet: "Ghi chú", Cells: map[[2]int]Cell{{1, 1}: {S: "không phải toa"}}, MaxRow: 1, MaxCol: 1}}
	parts, err := ParseAll(g, Options{})
	if err != nil || len(parts) != 1 || parts[0].PartOf != 0 || parts[0].SuggestedName() != "NhapHang_TanHoaLoi_11112026.xlsx" {
		t.Fatalf("file 1 toa bị đổi: %v %d", err, len(parts))
	}
}
