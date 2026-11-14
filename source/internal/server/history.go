package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gkm/internal/core"
)

// ExportRecord là một lần xuất file, lưu trong config\da_xuat.json để cảnh báo khi xuất trùng hóa đơn.
type ExportRecord struct {
	FP       string `json:"fp"`
	At       string `json:"at"` // giờ xuất, dạng 14/11/2026 19:30
	Source   string `json:"source"`
	Path     string `json:"path"`
	So       string `json:"so,omitempty"`
	Ngay     string `json:"ngay,omitempty"`
	Rows     int    `json:"rows"`
	Supplier string `json:"supplier,omitempty"`
}

const maxHistory = 5000

func (a *App) historyPath() string { return filepath.Join(a.Root, "config", "da_xuat.json") }

// fingerprint là dấu vân tay nội dung hóa đơn: số phiếu, ngày, nhà cung cấp và từng dòng sẽ nhập (mã, giá, %, số lượng).
// Cùng hóa đơn tải lại hay đổi tên file vẫn ra cùng dấu; hóa đơn khác dù cùng ngày thì khác dấu.
func fingerprint(p *core.Parsed) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%s\n", p.Profile, p.SupplierTag, p.Meta.So, p.Meta.Ngay)
	for _, r := range p.OutputRows() {
		q, _ := r.Qty.Float()
		ck, _ := r.CK.Float()
		fmt.Fprintf(h, "%s|%s|%s|%s\n", r.Code, strconv.FormatFloat(r.Price, 'f', -1, 64), strconv.FormatFloat(ck, 'f', -1, 64), strconv.FormatFloat(q, 'f', -1, 64))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (a *App) loadHistory() []ExportRecord {
	var hs []ExportRecord
	if b, err := os.ReadFile(a.historyPath()); err == nil {
		_ = json.Unmarshal(b, &hs)
	}
	return hs
}

// findExported trả lần xuất gần nhất có cùng dấu vân tay (nil nếu chưa xuất).
func (a *App) findExported(fp string) *ExportRecord {
	a.histMu.Lock()
	defer a.histMu.Unlock()
	hs := a.loadHistory()
	for i := len(hs) - 1; i >= 0; i-- {
		if hs[i].FP == fp {
			r := hs[i]
			return &r
		}
	}
	return nil
}

func (a *App) recordExport(p *core.Parsed, source, path string, rows int) {
	a.histMu.Lock()
	defer a.histMu.Unlock()
	hs := append(a.loadHistory(), ExportRecord{FP: fingerprint(p), At: time.Now().Format("02/01/2006 15:04"), Source: source,
		Path: path, So: p.Meta.So, Ngay: p.Meta.Ngay, Rows: rows, Supplier: p.Supplier})
	if len(hs) > maxHistory {
		hs = hs[len(hs)-maxHistory:]
	}
	if err := os.MkdirAll(filepath.Dir(a.historyPath()), 0o755); err != nil {
		a.Log.Printf("không lưu được lịch sử xuất: %v", err)
		return
	}
	b, _ := json.MarshalIndent(hs, "", " ")
	tmp := a.historyPath() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err == nil {
		err = os.Rename(tmp, a.historyPath())
		if err != nil {
			a.Log.Printf("không lưu được lịch sử xuất: %v", err)
		}
	} else {
		a.Log.Printf("không lưu được lịch sử xuất: %v", err)
	}
}
