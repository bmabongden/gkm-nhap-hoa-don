package core

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// Bộ đọc tối thiểu cho tệp OLE2/CFB (vỏ của .xls cũ). Chỉ đọc, không ghi.

type cfb struct {
	data       []byte
	secSize    int
	miniSize   int
	fat        []uint32
	miniFat    []uint32
	miniStream []byte
	miniCutoff uint32
	dirStart   uint32
}

const (
	cfbEnd  = 0xFFFFFFFE
	cfbFree = 0xFFFFFFFF
)

func (c *cfb) sector(n uint32) ([]byte, error) {
	off := (int(n) + 1) * c.secSize
	if off < 0 || off+c.secSize > len(c.data) {
		return nil, fmt.Errorf("sector %d nằm ngoài file", n)
	}
	return c.data[off : off+c.secSize], nil
}

func (c *cfb) chain(start uint32, fat []uint32) ([]uint32, error) {
	var out []uint32
	seen := map[uint32]bool{}
	for n := start; n != cfbEnd && n != cfbFree; {
		if seen[n] || int(n) >= len(fat) {
			return nil, fmt.Errorf("chuỗi sector bị hỏng")
		}
		seen[n] = true
		out = append(out, n)
		n = fat[n]
	}
	return out, nil
}

func (c *cfb) readChain(start uint32) ([]byte, error) {
	ch, err := c.chain(start, c.fat)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(ch)*c.secSize)
	for _, n := range ch {
		s, err := c.sector(n)
		if err != nil {
			return nil, err
		}
		out = append(out, s...)
	}
	return out, nil
}

func openCFB(data []byte) (*cfb, error) {
	if len(data) < 512 || string(data[:8]) != "\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1" {
		return nil, fmt.Errorf("không phải tệp Excel .xls (OLE2)")
	}
	le := binary.LittleEndian
	c := &cfb{data: data}
	c.secSize = 1 << le.Uint16(data[0x1E:])
	c.miniSize = 1 << le.Uint16(data[0x20:])
	if c.secSize != 512 && c.secSize != 4096 {
		return nil, fmt.Errorf("kích thước sector %d không hỗ trợ", c.secSize)
	}
	numFAT := le.Uint32(data[0x2C:])
	c.dirStart = le.Uint32(data[0x30:])
	c.miniCutoff = le.Uint32(data[0x38:])
	miniFatStart := le.Uint32(data[0x3C:])
	difatStart := le.Uint32(data[0x44:])
	numDifat := le.Uint32(data[0x48:])
	var fatSecs []uint32
	for i := 0; i < 109 && uint32(len(fatSecs)) < numFAT; i++ {
		fatSecs = append(fatSecs, le.Uint32(data[0x4C+4*i:]))
	}
	d := difatStart
	for i := uint32(0); i < numDifat && uint32(len(fatSecs)) < numFAT; i++ {
		s, err := c.sector(d)
		if err != nil {
			return nil, err
		}
		for j := 0; j < c.secSize/4-1 && uint32(len(fatSecs)) < numFAT; j++ {
			fatSecs = append(fatSecs, le.Uint32(s[4*j:]))
		}
		d = le.Uint32(s[c.secSize-4:])
	}
	for _, n := range fatSecs {
		s, err := c.sector(n)
		if err != nil {
			return nil, err
		}
		for j := 0; j < c.secSize/4; j++ {
			c.fat = append(c.fat, le.Uint32(s[4*j:]))
		}
	}
	if miniFatStart != cfbEnd && miniFatStart != cfbFree {
		b, err := c.readChain(miniFatStart)
		if err != nil {
			return nil, err
		}
		for j := 0; j+4 <= len(b); j += 4 {
			c.miniFat = append(c.miniFat, le.Uint32(b[j:]))
		}
	}
	return c, nil
}

// stream trả nội dung stream có tên cho trước (so khớp không phân biệt hoa thường).
func (c *cfb) stream(names ...string) ([]byte, error) {
	dir, err := c.readChain(c.dirStart)
	if err != nil {
		return nil, err
	}
	le := binary.LittleEndian
	type ent struct {
		name  string
		typ   byte
		start uint32
		size  uint32
	}
	var ents []ent
	for off := 0; off+128 <= len(dir); off += 128 {
		e := dir[off : off+128]
		nl := int(le.Uint16(e[0x40:]))
		if nl < 2 || nl > 64 {
			ents = append(ents, ent{})
			continue
		}
		u := make([]uint16, 0, nl/2)
		for j := 0; j+1 < nl-1; j += 2 {
			u = append(u, le.Uint16(e[j:]))
		}
		ents = append(ents, ent{string(utf16.Decode(u)), e[0x42], le.Uint32(e[0x74:]), le.Uint32(e[0x78:])})
	}
	if len(ents) == 0 || ents[0].typ != 5 {
		return nil, fmt.Errorf("thiếu Root Entry")
	}
	root := ents[0]
	if root.start != cfbEnd && root.start != cfbFree {
		ms, err := c.readChain(root.start)
		if err != nil {
			return nil, err
		}
		c.miniStream = ms
	}
	for _, e := range ents {
		if e.typ != 2 {
			continue
		}
		for _, n := range names {
			if equalFold(e.name, n) {
				return c.readStream(e.start, e.size)
			}
		}
	}
	return nil, fmt.Errorf("không có luồng Workbook trong tệp .xls")
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		x, y := a[i], b[i]
		if x >= 'A' && x <= 'Z' {
			x += 32
		}
		if y >= 'A' && y <= 'Z' {
			y += 32
		}
		if x != y {
			return false
		}
	}
	return true
}

func (c *cfb) readStream(start, size uint32) ([]byte, error) {
	if size < c.miniCutoff {
		ch, err := c.chain(start, c.miniFat)
		if err != nil {
			return nil, err
		}
		out := make([]byte, 0, len(ch)*c.miniSize)
		for _, n := range ch {
			off := int(n) * c.miniSize
			if off+c.miniSize > len(c.miniStream) {
				return nil, fmt.Errorf("mini sector %d nằm ngoài luồng", n)
			}
			out = append(out, c.miniStream[off:off+c.miniSize]...)
		}
		if int(size) <= len(out) {
			out = out[:size]
		}
		return out, nil
	}
	b, err := c.readChain(start)
	if err != nil {
		return nil, err
	}
	if int(size) <= len(b) {
		b = b[:size]
	}
	return b, nil
}
