package view

import (
	"fmt"
	"net"
	"time"

	model "snake/controller/modelC"

	"snake/model/proto"
	table "snake/view/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	WIN_W        = 2 * WIN_H
	WIN_H        = 600
	START_COLUMN = 3.
	START_ROWS   = 4.
	TABLE_IDENT  = 10.
)

var (
	appl           = app.New()
	mainWindow     fyne.Window
	gameTable      table.TableInterface
	infoSection    *fyne.Container
	tableContainer *fyne.Container

	raiting       *widget.TextGrid //
	exitButton    *widget.Button   //
	newGameButton *widget.Button   //
	currentGame   *widget.TextGrid //
	anouncesGames = make([]*container.TabItem, 0, 10)
	gridAnounce   = container.NewDocTabs()
)

func pushExitButtom() {}

func pushNewGameButton() {
	w, h, food := 50, 50, 5
	dur := time.Millisecond * 100
	model.NewGame(w, h, food, dur)

	gameTable = NewTable(w, h)
	updateTable()
	tableContainer.Refresh()
	model.UpdateTable(w, h)

	wString := fmt.Sprintln("Ширина: ", w)
	hString := fmt.Sprintln("Высота: ", h)
	foodString := fmt.Sprintf("Количество еды: %d+x1", food)
	durString := fmt.Sprintln("Задержка:", dur)
	currentGame.SetText(wString + hString + foodString + "\n" + durString)
	currentGame.Refresh()

}

func createRaitSection(localW, localH int) *fyne.Container {
	lableRating := widget.NewLabel("Рейтинг")
	lableRating.Move(fyne.NewPos(5, 0))
	lableRating.Resize(fyne.NewSize(float32(localW), float32(localH)/8))

	raiting = widget.NewTextGrid()
	raiting.SetText("1\n2\n3\n4\n5\n6\n7\n8\n9\n0\n1\n2\n4")
	raitingScroll := container.NewVScroll(raiting)
	raitingScroll.Move(fyne.NewPos(5, lableRating.Size().Height+5))
	raitingScroll.Resize(fyne.Size{Width: float32(localW - 10), Height: float32(localH-10) * 0.7})

	rPos := raitingScroll.Position()
	rSize := raitingScroll.Size()
	buttonPos := fyne.NewPos(5, rPos.Y+rSize.Height+5)
	buttonSize := fyne.NewSize((float32(localW)/2)-7.5, float32(localH)-buttonPos.Y-5)

	exitButton = widget.NewButton("Выход", pushExitButtom)
	exitButton.Move(buttonPos)
	exitButton.Resize(buttonSize)

	newGameButton = widget.NewButton("Новая игра", pushNewGameButton)
	newGameButton.Resize(buttonSize)
	newGameButton.Move(buttonPos.AddXY(buttonSize.Width+5, 0))

	return container.NewWithoutLayout(lableRating, raitingScroll, exitButton, newGameButton)
}

func createCurrentGameSection(localW, localH int) *fyne.Container {
	label := widget.NewLabel("Текущая игра")
	label.Move(fyne.NewPos(5, 0))

	currentGame = widget.NewTextGrid()
	currentGame.Move(label.Position().AddXY(0, label.MinSize().Height+5))
	currentGame.Resize(fyne.NewSize(float32(localW-10), float32(float64(localH)*0.7)))
	//currentGame.SetText("1\n2\n3\n4\n5\n6\n7")

	return container.NewWithoutLayout(label, currentGame)
}

type Anounce struct {
	Can_join                  bool
	Name                      string
	W, H                      int
	Food_static, Food_dynamic int
	RemoteAddres              *net.UDPAddr
	Duration                  time.Duration
	//players int,
}

func UpdateStats(str string) {
	raiting.SetText(str)
	raiting.Refresh()
}

func UpdateAnounces(anounces []Anounce) {
	anouncesGames = make([]*container.TabItem, 0, len(anounces))
	for _, v := range anounces {
		rGrid := container.NewGridWithRows(5,
			widget.NewLabel(v.Name),
			widget.NewLabel("Владелец: "+v.RemoteAddres.String()),
			widget.NewLabel(fmt.Sprintf("Размер: %dx%d", v.W, v.H)),
			widget.NewLabel(fmt.Sprintf("Возможность подключиться: %t", v.Can_join)),
			widget.NewLabel(fmt.Sprintf("Еда: %d+%dx", v.Food_static, v.Food_dynamic)),
		)

		cGrid := container.NewGridWithColumns(2,
			rGrid,
			widget.NewButton("Play", func() {
				gameTable = NewTable(v.W, v.H)
				updateTable()
				tableContainer.Refresh()
				model.UpdateTable(v.W, v.H)
				model.SendJoin(v.Name, v.RemoteAddres, time.Duration(v.Duration), v.W, v.H, v.Food_static)

				wString := fmt.Sprintln("Ширина: ", v.W)
				hString := fmt.Sprintln("Высота: ", v.H)
				foodString := fmt.Sprintf("Количество еды: %d+x1", v.Food_static)
				durString := fmt.Sprintln("Задержка:", v.Duration)
				currentGame.SetText(wString + hString + foodString + "\n" + durString)
				currentGame.Refresh()
			}),
		)

		anouncesGames = append(anouncesGames, container.NewTabItem(v.Name, cGrid))
		//fmt.Println(anouncesGames, "\n", gridAnounce)
	}
	gridAnounce.SetItems(anouncesGames)
	fmt.Println("Анонсы обновлены")
}

func createInfoSection() *fyne.Container {
	localW := WIN_W / 4
	localH := WIN_H / 2

	//секция с рейтингом и кнопками

	a := container.NewGridWithRows(2,
		container.NewGridWithColumns(2,
			createRaitSection(localW, localH),
			createCurrentGameSection(localW, localH),
		),
		gridAnounce,
	)
	return a
}

func getSide(x, y int) int {
	xSide := (WIN_W/2 - TABLE_IDENT*2) / x
	ySide := (WIN_H - TABLE_IDENT*2) / y
	tmp := min(xSide, ySide)

	return tmp - tmp%5
}

func getNullPos(x, y int) (int, int) {
	s := getSide(x, y)
	if x < y {
		return (WIN_W/2 - s*x) / 2, TABLE_IDENT
	} else {
		return TABLE_IDENT, (WIN_H - s*y) / 2
	}
}

func NewTable(x, y int) table.TableInterface {
	x0, y0 := getNullPos(x, y)
	return table.CreateTable(x, y, x0, y0, getSide(x, y))
}

func CreateApp(x, y int) {
	mainWindow = appl.NewWindow("Snake")
	mainWindow.Resize(fyne.Size{Width: WIN_W, Height: WIN_H})
	mainWindow.SetFixedSize(true)
	//x0, y0 := getNullPos(x, y)
	gameTable = NewTable(x, y)

	infoSection = createInfoSection()

	tableContainer = container.NewWithoutLayout()
	updateTable()

	tmp := container.NewGridWithColumns(2,
		tableContainer,
		infoSection,
	)

	mainWindow.SetContent(tmp)

	mainWindow.Canvas().SetOnTypedKey(func(event *fyne.KeyEvent) {
		switch event.Name {
		case fyne.KeyW:
			{
				model.SendSteer(proto.Direction_UP.Enum())
			}
		case fyne.KeyA:
			{
				model.SendSteer(proto.Direction_LEFT.Enum())
			}
		case fyne.KeyS:
			{
				model.SendSteer(proto.Direction_DOWN.Enum())
			}
		case fyne.KeyD:
			{
				model.SendSteer(proto.Direction_RIGHT.Enum())
			}
		}
	})

	mainWindow.ShowAndRun()
}

func updateTable() {
	tableContainer.Objects = nil
	x, y := gameTable.GetSize()
	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			tableContainer.Add(gameTable.GetData(i, j))
		}
	}
}

func SetEnemy(x, y int) {
	gameTable.SetEnemy(x, y)
}

func SetClear(x, y int) {
	gameTable.SetClear(x, y)
}

func SetMe(x, y int) {
	gameTable.SetMe(x, y)
}

func SetEat(x, y int) {
	gameTable.SetEat(x, y)
}
