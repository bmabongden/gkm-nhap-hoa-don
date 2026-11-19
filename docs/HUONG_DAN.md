# GKM - Nhập Hóa Đơn 1.9.0: hướng dẫn ngắn

Ứng dụng Windows độc lập (portable): **không cần Python, API, khóa/token**; xử lý hóa đơn chạy hoàn toàn trên máy, chỉ nút **Cập nhật** mới dùng Internet. Mọi thứ nằm trong thư mục dự án (trong tài liệu này ghi `E:\GKM_HoaDonNhapHang` làm ví dụ; thư mục thật là nơi anh chép dự án, ví dụ `D:\Data AI\App nhap hang\Chuan\GKM_HoaDonNhapHang`).

## 1. Cài đặt (một lần)
1. Chép nguyên thư mục `GKM_HoaDonNhapHang` (giải nén từ file zip) vào ổ/thư mục bạn muốn, ví dụ `E:\GKM_HoaDonNhapHang` hoặc `D:\Data AI\App nhap hang\Chuan\GKM_HoaDonNhapHang`. Từ bản 1.6.0 app tự dùng thư mục chứa `app\` làm thư mục dự án (không cần ổ E); giữ nguyên tên thư mục con `app`.
2. Chạy thử kiểm tra: bấm đúp `E:\GKM_HoaDonNhapHang\app\KiemTraMau.bat`. Kết quả phải là **ĐẠT** (3 mẫu gốc, các bản .xls của chúng, mẫu Tân Hòa Lợi, file gk 28-2.xls và phiếu xuất kho SL lấy).
3. Windows có thể hiện cảnh báo SmartScreen ("Windows protected your PC") vì file .exe chưa được ký số. Chọn **More info → Run anyway**. Chỉ làm khi file đến từ nguồn bạn tin cậy.

## 2. Dùng hằng ngày
1. Bấm đúp `E:\GKM_HoaDonNhapHang\app\GKM_NhapHoaDon.exe`. Một cửa sổ "GKM - Nhập Hóa Đơn" mở ra (dùng Microsoft Edge hoặc Chrome có sẵn trên máy; dữ liệu hóa đơn chỉ chạy trên máy bạn).
2. **Bước 1**: bấm vào khung để chọn **1** file hóa đơn Excel (`.xlsx` hoặc `.xls`) bằng hộp thoại của Windows, hoặc kéo thả file vào khung. Mỗi lần chỉ làm một hóa đơn.
3. **Bước 2 - Xem trước**: kiểm tra số dòng, hàng thường/khuyến mại, mục đối chiếu với tổng hóa đơn, cảnh báo và bảng dòng hàng. Dòng khuyến mại tô vàng. Hàng đổi/bảo hành (chỉ có số lượng) vẫn nhập với đơn giá 0 và có nhãn riêng; dòng có giá mà không có số lượng bị bỏ, nằm trong mục xổ xuống bên dưới. App **tự xác định đơn vị đơn giá**: mặc định là đồng; hóa đơn Tân Hòa Lợi và hóa đơn có ghi chú tính theo nghìn đồng (ví dụ "Đơn vị tính: nghìn đồng", "Đơn giá (1000đ)") thì app tự nhân 1.000 và ghi rõ lý do trên màn hình. Hóa đơn không có cột tên hàng thì cột Tên hàng để trống. Từ 1.8.1: số ghi dạng chữ đọc đúng dù phần mềm nguồn dùng dấu chấm hay dấu phẩy (250.000 và 250,000 đều là 250 nghìn; 0,35 và 0.35 là số lẻ); hóa đơn mà mọi đơn giá đều dưới 1.000 đ thì app hiểu là nghìn đồng, tự nhân 1.000 và cảnh báo để anh xem lại.
4. **Bước 3 - Xuất**: thư mục lưu mặc định là **thư mục chứa file hóa đơn vừa chọn** (đổi bằng ô đường dẫn hoặc nút **Chọn thư mục…**). Chọn file bằng hộp thoại thì app biết chắc thư mục. Kéo thả thì trình duyệt không cho biết file nằm ở đâu, nên app tự tìm file theo tên, kích thước và ngày sửa trong Tải xuống, Desktop, Documents, OneDrive, Zalo Received Files và thư mục dự án; tìm được thì dùng thư mục đó (có ghi chú "app tìm thấy file theo tên và kích thước"), không chắc thì dùng thư mục xuất lần trước. Muốn chắc chắn thì chọn file bằng hộp thoại. Nếu có cảnh báo, đọc rồi tích ô xác nhận. Bấm **Xuất file Excel**. File tên `NhapHang_<số phiếu>_<ngày>.xlsx`. App tự mở lại file vừa xuất để kiểm tra rồi mới báo đạt.
5. **Làm mới**: xong hóa đơn này thì bấm nút **Làm mới** (góc trên bên phải, cạnh nút **Cập nhật**) để xóa hóa đơn cũ, rồi chọn hóa đơn khác. Chưa bấm Làm mới thì app không nhận hóa đơn thứ hai. File đã xuất không bị xóa.
6. Đóng cửa sổ là app tự thoát.

### File có nhiều toa hàng (từ 1.8.0)
Một file Excel gửi nhiều toa (mỗi toa một sheet, hoặc nhiều toa xếp nối nhau trong một sheet, mỗi toa có dòng tổng riêng) thì app **tách ra mỗi toa một file**. Bước 2 hiện một nút cho từng toa (ví dụ "Toa 1 · sheet THL · 17 dòng", "Toa 2 · sheet HNP · 7 dòng"); bấm để xem trước và đối chiếu từng toa. Bước 3 có nút **Xuất N file Excel**, xuất tất cả toa vào cùng thư mục, mỗi toa một tên, ví dụ `NhapHang_TanHoaLoi_11112026.xlsx` và `NhapHang_HongNhatPhat_11112026.xlsx` (hai toa trùng tên thì thêm `_toa1`, `_toa2`). Toa nào có lỗi dữ liệu thì chưa xuất toa nào; cảnh báo ở bất kỳ toa nào thì tích ô xác nhận một lần cho cả file. Riêng phiếu xuất kho dưới đây vẫn chỉ lấy bảng cuối.

### Phiếu xuất kho kiêm giấy nhận nợ (cột "SL LẤY", "ĐƠN GIÁ BÁN")
Nhà cung cấp này hay gửi gộp cả chồng toa hàng trước đó trong một file. App **chỉ lấy bảng hóa đơn cuối cùng** (và ghi một dòng "File có N bảng hóa đơn" trong phần đối chiếu để bạn biết). Số lượng lấy từ cột **SL LẤY** (không phải SL ĐẶT), đơn giá lấy từ **ĐƠN GIÁ BÁN** (không phải ĐG ĐÃ CK). Dòng có SL LẤY bằng 0 hoặc trống bị bỏ như mọi dòng không có số lượng. Dòng "TỔNG PHẢI THU" (sau chiết khấu) chỉ hiện để tham khảo, không đưa vào file.

## 3. Quy tắc đang áp dụng
- Hóa đơn có giá gốc và % giảm giá (ví dụ KiotViet: Đơn giá, Giảm giá %, Giá bán): cột Đơn giá ghi **giá gốc**, cột Giảm giá (%) ghi **% giảm** như hóa đơn, không nhập thẳng giá đã trừ (từ 1.8.2). Dòng giảm 100% vẫn ghi giá gốc và 100%. Dòng giảm 0% để trống cột giảm giá.
- Mã hàng giữ nguyên như hóa đơn (kể cả số 0 đầu), không có quy tắc riêng cho hàng "ốc".
- Điền tên hàng. Không giảm giá thì để trống. Dòng chỉ có số lượng (hàng đổi, bảo hành, khuyến mại) ghi đơn giá 0. Dòng có giá mà không có số lượng thì bỏ.
- Chiết khấu cả hóa đơn (ví dụ "Tỷ lệ CK 5%" hay dòng "Giảm giá 1.070.100" dưới tổng) được chia đều vào cột Giảm giá (%) của từng dòng (từ 1.9.0). Dòng đã có % riêng thì gộp hai mức: 10% rồi 3,5% thành 13,15%. Chiết khấu ghi bằng số tiền được đổi ra %, làm tròn 2 chữ số; lệch vài đồng thì app báo số đồng lệch.
- Phiếu xuất kho có cột "ĐG ĐÃ CK 22%" (Mẫu 6): ghi 22% vào cột Giảm giá (%) của từng dòng (từ 1.9.0); app đối chiếu với cột giá đã chiết khấu, lệch thì cảnh báo.
- Thuế suất không đưa vào file (mẫu nhập hàng không có chỗ).
- Chi tiết: `docs\GKM_Quy_Tac_Anh_Xa.md`.

## 4. Cập nhật app qua Internet
1. Bấm nút **Cập nhật** ở góc trên. Ô tên kho GitHub có sẵn `bmabongden/gkm-nhap-hoa-don` (kho công khai, không cần mã truy cập; từ 1.9.0 đã bỏ ô dán mã GitHub).
2. Bấm **Kiểm tra bản mới**. Nếu có, bấm **Cập nhật ngay**: app tải file mới, kiểm tra **chữ ký số của Gia Khánh Motor** (từ 1.9.0; bản không có chữ ký hoặc chữ ký sai bị từ chối, kể cả khi ai đó sửa được kho GitHub), kích thước, mã SHA-256, dạng chương trình Windows và chạy thử `-version`; đạt hết mới thay file `.exe` (bản cũ giữ thành `GKM_NhapHoaDon.exe.old`, tự xóa ở lần mở sau) rồi tự khởi động lại. Hóa đơn đang mở mà chưa xuất sẽ bị bỏ khi khởi động lại, app có hỏi trước.
3. Nếu cập nhật lỗi, app giữ nguyên bản đang chạy; muốn quay về bản cũ thì đóng app, xóa `GKM_NhapHoaDon.exe` và đổi tên `GKM_NhapHoaDon.exe.old` về `GKM_NhapHoaDon.exe`.
4. Phía người phát hành: từ 19/11/2026 việc này TỰ ĐỘNG — gộp PR có nâng phiên bản vào `main` là GitHub Actions tự build, ký (khóa lưu trong GitHub Secrets tên `GKM_SIGN_KEY`, theo lựa chọn của chủ app ngày 19/11/2026) và đưa `update/` lên kho. Cách tay vẫn dùng được: sau `build\build.sh`, chạy `bash build/publish_update.sh "ghi chú"` để tạo `update\latest.json` và `update\GKM_NhapHoaDon.exe`, rồi commit và push hai file đó lên nhánh chính của kho. Lệnh này ký bản cập nhật bằng file khóa bí mật `gkm_update_ed25519.key` (biến `GKM_SIGN_KEY`, mặc định nằm trong thư mục dự án chung `GKM/khoa_ky_cap_nhat/`). **Không bao giờ đưa file khóa lên GitHub**; sao lưu nó ra USB. Mất khóa thì phải tạo khóa mới và cài tay bản mới trên từng máy. Vẫn nên bật đăng nhập hai bước cho tài khoản GitHub.

## 5. Khi app báo lỗi
| Thông báo | Cách xử lý |
|---|---|
| Chưa nhận biết mẫu / không tìm thấy dòng tiêu đề | Hóa đơn của nhà cung cấp mới, app không đoán. Chuyển file cho người hỗ trợ để thêm mẫu. |
| File không phải Excel / hỏng / có mật khẩu | Mở bằng Excel kiểm tra, bỏ mật khẩu rồi thả lại. File `.xls` đời 97-2003 đọc được; `.xls` đời rất cũ (5.0/95) thì Lưu thành `.xlsx`. |
| Lỗi dữ liệu (thiếu số lượng, đơn giá không đọc được) | Sửa trong hóa đơn gốc rồi thả lại. App không tự điền. |
| Không tạo được thư mục | Ổ hoặc thư mục đó không còn; bấm **Chọn thư mục…** để chọn nơi khác. |
| File này là file nhập hàng đã chuyển đổi | Anh thả nhầm file `NhapHang_...xlsx` app đã xuất. Bấm Làm mới rồi chọn file hóa đơn gốc. |
| Hóa đơn này đã xuất rồi | App nhớ các hóa đơn đã xuất (`config\da_xuat.json`). Thả lại cùng hóa đơn thì báo lúc nào, ra file nào. Nếu đúng là xuất lại có chủ ý thì tích ô xác nhận rồi xuất; không thì bấm Làm mới. |
| Cập nhật báo "không có chữ ký số hợp lệ" | Bản trên GitHub không do Gia Khánh Motor ký. Đừng cài; báo người hỗ trợ kiểm tra tài khoản GitHub. |
| Cảnh báo "không kiểm chứng được đơn giá là giá gốc" | Hóa đơn có cột % giảm giá nhưng không có cột giá sau giảm và thành tiền không khớp. Xem hóa đơn gốc: nếu đơn giá đã trừ giảm giá rồi thì đừng xuất (sẽ bị trừ hai lần), báo người hỗ trợ. |
| Cảnh báo "app hiểu hóa đơn ghi theo NGHÌN ĐỒNG" | Mọi đơn giá trên hóa đơn đều dưới 1.000 đ nên app đã nhân 1.000. Xem lại bảng đơn giá; đúng thì tích xác nhận rồi xuất. |
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
