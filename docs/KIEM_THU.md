# Kết quả kiểm thử app GKM - Nhập Hóa Đơn 1.6.0 (bản chính thức, sửa 22/10/2026)

File: `app/GKM_NhapHoaDon.exe`, PE32+ GUI x86-64 build bằng Go (cross-compile từ Linux), 11,9 MB (11.898.880 byte), SHA-256 `63f18e1e889095c70cc8cd2cdc856948002886986f74fe35022802ce15cb50fc` (build lại cho ra đúng mã này).
Khiêm yêu cầu chốt app ngày 02/10/2026. Chức năng gửi Telegram (1.3.x) đã gỡ hẳn theo yêu cầu của Khiêm; app chỉ còn đọc hóa đơn, xem trước và xuất file.

## Thay đổi so với 1.0.0 (4 yêu cầu của Khiêm)
1. Đọc được hóa đơn Tân Hòa Lợi (cột `SL`, đơn giá nghìn đồng, không có cột tên) và file `gk 28-2.xls` (cột `Sl đặt hàng`, `giá`, không có STT).
2. Nút **Làm mới**, chỉ làm một hóa đơn mỗi lần.
3. Đọc cả `.xls` (Excel 97-2003), Excel XML 2003 và tệp HTML đặt đuôi `.xls`.
4. Đơn vị đơn giá tự xác định (không có ô chọn): đồng, trừ Tân Hòa Lợi (luôn nghìn đồng) hoặc hóa đơn có ghi chú rõ nghìn đồng; hóa đơn không có cột tên thì cột Tên hàng để trống (Khiêm xác nhận).
5. Đã bỏ các chức năng từng thử: tự nhập KiotViet, gửi Zalo Web, gửi Telegram (Khiêm yêu cầu bỏ).
6. (Bản 1.5.0, 04/10/2026) Dòng chỉ có số lượng vẫn nhập với giá 0 (kể cả hàng đổi/bảo hành), dòng có giá không có số lượng bị bỏ; sửa nút "Mở thư mục chứa file"; thêm dòng bản quyền nội bộ ở cuối ứng dụng; logo mới (biểu tượng .exe, cửa sổ, đầu trang).
7. (Bản 1.6.0, 22/10/2026) Đọc được phiếu xuất kho kiêm giấy nhận nợ (SL LẤY thay SL ĐẶT, ĐƠN GIÁ BÁN thay ĐG ĐÃ CK, file gộp nhiều toa chỉ lấy bảng cuối, số lượng 0 thì bỏ dòng); thêm nút **Cập nhật** online qua kho GitHub; rà lỗi toàn app (xem "Lỗi tìm thấy khi rà soát").

## Lỗi tìm thấy khi rà soát (22/10/2026) và đã sửa
Chạy thử xáo trộn dữ liệu (xóa ô, thay bằng chữ/số lạ trong các hóa đơn mẫu, làm hỏng byte file .xls/.xlsx) và đọc lại mã tìm ra:
1. **Treo app** khi file hỏng khai ô ở dòng/cột cực xa: bước dò ghi chú nghìn đồng quét hàng tỷ ô. Sửa: chỉ duyệt ô có dữ liệu; sheet quá 50.000 dòng hoặc 200 cột bị từ chối.
2. **Tên kho kiểu `../x`** lọt qua bộ kiểm tra tên kho cập nhật (có thể đổi đường dẫn gọi GitHub). Sửa: chặn phần tên bắt đầu bằng dấu chấm.
3. Số lượng/đơn giá vô lý (ví dụ 1e21) làm file xuất sai số rồi mới bị chặn ở bước kiểm tra cuối. Sửa: báo lỗi ngay từ bước đọc (số lượng ≥ 1 tỷ, đơn giá ≥ 1.000 tỷ), cảnh báo số lượng > 100.000.
4. Ô chứa ký tự điều khiển (NUL...) và chữ `Inf`/`NaN` trong ô số làm file xuất không giống dữ liệu đọc. Sửa: bỏ ký tự điều khiển khi ghi tên/mã, từ chối Inf/NaN.
5. Lưu cấu hình từng ghi đè cả file `settings.json` (sẽ làm mất kho/mã cập nhật khi đổi thư mục xuất). Sửa: chỉ đổi trường được gửi lên, mã truy cập không bao giờ trả lại giao diện.
6. Lỗi bất ngờ trong một yêu cầu giờ trả thông báo tiếng Việt thay vì cắt kết nối; file tạm khi tải lên được xóa; các hàm chọn thư mục/mở thư mục chỉ nhận POST.
7. Thư mục dự án mặc định E:\ làm app báo "không tìm thấy thư mục dự án" trên máy không có ổ E (máy của Khiêm chỉ có ổ D). Sửa: dùng thư mục cha của `app\` chứa file .exe; có bài kiểm thử `TestRootFromExe`.
Không dùng ChatGPT: phiên làm việc này không có đường gọi dịch vụ đó nên việc rà soát do Claude tự làm; có thể đưa zip mã nguồn cho ChatGPT đọc thêm và gửi kết quả để xử lý.

## Đã kiểm chứng thực tế
| Hạng mục | Cách làm | Kết quả |
|---|---|---|
| Toàn bộ kiểm thử Go (1.6.0) | `go test ./internal/... ./cmd/...`: 30 bài kiểm thử, 0 lỗi (lõi, máy chủ giao diện, cập nhật) | Đạt |
| Mẫu 6 phiếu xuất kho (22/10) | `go test` và `--selftest`: 35 dòng (38 dòng hàng, 3 dòng SL LẤY = 0 bị bỏ), tổng tiền hàng 12.700.802,16 và tổng SL lấy 103 khớp, "Tổng phải thu 9.907.000" hiện để tham khảo; bài thử bảng xếp chồng, mỗi toa một sheet, tiêu đề lặp trang | Đạt |
| Xáo trộn dữ liệu | 300 hóa đơn mẫu bị xóa/đổi ô ngẫu nhiên (hạt giống cố định) và 60 file hỏng byte: không panic, không treo; file xuất luôn qua `VerifyExport` | Đạt |
| Cập nhật online (giả lập) | `go test`: máy chủ GitHub giả; so sánh phiên bản; thiếu kho/kho sai; kho riêng tư cần mã (404 không mã, đạt khi có mã, mã gửi trong tiêu đề Authorization); latest.json hỏng; SHA-256 sai, kích thước sai, không phải .exe, chạy thử lỗi (đều bị từ chối và xóa file tải, file đang chạy không bị đụng); thay file thành công giữ `.old`; khôi phục khi lỗi giữa chừng; endpoint kiểm tra/cài; app tự thoát để mở lại sau khi cài; chặn Host/Origin lạ | Đạt |
| Giao diện 1.6.0 | App bản Linux + Chromium: thả Mẫu 6 ra 35 dòng, xuất đạt; panel Cập nhật: báo thiếu kho, từ chối tên kho sai, lưu kho + mã (mã không hiện lại), báo lỗi GitHub dễ hiểu; không lỗi trình duyệt | Đạt |
| Chạy chính file .exe (1.6.0) | Wine: `--version` ra 1.6.0; `--selftest` đạt 11/11 mục khi gồm Mẫu 6 | Đạt |
| Lõi chuyển đổi | `go test ./internal/core/`: 3 mẫu gốc so từng ô, Mẫu 4, Mẫu 5, bản .xls/XML/HTML, file .xls 3000 dòng (qua nhiều đoạn CONTINUE) | Đạt |
| Chạy chính file .exe (1.5.0) | `GKM_NhapHoaDon.exe --selftest` bằng Wine 9.0 | Đạt 10/10 mục (xem dưới) |
| Giao diện | App bản Linux + Chromium (Playwright) với bản 1.4.0: Mẫu 5 (.xls), Làm mới, Mẫu 4, Làm mới, Mẫu 1, mỗi lần xuất xong đạt (281, 27, 44 dòng), không có ô chọn đơn vị giá/tên hàng, không còn chữ Telegram, không lỗi trình duyệt và không lỗi HTTP | Đạt |
| Quy tắc số lượng/giá (1.5.0) | `go test`: bảng thử có dòng thường, dòng chỉ có số lượng (vào file giá 0), dòng có giá không số lượng và dòng trống (bỏ), bảng mà mọi dòng thiếu số lượng (báo lỗi); Mẫu 3 có 9 hàng đổi vào file | Đạt |
| Giao diện 1.5.0 | App bản Linux + Chromium: Mẫu 3 hiện 64 dòng, bộ lọc "Hàng đổi" ra 9 dòng, xuất đạt 64 dòng, có logo và dòng bản quyền ở cuối, không lỗi trình duyệt/HTTP | Đạt |
| Biểu tượng .exe | Build có tài nguyên biểu tượng (16-256 px) và thông tin phiên bản 1.5.0; chạy `-version` và `--selftest` bằng Wine bình thường | Đạt (chưa nhìn biểu tượng trên Explorer/thanh tác vụ Windows thật) |
| Tự nhận đơn vị đơn giá | `go test`: các câu ghi chú nghìn đồng được nhận, số tiền bằng chữ ("... nghìn đồng chẵn") và câu thường không bị nhận nhầm; 8 hóa đơn mẫu không bị nhận nhầm | Đạt |

### `--selftest` (Wine)
| Mục | Kết quả |
|---|---|
| Mẫu 1 (xlsx và bản .xls) | 44 dòng, 1 ô khác kết quả cũ (A44 `PTXD0164`), 0 khác kiểu |
| Mẫu 2 (xlsx và bản .xls) | 80 dòng, 80 ô khác (cột Tên hàng nay được điền), 0 khác kiểu |
| Mẫu 3 (xlsx, bản .xls, bản XML 2003) | 64 dòng (37 thường, 18 khuyến mại, 9 hàng đổi đơn giá 0): 55 dòng đầu khớp kết quả mẫu từng ô (0 ô khác), 9 dòng hàng đổi có số lượng dương và đơn giá 0 |
| Mẫu 4 Tân Hòa Lợi (xlsx và bản .xls) | 27 dòng, đối chiếu tổng khớp (20.560), đơn giá dòng 1 = 7.500 |
| Mẫu 5 `gk 28-2.xls` (file .xls do Excel tạo) | 281 dòng, thành tiền dòng = SL × giá khớp toàn bộ, tổng 113.412.588 khớp |
| Mẫu 6 phiếu xuất kho (xlsx) | 35 dòng, tổng tiền hàng và tổng SL lấy khớp |

Giới hạn của phép thử: Mẫu 4 và 5 **chưa có kết quả mẫu** để so từng ô nên chỉ kiểm tra "đọc đúng, tổng khớp, xuất file đạt". Các bản .xls/XML/HTML của Mẫu 1-4 do công cụ Python (xlwt) tạo từ file gốc, không phải do Excel lưu; riêng `gk 28-2.xls` là file Excel thật. LibreOffice trong môi trường này không mở được bảng tính nên không dùng để tạo bản .xls.

## CHƯA kiểm chứng (cần thử trên máy Windows thật)
- **Cập nhật online chưa chạy với GitHub thật**: môi trường làm việc không gọi được GitHub API (gọi thử qua proxy của môi trường trả lỗi 403), nên mới thử với máy chủ giả mô phỏng đúng giao diện `contents` của GitHub. Lần đầu dùng thật: tạo kho, đẩy `update/`, mở app bản cũ, bấm Cập nhật. Việc đổi tên file `.exe` đang chạy rồi mở lại bản mới trên Windows thật cũng chưa thử (Windows cho phép đổi tên file đang chạy; Wine không phản ánh đủ).
- Mẫu 6 chưa có kết quả mẫu để so từng ô; bảng cuối là hóa đơn ngày 20/10/2026, các toa trước trong file bị bỏ theo yêu cầu.
- Chưa chạy trên Windows thật, chỉ chạy bằng Wine. Cần chạy `app\KiemTraMau.bat` trên máy Windows.
- Khiêm đã xác nhận (02/10/2026): Tân Hòa Lợi tính nghìn đồng, hóa đơn không có cột tên thì để trống. Việc tự nhận "hóa đơn ghi chú nghìn đồng" chỉ thử với các câu mẫu trong `go test` và 8 hóa đơn mẫu hiện có (không hóa đơn nào bị nhận nhầm); hóa đơn thật có cách ghi khác sẽ không được nhận và app chỉ cảnh báo nếu đơn giá cao nhất dưới 1.000 đ.
- Chưa mở cửa sổ giao diện kiểu ứng dụng bằng Edge/Chrome (`--app`) trên Windows; chưa thử hộp chọn thư mục (PowerShell), nút "Mở thư mục chứa file" và lưu thử vào ổ E thật.
- "Mở thư mục chứa file" chỉ thử được ở mức mã chạy qua kiểm tra biên dịch cho Windows; chưa bấm thử trên Windows thật. Nếu vẫn mở sai, báo kèm đường dẫn file hiển thị trong ô kết quả.
- .exe chưa được ký số nên Windows SmartScreen có thể cảnh báo lần đầu.
