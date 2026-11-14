package server

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// Xuất xong một hóa đơn rồi mở lại đúng hóa đơn đó (kể cả đổi tên file): app cảnh báo đã xuất và chỉ xuất lại khi xác nhận.
func TestDuplicateExportWarning(t *testing.T) {
	tpl, err := os.ReadFile(filepath.Join("..", "..", "cmd", "gkm", "template.xlsx"))
	if err != nil {
		t.Skip("thiếu template.xlsx")
	}
	a := newApp(t)
	a.Template = tpl
	h := a.Handler()
	out := t.TempDir()
	data := invoiceXLSX(t)
	m := postConvert(t, h, nil, "hd.xlsx", data)
	if m["ok"] != true || m["duplicate"] != nil {
		t.Fatalf("lần đầu không được báo trùng: %v", m)
	}
	if code, r := do(t, h, "POST", "/api/export", fmt.Sprintf(`{"id":%q,"dir":%q,"confirm":true}`, m["id"], out), nil); code != 200 {
		t.Fatalf("xuất lần đầu: %d %v", code, r)
	}
	m = postConvert(t, h, nil, "hd (1).xlsx", data)
	dup, _ := m["duplicate"].(map[string]any)
	if dup == nil || m["needConfirm"] != true || !strings.Contains(fmt.Sprint(dup["path"]), out) {
		t.Fatalf("lần hai phải báo đã xuất: %v", m)
	}
	if code, _ := do(t, h, "POST", "/api/export", fmt.Sprintf(`{"id":%q,"dir":%q,"confirm":false}`, m["id"], out), nil); code != 400 {
		t.Fatalf("xuất trùng chưa xác nhận phải bị chặn, được %d", code)
	}
	if code, _ := do(t, h, "POST", "/api/export", fmt.Sprintf(`{"id":%q,"dir":%q,"confirm":true}`, m["id"], out), nil); code != 200 {
		t.Fatalf("đã xác nhận thì phải xuất được, được %d", code)
	}
	// hóa đơn khác thì không báo trùng
	if m = postConvert(t, h, nil, "khac.xlsx", twoToaXLSX(t)); m["ok"] != true || m["duplicate"] != nil {
		t.Fatalf("hóa đơn khác không được báo trùng: %v", m["duplicate"])
	}
}
