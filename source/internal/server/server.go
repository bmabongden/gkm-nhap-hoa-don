// Package server là giao diện của app: máy chủ HTTP chỉ nghe 127.0.0.1, phục vụ trang web tiếng Việt nhúng sẵn
// trong file .exe. Không dùng Internet, API hay khóa nào.
package server

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gkm/internal/core"
	"gkm/internal/update"
)

//go:embed web/index.html
var indexHTML []byte

// Platform gom các chức năng riêng của hệ điều hành (chọn thư mục, mở thư mục).
type Platform interface {
	PickFolder(initial string) (string, error)
	OpenFolder(path string) error
}

// App chứa trạng thái của ứng dụng.
type App struct {
	Version   string
	Template  []byte
	Root      string // thư mục dự án, mặc định E:\GKM_HoaDonNhapHang
	Platform  Platform
	Log       *log.Logger
	MaxUpload int64
	// Updater cập nhật app từ kho GitHub (nil thì tắt chức năng cập nhật).
	Updater *update.Updater

	mu       sync.Mutex
	results  map[string]*entry
	order    []string
	lastPing time.Time
	closeAt  time.Time
	started  time.Time
	srv      *http.Server
	port     int
	quit     chan struct{}
	quitOnce sync.Once
	updating bool // đang tải/cài bản mới
	restart  bool // sau khi thoát phải mở lại bản vừa cập nhật
}

type entry struct {
	name   string
	format string
	grid   *core.Grid
	parsed *core.Parsed
}

// RowView là một dòng hiển thị ở bản xem trước.
type RowView struct {
	N     int     `json:"n"`
	STT   string  `json:"stt"`
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	CK    string  `json:"ck"`
	Qty   string  `json:"qty"`
	Kind  string  `json:"kind"`
	Note  string  `json:"note,omitempty"`
}

// ConvertResp là kết quả phân tích gửi cho giao diện.
type ConvertResp struct {
	ID            string       `json:"id"`
	FileName      string       `json:"fileName"`
	OK            bool         `json:"ok"`
	Error         string       `json:"error,omitempty"`
	Profile       string       `json:"profile"`
	ProfileName   string       `json:"profileName"`
	Known         bool         `json:"known"`
	So            string       `json:"so"`
	Ngay          string       `json:"ngay"`
	Rows          int          `json:"rows"`
	Normal        int          `json:"normal"`
	Promo         int          `json:"promo"`
	Exchange      int          `json:"exchange"`
	Excluded      int          `json:"excluded"`
	Items         []RowView    `json:"items"`
	ExcludedItems []RowView    `json:"excludedItems"`
	Checks        []core.Check `json:"checks"`
	Warnings      []string     `json:"warnings"`
	Errors        []string     `json:"errors"`
	NeedConfirm   bool         `json:"needConfirm"`
	SuggestedName string       `json:"suggestedName"`
	Format        string       `json:"format"`
	Scale         float64      `json:"scale"`
	HasName       bool         `json:"hasName"`
	ScaleNote     string       `json:"scaleNote"`
}

func New(version string, template []byte, root string, pl Platform, lg *log.Logger) *App {
	a := &App{Version: version, Template: template, Root: root, Platform: pl, Log: lg, MaxUpload: 25 << 20,
		results: map[string]*entry{}, started: time.Now(), quit: make(chan struct{})}
	a.Updater = &update.Updater{Current: version, Cfg: func() update.Config {
		s := a.loadSettings()
		return update.Config{Repo: s.UpdateRepo, Token: s.UpdateToken}
	}}
	return a
}

// DefaultOutputDir là thư mục xuất mặc định (nằm trong thư mục dự án, ổ E trên Windows).
func (a *App) DefaultOutputDir() string { return filepath.Join(a.Root, "output") }

func (a *App) settingsPath() string { return filepath.Join(a.Root, "config", "settings.json") }

type settings struct {
	OutputDir   string `json:"outputDir"`
	UpdateRepo  string `json:"updateRepo,omitempty"`  // kho GitHub chứa bản cập nhật, dạng tên-tài-khoản/tên-kho
	UpdateToken string `json:"updateToken,omitempty"` // mã truy cập chỉ đọc, chỉ cần khi kho riêng tư
}

func (a *App) loadSettings() settings {
	var s settings
	if b, err := os.ReadFile(a.settingsPath()); err == nil {
		_ = json.Unmarshal(b, &s)
	}
	if s.OutputDir == "" {
		s.OutputDir = a.DefaultOutputDir()
	}
	return s
}

func (a *App) saveSettings(s settings) {
	if err := os.MkdirAll(filepath.Dir(a.settingsPath()), 0o755); err != nil {
		a.Log.Printf("không lưu được cấu hình: %v", err)
		return
	}
	b, _ := json.MarshalIndent(s, "", "  ")
	if err := os.WriteFile(a.settingsPath(), b, 0o644); err != nil {
		a.Log.Printf("không lưu được cấu hình: %v", err)
	}
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"ok": false, "error": msg})
}

// guard chặn truy cập từ trang web khác (DNS rebinding / CSRF): chỉ nhận Host là localhost và POST có tiêu đề riêng.
func (a *App) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() { // lỗi bất ngờ trong một yêu cầu không được làm sập app hay để giao diện treo
			if e := recover(); e != nil && e != http.ErrAbortHandler {
				a.Log.Printf("lỗi bất ngờ ở %s: %v", r.URL.Path, e)
				fail(w, 500, "App gặp lỗi bất ngờ khi xử lý. Hãy bấm Làm mới và thử lại; nếu vẫn lỗi, gửi file hóa đơn và file nhật ký trong thư mục logs cho người hỗ trợ.")
			}
		}()
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}
		if host != "127.0.0.1" && host != "localhost" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			// sendBeacon (khi đóng cửa sổ) không gắn được tiêu đề riêng nên /api/closing chỉ kiểm tra Origin.
			if r.Header.Get("X-GKM") != "1" && r.URL.Path != "/api/closing" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			if o := r.Header.Get("Origin"); o != "" && !strings.HasPrefix(o, "http://127.0.0.1:") && !strings.HasPrefix(o, "http://localhost:") {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Handler trả bộ định tuyến của app.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(indexHTML)
	})
	mux.HandleFunc("/api/ping", a.handlePing)
	mux.HandleFunc("/api/closing", a.handleClosing)
	mux.HandleFunc("/api/config", a.handleConfig)
	mux.HandleFunc("/api/update/check", a.handleUpdateCheck)
	mux.HandleFunc("/api/update/apply", a.handleUpdateApply)
	mux.HandleFunc("/api/convert", a.handleConvert)
	mux.HandleFunc("/api/export", a.handleExport)
	mux.HandleFunc("/api/reset", a.handleReset)
	mux.HandleFunc("/api/pickfolder", a.handlePick)
	mux.HandleFunc("/api/openfolder", a.handleOpen)
	return a.guard(mux)
}

func (a *App) handlePing(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	a.lastPing = time.Now()
	a.closeAt = time.Time{}
	a.mu.Unlock()
	writeJSON(w, 200, map[string]any{"app": "GKM", "version": a.Version})
}

func (a *App) handleClosing(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	a.closeAt = time.Now().Add(6 * time.Second)
	a.mu.Unlock()
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (a *App) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var in struct {
			OutputDir   *string `json:"outputDir"`
			UpdateRepo  *string `json:"updateRepo"`
			UpdateToken *string `json:"updateToken"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
			fail(w, 400, "Yêu cầu không hợp lệ.")
			return
		}
		cur := a.loadSettings()
		if in.OutputDir != nil {
			if strings.TrimSpace(*in.OutputDir) == "" {
				fail(w, 400, "Thư mục xuất không hợp lệ.")
				return
			}
			cur.OutputDir = *in.OutputDir
		}
		if in.UpdateRepo != nil {
			repo := strings.TrimSpace(*in.UpdateRepo)
			if repo != "" && !update.ValidRepo(repo) {
				fail(w, 400, "Tên kho cập nhật phải có dạng tên-tài-khoản/tên-kho.")
				return
			}
			cur.UpdateRepo = repo
		}
		if in.UpdateToken != nil {
			cur.UpdateToken = strings.TrimSpace(*in.UpdateToken)
		}
		a.saveSettings(cur)
	}
	s := a.loadSettings()
	_, err := os.Stat(a.Root)
	repo := s.UpdateRepo
	if repo == "" {
		repo = update.DefaultRepo
	}
	writeJSON(w, 200, map[string]any{
		"ok": true, "version": a.Version, "root": a.Root, "rootExists": err == nil,
		"outputDir": s.OutputDir, "defaultOutputDir": a.DefaultOutputDir(), "os": runtime.GOOS,
		"updateRepo": repo, "hasUpdateToken": s.UpdateToken != "",
	})
}

// handleUpdateCheck hỏi kho GitHub xem có bản mới không (không tải gì về).
func (a *App) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	if a.Updater == nil {
		fail(w, 501, "Bản này không có chức năng cập nhật.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	info, err := a.Updater.Check(ctx)
	if err != nil {
		a.Log.Printf("kiểm tra cập nhật lỗi: %v", err)
		fail(w, 502, err.Error())
		return
	}
	a.Log.Printf("kiểm tra cập nhật: bản trên kho %s, bản đang chạy %s, mới hơn=%v", info.Version, a.Version, info.Newer)
	writeJSON(w, 200, map[string]any{"ok": true, "current": a.Version, "latest": info.Version, "newer": info.Newer, "notes": info.Notes})
}

// handleUpdateApply tải bản mới, kiểm tra, thay file .exe rồi cho app tự mở lại.
func (a *App) handleUpdateApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, "Sai phương thức.")
		return
	}
	if a.Updater == nil {
		fail(w, 501, "Bản này không có chức năng cập nhật.")
		return
	}
	a.mu.Lock()
	if a.updating {
		a.mu.Unlock()
		fail(w, 409, "Đang cập nhật, vui lòng đợi.")
		return
	}
	a.updating = true
	a.mu.Unlock()
	done := false
	defer func() {
		if !done {
			a.mu.Lock()
			a.updating = false
			a.mu.Unlock()
		}
	}()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()
	info, err := a.Updater.Check(ctx)
	if err == nil && !info.Newer {
		err = errors.New("đang dùng bản mới nhất, không cần cập nhật")
	}
	var tmp string
	if err == nil {
		tmp, err = a.Updater.Download(ctx, info)
	}
	if err == nil {
		err = a.Updater.Apply(tmp)
	}
	if err != nil {
		a.Log.Printf("cập nhật lỗi: %v", err)
		fail(w, 502, err.Error())
		return
	}
	done = true
	a.Log.Printf("đã cài bản %s (bản cũ giữ thành .old), mở lại app", info.Version)
	a.mu.Lock()
	a.restart = true
	a.mu.Unlock()
	writeJSON(w, 200, map[string]any{"ok": true, "version": info.Version})
	go func() {
		time.Sleep(800 * time.Millisecond)
		a.Quit()
	}()
}

// RestartRequested cho biết app thoát vì vừa cập nhật và cần được mở lại.
func (a *App) RestartRequested() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.restart
}

func rowView(it core.Item, n int, ck string) RowView {
	q := it.Qty.Text()
	return RowView{N: n, STT: it.STT, Code: it.Code, Name: it.Name, Price: it.Price, CK: ck, Qty: q, Kind: it.Kind, Note: it.Note}
}

// view dựng dữ liệu xem trước từ kết quả phân tích.
func view(id, name, format string, p *core.Parsed) *ConvertResp {
	resp := &ConvertResp{ID: id, FileName: name, Items: []RowView{}, ExcludedItems: []RowView{},
		Checks: []core.Check{}, Warnings: []string{}, Errors: []string{}}
	resp.OK = true
	resp.Format = format
	resp.Profile, resp.Known, resp.ProfileName = p.Profile, p.Known, core.ProfileNames[p.Profile]
	resp.So, resp.Ngay = p.Meta.So, p.Meta.Ngay
	resp.Normal, resp.Promo, resp.Exchange, resp.Excluded = p.Counts()
	n := 0
	for _, it := range p.Items {
		ck := ""
		if p.Profile == core.ProfileA {
			ck = it.CK.Text()
		}
		if it.Kind == core.KindExcluded {
			resp.ExcludedItems = append(resp.ExcludedItems, rowView(it, 0, ""))
			continue
		}
		n++
		resp.Items = append(resp.Items, rowView(it, n, ck))
	}
	resp.Rows = n
	resp.Checks = p.Checks
	resp.Warnings = append(resp.Warnings, p.Warnings...)
	resp.Errors = append(resp.Errors, p.Errors...)
	resp.NeedConfirm = p.NeedConfirm()
	resp.SuggestedName = p.SuggestedName()
	resp.Scale, resp.HasName, resp.ScaleNote = p.Scale, p.HasName, p.ScaleNote
	return resp
}

func failResp(name, msg string) *ConvertResp {
	return &ConvertResp{FileName: name, Error: msg, Items: []RowView{}, ExcludedItems: []RowView{},
		Checks: []core.Check{}, Warnings: []string{}, Errors: []string{}}
}

// readFile đọc hóa đơn .xlsx hoặc .xls (nhận dạng theo nội dung, không theo đuôi file).
func readFile(name string, data []byte) (*core.Grid, string, string) {
	g, kind, err := core.ReadGridAny(data)
	if err != nil {
		msg := err.Error()
		if kind == "" {
			msg = "File này không phải Excel (.xlsx hoặc .xls) mà app đọc được. Có thể file bị hỏng, có mật khẩu hoặc tải về chưa xong."
		}
		return nil, "", msg
	}
	return g, kind, ""
}

func parseMsg(err error) string {
	return err.Error() + ". App không đoán dữ liệu nên dừng ở đây; hãy gửi file này cho người hỗ trợ để thêm mẫu mới."
}

func (a *App) handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, "Sai phương thức.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, a.MaxUpload)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		fail(w, 400, "Không nhận được file (quá lớn hoặc bị lỗi khi tải lên).")
		return
	}
	defer r.MultipartForm.RemoveAll() // xóa file tạm nếu form lớn phải ghi ra đĩa
	f, hdr, err := r.FormFile("file")
	if err != nil {
		fail(w, 400, "Chưa chọn file.")
		return
	}
	defer f.Close()
	name := filepath.Base(hdr.Filename)
	data, err := io.ReadAll(f)
	if err != nil {
		fail(w, 400, "Không đọc được file.")
		return
	}
	g, kind, msg := readFile(name, data)
	var resp *ConvertResp
	if msg != "" {
		resp = failResp(name, msg)
	} else if p, err := core.Parse(g); err != nil {
		resp = failResp(name, parseMsg(err))
	} else {
		resp = view(newID(), name, kind, p)
		a.mu.Lock()
		a.results = map[string]*entry{resp.ID: {name: name, format: kind, grid: g, parsed: p}} // chỉ giữ 1 hóa đơn
		a.mu.Unlock()
		a.Log.Printf("phân tích %q (%s): mẫu %s, %d dòng (%d thường, %d khuyến mại, %d hàng đổi, %d bỏ vì không có số lượng), %d cảnh báo, %d lỗi", name, kind, resp.Profile, resp.Rows, resp.Normal, resp.Promo, resp.Exchange, resp.Excluded, len(resp.Warnings), len(resp.Errors))
	}
	if !resp.OK {
		a.Log.Printf("không phân tích được %q: %s", name, resp.Error)
	}
	writeJSON(w, 200, resp)
}

// handleReset xóa hóa đơn đang mở để chọn hóa đơn khác (không đụng tới file đã xuất).
func (a *App) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, "Sai phương thức.")
		return
	}
	a.mu.Lock()
	a.results = map[string]*entry{}
	a.mu.Unlock()
	a.Log.Printf("làm mới: đã xóa hóa đơn đang mở")
	writeJSON(w, 200, map[string]any{"ok": true})
}

type exportReq struct {
	ID      string `json:"id"`
	Dir     string `json:"dir"`
	Confirm bool   `json:"confirm"`
}

func (a *App) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, "Sai phương thức.")
		return
	}
	var in exportReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in); err != nil {
		fail(w, 400, "Yêu cầu không hợp lệ.")
		return
	}
	a.mu.Lock()
	e := a.results[in.ID]
	a.mu.Unlock()
	if e == nil {
		fail(w, 404, "Không tìm thấy hóa đơn đã phân tích. Hãy thả lại file hóa đơn.")
		return
	}
	p := e.parsed
	if len(p.Errors) > 0 {
		fail(w, 400, "Hóa đơn còn lỗi dữ liệu, chưa thể xuất: "+strings.Join(p.Errors, " "))
		return
	}
	if p.NeedConfirm() && !in.Confirm {
		fail(w, 400, "Hóa đơn có cảnh báo. Hãy xem lại rồi tích ô xác nhận trước khi xuất.")
		return
	}
	dir := strings.TrimSpace(in.Dir)
	if dir == "" || !filepath.IsAbs(dir) {
		fail(w, 400, "Hãy nhập đường dẫn thư mục đầy đủ (ví dụ D:\\GKM_HoaDonNhapHang\\output).")
		return
	}
	dir = filepath.Clean(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		hint := ""
		if runtime.GOOS == "windows" && len(dir) > 1 && dir[1] == ':' {
			if _, e2 := os.Stat(dir[:2] + `\`); e2 != nil {
				hint = " Ổ " + strings.ToUpper(dir[:2]) + " không tồn tại hoặc chưa cắm; hãy chọn thư mục khác."
			}
		}
		a.Log.Printf("không tạo được thư mục %s: %v", dir, err)
		fail(w, 400, "Không tạo được thư mục xuất: "+dir+"."+hint)
		return
	}
	rows := p.OutputRows()
	data, err := core.BuildWorkbook(a.Template, rows)
	if err != nil {
		a.Log.Printf("lỗi dựng file: %v", err)
		fail(w, 500, "Lỗi khi tạo file Excel: "+err.Error())
		return
	}
	if err := core.VerifyExport(data, rows); err != nil {
		a.Log.Printf("kiểm tra file xuất trong bộ nhớ không đạt: %v", err)
		fail(w, 500, "File xuất không đạt kiểm tra nên không lưu: "+err.Error())
		return
	}
	path := core.UniquePath(dir, p.SuggestedName())
	if err := core.WriteNew(path, data); err != nil {
		a.Log.Printf("không ghi được %s: %v", path, err)
		fail(w, 400, "Không ghi được file vào thư mục đã chọn (thư mục chỉ đọc hoặc không đủ quyền): "+err.Error())
		return
	}
	// mở lại file vừa ghi trên đĩa và kiểm tra lần nữa
	back, err := os.ReadFile(path)
	if err == nil {
		err = core.VerifyExport(back, rows)
	}
	if err != nil {
		a.Log.Printf("kiểm tra file trên đĩa không đạt %s: %v", path, err)
		fail(w, 500, "Đã ghi file nhưng kiểm tra lại không đạt: "+err.Error())
		return
	}
	cur := a.loadSettings()
	cur.OutputDir = dir
	a.saveSettings(cur)
	a.Log.Printf("đã xuất %s (%d dòng) từ %q", path, len(rows), e.name)
	warn := ""
	if runtime.GOOS == "windows" && strings.HasPrefix(strings.ToUpper(path), "C:") {
		warn = "File đang nằm trên ổ C. Mặc định của GKM là ổ E."
	}
	writeJSON(w, 200, map[string]any{"ok": true, "path": path, "name": filepath.Base(path), "rows": len(rows), "verified": true, "warning": warn})
}

func (a *App) handlePick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, "Sai phương thức.")
		return
	}
	var in struct {
		Initial string `json:"initial"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in)
	dir, err := a.Platform.PickFolder(in.Initial)
	if err != nil {
		fail(w, 501, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "dir": dir})
}

func (a *App) handleOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, "Sai phương thức.")
		return
	}
	var in struct {
		Path string `json:"path"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&in)
	if err := a.Platform.OpenFolder(in.Path); err != nil {
		fail(w, 501, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

// Listen mở cổng cục bộ: thử cổng cố định 47631 để chỉ chạy một bản; nếu bận thì dùng cổng ngẫu nhiên.
func Listen(preferred int) (net.Listener, bool, error) {
	if preferred > 0 {
		if l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", preferred)); err == nil {
			return l, false, nil
		}
		// cổng bận: có thể đã có một bản GKM đang chạy
		if IsGKM(preferred) {
			return nil, true, nil
		}
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	return l, false, err
}

// IsGKM kiểm tra cổng có đang chạy GKM không.
func IsGKM(port int) bool {
	c := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/api/ping", port))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	var v struct {
		App string `json:"app"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<12)).Decode(&v)
	return v.App == "GKM"
}

// Serve chạy máy chủ đến khi người dùng đóng cửa sổ (hoặc lâu không có tín hiệu).
func (a *App) Serve(l net.Listener, onReady func(url string)) error {
	a.port = l.Addr().(*net.TCPAddr).Port
	a.srv = &http.Server{Handler: a.Handler(), ReadHeaderTimeout: 10 * time.Second}
	errc := make(chan error, 1)
	go func() { errc <- a.srv.Serve(l) }()
	go a.watchdog()
	if onReady != nil {
		onReady(fmt.Sprintf("http://127.0.0.1:%d/", a.port))
	}
	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-a.quit:
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return a.srv.Shutdown(ctx)
	}
}

// Quit dừng máy chủ.
func (a *App) Quit() { a.quitOnce.Do(func() { close(a.quit) }) }

func (a *App) watchdog() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-a.quit:
			return
		case now := <-t.C:
			a.mu.Lock()
			closeAt, last := a.closeAt, a.lastPing
			a.mu.Unlock()
			switch {
			case !closeAt.IsZero() && now.After(closeAt):
				a.Log.Printf("cửa sổ đã đóng, thoát app")
				a.Quit()
			case last.IsZero() && now.Sub(a.started) > 5*time.Minute:
				a.Log.Printf("không có cửa sổ nào kết nối, thoát app")
				a.Quit()
			case !last.IsZero() && now.Sub(last) > 15*time.Minute:
				a.Log.Printf("lâu không có tín hiệu từ cửa sổ, thoát app")
				a.Quit()
			}
		}
	}
}
