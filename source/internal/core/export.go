package core

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// OutHeader là tiêu đề bắt buộc của MauFileNhapHang.xlsx.
var OutHeader = []string{"Mã hàng", "Tên hàng", "Đơn vị tính", "Đơn giá", "Giảm giá", "Giảm giá (%)", "Số lượng"}

// BuildWorkbook sao chép workbook mẫu và điền dữ liệu từ dòng 2 (không thêm cột/sheet, giữ định dạng mẫu).
func BuildWorkbook(template []byte, rows []OutRow) ([]byte, error) {
	f, err := excelize.OpenReader(bytes.NewReader(template))
	if err != nil {
		return nil, fmt.Errorf("không mở được file mẫu MauFileNhapHang.xlsx: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("file mẫu không có sheet")
	}
	sh := sheets[0]
	for i, h := range OutHeader { // bảo vệ: mẫu phải đúng tiêu đề
		axis, _ := excelize.CoordinatesToCellName(i+1, 1)
		if v, _ := f.GetCellValue(sh, axis); v != h {
			return nil, fmt.Errorf("file mẫu sai tiêu đề cột %s: có %q, cần %q", axis, v, h)
		}
	}
	setStr := func(col string, r int, s string) error {
		return f.SetCellStr(sh, col+strconv.Itoa(r), s)
	}
	setNum := func(col string, r int, v float64) error {
		axis := col + strconv.Itoa(r)
		if v == math.Trunc(v) && math.Abs(v) < 1e15 {
			return f.SetCellInt(sh, axis, int64(v))
		}
		return f.SetCellFloat(sh, axis, v, -1, 64)
	}
	setCell := func(col string, r int, c Cell) error {
		if c.Empty() {
			return nil
		}
		if c.Num {
			v, ok := c.Float()
			if !ok {
				return setStr(col, r, c.S)
			}
			return setNum(col, r, v)
		}
		return setStr(col, r, strings.TrimSpace(c.S))
	}
	for i, row := range rows {
		r := i + 2
		if err := setStr("A", r, row.Code); err != nil { // mã luôn là văn bản, giữ số 0 đầu
			return nil, err
		}
		if row.Name != "" {
			if err := setStr("B", r, row.Name); err != nil {
				return nil, err
			}
		}
		if row.Unit != "" {
			if err := setStr("C", r, row.Unit); err != nil {
				return nil, err
			}
		}
		if err := setNum("D", r, row.Price); err != nil {
			return nil, err
		}
		if err := setCell("F", r, row.CK); err != nil {
			return nil, err
		}
		if err := setCell("G", r, row.Qty); err != nil {
			return nil, err
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UniquePath trả đường dẫn chưa tồn tại (thêm _2, _3... nếu trùng) để không bao giờ ghi đè file có sẵn.
func UniquePath(dir, name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	p := filepath.Join(dir, name)
	for i := 2; ; i++ {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return p
		}
		p = filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
	}
}

// WriteNew ghi file mới, từ chối ghi đè (O_EXCL).
func WriteNew(path string, data []byte) error {
	fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := fh.Write(data); err != nil {
		fh.Close()
		os.Remove(path)
		return err
	}
	return fh.Close()
}

// VerifyExport mở lại file đã xuất và đối chiếu với dữ liệu dự kiến: tiêu đề, số dòng, từng ô A/B/D/F/G.
func VerifyExport(data []byte, rows []OutRow) error {
	g, err := ReadGrid(data)
	if err != nil {
		return err
	}
	for i, h := range OutHeader {
		if g.At(1, i+1).Text() != h {
			return fmt.Errorf("tiêu đề cột %d sai: %q", i+1, g.At(1, i+1).Text())
		}
	}
	if g.MaxRow-1 != len(rows) {
		return fmt.Errorf("số dòng file xuất %d khác dự kiến %d", g.MaxRow-1, len(rows))
	}
	if g.MaxCol > 7 {
		return fmt.Errorf("file xuất có cột thừa (đến cột %d)", g.MaxCol)
	}
	for i, row := range rows {
		r := i + 2
		if g.At(r, 1).S != row.Code || g.At(r, 1).Num {
			return fmt.Errorf("dòng %d: mã hàng %q khác dự kiến %q", r, g.At(r, 1).S, row.Code)
		}
		if g.At(r, 2).S != row.Name {
			return fmt.Errorf("dòng %d: tên hàng khác dự kiến", r)
		}
		d, ok := g.At(r, 4).Float()
		if !ok || d != row.Price {
			return fmt.Errorf("dòng %d: đơn giá %q khác dự kiến %v", r, g.At(r, 4).S, row.Price)
		}
		if strings.TrimSpace(g.At(r, 6).S) != strings.TrimSpace(row.CK.S) {
			return fmt.Errorf("dòng %d: giảm giá %% khác dự kiến", r)
		}
		if strings.TrimSpace(g.At(r, 7).S) != strings.TrimSpace(row.Qty.S) {
			return fmt.Errorf("dòng %d: số lượng %q khác dự kiến %q", r, g.At(r, 7).S, row.Qty.S)
		}
	}
	return nil
}
