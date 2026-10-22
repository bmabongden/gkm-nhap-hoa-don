package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fakeExe(tag string) []byte { return append([]byte("MZ"), []byte(strings.Repeat(tag, 50))...) }

type fakeRepo struct {
	manifest map[string]any
	exe      []byte
	token    string // nếu khác rỗng thì bắt buộc gửi đúng mã
	seenAuth string
}

func (f *fakeRepo) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.seenAuth = r.Header.Get("Authorization")
		if f.token != "" && f.seenAuth != "Bearer "+f.token {
			w.WriteHeader(404) // GitHub trả 404 cho kho riêng tư khi thiếu quyền
			return
		}
		switch r.URL.Path {
		case "/repos/acme/gkm/contents/update/latest.json":
			_ = json.NewEncoder(w).Encode(f.manifest)
		case "/repos/acme/gkm/contents/update/GKM_NhapHoaDon.exe":
			_, _ = w.Write(f.exe)
		default:
			w.WriteHeader(404)
		}
	})
}

func setup(t *testing.T, cur string, f *fakeRepo, cfg Config) (*Updater, string) {
	t.Helper()
	srv := httptest.NewServer(f.handler())
	t.Cleanup(srv.Close)
	dir := t.TempDir()
	exe := filepath.Join(dir, "GKM_NhapHoaDon.exe")
	if err := os.WriteFile(exe, fakeExe("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	u := &Updater{Current: cur, BaseURL: srv.URL, Cfg: func() Config { return cfg },
		ExePath: func() (string, error) { return exe, nil },
		Smoke:   func(path, want string) error { return nil }}
	return u, exe
}

func manifestFor(ver string, exe []byte) map[string]any {
	h := sha256.Sum256(exe)
	return map[string]any{"version": ver, "sha256": hex.EncodeToString(h[:]), "size": len(exe), "notes": "ghi chú"}
}

func TestNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.6.0", "1.5.0", true}, {"1.5.0", "1.5.0", false}, {"1.4.9", "1.5.0", false},
		{"1.10.0", "1.9.0", true}, {"2.0", "1.9.9", true}, {"v1.6.0", "1.5.0", true},
		{"abc", "1.5.0", false}, {"1.6.0", "xyz", false}, {"1.5.1 (thử)", "1.5.0", true},
	}
	for _, c := range cases {
		if got := Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%q,%q)=%v muốn %v", c.a, c.b, got, c.want)
		}
	}
}

func TestValidRepo(t *testing.T) {
	for _, s := range []string{"acme/gkm", "a-b/c_d.e"} {
		if !ValidRepo(s) {
			t.Errorf("%q phải hợp lệ", s)
		}
	}
	for _, s := range []string{"", "acme", "acme/gkm/x", "../x", "a b/c", "https://github.com/a/b", "a/b?x=1"} {
		if ValidRepo(s) {
			t.Errorf("%q phải không hợp lệ", s)
		}
	}
}

func TestCheckAndApplyOK(t *testing.T) {
	newExe := fakeExe("new")
	f := &fakeRepo{manifest: manifestFor("1.6.0", newExe), exe: newExe}
	u, exe := setup(t, "1.5.0", f, Config{Repo: "acme/gkm"})
	ctx := context.Background()
	info, err := u.Check(ctx)
	if err != nil || !info.Newer || info.Version != "1.6.0" {
		t.Fatalf("Check: %+v %v", info, err)
	}
	tmp, err := u.Download(ctx, info)
	if err != nil {
		t.Fatal(err)
	}
	if err := u.Apply(tmp); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(exe)
	if string(got) != string(newExe) {
		t.Fatal("file chưa được thay bằng bản mới")
	}
	if old, _ := os.ReadFile(exe + ".old"); string(old) != string(fakeExe("old")) {
		t.Fatal("bản cũ phải được giữ thành .old")
	}
	if _, err := os.Stat(exe + ".new"); !os.IsNotExist(err) {
		t.Fatal("file .new phải biến mất sau khi thay")
	}
	CleanupOld(exe)
	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Fatal("CleanupOld phải xóa .old")
	}
}

func TestNoNewerVersion(t *testing.T) {
	e := fakeExe("x")
	f := &fakeRepo{manifest: manifestFor("1.5.0", e), exe: e}
	u, _ := setup(t, "1.5.0", f, Config{Repo: "acme/gkm"})
	info, err := u.Check(context.Background())
	if err != nil || info.Newer {
		t.Fatalf("không được báo có bản mới: %+v %v", info, err)
	}
	if _, err := u.Download(context.Background(), info); err == nil {
		t.Fatal("không được tải khi không có bản mới")
	}
}

func TestDownloadRejectsBad(t *testing.T) {
	good := fakeExe("new")
	cases := map[string]struct {
		served []byte
		mutate func(m map[string]any)
		smoke  func(string, string) error
		want   string
	}{
		"sai sha":   {served: fakeExe("evl"), want: "SHA-256"},
		"sai cỡ":    {served: append(good, 'x'), want: "kích thước"},
		"không PE":  {served: []byte(strings.Repeat("a", len(good))), mutate: nil, want: "SHA-256"},
		"smoke lỗi": {served: good, smoke: func(string, string) error { return os.ErrInvalid }, want: "không chạy được"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f := &fakeRepo{manifest: manifestFor("1.6.0", good), exe: c.served}
			u, exe := setup(t, "1.5.0", f, Config{Repo: "acme/gkm"})
			if c.smoke != nil {
				u.Smoke = c.smoke
			}
			info, err := u.Check(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			_, err = u.Download(context.Background(), info)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("muốn lỗi chứa %q, được %v", c.want, err)
			}
			if _, e := os.Stat(exe + ".new"); !os.IsNotExist(e) {
				t.Fatal("file tải hỏng phải bị xóa")
			}
			if b, _ := os.ReadFile(exe); string(b) != string(fakeExe("old")) {
				t.Fatal("file đang chạy không được bị đụng tới")
			}
		})
	}
}

func TestNotPE(t *testing.T) {
	bad := []byte(strings.Repeat("a", 100))
	f := &fakeRepo{manifest: manifestFor("1.6.0", bad), exe: bad}
	u, _ := setup(t, "1.5.0", f, Config{Repo: "acme/gkm"})
	info, err := u.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := u.Download(context.Background(), info); err == nil || !strings.Contains(err.Error(), "không phải chương trình Windows") {
		t.Fatalf("muốn lỗi không phải .exe, được %v", err)
	}
}

func TestPrivateRepoNeedsToken(t *testing.T) {
	e := fakeExe("new")
	f := &fakeRepo{manifest: manifestFor("1.6.0", e), exe: e, token: "ghp_test"}
	u, _ := setup(t, "1.5.0", f, Config{Repo: "acme/gkm"})
	if _, err := u.Check(context.Background()); err == nil || !strings.Contains(err.Error(), "mã truy cập") {
		t.Fatalf("thiếu mã phải báo lỗi dễ hiểu: %v", err)
	}
	u2, _ := setup(t, "1.5.0", f, Config{Repo: "acme/gkm", Token: "ghp_test"})
	if _, err := u2.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	if f.seenAuth != "Bearer ghp_test" {
		t.Fatalf("tiêu đề Authorization sai: %q", f.seenAuth)
	}
}

func TestBadManifest(t *testing.T) {
	e := fakeExe("new")
	for name, m := range map[string]map[string]any{
		"thiếu sha":      {"version": "1.6.0", "size": len(e)},
		"sha không hex":  {"version": "1.6.0", "size": len(e), "sha256": strings.Repeat("z", 64)},
		"version lạ":     {"version": "abc", "size": len(e), "sha256": strings.Repeat("a", 64)},
		"size không hợp": {"version": "1.6.0", "size": 0, "sha256": strings.Repeat("a", 64)},
	} {
		t.Run(name, func(t *testing.T) {
			f := &fakeRepo{manifest: m, exe: e}
			u, _ := setup(t, "1.5.0", f, Config{Repo: "acme/gkm"})
			if _, err := u.Check(context.Background()); err == nil {
				t.Fatal("phải từ chối latest.json hỏng")
			}
		})
	}
}

func TestRepoRequired(t *testing.T) {
	defer func(v string) { DefaultRepo = v }(DefaultRepo)
	DefaultRepo = ""
	f := &fakeRepo{}
	u, _ := setup(t, "1.5.0", f, Config{})
	if _, err := u.Check(context.Background()); err == nil || !strings.Contains(err.Error(), "chưa nhập kho") {
		t.Fatalf("%v", err)
	}
	u, _ = setup(t, "1.5.0", f, Config{Repo: "../evil"})
	if _, err := u.Check(context.Background()); err == nil || !strings.Contains(err.Error(), "không hợp lệ") {
		t.Fatalf("%v", err)
	}
}

func TestApplyRollsBack(t *testing.T) {
	e := fakeExe("new")
	f := &fakeRepo{manifest: manifestFor("1.6.0", e), exe: e}
	u, exe := setup(t, "1.5.0", f, Config{Repo: "acme/gkm"})
	// file mới không tồn tại -> bước đổi tên thứ hai lỗi, bản cũ phải được trả về chỗ cũ
	if err := u.Apply(exe + ".new"); err == nil {
		t.Fatal("phải lỗi")
	}
	if b, _ := os.ReadFile(exe); string(b) != string(fakeExe("old")) {
		t.Fatal("phải khôi phục bản cũ")
	}
}
