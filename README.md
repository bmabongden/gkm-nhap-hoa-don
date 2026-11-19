# GKM_HoaDonNhapHang

App Windows độc lập "GKM - Nhập Hóa Đơn" cho Gia Khánh Motor: kéo thả 1 hóa đơn Excel (.xlsx/.xls), xem trước, nút Làm mới, xuất file `.xlsx` đúng `MauFileNhapHang.xlsx`; tự xác định đơn vị đơn giá. Có nút Cập nhật online qua kho GitHub. File nhiều toa thì mỗi toa xuất một file. Đọc đúng số dù NCC dùng dấu . hay ,. Hóa đơn có % giảm giá thì nhập giá gốc kèm %. Chiết khấu cả hóa đơn chia vào % từng dòng; cảnh báo khi xuất lại hóa đơn đã xuất; bản cập nhật có chữ ký số. Bản chính thức 1.9.1 (19/11/2026): đổi khóa chữ ký số, phát hành tự động khi gộp vào main.

- Dùng: chép thư mục này vào `E:\GKM_HoaDonNhapHang`, bấm đúp `app\GKM_NhapHoaDon.exe`. Xem `docs\HUONG_DAN.md`.
- Kiểm tra: `app\KiemTraMau.bat`. Kết quả kiểm thử đã làm: `docs\KIEM_THU.md` (kèm phần chưa kiểm chứng trên Windows thật).
- Build lại: `build\build.bat` (Windows, cần Go 1.23+) hoặc `build/build.sh`. Thư viện đã có sẵn trong `source\vendor`.
- Quy tắc: `docs\GKM_Quy_Tac_Anh_Xa.md`. Goal mới của Project: `docs\GOAL_MOI.md`.
- Cập nhật online: app đọc `update/latest.json` và `update/GKM_NhapHoaDon.exe` trong kho này (xem `docs\HUONG_DAN.md` mục 4). Phát hành bản mới (từ 19/11/2026 tự động): nâng `const version` trong `source/cmd/gkm/main.go`, gộp vào `main` là GitHub Actions (`.github/workflows/phat-hanh.yml`) tự build, ký bằng secret `GKM_SIGN_KEY` và commit thư mục `update/`; ghi chú lấy từ tiêu đề commit. Cách tay (`build/build.sh` rồi `bash build/publish_update.sh "ghi chú"`) vẫn dùng được.
- Thư mục `samples\` (hóa đơn mẫu thật của nhà cung cấp) KHÔNG nằm trong kho công khai; kiểm thử mẫu (`app\KiemTraMau.bat`, `--selftest`, `go test`) chỉ chạy đủ khi chép thư mục `samples` từ bản đầy đủ của Gia Khánh Motor vào, nếu không thì được bỏ qua.
- Ứng dụng chỉ sử dụng nội bộ Gia Khánh Motor, vui lòng không sao chép dưới mọi hình thức.
