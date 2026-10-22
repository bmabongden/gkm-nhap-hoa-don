//go:build !windows

package main

import (
	"fmt"
	"os"
)

// Bản không phải Windows chỉ dùng để phát triển/kiểm thử; app thật là bản Windows.
func defaultRoot() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h + "/GKM_HoaDonNhapHang"
	}
	return "./GKM_HoaDonNhapHang"
}

func fatal(msg string) { fmt.Fprintln(os.Stderr, msg); os.Exit(1) }

type devPlatform struct{}

func newPlatform(string) devPlatform { return devPlatform{} }

func (devPlatform) PickFolder(string) (string, error) {
	return "", fmt.Errorf("chức năng chọn thư mục chỉ có trên Windows; hãy gõ đường dẫn")
}
func (devPlatform) OpenFolder(string) error { return fmt.Errorf("chỉ có trên Windows") }

func openWindow(root, url string, noBrowser bool) { fmt.Println(url) }
