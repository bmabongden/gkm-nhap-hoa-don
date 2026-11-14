package core

import (
	"fmt"
	"strings"
)

// Nhà cung cấp nhận ra theo tên in trên toa (so sánh sau khi bỏ dấu và khoảng trắng).
var knownSuppliers = []struct{ key, tag, name string }{
	{"hongnhatphat", "HongNhatPhat", "Hồng Nhất Phát"},
	{"tanhoaloi", "TanHoaLoi", "Tân Hòa Lợi"},
}

// detectSupplier tìm tên nhà cung cấp trong các ô chữ của toa.
func detectSupplier(g *Grid) (tag, name string) {
	for _, k := range g.Positions() {
		c := g.Cells[k]
		if c.Num {
			continue
		}
		t := strings.Join(strings.Fields(stripAccents(strings.ToLower(c.S))), "")
		for _, s := range knownSuppliers {
			if strings.Contains(t, s.key) {
				return s.tag, s.name
			}
		}
	}
	return "", ""
}

// sliceGrid trả bản sao các dòng r1..r2 của sheet, giữ nguyên số dòng (để thông báo vẫn đúng dòng trong file gốc).
func sliceGrid(g *Grid, r1, r2 int) *Grid {
	out := &Grid{Sheet: g.Sheet, Cells: map[[2]int]Cell{}}
	for k, v := range g.Cells {
		if k[0] >= r1 && k[0] <= r2 {
			out.Cells[k] = v
			if k[0] > out.MaxRow {
				out.MaxRow = k[0]
			}
			if k[1] > out.MaxCol {
				out.MaxCol = k[1]
			}
		}
	}
	return out
}

func rowBlank(g *Grid, r int) bool {
	for c := 1; c <= g.MaxCol; c++ {
		if !g.At(r, c).Empty() {
			return false
		}
	}
	return true
}

type segment struct {
	g     *Grid
	label string
}

// splitToa chia file thành từng toa: mỗi sheet có bảng hàng là một toa, và trong một sheet mỗi bảng có dòng tổng riêng là một toa.
func splitToa(g *Grid) []segment {
	sheets := append([]*Grid{g}, g.Extra...)
	var segs []segment
	for _, sh := range sheets {
		if sh.MaxRow > maxGridRows || sh.MaxCol > maxGridCols {
			continue
		}
		hs := invoiceHeaders(sh)
		if len(hs) == 0 {
			continue
		}
		if len(hs) == 1 {
			segs = append(segs, segment{sh, "sheet " + sh.Sheet})
			continue
		}
		// điểm cắt: phần đầu toa (tên NCC, ngày...) là các dòng liền ngay trên dòng tiêu đề
		cuts := []int{1}
		for i := 1; i < len(hs); i++ {
			r := hs[i] - 1
			for r > hs[i-1] && !rowBlank(sh, r) {
				r--
			}
			cuts = append(cuts, r+1)
		}
		for i := range hs {
			end := sh.MaxRow
			if i+1 < len(cuts) {
				end = cuts[i+1] - 1
			}
			segs = append(segs, segment{sliceGrid(sh, cuts[i], end), fmt.Sprintf("sheet %s, bảng %d (từ dòng %d)", sh.Sheet, i+1, hs[i])})
		}
	}
	return segs
}

// ParseAll phân tích file hóa đơn và trả mỗi toa hàng một kết quả (mỗi toa xuất ra một file riêng).
// Phiếu xuất kho (mẫu F) giữ quy tắc riêng: NCC gửi kèm các toa cũ nên chỉ lấy bảng cuối cùng.
func ParseAll(g *Grid, opt Options) ([]*Parsed, error) {
	if hr, canon, raw := findHeader(g); hr != 0 && detectProfile(canon, raw) == ProfileF {
		p, err := ParseWith(g, opt)
		if err != nil {
			return nil, err
		}
		return []*Parsed{p}, nil
	}
	segs := splitToa(g)
	if len(segs) <= 1 {
		one := g
		if len(segs) == 1 {
			one = segs[0].g
		}
		p, err := ParseWith(one, opt)
		if err != nil {
			return nil, err
		}
		return []*Parsed{p}, nil
	}
	var out []*Parsed
	var fails []string
	var firstErr error
	for _, s := range segs {
		p, err := ParseWith(s.g, opt)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", s.label, err)
			}
			fails = append(fails, fmt.Sprintf("%s: %v", s.label, err))
			continue
		}
		p.Part = s.label
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, firstErr
	}
	for i, p := range out {
		p.PartNo, p.PartOf = i+1, len(out)
		if len(out) > 1 {
			p.Notes = append([]Check{{"Nhiều toa trong file", LevelOK, fmt.Sprintf("file có %d toa hàng; đây là toa %d (%s), xuất ra file riêng", len(out), i+1, p.Part)}}, p.Notes...)
			p.Checks = append([]Check{p.Notes[0]}, p.Checks...)
		}
		if len(fails) > 0 {
			p.Warnings = append(p.Warnings, "Có toa trong file app không đọc được nên không xuất: "+strings.Join(fails, "; ")+". Hãy kiểm tra hóa đơn gốc.")
		}
	}
	// hai toa trùng tên file thì thêm số thứ tự toa
	seen := map[string]int{}
	for _, p := range out {
		seen[p.SuggestedName()]++
	}
	for _, p := range out {
		if seen[p.SuggestedName()] > 1 {
			p.NameSuffix = fmt.Sprintf("_toa%d", p.PartNo)
		}
	}
	return out, nil
}

// ProfileName là tên mẫu hiển thị; toa cùng kiểu Tân Hòa Lợi của NCC khác thì hiện tên NCC đó.
func (p *Parsed) ProfileName() string {
	if p.Profile == ProfileD && p.Supplier != "" && p.SupplierTag != "TanHoaLoi" {
		return p.Supplier + " (Toa hàng cùng kiểu Tân Hòa Lợi: Mã hàng, Hiệu, SL, Đơn giá tính nghìn đồng)"
	}
	return ProfileNames[p.Profile]
}
