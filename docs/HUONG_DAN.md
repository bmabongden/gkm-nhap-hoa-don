# GKM - Nhập Hóa Đơn 1.6.0: hướng dẫn ngắn

Ứng dụng Windows độc lập (portable): **không cần Python, API, khóa/token**; xử lý hóa đơn chạy hoàn toàn trên máy, chỉ nút **Cập nhật** mới dùng Internet. Mọi thứ nằm trong thư mục dự án (trong tài liệu này ghi `E:\GKM_HoaDonNhapHang` làm ví dụ; thư mục thật là nơi anh chép dự án, ví dụ `D:\Data AI\App nhap hang\Chuan\GKM_HoaDonNhapHang`).

## 1. Cài đặt (một lần)
1. Chép nguyên thư mục `GKM_HoaDonNhapHang` (giải nén từ file zip) vào ổ/thư mục bạn muốn, ví dụ `E:\GKM_HoaDonNhapHang` hoặc `D:\Data AI\App nhap hang\Chuan\GKM_HoaDonNhapHang`. Từ bản 1.6.0 app tự dùng thư mục chứa `app\` làm thư mục dự án (không cần ổ E); giữ nguyên tên thư mục con `app`.
2. Chạy thử kiểm tra: bấm đúp `E:\GKM_HoaDonNhapHang\app\KiemTraMau.bat`. Kết quả phải là **ĐẠT** (3 mẫu gốc, các bản .xls của chúng, mẫu Tân Hòa Lợi, file gk 28-2.xls và phiếu xuất kho SL lấy).
3. Windows có thể hiện cảnh báo SmartScreen ("Windows protected your PC") vì file .exe chưa được ký số. Chọn **More info → Run anyway**. Chỉ làm khi file đến từ nguồn bạn tin cậy.

## 2. Dùng hằng ngày
1. Bấm đúp `E:\GKM_HoaDonNhapHang\app\GKM_NhapHoaDon.exe`. Một cửa sổ "GKM - Nhập Hóa Đơn" mở ra (dùng Microsoft Edge hoặc Chrome có sẵn trên máy; dữ liệu hóa đơn chỉ chạy trên máy bạn).
2. **Bước 1**: kéo thả **1** file hóa đơn Excel (`.xlsx` hoặc `.xls`) vào khung (hoặc bấm để chọn). Mỗi lần chỉ làm một hóa đơn.
3. **Bước 2 - Xem trước**: kiểm tra số dòng, hàng thường/khuyến mại, mục đối chiếu với tổng hóa đơn, cảnh báo và bảng dòng hàng. Dòng khuyến mại tô vàng. Hàng đổi/bảo hành (chỉ có số lượng) vẫn nhập với đơn giá 0 và có nhãn riêng; dòng có giá mà không có số lượng bị bỏ, nằm trong mục xổ xuống bên dưới. App **tự xác định đơn vị đơn giá**: mặc định là đồng; hóa đơn Tân Hòa Lợi và hóa đơn có ghi chú tính theo nghìn đồng (ví dụ "Đơn vị tính: nghìn đồng", "Đơn giá (1000đ)") thì app tự nhân 1.000 và ghi rõ lý do trên màn hình. Hóa đơn không có cột tên hàng thì cột Tên hàng để trống.
4. **Bước 3 - Xuất**: thư mục mặc định là `E:\GKM_HoaDonNhapHang\output` (đổi bằng ô đường dẫn hoặc nút **Chọn thư mục…**). Nếu có cảnh báo, đọc rồi tích ô xác nhận. Bấm **Xuất file Excel**. File tên `NhapHang_<số phiếu>_<ngày>.xlsx`. App tự mở lại file vừa xuất để kiểm tra rồi mới báo đạt.
5. **Làm mới**: xong hóa đơn này thì bấm nút **Làm mới** (góc trên bên phải, cạnh nút **Cập nhật**) để xóa hóa đơn cũ, rồi chọn hóa đơn khác. Chưa bấm Làm mới thì app không nhận hóa đơn thứ hai. File đã xuất không bị xóa.
6. Đóng cửa sổ là app tự thoát.

### Phiếu xuất kho kiêm giấy nhận nợ (cột "SL LẤY", "ĐƠN GIÁ BÁN")
Nhà cung cấp này hay gửi gộp cả chồng toa hàng trước đó trong một file. App **chỉ lấy bảng hóa đơn cuối cùng** (và ghi một dòng "File có N bảng hóa đơn" trong phần đối chiếu để bạn biết). Số lượng lấy từ cột **SL LẤY** (không phải SL ĐẶT), đơn giá lấy từ **ĐƠN GIÁ BÁN** (không phải ĐG ĐÃ CK). Dòng có SL LẤY bằng 0 hoặc trống bị bỏ như mọi dòng không có số lượng. Dòng "TỔNG PHẢI THU" (sau chiết khấu) chỉ hiện để tham khảo, không đưa vào file.

## 3. Quy tắc đang áp dụng
- Mã hàng giữ nguyên như hóa đơn (kể cả số 0 đầu), không có quy tắc riêng cho hàng "ốc".
- Điền tên hàng. Không giảm giá thì để trống. Dòng chỉ có số lượng (hàng đổi, bảo hành, khuyến mại) ghi đơn giá 0. Dòng có giá mà không có số lượng thì bỏ.
- Chiết khấu cấp hóa đơn và thuế suất không đưa vào file (mẫu nhập hàng không có chỗ cho chúng).
- Chi tiết: `docs\GKM_Quy_Tac_Anh_Xa.md`.

## 4. Cập nhật app qua Internet
1. Bấm nút **Cập nhật** ở góc trên. Lần đầu nhập tên kho GitHub chứa bản cập nhật (dạng `tên-tài-khoản/tên-kho`). Kho riêng tư thì mở mục "Kho riêng tư" và dán mã truy cập GitHub **chỉ đọc** (tạo ở GitHub, mục Fine-grained token, quyền Contents: Read-only chỉ cho đúng kho này). Mã được lưu nguyên văn trong `config\settings.json` trên máy, không được hiện lại trên màn hình.
2. Bấm **Kiểm tra bản mới**. Nếu có, bấm **Cập nhật ngay**: app tải file mới, kiểm tra kích thước, mã SHA-256, dạng chương trình Windows và chạy thử `-version`; đạt hết mới thay file `.exe` (bản cũ giữ thành `GKM_NhapHoaDon.exe.old`, tự xóa ở lần mở sau) rồi tự khởi động lại. Hóa đơn đang mở mà chưa xuất sẽ bị bỏ khi khởi động lại, app có hỏi trước.
3. Ô tên kho đã có sẵn `bmabongden/gkm-nhap-hoa-don` (kho riêng tư nên cần mã truy cập chỉ đọc). Nếu cập nhật lỗi, app giữ nguyên bản đang chạy; muốn quay về bản cũ thì đóng app, xóa `GKM_NhapHoaDon.exe` và đổi tên `GKM_NhapHoaDon.exe.old` về `GKM_NhapHoaDon.exe`.
4. Phía người phát hành: sau `build\build.sh`, chạy `bash build/publish_update.sh "ghi chú"` để tạo `update\latest.json` và `update\GKM_NhapHoaDon.exe`, rồi commit và push hai file đó lên nhánh chính của kho. Chỉ ai có quyền ghi vào kho mới phát hành được bản mới, vì vậy hãy bảo vệ tài khoản GitHub.

## 5. Khi app báo lỗi
| Thông báo | Cách xử lý |
|---|---|
| Chưa nhận biết mẫu / không tìm thấy dòng tiêu đề | Hóa đơn của nhà cung cấp mới, app không đoán. Chuyển file cho người hỗ trợ để thêm mẫu. |
| File không phải Excel / hỏng / có mật khẩu | Mở bằng Excel kiểm tra, bỏ mật khẩu rồi thả lại. File `.xls` đời 97-2003 đọc được; `.xls` đời rất cũ (5.0/95) thì Lưu thành `.xlsx`. |
| Lỗi dữ liệu (thiếu số lượng, đơn giá không đọc được) | Sửa trong hóa đơn gốc rồi thả lại. App không tự điền. |
| Không tạo được thư mục (ổ E chưa cắm) | Cắm ổ E hoặc chọn thư mục khác. App không tự lưu sang ổ C. |
| Cảnh báo "đơn giá nhỏ bất thường" | App không thấy ghi chú nghìn đồng trong hóa đơn nhưng đơn giá cao nhất dưới 1.000 đ. Kiểm tra hóa đơn gốc; nếu đúng là nghìn đồng, báo người hỗ trợ để thêm cách nhận ghi chú đó. |
| Cảnh báo đối chiếu tổng | Xem chi tiết, chỉ xuất khi đã tích xác nhận. |
| "File có N bảng hóa đơn; app chỉ lấy bảng cuối cùng" | Bình thường với phiếu xuất kho NCC gộp nhiều toa. Kiểm tra bảng cuối đúng là hóa đơn cần nhập. |
| Số lượng/đơn giá lớn vô lý, sheet kích thước bất thường | Ô bị nhập sai hoặc file hỏng; sửa trong hóa đơn gốc. |
| Cập nhật báo lỗi GitHub | Kiểm tra Internet, tên kho, mã truy cập chỉ đọc (kho riêng tư). |

App không bao giờ sửa file hóa đơn gốc hay file mẫu; nếu tên file xuất đã tồn tại, app thêm `_2`, `_3`… thay vì ghi đè. Nhật ký ở `E:\GKM_HoaDonNhapHang\logs`.

## 6. Cấu trúc thư mục
```
E:\GKM_HoaDonNhapHang
  app\        GKM_NhapHoaDon.exe, KiemTraMau.bat
  source\     mã nguồn Go (kèm vendor\, build được khi không có Internet)
  templates\  MauFileNhapHang.xlsx (mẫu đầu ra, cũng được nhúng sẵn trong .exe)
  samples\    workbook mẫu gốc (Mẫu 1/2/3, Kết quả mẫu 1/2/3, MauFileNhapHang), Mẫu 4 (Tân Hòa Lợi), Mẫu 5 (gk 28-2.xls), dinh-dang-khac\ (bản .xls/XML/HTML để thử)
  build\      build.bat (Windows), build.sh (Linux/macOS), gocache\ (cache build)
  docs\       hướng dẫn, quy tắc, kết quả kiểm thử, Goal mới của Project
  update\     latest.json + GKM_NhapHoaDon.exe (bản phát hành cho chức năng Cập nhật; dùng trong kho GitHub)
  logs\       nhật ký
  output\     file nhập hàng đã xuất (mặc định)
  config\     settings.json (thư mục xuất, kho cập nhật, mã truy cập nếu có), browser-profile\ (hồ sơ cửa sổ app)
```

## 7. Build lại từ mã nguồn (khi cần sửa)
Cài Go 1.23+ (bản cài đặt Go nằm ở nơi bạn chọn), rồi chạy `E:\GKM_HoaDonNhapHang\build\build.bat`. Script đặt toàn bộ cache build trong `build\`, chạy toàn bộ kiểm thử (lõi, giao diện máy chủ, cập nhật), build `app\GKM_NhapHoaDon.exe` và chạy lại kiểm tra bằng chính file .exe.

## 8. Dọn dẹp bản cũ (chỉ nếu đã chạy bản 1.2-1.3 có chức năng gửi Zalo/Telegram)
Bản 1.4.0 không còn chức năng gửi Zalo hay Telegram. Có thể xóa thư mục `E:\GKM_HoaDonNhapHang\config\scripts` và `config\zalo-profile` (nếu có); khối `telegram` còn trong `config\settings.json` bị bỏ qua, xóa hay giữ đều được.
