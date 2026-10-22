package core

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"unicode/utf16"
)

// Bộ đọc .xls (BIFF8, Excel 97-2003) tự viết: chỉ đọc giá trị ô của sheet đầu tiên.
// Giá trị công thức lấy theo kết quả Excel đã lưu sẵn trong file.

const (
	biffBOF      = 0x0809
	biffEOF      = 0x000A
	biffContinue = 0x003C
	biffSST      = 0x00FC
	biffBoundSh  = 0x0085
	biffFilePass = 0x002F
	biffLabelSST = 0x00FD
	biffLabel    = 0x0204
	biffRString  = 0x00D6
	biffNumber   = 0x0203
	biffRK       = 0x027E
	biffMulRK    = 0x00BD
	biffFormula  = 0x0006
	biffString   = 0x0207
	biffBoolErr  = 0x0205
)

type biffRec struct {
	id   uint16
	data []byte
	off  int
}

func biffRecords(b []byte, from int) []biffRec {
	var out []biffRec
	for off := from; off+4 <= len(b); {
		id := binary.LittleEndian.Uint16(b[off:])
		n := int(binary.LittleEndian.Uint16(b[off+2:]))
		if off+4+n > len(b) {
			break
		}
		out = append(out, biffRec{id, b[off+4 : off+4+n], off})
		off += 4 + n
	}
	return out
}

// strReader đọc chuỗi Unicode của BIFF8 qua nhiều đoạn (SST + CONTINUE).
type strReader struct {
	segs [][]byte
	seg  int
	pos  int
}

func (r *strReader) eof() bool {
	for r.seg < len(r.segs) && r.pos >= len(r.segs[r.seg]) {
		r.seg++
		r.pos = 0
	}
	return r.seg >= len(r.segs)
}

func (r *strReader) byte1() (byte, bool) {
	if r.eof() {
		return 0, false
	}
	b := r.segs[r.seg][r.pos]
	r.pos++
	return b, true
}

func (r *strReader) u16() (int, bool) {
	a, ok1 := r.byte1()
	b, ok2 := r.byte1()
	return int(a) | int(b)<<8, ok1 && ok2
}

func (r *strReader) skip(n int) bool {
	for n > 0 {
		if r.eof() {
			return false
		}
		k := len(r.segs[r.seg]) - r.pos
		if k > n {
			k = n
		}
		r.pos += k
		n -= k
	}
	return true
}

// unicodeString đọc một chuỗi (cch 2 byte, cờ, tùy chọn runs/ext, ký tự).
func (r *strReader) unicodeString() (string, bool) {
	cch, ok := r.u16()
	if !ok {
		return "", false
	}
	return r.stringBody(cch)
}

func (r *strReader) stringBody(cch int) (string, bool) {
	flags, ok := r.byte1()
	if !ok {
		return "", false
	}
	high := flags&1 != 0
	runs, ext := 0, 0
	if flags&8 != 0 {
		runs, ok = r.u16()
		if !ok {
			return "", false
		}
	}
	if flags&4 != 0 {
		a, ok1 := r.u16()
		b, ok2 := r.u16()
		if !ok1 || !ok2 {
			return "", false
		}
		ext = a | b<<16
	}
	u := make([]uint16, 0, cch)
	for len(u) < cch {
		if r.eof() {
			return "", false
		}
		if r.pos == 0 && len(u) > 0 {
			// đầu đoạn CONTINUE giữa chuỗi: có byte cờ mới
			f, _ := r.byte1()
			high = f&1 != 0
			if r.eof() {
				return "", false
			}
		}
		seg := r.segs[r.seg]
		if high {
			for r.pos+1 < len(seg) && len(u) < cch {
				u = append(u, binary.LittleEndian.Uint16(seg[r.pos:]))
				r.pos += 2
			}
			if r.pos < len(seg) && len(u) < cch { // lẻ 1 byte: dữ liệu hỏng
				return "", false
			}
		} else {
			for r.pos < len(seg) && len(u) < cch {
				u = append(u, uint16(seg[r.pos]))
				r.pos++
			}
		}
	}
	if !r.skip(runs*4 + ext) {
		return string(utf16.Decode(u)), true
	}
	return string(utf16.Decode(u)), true
}

func readSST(recs []biffRec, idx int) ([]string, error) {
	segs := [][]byte{recs[idx].data}
	for j := idx + 1; j < len(recs) && recs[j].id == biffContinue; j++ {
		segs = append(segs, recs[j].data)
	}
	if len(segs[0]) < 8 {
		return nil, fmt.Errorf("bảng chuỗi (SST) hỏng")
	}
	total := int(binary.LittleEndian.Uint32(segs[0][4:])) // số chuỗi duy nhất
	r := &strReader{segs: segs, pos: 8}
	out := make([]string, 0, total)
	for i := 0; i < total; i++ {
		s, ok := r.unicodeString()
		if !ok {
			break
		}
		out = append(out, s)
	}
	return out, nil
}

func rkValue(rk uint32) float64 {
	var v float64
	if rk&2 != 0 {
		v = float64(int32(rk) >> 2)
	} else {
		v = math.Float64frombits(uint64(rk&0xFFFFFFFC) << 32)
	}
	if rk&1 != 0 {
		v /= 100
	}
	return v
}

func fmtNum(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// ReadXLS đọc sheet đầu tiên của tệp .xls (BIFF8).
func ReadXLS(data []byte) (*Grid, error) {
	c, err := openCFB(data)
	if err != nil {
		return nil, err
	}
	wb, err := c.stream("Workbook", "Book")
	if err != nil {
		return nil, err
	}
	recs := biffRecords(wb, 0)
	if len(recs) == 0 || recs[0].id != biffBOF {
		return nil, fmt.Errorf("tệp .xls không đúng cấu trúc Excel")
	}
	if v := binary.LittleEndian.Uint16(recs[0].data); v != 0x0600 {
		return nil, fmt.Errorf("tệp .xls này là Excel đời rất cũ (5.0/95); hãy mở bằng Excel rồi Lưu thành .xlsx")
	}
	var sst []string
	sheetOff, sheetName := -1, ""
	depth := 0
	for i, r := range recs {
		switch r.id {
		case biffBOF:
			depth++
		case biffEOF:
			depth--
		case biffFilePass:
			return nil, fmt.Errorf("file Excel có mật khẩu/mã hóa, app không mở được; hãy bỏ mật khẩu rồi thử lại")
		case biffSST:
			if depth == 1 {
				sst, err = readSST(recs, i)
				if err != nil {
					return nil, err
				}
			}
		case biffBoundSh:
			if sheetOff < 0 && len(r.data) >= 8 && r.data[5] == 0 { // 0 = worksheet
				sheetOff = int(binary.LittleEndian.Uint32(r.data))
				sr := &strReader{segs: [][]byte{r.data}, pos: 6}
				n, _ := sr.byte1()
				if s, ok := sr.stringBody(int(n)); ok {
					sheetName = s
				}
			}
		}
		if depth == 0 && i > 0 {
			break
		}
	}
	if sheetOff < 0 || sheetOff >= len(wb) {
		return nil, fmt.Errorf("không tìm thấy sheet dữ liệu trong tệp .xls")
	}
	g := &Grid{Sheet: sheetName, Cells: map[[2]int]Cell{}}
	put := func(row, col int, c Cell) {
		if c.S == "" {
			return
		}
		r, cc := row+1, col+1
		g.Cells[[2]int{r, cc}] = c
		if r > g.MaxRow {
			g.MaxRow = r
		}
		if cc > g.MaxCol {
			g.MaxCol = cc
		}
	}
	srecs := biffRecords(wb, sheetOff)
	depth = 0
	var pendingFormula [2]int
	havePending := false
	for _, r := range srecs {
		d := r.data
		switch r.id {
		case biffBOF:
			depth++
			continue
		case biffEOF:
			depth--
			if depth <= 0 {
				goto done
			}
			continue
		}
		if depth != 1 {
			continue
		}
		switch r.id {
		case biffLabelSST:
			if len(d) >= 10 {
				if i := int(binary.LittleEndian.Uint32(d[6:])); i < len(sst) {
					put(int(binary.LittleEndian.Uint16(d)), int(binary.LittleEndian.Uint16(d[2:])), Cell{S: sst[i]})
				}
			}
		case biffLabel, biffRString:
			if len(d) >= 8 {
				sr := &strReader{segs: [][]byte{d}, pos: 6}
				if s, ok := sr.unicodeString(); ok {
					put(int(binary.LittleEndian.Uint16(d)), int(binary.LittleEndian.Uint16(d[2:])), Cell{S: s})
				}
			}
		case biffNumber:
			if len(d) >= 14 {
				put(int(binary.LittleEndian.Uint16(d)), int(binary.LittleEndian.Uint16(d[2:])),
					Cell{S: fmtNum(math.Float64frombits(binary.LittleEndian.Uint64(d[6:]))), Num: true})
			}
		case biffRK:
			if len(d) >= 10 {
				put(int(binary.LittleEndian.Uint16(d)), int(binary.LittleEndian.Uint16(d[2:])),
					Cell{S: fmtNum(rkValue(binary.LittleEndian.Uint32(d[6:]))), Num: true})
			}
		case biffMulRK:
			if len(d) >= 6 {
				row := int(binary.LittleEndian.Uint16(d))
				col := int(binary.LittleEndian.Uint16(d[2:]))
				for off := 4; off+6 <= len(d)-2; off += 6 {
					put(row, col, Cell{S: fmtNum(rkValue(binary.LittleEndian.Uint32(d[off+2:]))), Num: true})
					col++
				}
			}
		case biffBoolErr:
			if len(d) >= 8 && d[7] == 0 {
				v := "FALSE"
				if d[6] != 0 {
					v = "TRUE"
				}
				put(int(binary.LittleEndian.Uint16(d)), int(binary.LittleEndian.Uint16(d[2:])), Cell{S: v})
			}
		case biffFormula:
			havePending = false
			if len(d) >= 14 {
				row, col := int(binary.LittleEndian.Uint16(d)), int(binary.LittleEndian.Uint16(d[2:]))
				if d[12] == 0xFF && d[13] == 0xFF {
					if d[6] == 0 { // kết quả chuỗi nằm ở record STRING kế tiếp
						pendingFormula, havePending = [2]int{row, col}, true
					}
				} else {
					put(row, col, Cell{S: fmtNum(math.Float64frombits(binary.LittleEndian.Uint64(d[6:]))), Num: true})
				}
			}
		case biffString:
			if havePending {
				sr := &strReader{segs: [][]byte{d}}
				if s, ok := sr.unicodeString(); ok {
					put(pendingFormula[0], pendingFormula[1], Cell{S: s})
				}
				havePending = false
			}
		}
	}
done:
	if len(g.Cells) == 0 {
		return nil, fmt.Errorf("sheet đầu tiên của tệp .xls không có dữ liệu")
	}
	return g, nil
}
