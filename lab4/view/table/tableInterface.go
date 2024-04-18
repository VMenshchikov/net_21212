package table

import "fyne.io/fyne/v2/canvas"

type TableInterface interface {
	GetData(x, y int) *canvas.Rectangle
	GetSize() (int, int)
	SetClear(x, y int)
	SetMe(x, y int)
	SetEnemy(x, y int)
	SetEat(x, y int)
}
