package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gkm/internal/update"
)

type noPlat struct{}

func (noPlat) PickFolder(string) (string, error) { return "", nil }
func (noPlat) OpenFolder(string) error           { return nil }

func newApp(t *testing.T) *App {
	t.Helper()
	return New("1.5.0", nil, t.TempDir(), noPlat{}, log.New(io.Discard, "", 0))
}

func do(t *testing.T, h http.Handler, method, path, body string, hdr map[string]string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, "http://127.0.0.1:47631"+path, strings.NewReader(body))
	if method == http.MethodPost {
		req.Header.Set("X-GKM", "1")
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var m map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &m)
	return rec.Code, m
}

func TestConfigMergeAndTokenHidden(t *testing.T) {
	a := newApp(t)
	h := a.Handler()
	do(t, h, "POST", "/api/config", `{"outputDir":"E:\\out"}`, nil)
	do(t, h, "POST", "/api/config", `{"updateRepo":"acme/gkm","updateToken":"secret123"}`, nil)
	_, m := do(t, h, "GET", "/api/config", "", nil)
	if m["outputDir"] != `E:\out` {
		t.Fatalf("lưu repo không được làm mất thư mục xuất: %v", m["outputDir"])
	}
	if m["updateRepo"] != "acme/gkm" || m["hasUpdateToken"] != true {
		t.Fatalf("cấu hình cập nhật sai: %v", m)
	}
	b, _ := json.Marshal(m)
	if strings.Contains(string(b), "secret123") {
		t.Fatal("không được trả mã truy cập về giao diện")
	}
	// lưu thư mục xuất không được xóa kho/mã
	do(t, h, "POST", "/api/config", `{"outputDir":"E:\\out2"}`, nil)
	_, m = do(t, h, "GET", "/api/config", "", nil)
	if m["updateRepo"] != "acme/gkm" || m["hasUpdateToken"] != true {
		t.Fatalf("đổi thư mục xuất làm mất cấu hình cập nhật: %v", m)
	}
	if code, _ := do(t, h, "POST", "/api/config", `{"updateRepo":"../evil"}`, nil); code != 400 {
		t.Fatalf("kho sai phải bị từ chối, được %d", code)
	}
	if code, _ := do(t, h, "POST", "/api/config", `{"outputDir":"  "}`, nil); code != 400 {
		t.Fatalf("thư mục rỗng phải bị từ chối, được %d", code)
	}
}

func TestGuard(t *testing.T) {
	h := newApp(t).Handler()
	req := httptest.NewRequest("POST", "http://evil.com/api/update/apply", nil)
	req.Header.Set("X-GKM", "1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("Host lạ phải bị chặn, được %d", rec.Code)
	}
	req = httptest.NewRequest("POST", "http://127.0.0.1:1/api/update/apply", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("POST thiếu X-GKM phải bị chặn, được %d", rec.Code)
	}
	req = httptest.NewRequest("POST", "http://127.0.0.1:1/api/update/apply", nil)
	req.Header.Set("X-GKM", "1")
	req.Header.Set("Origin", "https://evil.com")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("Origin lạ phải bị chặn, được %d", rec.Code)
	}
}

func TestUpdateFlow(t *testing.T) {
	defer func(v string) { update.DefaultRepo = v }(update.DefaultRepo)
	update.DefaultRepo = ""
	newExe := append([]byte("MZ"), []byte(strings.Repeat("new", 40))...)
	sum := sha256.Sum256(newExe)
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/gkm/contents/update/latest.json":
			_ = json.NewEncoder(w).Encode(map[string]any{"version": "1.6.0", "sha256": hex.EncodeToString(sum[:]), "size": len(newExe), "notes": "bản thử"})
		case "/repos/acme/gkm/contents/update/GKM_NhapHoaDon.exe":
			_, _ = w.Write(newExe)
		default:
			w.WriteHeader(404)
		}
	}))
	defer gh.Close()
	a := newApp(t)
	dir := t.TempDir()
	exe := filepath.Join(dir, "GKM_NhapHoaDon.exe")
	_ = os.WriteFile(exe, []byte("MZold"), 0o755)
	a.Updater.BaseURL = gh.URL
	a.Updater.ExePath = func() (string, error) { return exe, nil }
	a.Updater.Smoke = func(string, string) error { return nil }
	h := a.Handler()

	// chưa nhập kho
	if code, m := do(t, h, "GET", "/api/update/check", "", nil); code != 502 || !strings.Contains(m["error"].(string), "chưa nhập kho") {
		t.Fatalf("thiếu kho: %d %v", code, m)
	}
	do(t, h, "POST", "/api/config", `{"updateRepo":"acme/gkm"}`, nil)
	code, m := do(t, h, "GET", "/api/update/check", "", nil)
	if code != 200 || m["newer"] != true || m["latest"] != "1.6.0" {
		t.Fatalf("check: %d %v", code, m)
	}
	code, m = do(t, h, "POST", "/api/update/apply", "", nil)
	if code != 200 || m["ok"] != true {
		t.Fatalf("apply: %d %v", code, m)
	}
	if b, _ := os.ReadFile(exe); string(b) != string(newExe) {
		t.Fatal("exe chưa được thay")
	}
	if !a.RestartRequested() {
		t.Fatal("phải đánh dấu cần mở lại")
	}
	select {
	case <-a.quit:
	case <-time.After(3 * time.Second):
		t.Fatal("app phải tự thoát sau khi cài để mở lại bản mới")
	}
}

func TestUpdateAlreadyLatest(t *testing.T) {
	e := []byte("MZxx")
	sum := sha256.Sum256(e)
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"version": "1.5.0", "sha256": hex.EncodeToString(sum[:]), "size": len(e)})
	}))
	defer gh.Close()
	a := newApp(t)
	a.Updater.BaseURL = gh.URL
	h := a.Handler()
	do(t, h, "POST", "/api/config", `{"updateRepo":"acme/gkm"}`, nil)
	if _, m := do(t, h, "GET", "/api/update/check", "", nil); m["newer"] != false {
		t.Fatalf("%v", m)
	}
	if code, _ := do(t, h, "POST", "/api/update/apply", "", nil); code != 502 {
		t.Fatalf("không có bản mới thì apply phải báo lỗi, được %d", code)
	}
	if a.RestartRequested() {
		t.Fatal("không được đòi mở lại")
	}
	_ = update.DefaultRepo
}
