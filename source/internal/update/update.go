// Package update kiểm tra và cài bản mới của app từ kho GitHub công khai.
//
// Kho chứa hai file trên nhánh mặc định:
//
//	update/latest.json          {"version":"1.9.0","sha256":"<64 ký tự hex>","size":12345678,"notes":"...","sig":"<chữ ký>"}
//	update/GKM_NhapHoaDon.exe   file .exe của đúng phiên bản đó
//
// "sig" là chữ ký số Ed25519 (base64) trên SignMessage(version, sha256, size), ký bằng khóa bí mật của Gia Khánh Motor
// cất NGOÀI GitHub; khóa công khai nhúng trong app (PublicKeyB64). Ai chiếm được kho GitHub cũng không làm ra được chữ ký,
// nên app từ chối mọi bản không có chữ ký đúng. App chỉ cập nhật khi người dùng bấm nút, kiểm tra chữ ký, SHA-256 và
// chạy thử "-version" của file tải về trước khi thay, giữ lại bản cũ thành GKM_NhapHoaDon.exe.old để quay lại được.
package update

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DefaultRepo là kho mặc định (owner/repo) khi người dùng chưa nhập gì trong màn hình Cập nhật. Có thể đổi khi build bằng -ldflags.
var DefaultRepo = "bmabongden/gkm-nhap-hoa-don"

// PublicKeyB64 là khóa công khai Ed25519 (base64) dùng để kiểm tra chữ ký bản cập nhật.
var PublicKeyB64 = "pXnIRJhe7yWiggNTkCML+8sdFjdVy2dAWDWYaIZ0atc="

// SignMessage là nội dung được ký cho một bản cập nhật.
func SignMessage(version, sha string, size int64) []byte {
	return []byte(fmt.Sprintf("GKM-NhapHoaDon update v1\nversion=%s\nsha256=%s\nsize=%d\n", version, strings.ToLower(sha), size))
}

// Sign ký bản cập nhật (dùng trong công cụ phát hành, không dùng trong app).
func Sign(priv ed25519.PrivateKey, version, sha string, size int64) string {
	return base64.StdEncoding.EncodeToString(ed25519.Sign(priv, SignMessage(version, sha, size)))
}

// verify kiểm tra chữ ký số của latest.json; sai hoặc thiếu thì không được cài.
func (u *Updater) verify(in *Info) error {
	pub, err := u.pubKey()
	if err != nil {
		return err
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in.Sig))
	if err != nil || len(sig) != ed25519.SignatureSize || !ed25519.Verify(pub, SignMessage(in.Version, in.SHA256, in.Size), sig) {
		return fmt.Errorf("bản %s trên kho không có chữ ký số hợp lệ của Gia Khánh Motor nên app KHÔNG cài (có thể kho bị người lạ sửa); hãy báo người hỗ trợ", in.Version)
	}
	return nil
}

func (u *Updater) pubKey() (ed25519.PublicKey, error) {
	if len(u.PublicKey) == ed25519.PublicKeySize {
		return u.PublicKey, nil
	}
	b, err := base64.StdEncoding.DecodeString(PublicKeyB64)
	if err != nil || len(b) != ed25519.PublicKeySize {
		return nil, errors.New("app thiếu khóa kiểm tra chữ ký nên không cập nhật được")
	}
	return ed25519.PublicKey(b), nil
}

const (
	manifestPath = "update/latest.json"
	exePath      = "update/GKM_NhapHoaDon.exe"
	maxExeSize   = 80 << 20
	apiBase      = "https://api.github.com"
)

// Config là cấu hình lưu trong config\settings.json.
type Config struct {
	Repo string `json:"repo"` // owner/repo (kho công khai, không cần mã truy cập)
}

var reRepo = regexp.MustCompile(`^[A-Za-z0-9._-]{1,100}/[A-Za-z0-9._-]{1,100}$`)

// ValidRepo kiểm tra dạng owner/repo.
func ValidRepo(s string) bool {
	if !reRepo.MatchString(s) {
		return false
	}
	for _, part := range strings.Split(s, "/") {
		if strings.HasPrefix(part, ".") { // chặn "." và ".." chui ra khỏi đường dẫn API
			return false
		}
	}
	return true
}

// Info là thông tin bản mới nhất trên kho.
type Info struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
	Size    int64  `json:"size"`
	Notes   string `json:"notes"`
	Sig     string `json:"sig,omitempty"`
	Newer   bool   `json:"newer"`
}

// Updater làm việc cập nhật. Các trường hàm cho phép thử nghiệm mà không cần mạng/Windows.
type Updater struct {
	Current string // phiên bản đang chạy, ví dụ "1.5.0"
	Cfg     func() Config
	BaseURL string // mặc định https://api.github.com (chỉ đổi khi thử nghiệm)
	HTTP    *http.Client
	// ExePath trả đường dẫn file .exe đang chạy (mặc định os.Executable).
	ExePath func() (string, error)
	// Smoke chạy thử file mới và xác nhận nó báo đúng phiên bản (mặc định chạy "<file> -version").
	Smoke func(path, wantVersion string) error
	// PublicKey thay khóa nhúng sẵn (chỉ dùng khi thử nghiệm).
	PublicKey ed25519.PublicKey
}

func (u *Updater) base() string {
	if u.BaseURL != "" {
		return strings.TrimRight(u.BaseURL, "/")
	}
	return apiBase
}

func (u *Updater) client() *http.Client {
	if u.HTTP != nil {
		return u.HTTP
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

func (u *Updater) repo() (Config, error) {
	c := u.Cfg()
	if c.Repo == "" {
		c.Repo = DefaultRepo
	}
	c.Repo = strings.TrimSpace(c.Repo)
	if c.Repo == "" {
		return c, errors.New("chưa nhập kho GitHub chứa bản cập nhật (dạng tên-tài-khoản/tên-kho)")
	}
	if !ValidRepo(c.Repo) {
		return c, fmt.Errorf("tên kho %q không hợp lệ, cần dạng tên-tài-khoản/tên-kho", c.Repo)
	}
	return c, nil
}

func (u *Updater) get(ctx context.Context, c Config, path string) (*http.Response, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/contents/%s", u.base(), c.Repo, path)
	if _, err := url.Parse(endpoint); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.raw+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "GKM-NhapHoaDon/"+u.Current)
	resp, err := u.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("không kết nối được GitHub (kiểm tra Internet): %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, errors.New("không thấy bản cập nhật trên kho (sai tên kho, kho chưa công khai, hoặc chưa có thư mục update)")
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("GitHub từ chối truy cập (mã %d): có thể đã hết lượt gọi tạm thời, hãy thử lại sau ít phút", resp.StatusCode)
		}
		return nil, fmt.Errorf("GitHub trả lỗi %d", resp.StatusCode)
	}
	return resp, nil
}

// Check đọc latest.json trên kho và so với phiên bản đang chạy.
func (u *Updater) Check(ctx context.Context) (*Info, error) {
	c, err := u.repo()
	if err != nil {
		return nil, err
	}
	resp, err := u.get(ctx, c, manifestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var in Info
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&in); err != nil {
		return nil, fmt.Errorf("file latest.json trên kho không đọc được: %w", err)
	}
	in.Version = strings.TrimPrefix(strings.TrimSpace(in.Version), "v")
	in.SHA256 = strings.ToLower(strings.TrimSpace(in.SHA256))
	if _, ok := parseVer(in.Version); !ok {
		return nil, fmt.Errorf("phiên bản %q trong latest.json không hợp lệ", in.Version)
	}
	if len(in.SHA256) != 64 || !isHex(in.SHA256) {
		return nil, errors.New("latest.json thiếu mã SHA-256 hợp lệ nên không thể cập nhật an toàn")
	}
	if in.Size <= 0 || in.Size > maxExeSize {
		return nil, errors.New("latest.json ghi kích thước file không hợp lệ")
	}
	newer := Newer(in.Version, u.Current)
	if newer {
		if err := u.verify(&in); err != nil {
			return nil, err
		}
	}
	in.Newer = newer
	return &in, nil
}

// Download tải file .exe của bản info vào thư mục của file đang chạy (tên .new), kiểm tra kích thước, SHA-256, dạng PE và chạy thử.
func (u *Updater) Download(ctx context.Context, info *Info) (string, error) {
	c, err := u.repo()
	if err != nil {
		return "", err
	}
	if !info.Newer {
		return "", errors.New("không có bản mới hơn để cập nhật")
	}
	if err := u.verify(info); err != nil {
		return "", err
	}
	cur, err := u.exe()
	if err != nil {
		return "", err
	}
	resp, err := u.get(ctx, c, exePath)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tmp := cur + ".new"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return "", fmt.Errorf("không ghi được file tải về cạnh app (%s): %w", filepath.Dir(cur), err)
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(resp.Body, maxExeSize+1))
	cerr := f.Close()
	fail := func(e error) (string, error) { _ = os.Remove(tmp); return "", e }
	if err != nil {
		return fail(fmt.Errorf("tải file bị gián đoạn: %w", err))
	}
	if cerr != nil {
		return fail(cerr)
	}
	if n > maxExeSize || n != info.Size {
		return fail(fmt.Errorf("file tải về có kích thước %d, khác kích thước ghi trong latest.json (%d)", n, info.Size))
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != info.SHA256 {
		return fail(errors.New("mã SHA-256 của file tải về không khớp latest.json nên không cài (file hỏng hoặc bị thay)"))
	}
	if err := checkPE(tmp); err != nil {
		return fail(err)
	}
	smoke := u.Smoke
	if smoke == nil {
		smoke = runVersion
	}
	if err := smoke(tmp, info.Version); err != nil {
		return fail(fmt.Errorf("file tải về không chạy được hoặc không đúng phiên bản: %w", err))
	}
	return tmp, nil
}

// Apply thay file đang chạy bằng file mới (đã Download). Bản cũ giữ thành <exe>.old. Lỗi giữa chừng thì trả lại như cũ.
func (u *Updater) Apply(newPath string) error {
	cur, err := u.exe()
	if err != nil {
		return err
	}
	old := cur + ".old"
	_ = os.Remove(old)
	if err := os.Rename(cur, old); err != nil {
		return fmt.Errorf("không đổi tên được file app đang chạy (%s): %w", cur, err)
	}
	if err := os.Rename(newPath, cur); err != nil {
		if e2 := os.Rename(old, cur); e2 != nil {
			return fmt.Errorf("thay file thất bại (%v) và không khôi phục được bản cũ (%v): hãy đổi tên %s về %s", err, e2, old, cur)
		}
		return fmt.Errorf("không thay được file app: %w", err)
	}
	return nil
}

// CleanupOld xóa file .old/.new còn sót sau lần cập nhật trước (gọi lúc khởi động).
func CleanupOld(exe string) {
	_ = os.Remove(exe + ".old")
	_ = os.Remove(exe + ".new")
}

func (u *Updater) exe() (string, error) {
	if u.ExePath != nil {
		return u.ExePath()
	}
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(p)
}

func checkPE(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b := make([]byte, 2)
	if _, err := io.ReadFull(f, b); err != nil || string(b) != "MZ" {
		return errors.New("file tải về không phải chương trình Windows (.exe)")
	}
	return nil
}

func runVersion(path, want string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "-version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if !strings.Contains(string(out), want) {
		return fmt.Errorf("báo phiên bản %q, cần %s", strings.TrimSpace(string(out)), want)
	}
	return nil
}

func isHex(s string) bool {
	for _, r := range s {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

func parseVer(s string) ([3]int, bool) {
	var v [3]int
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexAny(s, " -+("); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return v, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return v, false
		}
		v[i] = n
	}
	return v, true
}

// Newer cho biết a có mới hơn b (so từng số; phần chữ sau số, ví dụ "1.5.0 (bản chạy thử)", được bỏ qua).
func Newer(a, b string) bool {
	x, ok1 := parseVer(a)
	y, ok2 := parseVer(b)
	if !ok1 || !ok2 {
		return false
	}
	for i := 0; i < 3; i++ {
		if x[i] != y[i] {
			return x[i] > y[i]
		}
	}
	return false
}
