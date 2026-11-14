package server

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// twoToaXLSX: một file Excel có 2 toa hàng ở 2 sheet (như toa Tân Hòa Lợi + Hồng Nhất Phát gửi chung).
func twoToaXLSX(t *testing.T) []byte {
	t.Helper()
	f := excelize.NewFile()
	_ = f.SetSheetName("Sheet1", "THL")
	_, _ = f.NewSheet("HNP")
	put := func(sh string, rows [][]any) {
		for i, r := range rows {
			for j, v := range r {
				c, _ := excelize.CoordinatesToCellName(j+1, i+1)
				_ = f.SetCellValue(sh, c, v)
			}
		}
	}
	hdr := []any{"STT", "MÃ HÀNG", "HIỆU", "SL", "ĐƠN GIÁ", "THÀNH TIỀN"}
	put("THL", [][]any{{"", "", "TAN HOA LOI CO., LTD"}, {"", "", "", "", "", "11/11/2026"}, hdr,
		{1, "6004-ZZ", "KOYO", 20, 42, 840}, {2, "6201-ZZ", "KOYO", 30, 30, 900}, {"TỔNG TOA:", "", "", "", "", 1740}})
	put("HNP", [][]any{{"", "", "TOA HÀNG"}, {"", "", "", "", "", "Ngày: 11/11/2026"}, hdr,
		{1, "TC-30X52X10", "TQ", 10, 3.9, 39}, {"TỔNG SỐ TIỀN:", "", "", "", "", 39}, {"", "CÔNG TY TNHH HỒNG NHẤT PHÁT"}})
	var b bytes.Buffer
	if err := f.Write(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestConvertTwoToaExportsTwoFiles(t *testing.T) {
	tpl, err := os.ReadFile(filepath.Join("..", "..", "cmd", "gkm", "template.xlsx"))
	if err != nil {
		t.Skip("thiếu template.xlsx (build.sh chép vào khi đóng gói)")
	}
	a := newApp(t)
	a.Template = tpl
	h := a.Handler()
	m := postConvert(t, h, nil, "toa.xlsx", twoToaXLSX(t))
	parts, _ := m["parts"].([]any)
	if m["ok"] != true || len(parts) != 2 {
		t.Fatalf("cần 2 toa: %v", m)
	}
	out := t.TempDir()
	want := []string{"NhapHang_TanHoaLoi_11112026.xlsx", "NhapHang_HongNhatPhat_11112026.xlsx"}
	for i, x := range parts {
		p := x.(map[string]any)
		if p["suggestedName"] != want[i] {
			t.Fatalf("toa %d tên %v, cần %s", i+1, p["suggestedName"], want[i])
		}
		code, r := do(t, h, "POST", "/api/export", fmt.Sprintf(`{"id":%q,"dir":%q,"confirm":true}`, p["id"], out), nil)
		if code != 200 || r["ok"] != true {
			t.Fatalf("xuất toa %d: %d %v", i+1, code, r)
		}
		if _, err := os.Stat(filepath.Join(out, want[i])); err != nil {
			t.Fatalf("không thấy file toa %d: %v", i+1, err)
		}
	}
}
