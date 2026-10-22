package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const samples = "../../../samples"

func read(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(samples, name))
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("thiếu hóa đơn mẫu %q (thư mục samples không đưa lên kho công khai)", name)
		}
		t.Fatal(err)
	}
	return b
}

func TestSamples(t *testing.T) {
	tmpl := read(t, "MauFileNhapHang.xlsx")
	for _, r := range RunSelfTest(func(n string) ([]byte, error) { return os.ReadFile(filepath.Join(samples, n)) }, tmpl) {
		t.Log(r.Summary)
		if !r.OK {
			t.Errorf("Mẫu %d: %v", r.N, r.Problems)
		}
	}
}

func TestLeadingZeroKeptAsText(t *testing.T) {
	tmpl := read(t, "MauFileNhapHang.xlsx")
	out, err := BuildWorkbook(tmpl, []OutRow{{Code: "00123", Name: "Thử", Price: 1000, Qty: Cell{S: "2", Num: true}}})
	if err != nil {
		t.Fatal(err)
	}
	g, _ := ReadGrid(out)
	if c := g.At(2, 1); c.S != "00123" || c.Num {
		t.Fatalf("mã mất số 0 hoặc thành số: %+v", c)
	}
}

func TestNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	p1 := UniquePath(dir, "a.xlsx")
	if err := WriteNew(p1, []byte("x")); err != nil {
		t.Fatal(err)
	}
	p2 := UniquePath(dir, "a.xlsx")
	if p1 == p2 {
		t.Fatal("trùng đường dẫn")
	}
	if err := WriteNew(p1, []byte("y")); err == nil {
		t.Fatal("phải từ chối ghi đè")
	}
}

func TestUnknownFormat(t *testing.T) {
	// MauFileNhapHang (không có STT/Mã hàng kiểu hóa đơn) phải bị từ chối rõ ràng, không đoán.
	g, _ := ReadGrid(read(t, "MauFileNhapHang.xlsx"))
	if _, err := Parse(g); err == nil {
		t.Fatal("phải báo lỗi vì không có dòng tiêu đề hóa đơn")
	}
}

const mau4 = "Mau 4 - Toa hang Tan Hoa Loi.xlsx"

func TestMau4TanHoaLoi(t *testing.T) {
	g, err := ReadGrid(read(t, mau4))
	if err != nil {
		t.Fatal(err)
	}
	p, err := Parse(g)
	if err != nil {
		t.Fatal(err)
	}
	if p.Profile != ProfileD || len(p.Items) != 27 {
		t.Fatalf("profile %s, %d dòng", p.Profile, len(p.Items))
	}
	if p.Scale != 1000 || p.Items[0].Price != 7500 || p.Items[24].Price != 3100 || p.Items[26].Code != "60/22-ZZ" {
		t.Fatalf("giá/mã sai: %+v %+v", p.Items[0], p.Items[24])
	}
	if q, _ := p.Items[0].Qty.Float(); q != 20 || !p.Items[0].Qty.Num {
		t.Fatalf("số lượng sai: %+v", p.Items[0].Qty)
	}
	if !p.Known || p.HasName || p.ScaleNote != "" || p.Meta.Ngay != "01/04/2026" || p.SuggestedName() != "NhapHang_TanHoaLoi_01042026.xlsx" {
		t.Fatalf("meta: %+v %v %s", p.Meta.Ngay, p.Known, p.SuggestedName())
	}
	var okTotal bool
	tail := 0
	for _, c := range p.Checks {
		if c.Name == "Tổng tiền hàng" && c.Level == LevelOK {
			okTotal = true
		}
		if strings.HasPrefix(c.Name, "Dòng sau tổng:") {
			tail++
		}
	}
	if !okTotal || tail != 3 {
		t.Fatalf("đối chiếu: %+v", p.Checks)
	}
	rows := p.OutputRows()
	if len(rows) != 27 || rows[0].Name != "" || rows[0].Code != "26X37X10.5" {
		t.Fatalf("output: %+v", rows[0])
	}
	// Khiêm xác nhận: Tân Hòa Lợi tính nghìn đồng, không có cột tên thì để trống, không còn cảnh báo về hai điều này
	for _, w := range p.Warnings {
		if strings.Contains(w, "NGHÌN") || strings.Contains(w, "Tên hàng") {
			t.Fatalf("cảnh báo thừa: %s", w)
		}
	}
	if p2, _ := ParseWith(g, Options{PriceScale: 1}); p2.Items[0].Price != 7.5 {
		t.Fatalf("tùy chọn thử nghiệm: %+v", p2.Items[0])
	}
	// xuất rồi đọc lại
	out, err := BuildWorkbook(read(t, "MauFileNhapHang.xlsx"), rows)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyExport(out, rows); err != nil {
		t.Fatal(err)
	}
}

func TestOtherFormatsEqualXLSX(t *testing.T) {
	cases := []struct {
		src, file string
		typed     bool
	}{
		{"Mau 1.xlsx", "Mau1.xls", true}, {"Mau 2.xlsx", "Mau2.xls", true},
		{"Mau 3.xlsx", "Mau3.xls", true}, {mau4, "Mau4.xls", true},
		{"Mau 3.xlsx", "Mau3_xml2003.xls", true}, {mau4, "Mau4_xml2003.xls", true},
		{"Mau 3.xlsx", "Mau3_html.xls", false}, {mau4, "Mau4_html.xls", false},
	}
	for _, c := range cases {
		g0, err := ReadGrid(read(t, c.src))
		if err != nil {
			t.Fatal(err)
		}
		p0, err := Parse(g0)
		if err != nil {
			t.Fatal(err)
		}
		g1, kind, err := ReadGridAny(read(t, "dinh-dang-khac/"+c.file))
		if err != nil {
			t.Fatalf("%s: %v", c.file, err)
		}
		p1, err := Parse(g1)
		if err != nil {
			t.Fatalf("%s: %v", c.file, err)
		}
		a, b := p0.OutputRows(), p1.OutputRows()
		if len(a) != len(b) || p0.Profile != p1.Profile {
			t.Fatalf("%s (%s): %d/%d dòng, mẫu %s/%s", c.file, kind, len(a), len(b), p0.Profile, p1.Profile)
		}
		for i := range a {
			same := a[i].Code == b[i].Code && a[i].Name == b[i].Name && a[i].Price == b[i].Price &&
				a[i].CK.S == b[i].CK.S && a[i].Qty.S == b[i].Qty.S
			if c.typed {
				same = same && a[i].Qty.Num == b[i].Qty.Num
			}
			if !same {
				t.Fatalf("%s dòng %d khác: %+v / %+v", c.file, i+1, a[i], b[i])
			}
		}
		t.Logf("%s (%s): %d dòng khớp bản .xlsx", c.file, kind, len(a))
	}
}

func TestReadAnyRejectsGarbage(t *testing.T) {
	if _, _, err := ReadGridAny([]byte("khong phai excel")); err == nil {
		t.Fatal("phải từ chối tệp không phải bảng tính")
	}
	if _, err := ReadXLS([]byte("\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1" + string(make([]byte, 600)))); err == nil {
		t.Fatal("phải báo lỗi với .xls hỏng")
	}
}

func TestBigXLSContinueRecords(t *testing.T) {
	g, kind, err := ReadGridAny(read(t, "dinh-dang-khac/Lon_3000_dong.xls"))
	if err != nil || kind != KindXLS {
		t.Fatal(kind, err)
	}
	p, err := Parse(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) != 3000 {
		t.Fatalf("%d dòng", len(p.Items))
	}
	for _, i := range []int{1, 2, 1499, 2999, 3000} {
		it := p.Items[i-1]
		var want string
		if i%2 == 1 {
			want = strings.Repeat(fmt.Sprintf("Bạc đạn rất dài số %d ", i), 3)
		} else {
			want = strings.Repeat(fmt.Sprintf("Plain ascii name number %d ", i), 3)
		}
		if it.Code != fmt.Sprintf("%07d", i) || it.Name != strings.TrimSpace(want) || it.Price != float64(1000+i) {
			t.Fatalf("dòng %d sai: %q %q %v", i, it.Code, it.Name, it.Price)
		}
	}
}

func TestMau5RealXLS(t *testing.T) {
	g, kind, err := ReadGridAny(read(t, "Mau 5 - gk 28-2.xls"))
	if err != nil || kind != KindXLS {
		t.Fatal(kind, err)
	}
	p, err := Parse(g)
	if err != nil {
		t.Fatal(err)
	}
	if p.Profile != ProfileE || len(p.Items) != 281 || len(p.Errors) != 0 {
		t.Fatalf("mẫu %s, %d dòng, lỗi %v", p.Profile, len(p.Items), p.Errors)
	}
	if p.SuggestedName() != "NhapHang_DB26HA-00427_27022026.xlsx" || p.Meta.So != "DB26HA.00427" {
		t.Fatalf("tên file %s, số %s", p.SuggestedName(), p.Meta.So)
	}
	okTotal := false
	for _, c := range p.Checks {
		if c.Name == "Tổng tiền hàng" && c.Level == LevelOK {
			okTotal = true
		}
		if c.Level != LevelOK {
			t.Errorf("đối chiếu không đạt: %+v", c)
		}
	}
	if !okTotal {
		t.Fatal("tổng tiền hàng không khớp")
	}
	it := p.Items[0]
	if it.Code != "06430GCE910" || it.Name != "Bộ má phanh" || it.Qty.S != "40" || !it.Qty.Num || it.Price != 63892.8 {
		t.Fatalf("dòng 1 sai: %+v", it)
	}
	if last := p.Items[280]; last.Code != "9755031209K0SS" || last.Name != "Bộ nan hoa trước,ngoài A11x184.5" {
		t.Fatalf("dòng cuối sai: %+v", last)
	}
	rows := p.OutputRows()
	out, err := BuildWorkbook(read(t, "MauFileNhapHang.xlsx"), rows)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyExport(out, rows); err != nil {
		t.Fatal(err)
	}
}

func TestDetectThousandNote(t *testing.T) {
	yes := []string{"Đơn vị tính: nghìn đồng", "(ĐVT: 1.000đ)", "Đơn giá (nghìn đồng)", "Đơn giá (1000đ)", "Thành tiền (K)", "Giá (ngàn)", "ĐVT: 1000", "x1000 đồng", "Tiền tính bằng NGHÌN ĐỒNG"}
	no := []string{"Đơn vị tính: đồng", "Đvt", "Cái", "Giá khuyến mại", "Giá net", "BI 6000 (1000)", "Ghi chú: giao hàng 1.000 cái", "VNĐ", "Giá trị đơn hàng", "Sáu mươi bảy triệu tám trăm linh tám nghìn đồng chẵn.", "Một trăm nghìn đồng", "Số tiền viết bằng chữ: năm mươi nghìn đồng"}
	for _, x := range yes {
		g := &Grid{Cells: map[[2]int]Cell{{3, 2}: {S: x}}, MaxRow: 3, MaxCol: 2}
		if detectThousandNote(g) == "" {
			t.Errorf("phải nhận ra: %q", x)
		}
	}
	for _, x := range no {
		g := &Grid{Cells: map[[2]int]Cell{{3, 2}: {S: x}}, MaxRow: 3, MaxCol: 2}
		if n := detectThousandNote(g); n != "" {
			t.Errorf("không được nhận nhầm: %q", x)
		}
	}
	g := &Grid{Cells: map[[2]int]Cell{{3, 2}: {S: "1000", Num: true}}, MaxRow: 3, MaxCol: 2}
	if detectThousandNote(g) != "" {
		t.Error("ô số không được tính")
	}
}

func gridOf(rows [][]string) *Grid {
	g := &Grid{Cells: map[[2]int]Cell{}}
	for r, row := range rows {
		for c, v := range row {
			if v == "" {
				continue
			}
			_, err := strconv.ParseFloat(v, 64)
			g.Cells[[2]int{r + 1, c + 1}] = Cell{S: v, Num: err == nil}
			if c+1 > g.MaxCol {
				g.MaxCol = c + 1
			}
		}
	}
	g.MaxRow = len(rows)
	return g
}

// Khiêm 04/10/2026: chỉ có số lượng, không có giá -> vẫn nhập với đơn giá 0; có giá mà không có số lượng -> bỏ.
func TestQtyPriceRules(t *testing.T) {
	g := gridOf([][]string{
		{"STT", "Mã hàng", "Tên hàng", "Số lượng", "Đơn giá", "Thành tiền"},
		{"1", "A1", "Hàng thường", "2", "1000", "2000"},
		{"2", "B2", "Hàng bảo hành", "3", "", ""},
		{"3", "C3", "Có giá, không số lượng", "", "5000", ""},
		{"4", "D4", "Trống cả hai", "", "", ""},
		{"5", "E5", "Hàng thường 2", "1", "700", "700"},
	})
	p, err := Parse(g)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Errors) > 0 {
		t.Fatalf("không được báo lỗi: %v", p.Errors)
	}
	rows := p.OutputRows()
	got := ""
	for _, r := range rows {
		q, _ := r.Qty.Float()
		got += fmt.Sprintf("%s:%v:%v ", r.Code, r.Price, q)
	}
	if got != "A1:1000:2 B2:0:3 E5:700:1 " {
		t.Fatalf("file xuất sai: %s", got)
	}
	n, k, d, x := p.Counts()
	if n != 2 || k != 1 || d != 0 || x != 2 {
		t.Fatalf("đếm: %d %d %d %d", n, k, d, x)
	}
}

func TestAllRowsWithoutQtyRejected(t *testing.T) {
	g := gridOf([][]string{
		{"STT", "Mã hàng", "Tên hàng", "Số lượng", "Đơn giá"},
		{"1", "A1", "x", "", "1000"},
	})
	if _, err := Parse(g); err == nil || !strings.Contains(err.Error(), "không có số lượng") {
		t.Fatalf("phải từ chối: %v", err)
	}
}

var hdrF = []string{"STT", "MÃ PHỤ TÙNG", "TÊN PHỤ TÙNG", "SL ĐẶT", "SL LẤY", "ĐƠN GIÁ BÁN", "ĐG ĐÃ CK 22%", "THÀNH TIỀN"}

func toaF(day string, rows ...[]string) [][]string {
	out := [][]string{{"PHIẾU XUẤT KHO"}, {day}, hdrF}
	out = append(out, rows...)
	return append(out, []string{"", "", "TỔNG"})
}

// Khiêm 22/10/2026: có SL đặt và SL lấy thì dùng SL lấy; có đơn giá bán và đơn giá đã CK thì dùng đơn giá bán;
// NCC gửi cả chồng toa thì chỉ lấy bảng cuối; SL lấy = 0 thì bỏ dòng.
func TestProfileFLastTableSLLay(t *testing.T) {
	old := toaF("45000", []string{"1", "OLD1", "Cũ", "5", "5", "1000", "780", "5000"})
	cur := toaF("46315",
		[]string{"1", "NEW1", "Mới 1", "4", "3", "2000", "1560", "6000"},
		[]string{"2", "NEW2", "Mới 2", "2", "0", "500", "390", "0"},
		[]string{"3", "NEW3", "Mới 3", "1", "1", "249381.72000000003", "194517.74", "249381.72000000003"})
	// tính tổng để khớp
	cur[len(cur)-1] = []string{"", "", "TỔNG", "", "4", "", "", "255381.72"}
	// 1) cùng một sheet, bảng cũ phía trên
	g := gridOf(append(append([][]string{}, old...), cur...))
	p, err := Parse(g)
	if err != nil {
		t.Fatal(err)
	}
	if p.Profile != ProfileF || len(p.Errors) > 0 {
		t.Fatalf("profile %s lỗi %v", p.Profile, p.Errors)
	}
	got := ""
	for _, r := range p.OutputRows() {
		q, _ := r.Qty.Float()
		got += fmt.Sprintf("%s:%v:%v ", r.Code, r.Price, q)
	}
	if got != "NEW1:2000:3 NEW3:249381.72:1 " {
		t.Fatalf("chọn sai bảng/cột: %s", got)
	}
	if p.Meta.Ngay != "20/10/2026" || p.SuggestedName() != "NhapHang_PhieuXuatKho_20102026.xlsx" {
		t.Fatalf("ngày/tên: %q %s", p.Meta.Ngay, p.SuggestedName())
	}
	if len(p.Notes) != 1 || !strings.Contains(p.Notes[0].Detail, "2 bảng") {
		t.Fatalf("thiếu ghi chú nhiều bảng: %+v", p.Notes)
	}
	for _, c := range p.Checks {
		if c.Level != LevelOK {
			t.Fatalf("đối chiếu lệch: %+v", c)
		}
	}
	// 2) mỗi toa một sheet: lấy sheet cuối
	g1, g2 := gridOf(old), gridOf(cur)
	g1.Sheet, g2.Sheet = "T1", "T2"
	g1.Extra = []*Grid{g2}
	p2, err := Parse(g1)
	if err != nil || len(p2.OutputRows()) != 2 || p2.OutputRows()[0].Code != "NEW1" {
		t.Fatalf("nhiều sheet: %v %+v", err, p2)
	}
	// 3) tiêu đề lặp ở trang sau (không có dòng tổng ở giữa) thì KHÔNG phải hóa đơn mới
	page := [][]string{{"PHIẾU"}, {"46315"}, hdrF,
		{"1", "P1", "a", "1", "1", "100", "78", "100"},
		hdrF,
		{"2", "P2", "b", "1", "1", "100", "78", "100"},
		{"", "", "TỔNG", "", "2", "", "", "200"}}
	p3, err := Parse(gridOf(page))
	if err != nil || len(p3.OutputRows()) != 2 || len(p3.Notes) != 0 {
		t.Fatalf("tiêu đề lặp bị hiểu nhầm: %v %+v", err, p3)
	}
}
