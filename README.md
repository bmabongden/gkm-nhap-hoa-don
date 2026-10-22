# GKM_HoaDonNhapHang

App Windows độc lập "GKM - Nhập Hóa Đơn" cho Gia Khánh Motor: kéo thả 1 hóa đơn Excel (.xlsx/.xls), xem trước, nút Làm mới, xuất file `.xlsx` đúng `MauFileNhapHang.xlsx`; tự xác định đơn vị đơn giá. Có nút Cập nhật online qua kho GitHub. Bản chính thức 1.6.0 (sửa 22/10/2026).

- Dùng: chép thư mục này vào `E:\GKM_HoaDonNhapHang`, bấm đúp `app\GKM_NhapHoaDon.exe`. Xem `docs\HUONG_DAN.md`.
- Kiểm tra: `app\KiemTraMau.bat`. Kết quả kiểm thử đã làm: `docs\KIEM_THU.md` (kèm phần chưa kiểm chứng trên Windows thật).
- Build lại: `build\build.bat` (Windows, cần Go 1.23+) hoặc `build/build.sh`. Thư viện đã có sẵn trong `source\vendor`.
- Quy tắc: `docs\GKM_Quy_Tac_Anh_Xa.md`. Goal mới của Project: `docs\GOAL_MOI.md`.
- Cập nhật online: app đọc `update/latest.json` và `update/GKM_NhapHoaDon.exe` trong kho này (xem `docs\HUONG_DAN.md` mục 4). Phát hành bản mới: `build/build.sh` rồi `bash build/publish_update.sh "ghi chú"`, commit và push thư mục `update/`.
- Thư mục `samples\` (hóa đơn mẫu thật của nhà cung cấp) KHÔNG nằm trong kho công khai; kiểm thử mẫu (`app\KiemTraMau.bat`, `--selftest`, `go test`) chỉ chạy đủ khi chép thư mục `samples` từ bản đầy đủ của Gia Khánh Motor vào, nếu không thì được bỏ qua.
- Ứng dụng chỉ sử dụng nội bộ Gia Khánh Motor, vui lòng không sao chép dưới mọi hình thức.
