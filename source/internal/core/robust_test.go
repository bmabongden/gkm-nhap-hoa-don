package core

import (
	"math/rand"
	"testing"
)

// TestRobustMutatedGrids xáo trộn ô của các hóa đơn mẫu (xóa ô, thay bằng chữ/số lạ) và bảo đảm app
// không bao giờ panic: chỉ được trả lỗi hoặc cảnh báo. Hạt giống cố định nên chạy lại ra đúng kết quả cũ.
func TestRobustMutatedGrids(t *testing.T) {
	tpl := read(t, "MauFileNhapHang.xlsx")
	junk := []string{"", " ", "abc", "-1", "0", "1e308", "NaN", "Inf", "1,2,3", "-0", "999999999999999999999", "tổng", "TỔNG", "SL lấy", "nghìn đồng", "\x00", "ĐƠN GIÁ BÁN"}
	rng := rand.New(rand.NewSource(20261022))
	files := []string{"Mau 1.xlsx", "Mau 2.xlsx", "Mau 3.xlsx", "Mau 4 - Toa hang Tan Hoa Loi.xlsx", "Mau 6 - Phieu xuat kho SL lay.xlsx"}
	for _, name := range files {
		g0, err := ReadGrid(read(t, name))
		if err != nil {
			t.Fatal(err)
		}
		pos := g0.Positions()
		for iter := 0; iter < 60; iter++ {
			g := &Grid{Sheet: g0.Sheet, Cells: map[[2]int]Cell{}, MaxRow: g0.MaxRow, MaxCol: g0.MaxCol}
			for k, v := range g0.Cells {
				g.Cells[k] = v
			}
			for m := 0; m < 1+rng.Intn(12); m++ {
				k := pos[rng.Intn(len(pos))]
				switch rng.Intn(3) {
				case 0:
					delete(g.Cells, k)
				case 1:
					g.Cells[k] = Cell{S: junk[rng.Intn(len(junk))], Num: rng.Intn(2) == 0}
				default:
					g.Cells[k] = Cell{S: "12", Num: true}
				}
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic với %s lần %d: %v", name, iter, r)
					}
				}()
				p, err := Parse(g)
				if err != nil {
					return
				}
				_, _, _, _ = p.Counts()
				_ = p.SuggestedName()
				_ = p.NeedConfirm()
				rows := p.OutputRows()
				if len(p.Errors) == 0 && len(rows) > 0 {
					data, err := BuildWorkbook(tpl, rows)
					if err != nil {
						t.Fatalf("%s lần %d: dựng file lỗi: %v", name, iter, err)
					}
					if err := VerifyExport(data, rows); err != nil {
						t.Fatalf("%s lần %d: kiểm tra file xuất lỗi: %v", name, iter, err)
					}
				}
			}()
		}
	}
}

// TestRobustCorruptFiles cắt cụt / làm hỏng byte của file thật: ReadGridAny không được panic hay treo.
func TestRobustCorruptFiles(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for _, name := range []string{"Mau 1.xlsx", "Mau 5 - gk 28-2.xls", "Mau 6 - Phieu xuat kho SL lay.xlsx"} {
		data := read(t, name)
		for iter := 0; iter < 20; iter++ {
			b := append([]byte(nil), data...)
			switch iter % 3 {
			case 0:
				b = b[:rng.Intn(len(b))]
			case 1:
				for i := 0; i < 20; i++ {
					b[rng.Intn(len(b))] = byte(rng.Intn(256))
				}
			default:
				p := rng.Intn(len(b))
				b = append(b[:p], b[p+rng.Intn(len(b)-p):]...)
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("panic với %s lần %d: %v", name, iter, r)
					}
				}()
				if g, _, err := ReadGridAny(b); err == nil {
					_, _ = Parse(g)
				}
			}()
		}
	}
}
