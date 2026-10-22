# GKM – Quy tắc ánh xạ hóa đơn nhập hàng → file nhập hàng

Cập nhật: 02/10/2026 (đã áp dụng xác nhận của Khiêm cùng ngày, xem mục 4 và 5). Nguồn bằng chứng: 3 cặp `Mau N.xlsx` → `Ket qua mau N.xlsx`, đích là `MauFileNhapHang.xlsx`.
Mọi quy tắc dưới đây đã chạy thử và đối chiếu từng ô (xem mục 6). Chỗ nào chưa đủ bằng chứng được ghi ở mục 5, không tự đoán.

## 1. Vị trí lưu (thư mục dùng chung của Project, đọc được từ mọi thread)

| Nội dung | Đường dẫn |
|---|---|
| 7 workbook gốc | `/mnt/project-files/GKM/mau/` (Mau 1/2/3, Ket qua mau 1/2/3, MauFileNhapHang) |
| Công cụ chuyển đổi (Python + openpyxl, không API, không .exe) | `/mnt/project-files/GKM/gkm_convert.py` |
| Tài liệu này | `/mnt/project-files/GKM/GKM_Quy_Tac_Anh_Xa.md` |
| File thử nghiệm đã tạo | `/mnt/project-files/GKM/thu_nghiem/` |

Ổ E chỉ áp dụng khi lưu trên máy người dùng. Môi trường cloud không có ổ E nên mặc định lưu trong thư mục Project. Khi cần lưu local, hỏi/đặt đường dẫn bắt đầu bằng `E:\`.

## 2. File đích (MauFileNhapHang.xlsx)

Một sheet `PurchaseOrderTemplate`, dòng 1 là tiêu đề, dữ liệu từ dòng 2:

| Cột | Tên | Ghi chú |
|---|---|---|
| A | Mã hàng | Văn bản, giữ số 0 đầu |
| B | Tên hàng | |
| C | Đơn vị tính | Cả 3 kết quả mẫu đều để trống |
| D | Đơn giá | Số |
| E | Giảm giá | Cả 3 kết quả mẫu đều để trống |
| F | Giảm giá (%) | Chỉ Mẫu 1 có |
| G | Số lượng | |

Không có công thức. Không thêm cột/sheet. Quy trình: sao chép workbook mẫu rồi điền từ dòng 2 (dòng 2-4 giữ định dạng số của mẫu, các dòng sau để General, giống kết quả mẫu).

## 3. Nhận dạng nhà cung cấp (theo cấu trúc, không theo tên file)

Tìm dòng tiêu đề có `STT` + `Mã hàng`/`Mã vật tư`, lấy cột theo tên cột (không cố định số cột). Hóa đơn lặp tiêu đề ở trang 2 thì bỏ qua dòng lặp.

| Mẫu | Nhận dạng | Nhà cung cấp |
|---|---|---|
| A (Mẫu 1) | Tiêu đề có `CK(%)` và `Giá NET`; chân trang AMIS KẾ TOÁN | Công ty TNHH Đoàn Lâm Khoa, hóa đơn bán hàng PBH… |
| B (Mẫu 2) | Tiêu đề `Mã vật tư`, `Tên vật tư`, `DVT`; "Biên bản giao hàng kiêm giấy nhận nợ" | Công ty TNHH Bình Minh |
| C (Mẫu 3) | Tiêu đề `Đơn vị`, `Thành tiền` (không có CK%); "Phiếu xuất hàng"; chân trang MISA | Công ty TNHH TM SX & DV Quốc Quang – Việt Đài |

Mẫu không khớp một trong ba → dừng và hỏi người dùng, không áp quy tắc của mẫu khác.

## 4. Quy tắc ánh xạ (có bằng chứng)

Chung cho cả 3 mẫu:
- Mã hàng = cột mã của hóa đơn, giữ dạng văn bản (kể cả mã 13 chữ số như `8935039037289`).
- Đơn vị tính (C) và Giảm giá (E) để trống.
- Đơn giá (D) = đơn giá gốc trên hóa đơn (không phải Giá NET, không phải thành tiền). Chuỗi `492,000` đổi thành số 492000. Giá lẻ của nguồn (vd 21032.999999999996) giữ nguyên.
- Số lượng (G) = số lượng hóa đơn, giữ nguyên kiểu nguồn (Mẫu 3 nguồn là văn bản "1" nên kết quả cũng là văn bản).
- Cột STT, Thành tiền, thuế suất, thùng số, công nợ, tổng và các dòng trang/chân trang không đưa vào danh sách; chỉ dùng để đối chiếu.
- **Mã hàng giữ nguyên như hóa đơn, kể cả mặt hàng "Ốc"** (Khiêm chốt 02/10/2026: bỏ quy tắc ốc). Kết quả mẫu 1 ghi mã `o` cho "Ốc 6 dài" nhưng không áp dụng.
- Dòng chỉ có số lượng, không có giá (hàng đổi, hàng bảo hành, hàng khuyến mại) vẫn đưa vào file với đơn giá 0 (Khiêm 04/10/2026, thay quy tắc cũ "hàng đổi không nhập"). Dòng có giá mà không có số lượng thì bỏ, không đưa vào file. Giảm giá (%) không có thì để ô trống.
- Không gộp dòng trùng mã. Mẫu 3 có cùng mã xuất hiện ở dòng giá thường và dòng khuyến mại, kết quả mẫu giữ riêng cả hai dòng.

Theo từng mẫu:

| Quy tắc | Mẫu A (Đoàn Lâm Khoa) | Mẫu B (Bình Minh) | Mẫu C (Quốc Quang – Việt Đài) |
|---|---|---|---|
| Tên hàng (B) | Lấy từ hóa đơn | Lấy từ hóa đơn (Khiêm xác nhận 02/10: điền tên cho mọi mẫu; kết quả mẫu 2 để trống tên nhưng đã bị thay thế) | Lấy từ hóa đơn |
| Giảm giá (%) (F) | Lấy cột `CK(%)`, giữ nguyên văn bản như `2,00`, `1,49`; trống nếu hóa đơn trống | Trống | Trống (tỷ lệ CK 5% ở cấp hóa đơn **không** đưa vào dòng) |
| Hàng khuyến mại | Dòng có đơn giá trống: giữ dòng, Đơn giá = 0, SL giữ nguyên (mẫu có 1 dòng) | Không có | Dòng ghi `Khuyến mại` ở cột đơn giá và các dòng kế tiếp không có giá: giữ dòng, Đơn giá = 0 (mẫu có 18 dòng) |
| Hàng đổi | Không | Không | Dòng từ chỗ ghi `Hàng dổi` đến hết bảng (9 dòng, chỉ có số lượng): từ 04/10/2026 **có** đưa vào file với đơn giá 0 (trước đó Khiêm quy định không nhập) |
| Tổng đối chiếu | Cộng tiền hàng, Tiền chiết khấu, Tổng thanh toán | Tổng cộng (SL, thành tiền) | Tổng tiền hàng |

Tên file xuất: `NhapHang_<tenhoadon>_<ddMMyyyy>.xlsx`. `<tenhoadon>` là số phiếu (ký tự đặc biệt đổi thành `-`), nếu hóa đơn không có số thì dùng tên nhà cung cấp không dấu. Ví dụ `NhapHang_PBH002331_23072026.xlsx`, `NhapHang_BinhMinh_25072026.xlsx`, `NhapHang_08-26-03_01082026.xlsx`.

## 5. Điểm mơ hồ / chưa đủ bằng chứng

Đã xác nhận ngày 02/10/2026 (Khiêm; riêng hàng đổi đã đổi ngày 04/10/2026: nhập với đơn giá 0): điền Tên hàng cho mọi mẫu; Giảm giá (%) không có thì để trống; mã hàng giữ nguyên, không có quy tắc đặc biệt cho hàng ốc (từng bàn rồi bỏ).

Còn mở:
1. **Chiết khấu cấp hóa đơn (mẫu C: 5%)** và **thuế suất (mẫu B: 8%/10%)** không có chỗ trong file đích và kết quả mẫu bỏ qua. Công cụ bỏ qua tương tự và chỉ hiện trong báo cáo.
2. **Kiểu dữ liệu**: kết quả mẫu để `Giảm giá (%)` dạng văn bản (`2,00`) và Số lượng mẫu 3 dạng văn bản. Đang giữ đúng như vậy. Nếu phần mềm nhập hàng cần số thuần thì cần xác nhận để đổi.
3. **Mẫu 1 "Tiền chiết khấu"**: hóa đơn ghi 3.491.000 nhưng tính từ dòng ra 3.490.691 (lệch 309đ do hóa đơn làm tròn). Chỉ cảnh báo, không ảnh hưởng file.
4. **Đơn vị tính** luôn để trống ở cả 3 kết quả dù nguồn có đơn vị. Đang bỏ trống theo mẫu.
5. Chỉ có 3 nhà cung cấp làm mẫu. Nhà cung cấp mới phải được người dùng xác nhận trước khi áp quy tắc.

## 6. Kết quả chạy thử (so từng ô A1:G với `Ket qua mau N.xlsx`)

| Mẫu | Dòng ra / mẫu | Hàng thường / KM / loại | Ô khác giá trị | Ghi chú |
|---|---|---|---|---|
| 1 | 44 / 44 | 43 / 1 / 0 | **1** | A44: mã `PTXD0164` (giữ nguyên hóa đơn) so với `o` trong kết quả mẫu. Khác do chủ ý, Khiêm bỏ quy tắc ốc |
| 2 | 80 / 80 | 80 / 0 / 0 | **80** | 80 ô Tên hàng: kết quả mẫu để trống, theo xác nhận nay điền tên. Mã khớp hoàn toàn; 11 ô đơn giá chỉ khác sai số dấu phẩy động ≤ 1,2e-10 |
| 3 | 64 / 55 | 37 / 18 / 9 hàng đổi giá 0 | **0** (so 55 dòng đầu) | Kết quả mẫu 3 là bản cũ (hàng đổi không nhập). Từ 04/10/2026 thêm 9 dòng hàng đổi đơn giá 0 ở cuối file, 55 dòng đầu vẫn khớp từng ô |

Đối chiếu tổng: Mẫu 1 Cộng tiền hàng 71.299.000 khớp; mọi dòng Thành tiền = SL × Đơn giá; Mẫu 2 tổng tiền 47.132.326,5 và SL 439 khớp; Mẫu 3 tổng tiền hàng 46.871.000 khớp (37 dòng hàng thường).

## 7. Hướng dẫn dùng cho mỗi hóa đơn mới

Từ 02/10/2026 người dùng xử lý hóa đơn bằng app `GKM_NhapHoaDon.exe` (mục 8). Quy trình dưới đây là cách Claude xử lý khi người dùng gửi hóa đơn trong chat:
1. Chạy `python3 /mnt/project-files/GKM/gkm_convert.py "<hóa đơn>.xlsx" --out <thư mục>`; nếu cần thì cài `openpyxl` bằng `python3 -m pip install openpyxl`.
2. Công cụ nhận dạng mẫu, chuyển đổi, mở lại file xuất để kiểm tra tiêu đề và số dòng, rồi in báo cáo: số dòng, hàng thường, hàng khuyến mại, hàng đổi (giá 0), dòng bỏ vì không có số lượng, kết quả đối chiếu tổng.
3. Claude trả lại gọn trong chat: tên hóa đơn, số dòng, hàng thường, hàng khuyến mại, cảnh báo, và file Excel qua phần đính kèm. File bản chính nằm trong thư mục Project.
4. Nếu mẫu nhà cung cấp mới, thiếu dữ liệu bắt buộc (mã, số lượng, đơn giá không đọc được) hoặc tổng không khớp: giữ bản nháp ghi rõ "chưa đủ điều kiện nhập" và hỏi một câu gộp mọi điểm cần xác nhận. Không tự đoán mã hàng, giá, VAT, chiết khấu.
5. Không sửa các file trong `mau/`. Không gửi hóa đơn ra dịch vụ ngoài, không chia sẻ Project công khai.

## 8. App Windows GKM_NhapHoaDon.exe (thay thế yêu cầu "không build .exe", 02/10/2026)

Từ 02/10/2026 Khiêm yêu cầu app Windows độc lập để xử lý hóa đơn hằng ngày (không dùng Artifact/chat). Mã nguồn Go, bản build, hướng dẫn và kết quả kiểm thử nằm ở `GKM/GKM_HoaDonNhapHang/` (cùng cấu trúc với `E:\GKM_HoaDonNhapHang`). Quy tắc ở mục 4 được cài trong `source/internal/core`; sửa quy tắc phải sửa cả hai nơi (tài liệu này và mã Go) rồi chạy lại kiểm thử. `gkm_convert.py` (Python) chỉ còn là bản tham chiếu. Quy tắc "ốc" đã bỏ hẳn: mã hàng luôn giữ nguyên như hóa đơn.

## 9. Cập nhật 02/10/2026 (lần 2): mẫu mới, .xls, đơn vị giá

Nguồn bằng chứng thêm: `samples/Mau 4 - Toa hang Tan Hoa Loi.xlsx` và `samples/Mau 5 - gk 28-2.xls` (cả hai do Khiêm gửi). Hai mẫu này **chưa có kết quả mẫu** nên app luôn bắt tích xác nhận trước khi xuất.

- **Nhận cột theo nhãn, có bí danh**: `SL`, `Sl đặt hàng` = Số lượng; `giá` = Đơn giá; `Tên mặt hàng` = Tên hàng; `Đvt`; `Hiệu`. Lỗi cũ (app không đọc được cột số lượng/đơn giá) do tiêu đề `SL` không có trong danh sách nhãn.
- **Mẫu D, Tân Hòa Lợi (Toa hàng)**: cột STT, MÃ HÀNG, HIỆU, SL, ĐƠN GIÁ, THÀNH TIỀN. Không có cột tên hàng. **Đơn giá ghi theo nghìn đồng** (7,5 = 7.500 đ). Căn cứ: giá 7,5 / 39 / 285 cho vòng bi, phớt chỉ hợp lý nếu tính theo nghìn đồng (vòng bi 6000-2RS giá 39 nghĩa là 39.000 đ); tổng 20.560, "BI NHẬT KM 2%: 400" (2% của 20.560 là 411) và "HỖ TRỢ VẬN CHUYỂN: 100" cùng một đơn vị. **Khiêm đã xác nhận (02/10/2026): đúng, tính nghìn đồng.** App luôn nhân 1.000 cho mẫu này. Các dòng sau "TỔNG TOA" (KM 2%, hỗ trợ vận chuyển, còn lại) chỉ hiển thị để đối chiếu, không đưa vào file.
- **Tên hàng khi hóa đơn không có cột tên** (mẫu D): **để trống** (Khiêm xác nhận 02/10/2026). Đã bỏ tùy chọn ghi "HIỆU Mã hàng".
- **Đơn vị đơn giá tự xác định, không hỏi người dùng** (Khiêm, 02/10/2026): mặc định đồng; nghìn đồng khi là mẫu Tân Hòa Lợi hoặc khi một ô chữ trên hóa đơn ghi chú rõ ("nghìn đồng"/"ngàn đồng" đi kèm, "ĐVT: 1.000đ", "Đơn giá (nghìn)", "(x1000đ)"...). Số tiền bằng chữ ("... nghìn đồng chẵn") không tính. Trường hợp nhận theo ghi chú thì app cảnh báo kèm câu đã nhận ra; không có ghi chú mà đơn giá cao nhất dưới 1.000 đ thì cảnh báo "nhỏ bất thường". Mẫu D được coi là mẫu đã biết (không còn bắt xác nhận chỉ vì là mẫu mới).
- **Mẫu E, bảng kê đơn hàng từ phần mềm kế toán** (file .xls Excel thật): cột Ngày ct, Số ct, Mã khách, Tên khách, Trạng thái, Mã hàng, Tên mặt hàng, Đvt, Mã kho, Sl đặt hàng, giá và một cột thành tiền không có tên. Không có STT. Số chứng từ và ngày lấy ở dòng hàng đầu (ngày là số ngày của Excel, 46080 = 27/02/2026). Dòng tổng nằm ở cột thành tiền không có mã hàng. Đơn giá có phần thập phân (ví dụ 63892,8) giữ nguyên như hóa đơn, app cảnh báo vì KiotViet có thể làm tròn.
- **Đọc .xls**: đọc theo nội dung tệp (không theo đuôi): .xlsx, .xls Excel 97-2003 (BIFF8), Excel XML 2003 và tệp HTML đặt đuôi .xls. Excel 5.0/95 và tệp có mật khẩu bị từ chối kèm hướng dẫn.
- **Một hóa đơn mỗi lần**: nút Làm mới xóa hóa đơn đang mở; file đã xuất không bị xóa.
- **Đã bỏ các chức năng từng thử**: tự nhập KiotViet, gửi Zalo Web, gửi Telegram (Khiêm yêu cầu bỏ cả ba trong ngày 02/10/2026); app 1.4.0 chốt ngày 02/10/2026.

## 10. Cập nhật 04/10/2026: sửa theo yêu cầu của Khiêm (app 1.5.0)

- **Chỉ có số lượng, không có giá**: vẫn đưa vào file, đơn giá 0 (hàng đổi, hàng bảo hành, hàng khuyến mại). Mẫu 3: 9 dòng "hàng đổi" nay có trong file (64 dòng). Nếu dòng nằm trong khối "Hàng đổi" mà có giá thì app vẫn để 0 và cảnh báo.
- **Có giá, không có số lượng**: bỏ dòng đó (xem trong mục xổ xuống "dòng không có số lượng", đối chiếu có ghi số dòng bị bỏ). Dòng trống cả giá lẫn số lượng cũng bỏ. Hóa đơn mà mọi dòng đều không có số lượng thì báo lỗi (không có gì để nhập).
- **Mở thư mục chứa file**: sửa lỗi mở nhầm; app mở đúng thư mục chứa file vừa xuất bằng ShellExecute với đường dẫn đã chuẩn hóa (chưa thử trên Windows thật).
- **Dòng cuối ứng dụng**: "Ứng dụng được thiết kế và phát triển bởi Gia Khánh Motor. Chỉ sử dụng nội bộ, vui lòng không sao chép dưới mọi hình thức".
- **Logo mới**: chữ GK trắng trên nền xanh, có đường trục và hai bánh xe vàng; dùng làm biểu tượng file .exe, biểu tượng cửa sổ và logo ở đầu giao diện (thư mục `assets\`).

## 11. Cập nhật 22/10/2026: phiếu xuất kho SL lấy, cập nhật online (app 1.6.0)

Nguồn bằng chứng: `samples/Mau 6 - Phieu xuat kho SL lay.xlsx` (Khiêm gửi, file `GIA_KH_NH_2026.xlsx`). Lỗi cũ "không tìm thấy dòng tiêu đề" do tiêu đề `MÃ PHỤ TÙNG`, `TÊN PHỤ TÙNG`, `SL LẤY` chưa có trong danh sách nhãn.

- **Mẫu F, "Phiếu xuất kho kiêm giấy nhận nợ"** (nhận biết bằng cột `SL LẤY` / `SỐ LƯỢNG LẤY` / `SL THỰC LẤY`): khi có cả `SL ĐẶT` và `SL LẤY` thì **số lượng = SL LẤY**. Khi có cả `ĐƠN GIÁ BÁN` và `ĐG ĐÃ CK` thì **đơn giá = ĐƠN GIÁ BÁN**. Mã hàng = `MÃ PHỤ TÙNG`, tên hàng = `TÊN PHỤ TÙNG`.
- **Chỉ lấy bảng hóa đơn cuối cùng** trong file (NCC hay gửi gộp cả chồng toa trước): các bảng xếp chồng trong một sheet hoặc mỗi toa một sheet đều xử lý; tiêu đề lặp lại ở trang sau mà giữa hai tiêu đề không có dòng "TỔNG" được hiểu là lặp tiêu đề trang (cùng một hóa đơn), không phải hóa đơn mới. App ghi một dòng đối chiếu "File có N bảng hóa đơn; app chỉ lấy bảng cuối cùng" để người dùng biết.
- **Ngày** lấy ở khối phía trên tiêu đề của bảng cuối (số ngày Excel hoặc chữ); **đối chiếu**: tổng tiền hàng (dòng có ô đúng "TỔNG") và tổng số lượng lấy phải khớp phép tính từ các dòng; "TỔNG PHẢI THU" (sau chiết khấu cấp hóa đơn) chỉ hiện để tham khảo. Tên file xuất `NhapHang_PhieuXuatKho_<ddMMyyyy>.xlsx`.
- **Kết quả Mẫu 6**: 38 dòng hàng, 3 dòng có SL LẤY = 0 bị bỏ, còn 35 dòng; tổng tiền hàng 12.700.802,16 và tổng SL lấy 103 khớp; chưa có kết quả mẫu để so từng ô nên app bắt xác nhận nếu có cảnh báo.
- **Số lượng bằng 0** nay được coi như không có số lượng: dòng bị bỏ (trước đây chỉ cảnh báo), dù có giá hay không.
- **Đơn giá lẻ**: giữ nguyên đúng hóa đơn (làm tròn 6 chữ số để bỏ nhiễu dấu phẩy động); chỉ cảnh báo khi lẻ quá 2 chữ số thập phân.
- **Giá trị vô lý**: số lượng ≥ 1 tỷ hoặc đơn giá ≥ 1.000 tỷ báo lỗi (chặn xuất); số lượng > 100.000 chỉ cảnh báo. Sheet quá 50.000 dòng hoặc 200 cột bị từ chối (file hỏng).
- **Cập nhật online**: nút "Cập nhật" trong app đọc `update/latest.json` và `update/GKM_NhapHoaDon.exe` trong một kho GitHub, kiểm tra SHA-256/kích thước/MZ/`-version`, thay file, tự mở lại. Chi tiết trong `HUONG_DAN.md` mục 4. Không dùng ChatGPT hay dịch vụ AI nào trong app (Khiêm làm rõ 22/10/2026: ChatGPT chỉ dùng để rà lỗi mã nguồn, không phải nút trong app).
- **Thư mục dự án (22/10/2026)**: máy của Khiêm không có ổ E, app cũ nằm ở `D:\Data AI\App nhap hang\Chuan\GKM_HoaDonNhapHang`. Từ 1.6.0 app tự lấy thư mục cha của `app\` (nơi chứa .exe) làm thư mục dự án; chỉ khi không nằm trong `app\` mới dùng mặc định `E:\GKM_HoaDonNhapHang` (hoặc `--root`, biến `GKM_ROOT`).
