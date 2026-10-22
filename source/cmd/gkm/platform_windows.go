//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

const createNoWindow = 0x08000000

func defaultRoot() string { return `E:\GKM_HoaDonNhapHang` }

func hide(c *exec.Cmd) *exec.Cmd {
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	return c
}

// fatal hiện hộp thoại báo lỗi (app chạy không có cửa sổ console).
func fatal(msg string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	box := user32.NewProc("MessageBoxW")
	t, _ := syscall.UTF16PtrFromString("GKM - Nhập Hóa Đơn")
	m, _ := syscall.UTF16PtrFromString(msg)
	box.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), 0x10)
	os.Exit(1)
}

type winPlatform struct{ root string }

func newPlatform(root string) *winPlatform { return &winPlatform{root} }

func (winPlatform) PickFolder(initial string) (string, error) {
	script := `[Console]::OutputEncoding=[Text.Encoding]::UTF8;` +
		`Add-Type -AssemblyName System.Windows.Forms;` +
		`$f=New-Object System.Windows.Forms.Form;$f.TopMost=$true;$f.ShowInTaskbar=$false;$f.WindowState='Minimized';` +
		`$d=New-Object System.Windows.Forms.FolderBrowserDialog;` +
		`$d.Description='Chọn thư mục lưu file nhập hàng';$d.ShowNewFolderButton=$true;` +
		`if($env:GKM_INITIAL -and (Test-Path -LiteralPath $env:GKM_INITIAL)){$d.SelectedPath=$env:GKM_INITIAL};` +
		`$f.Show();$f.Hide();if($d.ShowDialog($f) -eq 'OK'){[Console]::Out.Write($d.SelectedPath)}`
	c := hide(exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-STA", "-ExecutionPolicy", "Bypass", "-Command", script))
	c.Env = append(os.Environ(), "GKM_INITIAL="+initial)
	out, err := c.Output()
	if err != nil {
		return "", fmt.Errorf("không mở được hộp chọn thư mục của Windows; hãy gõ đường dẫn vào ô thư mục: %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// OpenFolder mở thư mục chứa file (hoặc chính thư mục nếu path là thư mục) trong Explorer.
// Dùng ShellExecute với đường dẫn đã chuẩn hóa (dấu \\, đường dẫn tuyệt đối) thay cho "explorer /select,": cách cũ dễ bị
// Explorer hiểu sai (mở nhầm Documents/This PC) khi đường dẫn có dấu cách, dấu phẩy hoặc dấu /.
func (winPlatform) OpenFolder(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("thiếu đường dẫn")
	}
	p, err := filepath.Abs(filepath.Clean(filepath.FromSlash(path)))
	if err != nil {
		return fmt.Errorf("đường dẫn không hợp lệ: %s", path)
	}
	dir := p
	if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
		dir = filepath.Dir(p)
	} else if err != nil {
		dir = filepath.Dir(p) // file không còn: mở thư mục chứa nó
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return fmt.Errorf("thư mục không tồn tại: %s", dir)
	}
	verb, _ := syscall.UTF16PtrFromString("open")
	target, _ := syscall.UTF16PtrFromString(dir)
	shell := syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW")
	r, _, _ := shell.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(target)), 0, 0, 1) // 1 = SW_SHOWNORMAL
	if r <= 32 {
		return fmt.Errorf("Windows không mở được thư mục %s (mã %d)", dir, r)
	}
	return nil
}

func firstExisting(paths ...string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// openWindow mở cửa sổ giao diện: ưu tiên Edge/Chrome ở chế độ ứng dụng (không có thanh địa chỉ),
// hồ sơ trình duyệt đặt trong thư mục dự án để không ghi sang ổ C; không có thì dùng trình duyệt mặc định.
func openWindow(root, url string, noBrowser bool) {
	if noBrowser {
		fmt.Println(url)
		return
	}
	pf, pf86, local := os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("LocalAppData")
	browser := firstExisting(
		filepath.Join(pf86, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pf, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pf, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(pf86, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(local, `Google\Chrome\Application\chrome.exe`),
	)
	if browser != "" {
		profile := filepath.Join(root, "config", "browser-profile")
		if err := os.MkdirAll(profile, 0o755); err != nil {
			if exe, e := os.Executable(); e == nil {
				profile = filepath.Join(filepath.Dir(exe), "browser-profile")
				_ = os.MkdirAll(profile, 0o755)
			}
		}
		c := exec.Command(browser, "--app="+url, "--user-data-dir="+profile, "--no-first-run", "--no-default-browser-check", "--window-size=1180,900")
		if err := c.Start(); err == nil {
			return
		}
	}
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
