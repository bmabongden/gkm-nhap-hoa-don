package core

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var rePlainNum = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

// Nhiều phần mềm kế toán xuất ra tệp đuôi .xls nhưng thật ra là XML hoặc HTML.
// Hai bộ đọc dưới đây đọc những tệp đó (chỉ đọc, không ghi).

func decodeText(data []byte) []byte {
	switch {
	case bytes.HasPrefix(data, []byte{0xEF, 0xBB, 0xBF}):
		return data[3:]
	case bytes.HasPrefix(data, []byte{0xFF, 0xFE}), bytes.HasPrefix(data, []byte{0xFE, 0xFF}):
		le := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
		if out, _, err := transform.Bytes(le.NewDecoder(), data); err == nil {
			return out
		}
	}
	return data
}

func isSpreadsheetML(head string) bool {
	return strings.Contains(head, "urn:schemas-microsoft-com:office:spreadsheet")
}

// ReadSpreadsheetML đọc Excel 2003 XML (sheet đầu tiên).
func ReadSpreadsheetML(data []byte) (*Grid, error) {
	dec := xml.NewDecoder(bytes.NewReader(decodeText(data)))
	dec.Strict = false
	dec.CharsetReader = charset.NewReaderLabel
	g := &Grid{Cells: map[[2]int]Cell{}}
	attr := func(se xml.StartElement, name string) string {
		for _, a := range se.Attr {
			if a.Name.Local == name {
				return a.Value
			}
		}
		return ""
	}
	inSheet, inRow, inData := false, false, false
	row, col := 0, 0
	var typ string
	var txt strings.Builder
	var cellCol int
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("không đọc được tệp Excel XML: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "Worksheet":
				if inSheet {
					goto done
				}
				inSheet = true
				g.Sheet = attr(t, "Name")
			case "Row":
				if inSheet {
					inRow = true
					if v, e := strconv.Atoi(attr(t, "Index")); e == nil && v > 0 {
						row = v
					} else {
						row++
					}
					col = 0
				}
			case "Cell":
				if inRow {
					if v, e := strconv.Atoi(attr(t, "Index")); e == nil && v > 0 {
						col = v
					} else {
						col++
					}
					cellCol = col
					if m, e := strconv.Atoi(attr(t, "MergeAcross")); e == nil && m > 0 {
						col += m
					}
				}
			case "Data":
				if inRow {
					inData = true
					typ = attr(t, "Type")
					txt.Reset()
				}
			}
		case xml.CharData:
			if inData {
				txt.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "Data":
				if inData {
					inData = false
					s := txt.String()
					if strings.TrimSpace(s) != "" {
						g.Cells[[2]int{row, cellCol}] = Cell{S: s, Num: typ == "Number" || typ == "DateTime"}
						if row > g.MaxRow {
							g.MaxRow = row
						}
						if cellCol > g.MaxCol {
							g.MaxCol = cellCol
						}
					}
				}
			case "Row":
				inRow = false
			}
		}
	}
done:
	if len(g.Cells) == 0 {
		return nil, fmt.Errorf("không có dữ liệu trong tệp Excel XML")
	}
	return g, nil
}

// ReadHTMLTable đọc bảng lớn nhất trong tệp HTML đuôi .xls.
func ReadHTMLTable(data []byte) (*Grid, error) {
	raw := decodeText(data)
	var r io.Reader
	if utf8.Valid(raw) {
		r = bytes.NewReader(raw)
	} else {
		cr, err := charset.NewReader(bytes.NewReader(raw), "text/html")
		if err != nil {
			return nil, err
		}
		r = cr
	}
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("không đọc được tệp HTML: %w", err)
	}
	var tables []*html.Node
	var find func(*html.Node)
	find = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tables = append(tables, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(doc)
	if len(tables) == 0 {
		return nil, fmt.Errorf("không thấy bảng dữ liệu trong tệp")
	}
	// Bảng chọn: nhiều ô nhất tính riêng (không tính bảng lồng bên trong).
	var best *Grid
	for _, tb := range tables {
		g := tableGrid(tb)
		if best == nil || len(g.Cells) > len(best.Cells) {
			best = g
		}
	}
	if best == nil || len(best.Cells) == 0 {
		return nil, fmt.Errorf("bảng trong tệp không có dữ liệu")
	}
	return best, nil
}

func nodeText(n *html.Node, b *strings.Builder) {
	if n.Type == html.TextNode {
		b.WriteString(n.Data)
	}
	if n.Type == html.ElementNode && (n.Data == "br" || n.Data == "p" || n.Data == "div") {
		b.WriteByte(' ')
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "table" {
			continue
		}
		nodeText(c, b)
	}
}

func tableGrid(tb *html.Node) *Grid {
	g := &Grid{Cells: map[[2]int]Cell{}}
	occupied := map[[2]int]bool{}
	row := 0
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != html.ElementNode {
				continue
			}
			switch c.Data {
			case "table":
				continue
			case "tr":
				row++
				col := 0
				for td := c.FirstChild; td != nil; td = td.NextSibling {
					if td.Type != html.ElementNode || (td.Data != "td" && td.Data != "th") {
						continue
					}
					col++
					for occupied[[2]int{row, col}] {
						col++
					}
					cs, rs := 1, 1
					for _, a := range td.Attr {
						switch a.Key {
						case "colspan":
							if v, e := strconv.Atoi(a.Val); e == nil && v > 0 && v < 100 {
								cs = v
							}
						case "rowspan":
							if v, e := strconv.Atoi(a.Val); e == nil && v > 0 && v < 1000 {
								rs = v
							}
						}
					}
					var sb strings.Builder
					nodeText(td, &sb)
					s := strings.TrimSpace(strings.Join(strings.Fields(strings.ReplaceAll(sb.String(), " ", " ")), " "))
					if s != "" {
						g.Cells[[2]int{row, col}] = Cell{S: s, Num: rePlainNum.MatchString(s)}
						if row > g.MaxRow {
							g.MaxRow = row
						}
						if col > g.MaxCol {
							g.MaxCol = col
						}
					}
					for dr := 0; dr < rs; dr++ {
						for dc := 0; dc < cs; dc++ {
							if dr > 0 || dc > 0 {
								occupied[[2]int{row + dr, col + dc}] = true
							}
						}
					}
					col += cs - 1
				}
			default:
				walk(c)
				continue
			}
		}
	}
	walk(tb)
	return g
}
