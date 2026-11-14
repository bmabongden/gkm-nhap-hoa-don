# Kết quả kiểm thử app GKM - Nhập Hóa Đơn 1.9.0 (bản chính thức, sửa 14/11/2026)

File: `app/GKM_NhapHoaDon.exe`, PE32+ GUI x86-64 build bằng Go (cross-compile từ Linux), 12,0 MB (11.980.288 byte), SHA-256 `67d7fc4fd9592073bc636e5e3d3341e5abb6f48aadc416732de6f01d28c205a8` (build lại cho ra đúng mã này).
Khiêm yêu cầu chốt app ngày 02/10/2026. Chức năng gửi Telegram (1.3.x) đã gỡ hẳn theo yêu cầu của Khiêm; app chỉ còn đọc hóa đơn, xem trước và xuất file.

## Thay đổi so với 1.0.0 (4 yêu cầu của Khiêm)
1. Đọc được hóa đơn Tân Hòa Lợi (cột `SL`, đơn giá nghìn đồng, không có cột tên) và file `gk 28-2.xls` (cột `Sl đặt hàng`, `giá`, không có STT).
2. Nút **Làm mới**, chỉ làm một hóa đơn mỗi lần.
3. Đọc cả `.xls` (Excel 97-2003), Excel XML 2003 và tệp HTML đặt đuôi `.xls`.
4. Đơn vị đơn giá tự xác định (không có ô chọn): đồng, trừ Tân Hòa Lợi (luôn nghìn đồng) hoặc hóa đơn có ghi chú rõ nghìn đồng; hóa đơn không có cột tên thì cột Tên hàng để trống (Khiêm xác nhận).
5. Đã bỏ các chức năng từng thử: tự nhập KiotViet, gửi Zalo Web, gửi Telegram (Khiêm yêu cầu bỏ).
6. (Bản 1.5.0, 04/10/2026) Dòng chỉ có số lượng vẫn nhập với giá 0 (kể cả hàng đổi/bảo hành), dòng có giá không có số lượng bị bỏ; sửa nút "Mở thư mục chứa file"; thêm dòng bản quyền nội bộ ở cuối ứng dụng; logo mới (biểu tượng .exe, cửa sổ, đầu trang).
12. (Bản 1.9.0, 14/11/2026, làm đủ 5 gợi ý theo yêu cầu của Khiêm) (1) Phiếu xuất kho có cột "ĐG ĐÃ CK 22%": ghi 22% vào Giảm giá (%), đối chiếu với giá đã CK; (2) chiết khấu cả hóa đơn chia vào % từng dòng (Mẫu 3: 5% ở cột F cho 37 dòng; Mẫu 9: dòng 10% thành 13,15%, lệch làm tròn 77 đ có cảnh báo); (3) app nhớ hóa đơn đã xuất (`config/da_xuat.json`) và cảnh báo khi xuất lại; (4) bỏ ô dán mã GitHub; (5) bản cập nhật phải có chữ ký số Ed25519 của Gia Khánh Motor (khóa công khai gắn trong app, khóa bí mật không lên GitHub). Selftest báo BỎ QUA thay vì ĐẠT khi thiếu file mẫu.
11. (Bản 1.8.3, 14/11/2026, rà soát toàn app theo yêu cầu của Khiêm) Bỏ cảnh báo sai "File đang nằm trên ổ C, mặc định là ổ E" (máy Khiêm không có ổ E, thư mục xuất nay theo file hóa đơn); cảnh báo khi hóa đơn có % giảm giá mà không kiểm chứng được đơn giá là giá gốc (tránh trừ hai lần); từ chối rõ ràng khi thả nhầm file nhập hàng đã chuyển đổi; sửa chữ và tài liệu cũ.
10. (Bản 1.8.2, 14/11/2026) Hóa đơn có giá gốc và % giảm giá: nhập giá gốc vào Đơn giá và % vào Giảm giá (%), không nhập giá đã trừ (Khiêm báo với hóa đơn KiotViet "2.xlsx"). Số hóa đơn KiotViet lấy từ tên sheet.
9. (Bản 1.8.1, 14/11/2026) Sửa lỗi Khiêm báo: đơn giá thiếu hệ số 1.000. Đọc số dạng chữ không phụ thuộc dấu . hay , (dấu đứng sau là thập phân; một loại dấu với các nhóm đúng 3 chữ số là phân cách nghìn); hóa đơn mà đơn giá cao nhất dưới 1.000 đ thì nhân 1.000 kèm cảnh báo; đọc cột "Tên hàng hóa, dịch vụ" và số phiếu ở ô "Số:" của mẫu lạ (phiếu giao hàng MISA, Mẫu 8).
8. (Bản 1.8.0, 14/11/2026) File có nhiều toa hàng (nhiều sheet, hoặc nhiều toa nối nhau trong một sheet) thì mỗi toa xuất ra một file riêng; nhận tên Hồng Nhất Phát; đọc được dòng "TỔNG SỐ TIỀN" và ô "Ngày: dd/mm/yyyy". Phiếu xuất kho (Mẫu 6) giữ quy tắc chỉ lấy bảng cuối.
7. (Bản 1.7.0, 29/10/2026) Thư mục xuất mặc định là thư mục chứa file hóa đơn (hộp chọn file Windows; kéo thả thì dò theo tên + kích thước + ngày sửa). Bản 1.6.0 (22/10/2026): Đọc được phiếu xuất kho kiêm giấy nhận nợ (SL LẤY thay SL ĐẶT, ĐƠN GIÁ BÁN thay ĐG ĐÃ CK, file gộp nhiều toa chỉ lấy bảng cuối, số lượng 0 thì bỏ dòng); thêm nút **Cập nhật** online qua kho GitHub; rà lỗi toàn app (xem "Lỗi tìm thấy khi rà soát").

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
| Toàn bộ kiểm thử Go (1.6.0) | `go test ./internal/... ./cmd/...`: 34 bài kiểm thử, 0 lỗi (lõi, máy chủ giao diện, cập nhật) | Đạt |
| Mẫu 6 phiếu xuất kho (22/10) | `go test` và `--selftest`: 35 dòng (38 dòng hàng, 3 dòng SL LẤY = 0 bị bỏ), tổng tiền hàng 12.700.802,16 và tổng SL lấy 103 khớp, "Tổng phải thu 9.907.000" hiện để tham khảo; bài thử bảng xếp chồng, mỗi toa một sheet, tiêu đề lặp trang | Đạt |
| Xáo trộn dữ liệu | 300 hóa đơn mẫu bị xóa/đổi ô ngẫu nhiên (hạt giống cố định) và 60 file hỏng byte: không panic, không treo; file xuất luôn qua `VerifyExport` | Đạt |
| Cập nhật online (giả lập) | `go test`: máy chủ GitHub giả; so sánh phiên bản; thiếu kho/kho sai; kho riêng tư cần mã (404 không mã, đạt khi có mã, mã gửi trong tiêu đề Authorization); latest.json hỏng; SHA-256 sai, kích thước sai, không phải .exe, chạy thử lỗi (đều bị từ chối và xóa file tải, file đang chạy không bị đụng); thay file thành công giữ `.old`; khôi phục khi lỗi giữa chừng; endpoint kiểm tra/cài; app tự thoát để mở lại sau khi cài; chặn Host/Origin lạ | Đạt |
| Bản 1.9.0 | `go vet`, `go test` toàn bộ (thêm kiểm thử chia chiết khấu, ghi 22%, cảnh báo đã xuất, chữ ký: thiếu, khóa lạ, sửa phiên bản sau khi ký, chữ ký hỏng đều bị từ chối, kể cả khi tải); giao diện Chromium: xuất Mẫu 9 hai lần, lần hai hiện "Hóa đơn này đã xuất rồi" và nút Xuất khóa đến khi xác nhận, bảng Cập nhật không còn ô mã; **file .exe chạy bằng Wine**: `--version` 1.9.0, `--selftest` 15 ĐẠT | Đạt (trên Windows thật, bản 1.6.0 → 1.8.3 qua nút Cập nhật đã thử trên máy Khiêm: SHA đúng, tự khởi động lại, không sót file .old/.new) |
| Rà soát 1.8.3 | `go vet`, `go test` (kể cả `-race` cho máy chủ và cập nhật): đạt; `ParseAll` trên toàn bộ 22 file mẫu và bản .xls/XML/HTML: đúng số toa, số dòng, tên file; giao diện Linux + Chromium chạy lần lượt 19 file: không lỗi trình duyệt, xuất và mở lại kiểm tra đạt; **chạy chính file .exe bằng Wine**: `--version` và `--selftest` 15/15 đạt; file .exe trên GitHub tải về có SHA-256 đúng như latest.json | Đạt (hộp chọn file Windows và nút Cập nhật trên Windows thật vẫn cần Khiêm thử) |
| Giá gốc và % giảm giá (1.8.2) | `go test`: Đơn giá + Giảm giá % + Giá bán ra giá gốc và %; giảm 100% giữ giá gốc + 100%; giảm 0% để trống; số hóa đơn từ tên sheet; `--selftest` Mẫu 9 (file Khiêm gửi): 12 dòng, 317.000 giảm 10% (trước đây nhập 285.300), 5 dòng giảm 100% ghi giá gốc, tổng 30.572.100 khớp; so từng ô Mẫu 1–3 không đổi (1/80/0 ô như trước); 15/15 mục đạt | Đạt (chưa chạy .exe trên Windows thật) |
| Đơn giá nghìn đồng, dấu . và , (1.8.1) | `go test`: 20 cách viết số (250.000, 250,000, 1.234.567,5, 1,234,567.5, 0,35, 0.350, (2.000)...) và 7 chuỗi sai bị từ chối; hóa đơn có mọi đơn giá dưới 1.000 đ được nhân 1.000 và bắt xác nhận; `--selftest` Mẫu 8 (phiếu giao hàng 13/11 Khiêm gửi): 12 dòng, 250 thành 250.000, 0,35 thành 350, tên hàng được điền, tổng 87.650 khớp, tên file NhapHang_BH01489_13112026.xlsx; 14/14 mục đạt | Đạt (chưa chạy .exe trên Windows thật) |
| File nhiều toa (1.8.0) | `go test` và `--selftest` với Mẫu 7 (file Khiêm gửi 14/11: sheet THL 17 dòng, tổng toa 12.247,5 khớp; sheet HNP Hồng Nhất Phát 7 dòng, tổng 4.606 khớp; đơn giá nhân 1.000); bài thử 2 toa nối nhau trong 1 sheet, file 1 toa không đổi, máy chủ xuất ra 2 file; giao diện Linux + Chromium: hiện 2 nút toa, "Xuất 2 file Excel", ghi 2 file vào thư mục hóa đơn và mở lại kiểm tra đạt | Đạt (chưa chạy .exe trên Windows thật) |
| Thư mục xuất theo file hóa đơn (1.7.0) | `go test`: chọn file bằng đường dẫn trả đúng thư mục và nhớ lại; từ chối đuôi lạ/đường dẫn tương đối/file không có; dò file kéo thả (trùng tên không phân biệt hoa thường, sai kích thước hoặc ngày sửa thì không thấy, hai nơi trùng thì không đoán, quá sâu bỏ qua); giao diện Linux + Chromium: chọn file ra thư mục Tải xuống, xuất file vào đó, không tìm thấy thì quay về thư mục xuất lần trước; Wine `--selftest` 11/11 | Đạt (hộp chọn file Windows thật CHƯA thử) |
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
- **Hộp chọn file của Windows (1.7.0)** chạy bằng PowerShell giống hộp chọn thư mục; mới biên dịch và thử logic bằng bản Linux giả lập, chưa bấm trên Windows thật. Nếu hộp không mở, app tự quay về cách chọn file của trình duyệt (khi đó thư mục xuất dò theo tên file).
- **Cập nhật online chưa chạy với GitHub thật**: môi trường làm việc không gọi được GitHub API (gọi thử qua proxy của môi trường trả lỗi 403), nên mới thử với máy chủ giả mô phỏng đúng giao diện `contents` của GitHub. Lần đầu dùng thật: tạo kho, đẩy `update/`, mở app bản cũ, bấm Cập nhật. Việc đổi tên file `.exe` đang chạy rồi mở lại bản mới trên Windows thật cũng chưa thử (Windows cho phép đổi tên file đang chạy; Wine không phản ánh đủ).
- Mẫu 6 chưa có kết quả mẫu để so từng ô; bảng cuối là hóa đơn ngày 20/10/2026, các toa trước trong file bị bỏ theo yêu cầu.
- Chưa chạy trên Windows thật, chỉ chạy bằng Wine. Cần chạy `app\KiemTraMau.bat` trên máy Windows.
- Khiêm đã xác nhận (02/10/2026): Tân Hòa Lợi tính nghìn đồng, hóa đơn không có cột tên thì để trống. Việc tự nhận "hóa đơn ghi chú nghìn đồng" chỉ thử với các câu mẫu trong `go test` và 8 hóa đơn mẫu hiện có (không hóa đơn nào bị nhận nhầm); hóa đơn thật có cách ghi khác sẽ không được nhận và app chỉ cảnh báo nếu đơn giá cao nhất dưới 1.000 đ.
- Chưa mở cửa sổ giao diện kiểu ứng dụng bằng Edge/Chrome (`--app`) trên Windows; chưa thử hộp chọn thư mục (PowerShell), nút "Mở thư mục chứa file" và lưu thử vào ổ E thật.
- "Mở thư mục chứa file" chỉ thử được ở mức mã chạy qua kiểm tra biên dịch cho Windows; chưa bấm thử trên Windows thật. Nếu vẫn mở sai, báo kèm đường dẫn file hiển thị trong ô kết quả.
- .exe chưa được ký số nên Windows SmartScreen có thể cảnh báo lần đầu.
