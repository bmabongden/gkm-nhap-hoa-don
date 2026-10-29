package server

import (
	"os"
	"path/filepath"
	"strings"
)

// locateSource tìm trên đĩa file hóa đơn người dùng vừa kéo thả vào cửa sổ. Trình duyệt chỉ cho biết tên, kích thước và
// ngày sửa của file chứ không cho biết nó nằm ở đâu, nên app dò các thư mục hay chứa hóa đơn (Tải xuống, Desktop,
// Documents, OneDrive, thư mục dự án; sâu tối đa 3 cấp) tìm file trùng tên, trùng kích thước (và trùng ngày sửa nếu có).
// Trả thư mục chứa file, hoặc "" nếu không tìm thấy hoặc có nhiều nơi cùng khớp mà không phân biệt được.
func locateSource(name string, size, mtimeMs int64, roots []string) string {
	if name == "" || size <= 0 {
		return ""
	}
	const maxVisit = 20000
	visited := 0
	seen := map[string]bool{}
	var found []string
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if visited > maxVisit || seen[dir] {
			return
		}
		seen[dir] = true
		ents, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range ents {
			visited++
			if visited > maxVisit {
				return
			}
			p := filepath.Join(dir, e.Name())
			if e.IsDir() {
				if depth < 3 && !strings.HasPrefix(e.Name(), ".") {
					walk(p, depth+1)
				}
				continue
			}
			if !strings.EqualFold(e.Name(), name) {
				continue
			}
			fi, err := e.Info()
			if err != nil || fi.Size() != size {
				continue
			}
			if mtimeMs > 0 {
				d := fi.ModTime().UnixMilli() - mtimeMs
				if d < 0 {
					d = -d
				}
				if d > 2000 {
					continue
				}
			}
			found = append(found, dir)
		}
	}
	for _, r := range roots {
		if r != "" {
			walk(r, 0)
		}
	}
	uniq := map[string]bool{}
	for _, d := range found {
		uniq[d] = true
	}
	if len(uniq) == 1 {
		return found[0]
	}
	return ""
}

// sourceRoots là các thư mục hay chứa hóa đơn tải về.
func sourceRoots(projectRoot string) []string {
	var roots []string
	home, _ := os.UserHomeDir()
	for _, b := range []string{home, os.Getenv("OneDrive"), os.Getenv("OneDriveConsumer"), os.Getenv("OneDriveCommercial")} {
		if b == "" {
			continue
		}
		for _, sub := range []string{"Downloads", "Desktop", "Documents", "Tài liệu", "Zalo Received Files"} {
			roots = append(roots, filepath.Join(b, sub))
		}
	}
	return append(roots, projectRoot)
}
