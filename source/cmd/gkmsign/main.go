// Lệnh gkmsign ký bản cập nhật của GKM - Nhập Hóa Đơn bằng khóa Ed25519 (chỉ dùng khi phát hành, không nằm trong app).
//
//	go run ./cmd/gkmsign -genkey <file khóa>          tạo khóa mới, in khóa công khai (base64) để nhúng vào app
//	go run ./cmd/gkmsign -key <file khóa> -pub        in khóa công khai của file khóa
//	go run ./cmd/gkmsign -key <file khóa> -version 1.9.0 -sha <sha256> -size <byte>   in chữ ký (base64)
//
// File khóa chứa 32 byte hạt giống (base64). Đây là khóa BÍ MẬT: không bao giờ đưa lên GitHub.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"

	"gkm/internal/update"
)

func main() {
	gen := flag.String("genkey", "", "tạo khóa mới vào file này (không ghi đè file có sẵn)")
	key := flag.String("key", "", "file khóa bí mật")
	pub := flag.Bool("pub", false, "in khóa công khai")
	ver := flag.String("version", "", "phiên bản")
	sha := flag.String("sha", "", "SHA-256 của file .exe")
	size := flag.Int64("size", 0, "kích thước file .exe (byte)")
	flag.Parse()
	if *gen != "" {
		seed := make([]byte, ed25519.SeedSize)
		if _, err := rand.Read(seed); err != nil {
			fail(err)
		}
		f, err := os.OpenFile(*gen, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			fail(err)
		}
		fmt.Fprintln(f, base64.StdEncoding.EncodeToString(seed))
		if err := f.Close(); err != nil {
			fail(err)
		}
		fmt.Println(base64.StdEncoding.EncodeToString(ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)))
		return
	}
	b, err := os.ReadFile(*key)
	if err != nil {
		fail(err)
	}
	seed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(b)))
	if err != nil || len(seed) != ed25519.SeedSize {
		fail(fmt.Errorf("file khóa %s không đúng dạng", *key))
	}
	priv := ed25519.NewKeyFromSeed(seed)
	if *pub {
		fmt.Println(base64.StdEncoding.EncodeToString(priv.Public().(ed25519.PublicKey)))
		return
	}
	if *ver == "" || len(*sha) != 64 || *size <= 0 {
		fail(fmt.Errorf("cần -version, -sha (64 ký tự) và -size"))
	}
	fmt.Println(update.Sign(priv, *ver, *sha, *size))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "gkmsign:", err)
	os.Exit(1)
}
