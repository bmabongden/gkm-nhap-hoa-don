// GKM - Nhập Hóa Đơn: app Windows độc lập (portable .exe) chuyển hóa đơn nhập hàng Excel sang file nhập hàng
// theo MauFileNhapHang.xlsx. Không cần Python, API, khóa hay Internet.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gkm/internal/core"
	"gkm/internal/server"
	"gkm/internal/update"
)

const version = "1.6.0"

//go:embed template.xlsx
var embeddedTemplate []byte

func main() {
	var (
		cli       = flag.Bool("cli", false, "chế độ dòng lệnh: chuyển các file hóa đơn đưa vào (không mở cửa sổ)")
		selftest  = flag.Bool("selftest", false, "chạy thử các mẫu trong thư mục samples và so từng ô với kết quả mẫu")
		samples   = flag.String("samples", "", "thư mục chứa 7 workbook mẫu (mặc định <root>\\samples)")
		out       = flag.String("out", "", "thư mục xuất (chế độ --cli); mặc định <root>\\output")
		root      = flag.String("root", "", "thư mục dự án (mặc định: thư mục cha của app\\ chứa file .exe, nếu không có thì E:\\GKM_HoaDonNhapHang; hoặc biến GKM_ROOT)")
		port      = flag.Int("port", 47631, "cổng cục bộ ưu tiên (chỉ chạy một bản)")
		noBrowser = flag.Bool("no-browser", false, "không tự mở cửa sổ; chỉ in địa chỉ")
		showVer   = flag.Bool("version", false, "in phiên bản")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "GKM - Nhập Hóa Đơn %s\n\nCách dùng:\n  GKM_NhapHoaDon.exe                 mở giao diện\n  GKM_NhapHoaDon.exe --selftest      kiểm tra các mẫu\n  GKM_NhapHoaDon.exe --cli hoadon.xlsx (hoặc .xls) [--out thư_mục]\n\n", version)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVer {
		fmt.Println("GKM - Nhập Hóa Đơn", version)
		return
	}
	rootDir := resolveRoot(*root)
	lg, closeLog := openLog(rootDir)
	defer closeLog()
	lg.Printf("khởi động GKM %s, thư mục dự án %s", version, rootDir)

	switch {
	case *selftest:
		os.Exit(runSelfTest(lg, rootDir, *samples))
	case *cli:
		os.Exit(runCLI(lg, rootDir, *out, flag.Args()))
	}

	if exe, err := os.Executable(); err == nil {
		update.CleanupOld(exe) // dọn file .old/.new của lần cập nhật trước
	}
	app := server.New(version, embeddedTemplate, rootDir, newPlatform(rootDir), lg)
	l, already, err := server.Listen(*port)
	if err != nil {
		fatal("Không mở được cổng cục bộ cho ứng dụng: " + err.Error())
	}
	if already {
		lg.Printf("đã có một bản GKM đang chạy, mở lại cửa sổ")
		openWindow(rootDir, fmt.Sprintf("http://127.0.0.1:%d/", *port), *noBrowser)
		return
	}
	err = app.Serve(l, func(url string) {
		lg.Printf("giao diện tại %s", url)
		openWindow(rootDir, url, *noBrowser)
	})
	if err != nil {
		lg.Printf("lỗi máy chủ: %v", err)
		fatal("Ứng dụng gặp lỗi: " + err.Error())
	}
	if app.RestartRequested() {
		relaunch(lg)
	}
	lg.Printf("đã thoát")
}

// relaunch mở lại chính file .exe (đã được thay bằng bản mới). Cửa sổ cũ tự tải lại trang nên không mở thêm cửa sổ.
func relaunch(lg *log.Logger) {
	exe, err := os.Executable()
	if err != nil {
		lg.Printf("không mở lại được app sau cập nhật: %v", err)
		return
	}
	args := os.Args[1:]
	hasNB := false
	for _, a := range args {
		if a == "-no-browser" || a == "--no-browser" || a == "-no-browser=true" || a == "--no-browser=true" {
			hasNB = true
		}
	}
	if !hasNB {
		args = append(append([]string{}, args...), "--no-browser")
	}
	cmd := exec.Command(exe, args...)
	if err := cmd.Start(); err != nil {
		lg.Printf("không mở lại được app sau cập nhật: %v", err)
		return
	}
	lg.Printf("đã mở lại app bản mới (pid %d)", cmd.Process.Pid)
}

func resolveRoot(flagRoot string) string {
	if flagRoot != "" {
		return flagRoot
	}
	if v := os.Getenv("GKM_ROOT"); v != "" {
		return v
	}
	if exe, err := os.Executable(); err == nil {
		if r := rootFromExe(exe); r != "" {
			return r
		}
	}
	return defaultRoot()
}

// rootFromExe: file .exe nằm trong thư mục app\ của dự án thì thư mục dự án là thư mục cha của app\.
// Nhờ vậy chép dự án sang ổ/thư mục nào (ví dụ D:\...\GKM_HoaDonNhapHang) app cũng dùng đúng thư mục đó, không cần ổ E.
func rootFromExe(exe string) string {
	dir := filepath.Dir(exe)
	if !strings.EqualFold(filepath.Base(dir), "app") {
		return ""
	}
	parent := filepath.Dir(dir)
	if _, err := os.Stat(filepath.Join(parent, "templates")); err != nil {
		if _, err2 := os.Stat(filepath.Join(parent, "docs")); err2 != nil {
			return "" // không giống thư mục dự án GKM
		}
	}
	return parent
}

// openLog ghi nhật ký vào <root>\logs; nếu không ghi được thì dùng thư mục cạnh file .exe (không dùng ổ C trừ khi .exe nằm ở đó).
func openLog(root string) (*log.Logger, func()) {
	dirs := []string{filepath.Join(root, "logs")}
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exe), "logs"))
	}
	name := "gkm_" + time.Now().Format("20060102") + ".log"
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			continue
		}
		f, err := os.OpenFile(filepath.Join(d, name), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err == nil {
			return log.New(f, "", log.LstdFlags), func() { f.Close() }
		}
	}
	return log.New(os.Stderr, "", log.LstdFlags), func() {}
}

func runSelfTest(lg *log.Logger, root, samples string) int {
	if samples == "" {
		samples = filepath.Join(root, "samples")
		if _, err := os.Stat(samples); err != nil {
			if exe, e2 := os.Executable(); e2 == nil {
				samples = filepath.Join(filepath.Dir(exe), "..", "samples")
			}
		}
	}
	fmt.Println("GKM - kiểm tra các mẫu, thư mục mẫu:", samples)
	res := core.RunSelfTest(func(n string) ([]byte, error) { return os.ReadFile(filepath.Join(samples, n)) }, embeddedTemplate)
	bad := 0
	for _, r := range res {
		if r.OK {
			fmt.Println("ĐẠT  ", r.Summary)
		} else {
			bad++
			fmt.Printf("LỖI   Mẫu %d: %s\n", r.N, strings.Join(r.Problems, "; "))
		}
		lg.Printf("selftest mẫu %d ok=%v %s %v", r.N, r.OK, r.Summary, r.Problems)
	}
	if bad == 0 {
		fmt.Println("Tất cả các mẫu đạt (Mẫu 1 khác kết quả cũ 1 ô mã PTXD0164; Mẫu 2 khác 80 ô tên hàng được điền; Mẫu 3 khớp; Mẫu 4 và 5 đọc đúng, tổng khớp).")
		return 0
	}
	return 1
}

func runCLI(lg *log.Logger, root, out string, files []string) int {
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Chưa đưa file hóa đơn nào. Ví dụ: GKM_NhapHoaDon.exe --cli hoadon.xlsx --out D:\\thu_muc_xuat")
		return 2
	}
	if out == "" {
		out = filepath.Join(root, "output")
	}
	code := 0
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("LỖI  %s: không đọc được file: %v\n", f, err)
			code = 1
			continue
		}
		g, _, err := core.ReadGridAny(data)
		var p *core.Parsed
		if err == nil {
			p, err = core.Parse(g)
		}
		if err != nil {
			fmt.Printf("LỖI  %s: %v\n", filepath.Base(f), err)
			code = 1
			continue
		}
		n, k, d, x := p.Counts()
		fmt.Printf("%s: %s | %d dòng (%d thường, %d khuyến mại, %d hàng đổi; bỏ %d dòng không có số lượng)\n", filepath.Base(f), core.ProfileNames[p.Profile], n+k+d, n, k, d, x)
		for _, c := range p.Checks {
			fmt.Printf("   [%s] %s: %s\n", c.Level, c.Name, c.Detail)
		}
		for _, w := range p.Warnings {
			fmt.Println("   [cảnh báo]", w)
		}
		if len(p.Errors) > 0 {
			for _, e := range p.Errors {
				fmt.Println("   [lỗi]", e)
			}
			code = 1
			continue
		}
		if p.NeedConfirm() {
			fmt.Println("   Có cảnh báo: dùng giao diện để xem và xác nhận. Chế độ --cli không xuất file này.")
			code = 1
			continue
		}
		rows := p.OutputRows()
		data2, err := core.BuildWorkbook(embeddedTemplate, rows)
		if err == nil {
			err = core.VerifyExport(data2, rows)
		}
		if err != nil {
			fmt.Println("   [lỗi]", err)
			code = 1
			continue
		}
		if err := os.MkdirAll(out, 0o755); err != nil {
			fmt.Println("   [lỗi] không tạo được thư mục xuất:", err)
			code = 1
			continue
		}
		path := core.UniquePath(out, p.SuggestedName())
		if err := core.WriteNew(path, data2); err != nil {
			fmt.Println("   [lỗi]", err)
			code = 1
			continue
		}
		fmt.Println("   Đã xuất:", path)
		lg.Printf("cli: xuất %s từ %s", path, f)
	}
	return code
}
