package core

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func pctCell(v float64) Cell {
	return Cell{S: strconv.FormatFloat(math.Round(v*1e4)/1e4, 'f', -1, 64), Num: true}
}

var rePct = regexp.MustCompile(`(\d+(?:[.,]\d+)?)\s*%`)

// applyNetPercent (phiếu xuất kho, mẫu F): cột "ĐG ĐÃ CK 22%" cho biết % chiết khấu của từng dòng. Đơn giá vẫn là
// ĐƠN GIÁ BÁN (giá gốc) và % ghi vào cột Giảm giá (%) (Khiêm 14/11/2026: nhập giá gốc kèm %).
// Mỗi dòng phải khớp: ĐG đã CK = Đơn giá bán × (100% - %), lệch thì không ghi % và cảnh báo.
func applyNetPercent(g *Grid, p *Parsed, hr int) {
	netCol, head := 0, 0.0
	for c := 1; c <= g.MaxCol; c++ {
		t := normText(g.At(hr, c).S)
		if strings.HasPrefix(t, "đg đã ck") || strings.HasPrefix(t, "đơn giá đã ck") || strings.HasPrefix(t, "giá đã ck") {
			netCol = c
			if m := rePct.FindStringSubmatch(t); m != nil {
				head, _ = parseVNNumber(m[1])
			}
		}
	}
	if netCol == 0 {
		return
	}
	var bad []string
	n := 0
	for i := range p.Items {
		it := &p.Items[i]
		if it.Kind != KindNormal {
			continue
		}
		pr, ok1 := parseMoney(it.PriceRaw)
		net, ok2 := g.At(it.Row, netCol).Float()
		if !ok1 || !ok2 || pr <= 0 {
			continue
		}
		pct := (1 - net/pr) * 100
		if head > 0 && math.Abs(pct-head) < 0.01 {
			pct = head
		} else {
			pct = math.Round(pct*100) / 100
			if math.Abs(pr*(1-pct/100)-net) > math.Max(1, net*0.001) || pct < 0 || pct >= 100 {
				bad = append(bad, it.STT)
				continue
			}
		}
		if pct > 0 {
			it.CK, it.Net = pctCell(pct), g.At(it.Row, netCol)
			n++
		}
	}
	if n > 0 {
		lbl := "theo cột giá đã chiết khấu"
		if head > 0 {
			lbl = fmt.Sprintf("%s%% theo tiêu đề cột giá đã chiết khấu", vnNum(head))
		}
		p.Notes = append(p.Notes, Check{"Chiết khấu từng dòng", LevelOK, fmt.Sprintf("%d dòng ghi đơn giá bán (giá gốc) kèm giảm giá %s; từng dòng đã đối chiếu đúng với giá đã chiết khấu", n, lbl)})
	}
	if len(bad) > 0 {
		p.Warnings = append(p.Warnings, fmt.Sprintf("Dòng STT %s: giá đã chiết khấu không khớp đơn giá bán × %% chiết khấu nên app không ghi %% giảm giá cho các dòng này. Cần kiểm tra.", strings.Join(bad, ", ")))
	}
}

var (
	invPctLabels = []string{"tỷ lệ ck", "tỷ lệ chiết khấu", "giảm giá %", "giảm giá (%)", "chiết khấu %", "chiết khấu (%)", "% chiết khấu", "ck %", "ck(%)"}
	invAmtLabels = []string{"số tiền chiết khấu", "tiền chiết khấu", "chiết khấu", "giảm giá", "chiết khấu giảm giá", "ck"}
)

// invoiceDiscount đọc chiết khấu cả hóa đơn ở các dòng sau bảng hàng: trả % (nếu hóa đơn ghi %) và số tiền.
func invoiceDiscount(g *Grid, from int) (pct, amt float64) {
	for r := from; r <= g.MaxRow && r <= from+12; r++ {
		for c := 1; c <= g.MaxCol; c++ {
			cell := g.At(r, c)
			if cell.Empty() || cell.Num {
				continue
			}
			t := strings.Trim(normText(cell.S), " :.")
			isPct, isAmt := false, false
			for _, l := range invPctLabels {
				isPct = isPct || t == l
			}
			for _, l := range invAmtLabels {
				isAmt = isAmt || t == l
			}
			if !isPct && !isAmt {
				continue
			}
			for c2 := c + 1; c2 <= g.MaxCol; c2++ {
				v := g.At(r, c2)
				if v.Empty() {
					continue
				}
				f, ok := v.Float()
				if !ok {
					break // gặp nhãn khác trước khi có số
				}
				if isPct && !strings.Contains(t, "%") && !strings.Contains(v.S, "%") && !strings.HasPrefix(t, "tỷ lệ") {
					break
				}
				if isPct {
					pct = f
				} else {
					amt = f
				}
				break
			}
		}
	}
	return pct, amt
}

// spreadInvoiceDiscount chia chiết khấu cả hóa đơn (ví dụ "Tỷ lệ CK: 5 %" hoặc "Giảm giá: 1.070.100" dưới dòng tổng)
// vào cột Giảm giá (%) của từng dòng có giá (Khiêm 14/11/2026). Dòng đã có % riêng thì gộp: 1 - (1-a)(1-b).
// Hóa đơn chỉ ghi số tiền thì đổi ra % trên tổng tiền hàng (làm tròn 2 chữ số) và cảnh báo phần lệch do làm tròn.
func spreadInvoiceDiscount(g *Grid, p *Parsed, stopRow int) {
	from := stopRow
	if from == 0 {
		for _, it := range p.Items {
			if it.Row >= from {
				from = it.Row + 1
			}
		}
	}
	if from == 0 {
		return
	}
	pct, amt := invoiceDiscount(g, from)
	if pct <= 0 && amt <= 0 {
		return
	}
	// tổng tiền hàng sau giảm giá từng dòng (cơ sở để tính chiết khấu cả hóa đơn)
	base := 0.0
	for _, it := range p.Items {
		if it.Kind != KindNormal {
			continue
		}
		q, _ := it.Qty.Float()
		ck, _ := it.CK.Float()
		base += q * it.Price * (1 - ck/100)
	}
	if base <= 0 {
		return
	}
	scale := p.Scale
	if scale == 0 {
		scale = 1
	}
	exact := pct > 0
	if !exact {
		pct = math.Round(amt*scale/base*100*100) / 100
	}
	if pct <= 0 || pct >= 100 {
		p.Warnings = append(p.Warnings, fmt.Sprintf("Chiết khấu cả hóa đơn %s%% bất thường nên app không chia vào từng dòng. Cần kiểm tra.", vnNum(pct)))
		return
	}
	n := 0
	for i := range p.Items {
		it := &p.Items[i]
		if it.Kind != KindNormal || it.Price == 0 {
			continue
		}
		a, _ := it.CK.Float()
		it.CK = pctCell((1 - (1-a/100)*(1-pct/100)) * 100)
		n++
	}
	p.InvoiceCK = pct
	for i, t := range p.Meta.Tail { // dòng chiết khấu dưới tổng nay đã đưa vào file
		l := strings.Trim(normText(strings.TrimPrefix(t.Name, "Dòng sau tổng: ")), " :.")
		for _, x := range append(append([]string{}, invPctLabels...), invAmtLabels...) {
			if l == x && !strings.HasPrefix(t.Detail, "0 ") {
				p.Meta.Tail[i].Detail = "đã chia vào cột Giảm giá (%) của từng dòng (xem mục Chiết khấu cả hóa đơn)"
			}
		}
	}
	src := fmt.Sprintf("hóa đơn ghi chiết khấu %s%%", vnNum(pct))
	if !exact {
		src = fmt.Sprintf("hóa đơn ghi chiết khấu %s đ trên tổng tiền hàng %s đ, tức %s%%", vnNum(amt*scale), vnNum(base), vnNum(pct))
	}
	p.Notes = append(p.Notes, Check{"Chiết khấu cả hóa đơn", LevelOK, fmt.Sprintf("%s; đã chia vào cột Giảm giá (%%) của %d dòng (dòng có giảm giá riêng thì gộp hai mức)", src, n)})
	if !exact {
		if diff := base*pct/100 - amt*scale; math.Abs(diff) >= 1 {
			p.Warnings = append(p.Warnings, fmt.Sprintf("Chiết khấu cả hóa đơn ghi bằng số tiền (%s đ) nên app đổi ra %s%% để chia vào từng dòng; do làm tròn, tổng chiết khấu khi nhập kho lệch %s đ so với hóa đơn. Cần xác nhận.", vnNum(amt*scale), vnNum(pct), vnNum(math.Round(diff))))
		}
	}
}
