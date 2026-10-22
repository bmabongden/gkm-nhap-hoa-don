package core

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Mã mẫu nhà cung cấp đã có bằng chứng (3 cặp mẫu). "X" là mẫu lạ: đọc chung, bắt buộc người dùng xác nhận.
const (
	ProfileA = "A" // AMIS - Công ty TNHH Đoàn Lâm Khoa (Mẫu 1)
	ProfileB = "B" // Công ty TNHH Bình Minh (Mẫu 2)
	ProfileC = "C" // Quốc Quang - Việt Đài, MISA (Mẫu 3)
	ProfileD = "D" // Tân Hòa Lợi - Toa hàng (Mẫu 4, chưa có kết quả mẫu để đối chiếu)
	ProfileE = "E" // Bảng kê xuất từ phần mềm kế toán (đơn đặt hàng: Ngày ct, Số ct, Mã hàng, Tên mặt hàng, Sl đặt hàng, giá), file .xls
	ProfileF = "F" // Phiếu xuất kho kiêm giấy nhận nợ (MÃ PHỤ TÙNG, SL ĐẶT, SL LẤY, ĐƠN GIÁ BÁN, ĐG ĐÃ CK): lấy SL lấy, ĐG bán, chỉ bảng cuối cùng
	ProfileX = "X" // chưa nhận biết
)

// Options chỉ dùng để thử nghiệm: mặc định app tự xác định đơn vị đơn giá (xem detectScale).
type Options struct {
	// PriceScale: 0 = tự xác định; 1 = đồng; 1000 = hóa đơn ghi nghìn đồng.
	PriceScale float64
}

var ProfileNames = map[string]string{
	ProfileA: "AMIS - Công ty TNHH Đoàn Lâm Khoa (Hóa đơn bán hàng, CK%/Giá NET)",
	ProfileB: "Công ty TNHH Bình Minh (Biên bản giao hàng kiêm giấy nhận nợ)",
	ProfileC: "Quốc Quang - Việt Đài (Phiếu xuất hàng MISA)",
	ProfileE: "Bảng kê đơn hàng từ phần mềm kế toán (Ngày ct, Số ct, Mã hàng, Sl đặt hàng, giá)",
	ProfileD: "Tân Hòa Lợi (Toa hàng: Mã hàng, Hiệu, SL, Đơn giá tính nghìn đồng)",
	ProfileF: "Phiếu xuất kho kiêm giấy nhận nợ (Mã phụ tùng, SL đặt, SL lấy, Đơn giá bán, ĐG đã CK)",
	ProfileX: "Chưa nhận biết mẫu hóa đơn",
}

// Loại dòng.
const (
	KindNormal   = "thuong" // hàng thường
	KindPromo    = "km"     // hàng khuyến mại (đơn giá 0)
	KindExchange = "doi"    // hàng đổi/bảo hành: có số lượng, không có giá, vẫn nhập với đơn giá 0
	KindExcluded = "loai"   // dòng không có số lượng (dù có giá): bỏ, không nhập
)

// Mức của một mục đối chiếu/cảnh báo.
const (
	LevelOK    = "ok"
	LevelWarn  = "warn"
	LevelError = "error"
)

// Item là một dòng hàng đọc từ hóa đơn.
type Item struct {
	Row      int     `json:"row"`
	STT      string  `json:"stt"`
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Unit     string  `json:"unit"`
	Kind     string  `json:"kind"`
	Qty      Cell    `json:"-"`
	PriceRaw Cell    `json:"-"`
	Amount   Cell    `json:"-"`
	CK       Cell    `json:"-"`
	Net      Cell    `json:"-"`
	Price    float64 `json:"price"`
	Note     string  `json:"note,omitempty"`
}

// Check là một mục đối chiếu hiển thị cho người dùng.
type Check struct {
	Name   string `json:"name"`
	Level  string `json:"level"`
	Detail string `json:"detail"`
}

// Meta là thông tin đầu/cuối hóa đơn dùng để đặt tên file và đối chiếu.
type Meta struct {
	So            string
	Ngay          string
	TongTienHang  *float64
	ChietKhau     *float64
	TongThanhToan *float64
	TongSL        *float64
	Tail          []Check // các dòng "nhãn: số" ngay sau dòng tổng (mẫu D/X), không đưa vào file
}

// Parsed là kết quả phân tích một hóa đơn.
type Parsed struct {
	Profile   string
	Known     bool
	Items     []Item
	Meta      Meta
	Checks    []Check
	Errors    []string // lỗi chặn xuất
	Warnings  []string
	HeaderRow int
	// Scale là hệ số đã nhân vào đơn giá (1 hoặc 1000).
	Scale     float64
	ScaleNote string  // câu trên hóa đơn cho biết đơn giá tính nghìn đồng (nếu app nhận ra)
	HasName   bool    // hóa đơn có cột tên hàng (không có thì cột Tên hàng để trống)
	Notes     []Check // ghi chú thông tin (ví dụ file có nhiều bảng, chỉ lấy bảng cuối)
}

// OutRow là một dòng của file nhập hàng (A..G).
type OutRow struct {
	Code string
	Name string
	Unit string
	// Price luôn có (0 với hàng khuyến mại).
	Price float64
	Kind  string // loại dòng nguồn (KindNormal, KindPromo, KindExchange)
	CK    Cell   // Giảm giá (%): giữ nguyên văn bản nguồn, trống nếu không giảm
	Qty   Cell   // giữ nguyên kiểu nguồn (số hoặc văn bản)
}

type colMap struct {
	stt, code, name, unit, qty, price, amount, ck, net, brand, soCt, ngayCt int
	blockStart                                                              int // mẫu F: dòng tiêu đề của bảng liền trước (0 nếu không có)
}

// Nhãn cột đã chuẩn hóa (chữ thường) -> tên chuẩn. Hóa đơn mới chỉ cần thêm nhãn vào đây.
var headerAliases = map[string]string{
	"stt": "stt", "số tt": "stt", "s.t.t": "stt",
	"mã hàng": "code", "mã vật tư": "code", "mã": "code", "mã sp": "code", "mã số": "code", "mã hh": "code",
	"mã hàng hóa": "code", "mã hàng hoá": "code", "mã sản phẩm": "code", "mã phụ tùng": "code", "mã pt": "code",
	"tên hàng": "name", "tên vật tư": "name", "tên": "name", "tên sản phẩm": "name", "tên hàng hóa": "name",
	"tên hàng hoá": "name", "tên hh": "name", "tên phụ tùng": "name", "diễn giải": "name", "tên mặt hàng": "name",
	"đv": "unit", "dvt": "unit", "đvt": "unit", "đơn vị": "unit", "đơn vị tính": "unit",
	"số lượng": "qty", "sl": "qty", "s.lượng": "qty", "số lg": "qty", "slượng": "qty", "số lượng thực xuất": "qty", "sl đặt hàng": "qty", "sl đặt": "qty", "số lượng đặt": "qty", "sl xuất": "qty", "sl lấy": "qty", "số lượng lấy": "qty", "sl thực lấy": "qty",
	"đơn giá": "price", "đg": "price", "giá": "price", "đơn giá bán": "price", "đ.giá": "price", "giá bán": "price", "đg bán": "price",
	"thành tiền": "amount", "t.tiền": "amount", "thành tiền (vnđ)": "amount",
	"ck(%)": "ck", "giá net": "net",
	"hiệu": "brand", "nhãn hiệu": "brand", "hãng": "brand", "thương hiệu": "brand",
}

// Giới hạn kích thước sheet hợp lý của một hóa đơn; vượt quá thì coi là file hỏng (tránh quét hàng tỷ ô).
const (
	maxGridRows = 50000
	maxGridCols = 200
)

func canonLabel(s string) string {
	t := strings.Trim(normText(s), " :.")
	if k, ok := headerAliases[t]; ok {
		return k
	}
	return ""
}

// headerAt kiểm tra dòng r có phải dòng tiêu đề bảng hàng không. canon: tên chuẩn -> cột; raw: nhãn gốc đã chuẩn hóa -> cột.
func headerAt(g *Grid, r int) (canon map[string]int, raw map[string]int, ok bool) {
	canon, raw = map[string]int{}, map[string]int{}
	for c := 1; c <= g.MaxCol; c++ {
		cell := g.At(r, c)
		if cell.Empty() || cell.Num {
			continue
		}
		raw[normText(cell.S)] = c // nhãn trùng: lấy cột sau cùng
		if k := canonLabel(cell.S); k != "" {
			canon[k] = c
		}
	}
	// Có hai cột số lượng (SL đặt, SL lấy) thì số lượng nhập chuẩn là SL lấy; có cả "Đơn giá bán" và
	// "Đơn giá đã CK" thì lấy Đơn giá bán (Khiêm 22/10/2026).
	for _, k := range []string{"sl lấy", "số lượng lấy", "sl thực lấy"} {
		if c, ok := raw[k]; ok {
			canon["qty"] = c
		}
	}
	for _, k := range []string{"đơn giá bán", "đg bán"} {
		if c, ok := raw[k]; ok {
			canon["price"] = c
		}
	}
	if canon["code"] == 0 {
		return nil, nil, false
	}
	n := 0
	for _, k := range []string{"stt", "qty", "price", "name", "amount"} {
		if canon[k] > 0 {
			n++
		}
	}
	return canon, raw, n >= 2
}

// findHeader tìm dòng tiêu đề bảng hàng đầu tiên.
func findHeader(g *Grid) (row int, canon map[string]int, raw map[string]int) {
	for r := 1; r <= g.MaxRow; r++ {
		if c, rw, ok := headerAt(g, r); ok {
			return r, c, rw
		}
	}
	return 0, nil, nil
}

// isTotalRow: dòng tổng của một hóa đơn ("TỔNG", "TỔNG PHẢI THU", "Tổng cộng", "Tổng tiền hàng"...).
func isTotalRow(g *Grid, r int) bool {
	for c := 1; c <= g.MaxCol; c++ {
		cell := g.At(r, c)
		if cell.Num {
			continue
		}
		t := normText(cell.S)
		if t == "tổng" || strings.HasPrefix(t, "tổng phải thu") || strings.HasPrefix(t, "tổng cộng") || strings.HasPrefix(t, "tổng tiền hàng") {
			return true
		}
	}
	return false
}

// invoiceHeaders trả các dòng tiêu đề mở đầu một hóa đơn: tiêu đề đầu tiên, và mỗi tiêu đề sau chỉ tính là hóa đơn mới
// khi giữa nó và tiêu đề trước có dòng tổng (tiêu đề lặp ở trang sau thì không tính).
func invoiceHeaders(g *Grid) []int {
	var out []int
	prev := 0
	for r := 1; r <= g.MaxRow; r++ {
		if _, _, ok := headerAt(g, r); !ok {
			continue
		}
		if prev == 0 {
			out = append(out, r)
		} else {
			for x := prev + 1; x < r; x++ {
				if isTotalRow(g, x) {
					out = append(out, r)
					break
				}
			}
		}
		prev = r
	}
	return out
}

func detectProfile(canon, raw map[string]int) string {
	has := func(k string) bool { _, ok := raw[k]; return ok }
	switch {
	case has("mã vật tư"):
		return ProfileB
	case has("ck(%)") || has("giá net"):
		return ProfileA
	case has("thành tiền") && has("đơn vị"):
		return ProfileC
	case canon["stt"] == 0 && has("số ct"):
		return ProfileE
	case has("sl lấy") || has("số lượng lấy") || has("sl thực lấy"):
		return ProfileF
	case canon["brand"] > 0 && canon["name"] == 0:
		return ProfileD
	}
	return ProfileX
}

var (
	promoMarks    = map[string]bool{"khuyến mại": true, "khuyến mãi": true}
	excludedMarks = map[string]bool{"hàng dổi": true, "hàng đổi": true}
)

func rowHasTotals(g *Grid, r int) bool {
	for c := 1; c <= g.MaxCol; c++ {
		t := normText(g.At(r, c).S)
		if t == "tổng" || strings.HasPrefix(t, "tổng tiền hàng") || strings.HasPrefix(t, "cộng tiền hàng") || strings.HasPrefix(t, "tổng cộng") || strings.HasPrefix(t, "tổng toa") || strings.HasPrefix(t, "tổng phải thu") {
			return true
		}
	}
	return false
}

// Parse phân tích hóa đơn với lựa chọn mặc định.
func Parse(g *Grid) (*Parsed, error) { return ParseWith(g, Options{}) }

// ParseWith phân tích hóa đơn: nhận dạng mẫu, đọc dòng hàng, đối chiếu tổng.
func ParseWith(g *Grid, opt Options) (*Parsed, error) {
	if g.MaxRow > maxGridRows || g.MaxCol > maxGridCols {
		return nil, fmt.Errorf("sheet có kích thước bất thường (%d dòng, %d cột), không phải hóa đơn bình thường; có thể file bị hỏng", g.MaxRow, g.MaxCol)
	}
	hr, canon, raw := findHeader(g)
	if hr == 0 {
		return nil, fmt.Errorf("không tìm thấy dòng tiêu đề của bảng hàng (cần có cột Mã hàng cùng với STT, Số lượng/SL hoặc Đơn giá)")
	}
	prof := detectProfile(canon, raw)
	blockStart, tables, note := 0, 1, ""
	if prof == ProfileF {
		// NCC này hay gửi cả chồng toa hàng cũ: chỉ lấy bảng hóa đơn cuối cùng trong file (sheet cuối, rồi bảng cuối trong sheet).
		cands := append([]*Grid{g}, g.Extra...)
		tables = 0
		last := -1
		for i, cg := range cands {
			if h, c, r := findHeader(cg); h != 0 && detectProfile(c, r) == ProfileF {
				last = i
				tables += len(invoiceHeaders(cg))
			}
		}
		if last > 0 {
			g = cands[last]
			note = fmt.Sprintf("sheet \"%s\", ", g.Sheet)
		}
		hs := invoiceHeaders(g)
		hr = hs[len(hs)-1]
		if len(hs) > 1 {
			blockStart = hs[len(hs)-2]
		}
		canon, raw, _ = headerAt(g, hr)
		if tables > 1 {
			note = fmt.Sprintf("File có %d bảng hóa đơn; app chỉ lấy bảng cuối cùng (%sdòng tiêu đề %d), các bảng trước bị bỏ qua.", tables, note, hr)
		} else {
			note = ""
		}
	}
	p := &Parsed{Profile: prof, HeaderRow: hr}
	if note != "" {
		p.Notes = append(p.Notes, Check{"Nhiều hóa đơn trong file", LevelOK, strings.TrimPrefix(note, "")})
	}
	p.Known = p.Profile != ProfileX
	cm := colMap{
		stt: canon["stt"], code: canon["code"], name: canon["name"], unit: canon["unit"],
		qty: canon["qty"], price: canon["price"], amount: canon["amount"],
		ck: canon["ck"], net: canon["net"], brand: canon["brand"],
		soCt: raw["số ct"], ngayCt: raw["ngày ct"], blockStart: blockStart,
	}
	if cm.amount == 0 {
		cm.amount = inferAmountCol(g, hr, cm)
	}
	if cm.qty == 0 || cm.price == 0 {
		var miss []string
		if cm.qty == 0 {
			miss = append(miss, "Số lượng (hoặc SL)")
		}
		if cm.price == 0 {
			miss = append(miss, "Đơn giá")
		}
		return nil, fmt.Errorf("bảng hàng thiếu cột %s ở dòng tiêu đề (dòng %d), không thể đọc an toàn", strings.Join(miss, " và "), hr)
	}
	p.HasName = cm.name > 0
	p.Scale, p.ScaleNote = opt.PriceScale, ""
	if p.Scale == 0 {
		p.Scale = 1
		if p.Profile == ProfileD { // Tân Hòa Lợi ghi nghìn đồng (Khiêm xác nhận 02/10/2026)
			p.Scale = 1000
		} else if note := detectThousandNote(g); note != "" {
			p.Scale, p.ScaleNote = 1000, note
		}
	}
	if !p.Known {
		p.Warnings = append(p.Warnings, "Mẫu hóa đơn này chưa có kết quả mẫu để đối chiếu (chỉ có 3 mẫu đã kiểm chứng). App đọc theo tên cột; hãy xem kỹ bản xem trước rồi tích xác nhận trước khi xuất.")
	}
	if p.ScaleNote != "" {
		p.Warnings = append(p.Warnings, fmt.Sprintf("Hóa đơn có ghi chú \"%s\" nên app hiểu đơn giá tính NGHÌN ĐỒNG và nhân %s khi xuất (ví dụ 7,5 thành 7.500 đ). Hãy xem lại đơn giá trong bảng bên dưới.", p.ScaleNote, vnNum(p.Scale)))
	}

	mode := "normal"
	stopRow := 0
	for r := hr + 1; r <= g.MaxRow; r++ {
		stt := g.At(r, cm.stt).Text()
		code := g.At(r, cm.code)
		if normText(stt) == "stt" { // tiêu đề lặp ở trang sau
			continue
		}
		isRow := isDigits(strings.ReplaceAll(normText(stt), ".", "")) && !code.Empty()
		if cm.stt == 0 { // hóa đơn không có cột STT: dòng hàng là dòng có mã và số lượng là số
			_, qok := g.At(r, cm.qty).Float()
			isRow = !code.Empty() && qok
			stt = strconv.Itoa(r - hr)
		}
		if !isRow {
			if normText(stt) == "tổng cộng" || rowHasTotals(g, r) {
				stopRow = r
				break
			}
			continue
		}
		name := g.At(r, cm.name).Text()
		it := Item{
			Row: r, STT: stt, Code: code.Text(), Name: name, Unit: g.At(r, cm.unit).Text(),
			Qty: g.At(r, cm.qty), PriceRaw: g.At(r, cm.price), Amount: g.At(r, cm.amount),
			CK: g.At(r, cm.ck), Net: g.At(r, cm.net),
		}
		if code.Num {
			p.Warnings = append(p.Warnings, fmt.Sprintf("Dòng STT %s: mã hàng %s được lưu dạng số trong Excel, có thể đã mất số 0 ở đầu.", stt, it.Code))
		}
		pr := it.PriceRaw
		if !pr.Num && !pr.Empty() {
			switch t := normText(pr.S); {
			case promoMarks[t]:
				mode = "promo"
			case excludedMarks[t]:
				mode = "exchange"
			}
		}
		price, priceOK := parseMoney(pr)
		price = math.Round(price*p.Scale*1e6) / 1e6 // bỏ nhiễu dấu phẩy động (249381.72000000003 -> 249381.72)
		it.Price = price
		if qv, qok := it.Qty.Float(); it.Qty.Empty() || (qok && qv == 0) { // không có số lượng (trống hoặc 0, ví dụ SL lấy = 0) thì bỏ dòng, dù có hay không có giá (Khiêm 04/10/2026, 22/10/2026)
			it.Kind, it.Note = KindExcluded, "Không có số lượng nên không nhập"
			p.Items = append(p.Items, it)
			continue
		}
		switch {
		case mode == "exchange":
			it.Kind = KindExchange
			if priceOK && price != 0 {
				it.Note = "Nằm trong khối Hàng đổi nhưng có đơn giá trên hóa đơn; đã để đơn giá 0"
				p.Warnings = append(p.Warnings, fmt.Sprintf("Dòng STT %s (%s): trong khối Hàng đổi nhưng hóa đơn có đơn giá %s, app để 0 như các hàng đổi khác. Cần xác nhận.", stt, it.Code, vnNum(price)))
			}
			it.Price = 0
		case mode == "promo":
			it.Kind = KindPromo
			if priceOK && price != 0 {
				it.Note = "Nằm trong khối Khuyến mại nhưng có đơn giá trên hóa đơn; đã để đơn giá 0"
				p.Warnings = append(p.Warnings, fmt.Sprintf("Dòng STT %s (%s): trong khối Khuyến mại nhưng hóa đơn có đơn giá %s, app để 0 theo quy tắc khuyến mại. Cần xác nhận.", stt, it.Code, vnNum(price)))
			}
			it.Price = 0
		case !priceOK && pr.Empty():
			it.Kind = KindPromo // đơn giá trống = hàng khuyến mại (Mẫu 1)
			it.Price = 0
			if !p.Known {
				it.Note = "Không có đơn giá, tạm coi là khuyến mại"
				p.Warnings = append(p.Warnings, fmt.Sprintf("Dòng STT %s (%s): không có đơn giá, app tạm coi là hàng khuyến mại (đơn giá 0). Cần kiểm tra.", stt, it.Code))
			}
		case !priceOK:
			it.Kind = KindNormal
			p.Errors = append(p.Errors, fmt.Sprintf("Dòng STT %s (%s): đơn giá \"%s\" không đọc được thành số.", stt, it.Code, pr.Text()))
		case price == 0:
			it.Kind = KindPromo
		default:
			it.Kind = KindNormal
		}
		if it.Kind != KindExcluded {
			q, qok := it.Qty.Float()
			switch {
			case !qok:
				p.Errors = append(p.Errors, fmt.Sprintf("Dòng STT %s (%s): số lượng \"%s\" không đọc được thành số.", stt, it.Code, it.Qty.Text()))
			case q <= 0:
				p.Warnings = append(p.Warnings, fmt.Sprintf("Dòng STT %s (%s): số lượng %s không dương.", stt, it.Code, vnNum(q)))
			case q >= 1e9:
				p.Errors = append(p.Errors, fmt.Sprintf("Dòng STT %s (%s): số lượng %s lớn vô lý, có thể ô bị nhập sai.", stt, it.Code, vnNum(q)))
			case q > 1e5:
				p.Warnings = append(p.Warnings, fmt.Sprintf("Dòng STT %s (%s): số lượng %s lớn bất thường. Cần xác nhận.", stt, it.Code, vnNum(q)))
			}
			if it.Price >= 1e12 {
				p.Errors = append(p.Errors, fmt.Sprintf("Dòng STT %s (%s): đơn giá %s lớn vô lý, có thể ô bị nhập sai.", stt, it.Code, vnNum(it.Price)))
			}
		}
		p.Items = append(p.Items, it)
	}
	if len(p.Items) == 0 {
		return nil, fmt.Errorf("không có dòng hàng nào dưới dòng tiêu đề")
	}
	if n, _, _, ex := p.Counts(); n+ex == len(p.Items) && n == 0 && ex > 0 {
		return nil, fmt.Errorf("tất cả %d dòng hàng đều không có số lượng, không có gì để nhập (kiểm tra cột Số lượng/SL)", ex)
	}
	if p.Scale == 1 {
		maxP := 0.0
		for _, it := range p.Items {
			if it.Kind == KindNormal && it.Price > maxP {
				maxP = it.Price
			}
		}
		if maxP > 0 && maxP < 1000 {
			p.Warnings = append(p.Warnings, fmt.Sprintf("Đơn giá cao nhất trên hóa đơn chỉ là %s đ, nhỏ bất thường. Hóa đơn có thể ghi đơn giá theo nghìn đồng mà app không nhận ra: hãy kiểm tra hóa đơn gốc trước khi xuất.", vnNum(maxP)))
		}
	}
	frac := 0
	for _, it := range p.Items {
		if it.Kind != KindExcluded && math.Abs(it.Price*100-math.Round(it.Price*100)) > 0.01 { // chỉ cảnh báo khi lẻ quá 2 chữ số thập phân
			frac++
		}
	}
	if frac > 0 {
		p.Warnings = append(p.Warnings, fmt.Sprintf("%d dòng có đơn giá lẻ quá 2 chữ số thập phân (ví dụ 69400,6667). App giữ nguyên đúng như hóa đơn; KiotViet có thể tự làm tròn khi nhập.", frac))
	}
	if cm.soCt > 0 {
		docs := map[string]bool{}
		for _, it := range p.Items {
			docs[g.At(it.Row, cm.soCt).Text()] = true
		}
		if len(docs) > 1 {
			p.Warnings = append(p.Warnings, fmt.Sprintf("File có %d số chứng từ khác nhau (cột Số ct); app gộp tất cả vào một file nhập hàng. Hãy kiểm tra.", len(docs)))
		}
	}
	p.Meta = readMeta(g, p.Profile, cm, hr)
	if p.Profile == ProfileD || p.Profile == ProfileX {
		p.Meta.Tail = readTail(g, stopRow)
	}
	p.Checks = check(p)
	return p, nil
}

func readMeta(g *Grid, prof string, cm colMap, hr int) Meta {
	var m Meta
	pos := g.Positions()
	after := func(pred func(string) bool, numeric bool) (Cell, bool) {
		for _, k := range pos {
			c := g.Cells[k]
			if c.Num || !pred(normText(c.S)) {
				continue
			}
			for _, k2 := range pos {
				if k2[0] == k[0] && k2[1] > k[1] {
					v := g.Cells[k2]
					if !numeric {
						return v, true
					}
					if _, ok := v.Float(); ok {
						return v, true
					}
				}
			}
		}
		return Cell{}, false
	}
	fptr := func(c Cell, ok bool) *float64 {
		if !ok {
			return nil
		}
		f, ok2 := c.Float()
		if !ok2 {
			return nil
		}
		return &f
	}
	eq := func(s string) func(string) bool { return func(t string) bool { return t == s } }
	pre := func(s string) func(string) bool { return func(t string) bool { return strings.HasPrefix(t, s) } }
	switch prof {
	case ProfileA, ProfileC:
		if c, ok := after(eq("ngày:"), false); ok {
			m.Ngay = c.Text()
		}
		if c, ok := after(eq("số:"), false); ok {
			m.So = c.Text()
		}
	case ProfileB:
		for _, k := range pos {
			if mm := reDateWords.FindStringSubmatch(normText(g.Cells[k].S)); mm != nil {
				m.Ngay = padDate(mm[1], mm[2], mm[3])
			}
		}
	}
	switch prof {
	case ProfileD, ProfileX:
		for _, k := range pos {
			if k[0] >= hr {
				break
			}
			if _, _, _, ok := dateParts(g.Cells[k].S); ok && !g.Cells[k].Num {
				m.Ngay = strings.TrimSpace(g.Cells[k].S)
				break
			}
		}
		m.TongTienHang = fptr(after(func(t string) bool {
			return strings.HasPrefix(t, "tổng toa") || strings.HasPrefix(t, "tổng tiền hàng") || strings.HasPrefix(t, "cộng tiền hàng")
		}, true))
	case ProfileF:
		// ngày: ô ngày nằm trong khối ngay trên bảng (sau bảng trước đó nếu file có nhiều bảng)
		for _, k := range pos {
			if k[0] <= cm.blockStart {
				continue
			}
			if k[0] >= hr {
				break
			}
			c := g.Cells[k]
			if c.Num {
				if f, ok := c.Float(); ok && f > 20000 && f < 80000 {
					m.Ngay = excelDate(f)
					break
				}
			} else if _, _, _, ok := dateParts(c.S); ok {
				m.Ngay = c.Text()
				break
			}
		}
		for r := hr + 1; r <= g.MaxRow; r++ {
			if _, ok := g.At(r, cm.code).Float(); ok || !g.At(r, cm.code).Empty() {
				continue
			}
			if isTotalRow(g, r) && normText(g.At(r, cm.stt).S) != "stt" {
				for c := 1; c <= g.MaxCol; c++ {
					t := normText(g.At(r, c).S)
					if t == "tổng" && !g.At(r, c).Num {
						m.TongTienHang = fptr(g.At(r, cm.amount), cm.amount > 0)
						m.TongSL = fptr(g.At(r, cm.qty), true)
					}
				}
				if m.TongTienHang != nil {
					m.TongThanhToan = fptr(after(func(t string) bool { return strings.HasPrefix(t, "tổng phải thu") }, true))
					break
				}
			}
		}
	case ProfileE:
		// số chứng từ, ngày chứng từ: lấy ở dòng hàng đầu tiên
		for r := hr + 1; r <= g.MaxRow; r++ {
			if g.At(r, cm.code).Empty() {
				continue
			}
			m.So = g.At(r, cm.soCt).Text()
			if c := g.At(r, cm.ngayCt); c.Num {
				if f, ok := c.Float(); ok && f > 20000 && f < 80000 {
					m.Ngay = excelDate(f)
				}
			} else if _, _, _, ok := dateParts(c.S); ok {
				m.Ngay = c.Text()
			}
			break
		}
		// dòng tổng: dòng không có mã hàng, chỉ có số ở cột thành tiền
		if cm.amount > 0 {
			for r := g.MaxRow; r > hr; r-- {
				if g.At(r, cm.code).Empty() && !g.At(r, cm.amount).Empty() {
					m.TongTienHang = fptr(g.At(r, cm.amount), true)
					break
				}
				if !g.At(r, cm.code).Empty() {
					break
				}
			}
		}
	case ProfileA:
		m.TongTienHang = fptr(after(pre("cộng tiền hàng"), true))
		m.ChietKhau = fptr(after(pre("tiền chiết khấu"), true))
		m.TongThanhToan = fptr(after(pre("tổng tiền thanh toán"), true))
	case ProfileC:
		m.TongTienHang = fptr(after(pre("tổng tiền hàng"), true))
		m.ChietKhau = fptr(after(pre("số tiền chiết khấu"), true))
		m.TongThanhToan = fptr(after(pre("cộng tiền hàng (đã trừ ck)"), true))
	case ProfileB:
		for _, k := range pos {
			c := g.Cells[k]
			if !c.Num && normText(c.S) == "tổng cộng" {
				m.TongSL = fptr(g.At(k[0], cm.qty), true)
				m.TongTienHang = fptr(g.At(k[0], cm.amount), cm.amount > 0)
			}
		}
	}
	return m
}

// readTail đọc vài dòng "nhãn: số" ngay sau dòng tổng (ví dụ "BI NHẬT KM 2%: 400").
func readTail(g *Grid, stopRow int) []Check {
	if stopRow == 0 {
		return nil
	}
	var out []Check
	for r := stopRow + 1; r <= g.MaxRow && r <= stopRow+8; r++ {
		label := ""
		var val *Cell
		for c := 1; c <= g.MaxCol; c++ {
			cell := g.At(r, c)
			if cell.Empty() {
				continue
			}
			if label == "" && !cell.Num {
				label = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(cell.S), ":"))
				continue
			}
			if label != "" {
				if _, ok := cell.Float(); ok {
					cc := cell
					val = &cc
					break
				}
			}
		}
		if label == "" || val == nil {
			break
		}
		f, _ := val.Float()
		out = append(out, Check{"Dòng sau tổng: " + label, LevelOK, vnNum(f) + " (theo đơn vị trên hóa đơn; mẫu nhập hàng không có chỗ cho khoản này nên không đưa vào file)"})
	}
	return out
}

// inferAmountCol tìm cột thành tiền không có tên: cột ngay bên phải cột đơn giá nếu phần lớn dòng bằng SL × đơn giá.
func inferAmountCol(g *Grid, hr int, cm colMap) int {
	c := cm.price + 1
	if c > g.MaxCol || !g.At(hr, c).Empty() {
		return 0
	}
	match, total := 0, 0
	for r := hr + 1; r <= g.MaxRow; r++ {
		if g.At(r, cm.code).Empty() {
			continue
		}
		q, ok1 := g.At(r, cm.qty).Float()
		pr, ok2 := parseMoney(g.At(r, cm.price))
		a, ok3 := g.At(r, c).Float()
		total++
		if ok1 && ok2 && ok3 && math.Abs(q*pr-a) <= 0.5 {
			match++
		}
	}
	if total > 0 && match*10 >= total*9 {
		return c
	}
	return 0
}

// excelDate đổi số ngày của Excel (1900) thành dd/mm/yyyy.
func excelDate(serial float64) string {
	t := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(serial))
	return t.Format("02/01/2006")
}

func padDate(d, mo, y string) string {
	var di, mi int
	fmt.Sscanf(d, "%d", &di)
	fmt.Sscanf(mo, "%d", &mi)
	return fmt.Sprintf("%02d/%02d/%s", di, mi, y)
}

func f0(c Cell) float64 {
	f, _ := c.Float()
	return f
}

func check(p *Parsed) []Check {
	var res []Check
	add := func(name string, ok bool, detail string) {
		lv := LevelOK
		if !ok {
			lv = LevelWarn
		}
		res = append(res, Check{name, lv, detail})
	}
	// 1. Thành tiền dòng = SL × đơn giá (hàng thường có cột Thành tiền)
	var bad []string
	haveAmount := false
	for _, it := range p.Items {
		if it.Kind == KindNormal && !it.Amount.Empty() {
			haveAmount = true
			q, qok := it.Qty.Float()
			pr, pok := parseMoney(it.PriceRaw)
			a, aok := it.Amount.Float()
			if !qok || !pok || !aok || math.Abs(q*pr-a) > 0.5 {
				bad = append(bad, it.STT)
			}
		}
	}
	if haveAmount {
		d := "khớp toàn bộ"
		if len(bad) > 0 {
			d = "dòng lệch (STT): " + strings.Join(bad, ", ")
		}
		add("Thành tiền dòng = SL × đơn giá", len(bad) == 0, d)
	}
	// 2. Dòng trùng (mã + giá + loại)
	seen := map[string]bool{}
	var dup []string
	for _, it := range p.Items {
		if it.Kind == KindExcluded {
			continue
		}
		k := fmt.Sprintf("%s|%v|%s", it.Code, it.Price, it.Kind)
		if seen[k] {
			dup = append(dup, it.Code)
		}
		seen[k] = true
	}
	d := "không có"
	if len(dup) > 0 {
		d = "trùng: " + strings.Join(dup, ", ")
	}
	add("Dòng trùng (mã + giá + loại)", len(dup) == 0, d)
	// 3. Tổng theo hóa đơn
	var sumAmt, sumQty, sumCK float64
	for _, it := range p.Items {
		if it.Kind == KindNormal {
			amt := f0(it.Amount)
			if it.Amount.Empty() {
				pr, _ := parseMoney(it.PriceRaw)
				amt = f0(it.Qty) * pr
			}
			sumAmt += amt
			sumQty += f0(it.Qty)
			sumCK += amt - f0(it.Net)*f0(it.Qty)
		}
	}
	m := p.Meta
	missing := func(name string) {
		res = append(res, Check{name, LevelWarn, "không đọc được số tổng trên hóa đơn để đối chiếu"})
	}
	switch p.Profile {
	case ProfileA:
		if m.TongTienHang == nil || m.ChietKhau == nil || m.TongThanhToan == nil {
			missing("Tổng hóa đơn")
			break
		}
		add("Cộng tiền hàng", math.Abs(sumAmt-*m.TongTienHang) < 0.5, fmt.Sprintf("tính %s / hóa đơn %s", vnNum(sumAmt), vnNum(*m.TongTienHang)))
		dd := math.Abs(sumCK - *m.ChietKhau)
		extra := ""
		if dd >= 1 {
			extra = fmt.Sprintf(" (hóa đơn làm tròn, lệch %sđ, chỉ cảnh báo, không ảnh hưởng file)", vnNum(math.Round(dd)))
		}
		add("Tiền chiết khấu", dd < 1000, fmt.Sprintf("tính %s / hóa đơn %s%s", vnNum(math.Round(sumCK)), vnNum(*m.ChietKhau), extra))
		add("Tổng thanh toán", math.Abs(*m.TongTienHang-*m.ChietKhau-*m.TongThanhToan) < 0.5,
			fmt.Sprintf("%s - %s = %s", vnNum(*m.TongTienHang), vnNum(*m.ChietKhau), vnNum(*m.TongThanhToan)))
	case ProfileB:
		if m.TongTienHang == nil || m.TongSL == nil {
			missing("Tổng cộng")
			break
		}
		add("Tổng tiền hàng", math.Abs(sumAmt-*m.TongTienHang) < 0.01, fmt.Sprintf("tính %s / hóa đơn %s", vnNum(sumAmt), vnNum(*m.TongTienHang)))
		add("Tổng số lượng", sumQty == *m.TongSL, fmt.Sprintf("tính %s / hóa đơn %s", vnNum(sumQty), vnNum(*m.TongSL)))
	case ProfileF:
		if m.TongTienHang == nil {
			missing("Tổng tiền hàng")
			break
		}
		add("Tổng tiền hàng", math.Abs(sumAmt-*m.TongTienHang) < 0.5, fmt.Sprintf("tính %s / hóa đơn %s", vnNum(sumAmt), vnNum(*m.TongTienHang)))
		if m.TongSL != nil {
			add("Tổng số lượng lấy", sumQty == *m.TongSL, fmt.Sprintf("tính %s / hóa đơn %s", vnNum(sumQty), vnNum(*m.TongSL)))
		}
		if m.TongThanhToan != nil {
			res = append(res, Check{"Tổng phải thu", LevelOK, fmt.Sprintf("hóa đơn ghi %s (đã trừ chiết khấu); mẫu nhập hàng không có chỗ cho chiết khấu cấp hóa đơn nên không đưa vào file", vnNum(*m.TongThanhToan))})
		}
	case ProfileD, ProfileE, ProfileX:
		if m.TongTienHang == nil {
			missing("Tổng tiền hàng")
			break
		}
		add("Tổng tiền hàng", math.Abs(sumAmt-*m.TongTienHang) < 0.5, fmt.Sprintf("tính %s / hóa đơn %s (đơn vị trên hóa đơn)", vnNum(sumAmt), vnNum(*m.TongTienHang)))
		res = append(res, m.Tail...)
	case ProfileC:
		if m.TongTienHang == nil {
			missing("Tổng tiền hàng")
			break
		}
		add("Tổng tiền hàng", math.Abs(sumAmt-*m.TongTienHang) < 0.5, fmt.Sprintf("tính %s / hóa đơn %s", vnNum(sumAmt), vnNum(*m.TongTienHang)))
		if m.ChietKhau != nil {
			res = append(res, Check{"Chiết khấu cấp hóa đơn", LevelOK, fmt.Sprintf("hóa đơn có CK %s; mẫu nhập hàng không có chỗ cho CK cấp hóa đơn nên không đưa vào dòng", vnNum(*m.ChietKhau))})
		}
	}
	if len(p.Items) > 0 {
		_, _, exch, excl := p.Counts()
		if exch > 0 {
			res = append(res, Check{"Hàng đổi/bảo hành", LevelOK, fmt.Sprintf("%d dòng chỉ có số lượng, không có giá: vẫn đưa vào file nhập với đơn giá 0", exch)})
		}
		if excl > 0 {
			res = append(res, Check{"Dòng không có số lượng", LevelOK, fmt.Sprintf("%d dòng không có số lượng nên bỏ, không đưa vào file nhập", excl)})
		}
	}
	return append(p.Notes, res...)
}

// OutputRows trả các dòng sẽ ghi vào file nhập hàng (đã bỏ dòng không có số lượng).
func (p *Parsed) OutputRows() []OutRow {
	var rows []OutRow
	for _, it := range p.Items {
		if it.Kind == KindExcluded {
			continue
		}
		r := OutRow{Code: cleanCtl(it.Code), Name: cleanCtl(it.Name), Price: it.Price, Qty: it.Qty, Kind: it.Kind}
		if p.Profile == ProfileA && !it.CK.Empty() {
			r.CK = it.CK // giữ nguyên văn bản như "2,00"
		}
		rows = append(rows, r)
	}
	return rows
}

// Counts trả số dòng theo loại.
func (p *Parsed) Counts() (normal, promo, exchange, excluded int) {
	for _, it := range p.Items {
		switch it.Kind {
		case KindNormal:
			normal++
		case KindPromo:
			promo++
		case KindExchange:
			exchange++
		case KindExcluded:
			excluded++
		}
	}
	return
}

var reBad = regexp.MustCompile(`[^0-9A-Za-z]+`)

// SuggestedName: NhapHang_<tenhoadon>_<ddMMyyyy>.xlsx
func (p *Parsed) SuggestedName() string {
	ten := ""
	if p.Meta.So != "" {
		ten = strings.Trim(reBad.ReplaceAllString(stripAccents(p.Meta.So), "-"), "-")
	}
	if ten == "" {
		ten = map[string]string{ProfileB: "BinhMinh", ProfileD: "TanHoaLoi", ProfileE: "DonHang", ProfileF: "PhieuXuatKho"}[p.Profile]
		if ten == "" {
			ten = "HoaDon"
		}
	}
	ngay := "khongngay"
	if d, m, y, ok := dateParts(p.Meta.Ngay); ok {
		ngay = d + m + y
	}
	return fmt.Sprintf("NhapHang_%s_%s.xlsx", ten, ngay)
}

// NeedConfirm: cần người dùng tích xác nhận trước khi xuất (mẫu lạ hoặc có mục đối chiếu lệch/cảnh báo).
func (p *Parsed) NeedConfirm() bool {
	if !p.Known || len(p.Warnings) > 0 {
		return true
	}
	for _, c := range p.Checks {
		if c.Level != LevelOK {
			return true
		}
	}
	return false
}

// reThousand nhận các cách hóa đơn ghi chú "đơn giá tính nghìn đồng" (đã bỏ dấu, chữ thường).
var reThousand = regexp.MustCompile(`(nghin|ngan)\s*(dong|vnd|d)\b|\b(dvt|don vi( tinh)?)\s*[:=]?\s*(nghin|ngan|1\.?000)(dong|vnd|d)?\b|\(\s*(x\s*)?1\.?000\s*(dong|vnd|d)\s*\)|\bx\s*1\.?000\s*(dong|vnd|d)\b|\b(don gia|gia|thanh tien)\s*\(\s*(nghin|ngan|k|1\.?000\s*(dong|vnd|d)?)\s*\)`)

var reAmountWords = regexp.MustCompile(`\b(chan|trieu|tram|muoi|linh|ty|le)\b`)

// detectThousandNote tìm trong các ô chữ của hóa đơn (không tính ô số) câu cho biết tiền ghi theo nghìn đồng.
// Trả câu đó (đã cắt gọn) hoặc "" nếu không có: khi đó đơn giá hiểu là đồng.
func detectThousandNote(g *Grid) string {
	for _, pos := range g.Positions() {
		cell := g.At(pos[0], pos[1])
		if cell.Num {
			continue
		}
		t := cell.Text()
		if len([]rune(t)) > 80 {
			continue
		}
		f := strings.ToLower(stripAccents(normText(t)))
		if reAmountWords.MatchString(f) { // "Sáu mươi bảy triệu ... nghìn đồng chẵn" là số tiền bằng chữ, không phải ghi chú đơn vị
			continue
		}
		if reThousand.MatchString(f) {
			return t
		}
	}
	return ""
}

// cleanCtl bỏ ký tự điều khiển (NUL, ESC...) mà file Excel/XML không lưu được; giữ tab và xuống dòng.
func cleanCtl(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\t' && r != '\n' && r != '\r' || r == 0x7f || r == 0xfffe || r == 0xffff || r == 0xfffd {
			return -1
		}
		return r
	}, s)
}
