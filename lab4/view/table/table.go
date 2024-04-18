package table

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	clearColor = color.RGBA{R: uint8(127), G: uint8(245), B: uint8(37), A: uint8(255)}
	meColor    = color.RGBA{R: uint8(19), G: uint8(112), B: uint8(24), A: uint8(255)}
	enemyColor = color.RGBA{R: uint8(184), G: uint8(37), B: uint8(20), A: uint8(255)}
	eatColor   = color.RGBA{R: uint8(20), G: uint8(50), B: uint8(184), A: uint8(255)}
)

type Table struct {
	data [][]*canvas.Rectangle
	side int
}

func (t Table) GetData(x, y int) *canvas.Rectangle {
	return t.data[x][y]
}

func (t Table) GetSize() (int, int) {
	return len(t.data), len(t.data[0])
}

func CreateTable(w, h int, x0, y0, side int) Table {
	tmp := make([][]*canvas.Rectangle, w)
	for i := range tmp {
		tmp[i] = make([]*canvas.Rectangle, h)
	}

	t := Table{data: tmp, side: side}

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			t.data[i][j] = createRect(i, j, x0, y0, side)
		}
	}

	return t
}

func createRect(x, y int, x0, y0, side int) *canvas.Rectangle {
	rec := canvas.NewRectangle(clearColor)
	rec.StrokeColor = color.Black
	rec.StrokeWidth = 0.
	tmp := fyne.Position{X: float32(x0 + x*side), Y: float32(y0 + y*side)}
	//fmt.Println(tmp)
	rec.Move(tmp)
	rec.Resize(fyne.Size{Width: float32(side), Height: float32(side)})
	return rec
}

func (t Table) SetClear(x, y int) {
	rec := t.data[x][y]
	rec.FillColor = clearColor
	rec.Refresh()
}

func (t Table) SetMe(x, y int) {
	rec := t.data[x][y]
	rec.FillColor = meColor
	rec.Refresh()
}

func (t Table) SetEnemy(x, y int) {
	rec := t.data[x][y]
	rec.FillColor = enemyColor
	rec.Refresh()
}

func (t Table) SetEat(x, y int) {
	rec := t.data[x][y]
	rec.FillColor = eatColor
	rec.Refresh()
}
