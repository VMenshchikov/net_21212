package snakeTable

import (
	"snake/model/proto"
)

// 0 - clean
// 1 - eat
// 2 - me
// 3 - enemy

const (
	CLEAR = byte(0)
	EAT   = byte(1)
	ME    = byte(2)
	ENEMY = byte(3)
)

type table struct {
	gameZone [][]byte
}

var (
	tab = table{}
)

func GetTable() table {
	return tab
}

func (t table) GetGameZone() [][]byte {
	return t.gameZone
}

func NewTable(w, h int) {
	t := table{
		gameZone: [][]byte{},
	}
	t.gameZone = make([][]byte, w)
	for i := range t.gameZone {
		t.gameZone[i] = make([]byte, h)
	}

	tab = t
}

func (t table) cleanTable() {
	for i := range t.gameZone {
		for j := range t.gameZone[i] {
			t.gameZone[i][j] = CLEAR
		}
	}
}

func abs(i int) int {
	if i >= 0 {
		return i
	} else {
		return -i
	}
}

func (t table) setSnakes(snakes []*proto.GameState_Snake, mySnakeId int) {
	for i := range snakes {
		snakeType := ENEMY
		if snakes[i].GetPlayerId() == int32(mySnakeId) {
			snakeType = ME
		}

		var cx, cy int
		for j, coord := range snakes[i].Points {
			if j == 0 {
				cx, cy = int(coord.GetX()), int(coord.GetY())
				t.gameZone[cx][cy] = snakeType

			} else {
				x, y := int(coord.GetX()), int(coord.GetY())
				var nx, ny, count int
				if x != 0 {
					count = abs(x)
					nx = x / count
					ny = 0
				} else {

					count = abs(y)
					nx = 0
					ny = y / count
				}
				for i := 0; i < count; i++ {
					cx = (cx + nx + len(t.gameZone)) % len(t.gameZone)
					cy = (cy + ny + len(t.gameZone[0])) % len(t.gameZone[0])
					t.gameZone[cx][cy] = snakeType
				}
			}
		}
	}
}

func (t table) setEat(eats []*proto.GameState_Coord) {
	for i := range eats {
		t.gameZone[int(eats[i].GetX())][int(eats[i].GetY())] = EAT
	}
}

func (t table) UpdateTable(state *proto.GameState, myId int) {
	t.cleanTable()
	t.setSnakes(state.Snakes, myId)
	t.setEat(state.Foods)
}
