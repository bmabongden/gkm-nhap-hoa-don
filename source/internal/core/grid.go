// Package core đọc hóa đơn Excel, áp dụng quy tắc ánh xạ GKM và xuất file nhập hàng theo MauFileNhapHang.xlsx.
// Quy tắc lấy từ docs/GKM_Quy_Tac_Anh_Xa.md (3 cặp mẫu + các xác nhận của Khiêm ngày 02/10/2026).
package core

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Cell là một ô đọc từ Excel: giữ chuỗi thô và biết ô là số hay văn bản
// (cần để giữ "1" dạng văn bản khác 1 dạng số, như kết quả mẫu).
type Cell struct {
	S   string
	Num bool
}

// Empty cho biết ô trống (hoặc chỉ có khoảng trắng).
func (c Cell) Empty() bool { return strings.TrimSpace(c.S) == "" }

// Text trả chuỗi đã cắt khoảng trắng.
func (c Cell) Text() string { return strings.TrimSpace(c.S) }

// Float đọc ô thành số (số thật hoặc chuỗi kiểu "492,000", "2,00").
func (c Cell) Float() (float64, bool) {
	if c.Empty() {
		return 0, false
	}
	if c.Num {
		f, err := strconv.ParseFloat(strings.TrimSpace(c.S), 64)
		return f, err == nil && !math.IsNaN(f) && !math.IsInf(f, 0)
	}
	return numFromText(c.S)
}

// Grid là toàn bộ sheet đầu tiên của một workbook.
type Grid struct {
	Sheet  string
	Cells  map[[2]int]Cell // [dòng, cột], đánh số từ 1
	MaxRow int
	MaxCol int
	// Extra là các sheet có dữ liệu phía sau sheet đầu (chỉ .xlsx). Mẫu NCC gửi cả chồng toa hàng dùng đến.
	Extra []*Grid
}

// At trả ô tại dòng r, cột c (rỗng nếu không có).
func (g *Grid) At(r, c int) Cell {
	if c <= 0 {
		return Cell{}
	}
	return g.Cells[[2]int{r, c}]
}

// Positions trả danh sách ô có dữ liệu theo thứ tự dòng rồi cột.
func (g *Grid) Positions() [][2]int {
	keys := make([][2]int, 0, len(g.Cells))
	for k, v := range g.Cells {
		if !v.Empty() {
			keys = append(keys, k)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})
	return keys
}

// ReadGrid đọc sheet đầu tiên của workbook .xlsx (dữ liệu nằm trong bộ nhớ, không ghi gì ra đĩa).
func ReadGrid(data []byte) (*Grid, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("không mở được file Excel (có thể file hỏng, có mật khẩu hoặc không phải .xlsx): %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("file Excel không có sheet nào")
	}
	g, err := readSheet(f, sheets[0])
	if err != nil {
		return nil, err
	}
	for _, name := range sheets[1:] {
		if x, err := readSheet(f, name); err == nil && len(x.Cells) > 0 {
			g.Extra = append(g.Extra, x)
		}
	}
	return g, nil
}

func readSheet(f *excelize.File, sheet string) (*Grid, error) {
	rows, err := f.Rows(sheet)
	if err != nil {
		return nil, fmt.Errorf("không đọc được sheet %q: %w", sheet, err)
	}
	defer rows.Close()
	g := &Grid{Sheet: sheet, Cells: map[[2]int]Cell{}}
	r := 0
	for rows.Next() {
		r++
		cols, err := rows.Columns(excelize.Options{RawCellValue: true})
		if err != nil {
			return nil, fmt.Errorf("lỗi đọc dòng %d: %w", r, err)
		}
		for ci, v := range cols {
			if v == "" {
				continue
			}
			axis, _ := excelize.CoordinatesToCellName(ci+1, r)
			t, _ := f.GetCellType(sheet, axis)
			// Ô số trong Excel thường không có thuộc tính t nên excelize trả CellTypeUnset.
			isNum := t == excelize.CellTypeNumber || t == excelize.CellTypeDate || t == excelize.CellTypeUnset
			g.Cells[[2]int{r, ci + 1}] = Cell{S: v, Num: isNum}
			if r > g.MaxRow {
				g.MaxRow = r
			}
			if ci+1 > g.MaxCol {
				g.MaxCol = ci + 1
			}
		}
	}
	return g, nil
}
