package server

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

// invoiceXLSX dựng một hóa đơn nhỏ đủ cột để app nhận biết và chuyển đổi.
func invoiceXLSX(t *testing.T) []byte {
	t.Helper()
	f := excelize.NewFile()
	rows := [][]any{
		{"Hóa đơn thử"},
		{"STT", "Mã hàng", "Tên hàng", "ĐVT", "Số lượng", "Đơn giá", "Thành tiền"},
		{1, "AB100", "Bố thắng", "Cái", 2, 50000, 100000},
		{2, "AB200", "Nhông xích", "Bộ", 1, 120000, 120000},
		{"", "", "Tổng cộng", "", 3, "", 220000},
	}
	for i, r := range rows {
		for j, v := range r {
			c, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = f.SetCellValue("Sheet1", c, v)
		}
	}
	var b bytes.Buffer
	if err := f.Write(&b); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func postConvert(t *testing.T, h http.Handler, fields map[string]string, fileName string, fileData []byte) map[string]any {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	if fileData != nil {
		fw, _ := mw.CreateFormFile("file", fileName)
		_, _ = fw.Write(fileData)
	}
	_ = mw.Close()
	req := httptest.NewRequest("POST", "http://127.0.0.1:47631/api/convert", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-GKM", "1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("không phải JSON: %s", rec.Body.String())
	}
	return m
}

func TestConvertFromPathGivesSrcDir(t *testing.T) {
	a := newApp(t)
	h := a.Handler()
	dir := filepath.Join(t.TempDir(), "Hóa đơn tháng 10")
	_ = os.MkdirAll(dir, 0o755)
	file := filepath.Join(dir, "hd1.xlsx")
	_ = os.WriteFile(file, invoiceXLSX(t), 0o644)
	m := postConvert(t, h, map[string]string{"path": file}, "", nil)
	if m["ok"] != true {
		t.Fatalf("convert: %v", m)
	}
	if m["srcDir"] != dir || m["srcExact"] != true {
		t.Fatalf("srcDir/srcExact sai: %v %v", m["srcDir"], m["srcExact"])
	}
	// thư mục được nhớ cho lần mở hộp chọn file sau
	if got := a.loadSettings().LastSrcDir; got != dir {
		t.Fatalf("LastSrcDir=%q", got)
	}
}

func TestConvertPathRejectsNonExcel(t *testing.T) {
	h := newApp(t).Handler()
	d := t.TempDir()
	for name, path := range map[string]string{
		"đuôi lạ":             filepath.Join(d, "a.txt"),
		"đường dẫn tương đối": "hd.xlsx",
		"không tồn tại":       filepath.Join(d, "none.xlsx"),
	} {
		if name == "đuôi lạ" {
			_ = os.WriteFile(path, []byte("x"), 0o644)
		}
		m := postConvert(t, h, map[string]string{"path": path}, "", nil)
		if m["ok"] != false {
			t.Fatalf("%s: phải bị từ chối: %v", name, m)
		}
	}
}

func TestLocateSource(t *testing.T) {
	base := t.TempDir()
	dl := filepath.Join(base, "Downloads")
	sub := filepath.Join(dl, "Zalo", "Nhan")
	_ = os.MkdirAll(sub, 0o755)
	data := invoiceXLSX(t)
	file := filepath.Join(sub, "hd.xlsx")
	_ = os.WriteFile(file, data, 0o644)
	mt := time.Now().Add(-time.Hour)
	_ = os.Chtimes(file, mt, mt)
	size := int64(len(data))

	if got := locateSource("hd.xlsx", size, mt.UnixMilli(), []string{dl}); got != sub {
		t.Fatalf("muốn %q, được %q", sub, got)
	}
	if got := locateSource("HD.XLSX", size, 0, []string{dl}); got != sub {
		t.Fatalf("không phân biệt hoa thường: %q", got)
	}
	if got := locateSource("hd.xlsx", size+1, 0, []string{dl}); got != "" {
		t.Fatalf("sai kích thước phải không tìm thấy, được %q", got)
	}
	if got := locateSource("hd.xlsx", size, mt.Add(time.Hour).UnixMilli(), []string{dl}); got != "" {
		t.Fatalf("sai ngày sửa phải không tìm thấy, được %q", got)
	}
	// hai nơi cùng khớp thì không đoán
	other := filepath.Join(dl, "cu")
	_ = os.MkdirAll(other, 0o755)
	_ = os.WriteFile(filepath.Join(other, "hd.xlsx"), data, 0o644)
	if got := locateSource("hd.xlsx", size, 0, []string{dl}); got != "" {
		t.Fatalf("hai nơi khớp phải trả rỗng, được %q", got)
	}
	// quá sâu (4 cấp dưới gốc) thì bỏ qua
	deep := filepath.Join(base, "R", "a", "b", "c", "d")
	_ = os.MkdirAll(deep, 0o755)
	_ = os.WriteFile(filepath.Join(deep, "x.xlsx"), data, 0o644)
	if got := locateSource("x.xlsx", size, 0, []string{filepath.Join(base, "R")}); got != "" {
		t.Fatalf("quá sâu phải bỏ qua, được %q", got)
	}
}

func TestConvertUploadLocatesByNameAndSize(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	dl := filepath.Join(home, "Downloads")
	_ = os.MkdirAll(dl, 0o755)
	data := invoiceXLSX(t)
	file := filepath.Join(dl, "hdkeo.xlsx")
	_ = os.WriteFile(file, data, 0o644)
	st, _ := os.Stat(file)
	h := newApp(t).Handler()
	m := postConvert(t, h, map[string]string{"size": strconv.Itoa(len(data)), "mtime": strconv.FormatInt(st.ModTime().UnixMilli(), 10)}, "hdkeo.xlsx", data)
	if m["ok"] != true {
		t.Fatalf("convert: %v", m)
	}
	if m["srcDir"] != dl || m["srcExact"] != false {
		t.Fatalf("muốn đoán ra %q, được %v exact=%v", dl, m["srcDir"], m["srcExact"])
	}
	// file không có ở đâu: srcDir rỗng
	m = postConvert(t, h, nil, "khongco.xlsx", data)
	if m["ok"] != true || m["srcDir"] != "" {
		t.Fatalf("muốn srcDir rỗng: %v", m["srcDir"])
	}
	_ = strings.TrimSpace
}
