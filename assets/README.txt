Logo GKM (Gia Khánh Motor): logo.svg (gốc, vẽ bằng đường nét, không dùng font), logo-256.png, logo.ico (16-256 px).
Logo cũng nằm trong giao diện (index.html, SVG nhúng) và trong file .exe (source/cmd/gkm/rsrc_windows_amd64.syso).
Tạo lại .syso khi đổi logo/phiên bản (cần go-winres, chạy trên máy có Internet):
  go install github.com/tc-hib/go-winres@latest
  go-winres simply --arch amd64 --icon assets/logo.ico --manifest gui --file-version X.Y.Z.0 --product-version X.Y.Z.0 \
    --product-name "GKM - Nhập Hóa Đơn" --file-description "GKM - Nhập Hóa Đơn (Gia Khánh Motor)" \
    --copyright "Gia Khánh Motor, chỉ sử dụng nội bộ" --original-filename GKM_NhapHoaDon.exe --out source/cmd/gkm/rsrc
