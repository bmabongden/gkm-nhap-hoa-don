package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRootFromExe(t *testing.T) {
	base := t.TempDir()
	proj := filepath.Join(base, "Data AI", "GKM_HoaDonNhapHang")
	_ = os.MkdirAll(filepath.Join(proj, "app"), 0o755)
	_ = os.MkdirAll(filepath.Join(proj, "templates"), 0o755)
	if got := rootFromExe(filepath.Join(proj, "app", "GKM_NhapHoaDon.exe")); got != proj {
		t.Fatalf("muốn %q, được %q", proj, got)
	}
	other := filepath.Join(base, "khac", "app")
	_ = os.MkdirAll(other, 0o755)
	if got := rootFromExe(filepath.Join(other, "x.exe")); got != "" {
		t.Fatalf("thư mục app không thuộc dự án phải trả rỗng, được %q", got)
	}
	if got := rootFromExe(filepath.Join(proj, "GKM.exe")); got != "" {
		t.Fatalf("exe không nằm trong app\\ phải trả rỗng, được %q", got)
	}
}
