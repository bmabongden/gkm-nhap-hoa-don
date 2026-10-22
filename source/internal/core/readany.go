package core

import (
	"bytes"
	"fmt"
	"strings"
)

// Kind là loại tệp đã nhận dạng theo nội dung (không tin vào đuôi tệp).
const (
	KindXLSX = "xlsx"
	KindXLS  = "xls"
	KindXML  = "xml2003"
	KindHTML = "html"
)

// SniffFormat nhận dạng loại tệp theo nội dung; "" nếu không phải bảng tính hỗ trợ.
func SniffFormat(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte("PK\x03\x04")):
		return KindXLSX
	case bytes.HasPrefix(data, []byte("\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1")):
		return KindXLS
	}
	n := len(data)
	if n > 4096 {
		n = 4096
	}
	head := strings.ToLower(string(decodeText(data[:n])))
	switch {
	case isSpreadsheetML(head):
		return KindXML
	case strings.Contains(head, "<table") || strings.Contains(head, "<html") || strings.Contains(head, "<!doctype html"):
		return KindHTML
	}
	return ""
}

// ReadGridAny đọc hóa đơn .xlsx/.xlsm, .xls (Excel 97-2003) hoặc tệp đuôi .xls thật ra là XML/HTML.
func ReadGridAny(data []byte) (*Grid, string, error) {
	k := SniffFormat(data)
	var g *Grid
	var err error
	switch k {
	case KindXLSX:
		g, err = ReadGrid(data)
	case KindXLS:
		g, err = ReadXLS(data)
	case KindXML:
		g, err = ReadSpreadsheetML(data)
	case KindHTML:
		g, err = ReadHTMLTable(data)
	default:
		return nil, "", fmt.Errorf("tệp này không phải file Excel (.xlsx hoặc .xls) mà app đọc được")
	}
	return g, k, err
}
