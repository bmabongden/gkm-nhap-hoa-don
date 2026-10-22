package core

import (
	"fmt"
	"math"
	"strings"
)

// CellDiff là một ô khác nhau giữa hai file.
type CellDiff struct {
	Cell string `json:"cell"`
	Got  string `json:"got"`
	Want string `json:"want"`
}

// CompareResult là kết quả so từng ô A1:G của hai file nhập hàng.
type CompareResult struct {
	RowsGot, RowsWant int
	ValueDiffs        []CellDiff // khác giá trị
	TypeDiffs         int        // cùng giá trị nhưng khác kiểu (số/văn bản)
	FloatNoise        int        // chỉ lệch sai số dấu phẩy động < 1e-6
}

func colName(c int) string { return string(rune('A' + c - 1)) }

// Compare so từng ô A1:G(max) giữa file xuất và file kết quả mong đợi. Ô trống và chuỗi rỗng coi như nhau.
func Compare(got, want []byte) (*CompareResult, error) {
	a, err := ReadGrid(got)
	if err != nil {
		return nil, err
	}
	b, err := ReadGrid(want)
	if err != nil {
		return nil, err
	}
	res := &CompareResult{RowsGot: a.MaxRow - 1, RowsWant: b.MaxRow - 1}
	n := a.MaxRow
	if b.MaxRow > n {
		n = b.MaxRow
	}
	for r := 1; r <= n; r++ {
		for c := 1; c <= 7; c++ {
			x, y := a.At(r, c), b.At(r, c)
			xe, ye := strings.TrimSpace(x.S) == "", strings.TrimSpace(y.S) == ""
			if xe && ye {
				continue
			}
			if x.Num && y.Num {
				fx, _ := x.Float()
				fy, _ := y.Float()
				if fx == fy {
					continue
				}
				if math.Abs(fx-fy) < 1e-6 {
					res.FloatNoise++
					continue
				}
			} else if x.S == y.S && !xe {
				if x.Num != y.Num {
					res.TypeDiffs++
				}
				continue
			}
			res.ValueDiffs = append(res.ValueDiffs, CellDiff{fmt.Sprintf("%s%d", colName(c), r), x.S, y.S})
		}
	}
	return res, nil
}
