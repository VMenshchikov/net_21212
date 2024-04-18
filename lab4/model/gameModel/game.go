package gamemodel

import (
	"fmt"
	"math"
	mrand "math/rand"
	"net"
	"snake/model/proto"
	"sync"

	err "errors"
)

type GameStruct struct {
	snakes      []snake
	foods       []food
	players     []player
	stateOrder  int32
	static_food int
	w, h        int

	lock sync.Mutex
}

type coord struct {
	X, Y int
}
type snake struct {
	status *proto.GameState_Snake_SnakeState
	id     int32
	coords []coord
	dir    proto.Direction
}

type food struct {
	c     coord
	eated bool
}

type player struct {
	name  string
	id    int32
	addr  *net.UDPAddr
	prole *proto.PlayerType
	nrole *proto.NodeRole
	score int32
}

func CreateNew(gw, gh, gstatic int, hostAddr *net.UDPAddr) *GameStruct {
	gs := GameStruct{
		snakes:      []snake{},
		foods:       []food{},
		players:     []player{},
		stateOrder:  0,
		static_food: gstatic,
		w:           gw,
		h:           gh,
		lock:        sync.Mutex{},
	}
	return &gs
}

func CreateWithGamestate(gw, gh, gstatic int, hostAddr *net.UDPAddr, state *proto.GameMessage_StateMsg) *GameStruct {

	f := make([]food, 0)
	for _, v := range state.State.GetFoods() {
		f = append(f, food{
			c: coord{
				X: int(v.GetX()),
				Y: int(v.GetY())},
			eated: false,
		},
		)

	}

	players := make([]player, 0)
	for _, v := range state.GetState().GetPlayers().GetPlayers() {

		pAddr := *&net.UDPAddr{
			IP:   net.ParseIP(*v.IpAddress),
			Port: int(v.GetPort()),
			Zone: "",
		}
		players = append(players, player{
			name:  v.GetName(),
			id:    v.GetId(),
			addr:  &pAddr,
			prole: v.GetType().Enum(),
			nrole: v.GetRole().Enum(),
			score: *v.Score,
		})
	}

	snakes := make([]snake, 0)

	for _, v := range state.GetState().GetSnakes() {
		snakeCoords := make([]coord, 0)
		for i, k := range v.GetPoints() {
			if i == 0 {
				snakeCoords = append(snakeCoords, coord{
					X: int(k.GetX()),
					Y: int(k.GetY()),
				})
			} else {
				snakeCoords = append(snakeCoords, coord{
					X: snakeCoords[len(snakeCoords)-1].X + int(k.GetX()),
					Y: snakeCoords[len(snakeCoords)-1].Y + int(k.GetY()),
				})
			}
		}
		snakes = append(snakes, snake{
			status: v.GetState().Enum(),
			id:     v.GetPlayerId(),
			coords: snakeCoords,
			dir:    v.GetHeadDirection(),
		})
	}

	gs := GameStruct{
		snakes:      snakes,
		foods:       f,
		players:     players,
		stateOrder:  state.State.GetStateOrder(),
		static_food: gstatic,
		w:           gw,
		h:           gh,
		lock:        sync.Mutex{},
	}

	return &gs
}

func (prev *GameStruct) CreateCopy() *GameStruct {
	snakesCopy := make([]snake, len(prev.snakes))
	copy(snakesCopy, prev.snakes)

	foodsCopy := make([]food, len(prev.foods))
	copy(foodsCopy, prev.foods)

	playersCopy := make([]player, len(prev.players))
	copy(playersCopy, prev.players)

	return &GameStruct{
		snakes:      snakesCopy,
		foods:       foodsCopy,
		players:     playersCopy,
		stateOrder:  prev.stateOrder,
		static_food: prev.static_food,
		w:           prev.w,
		h:           prev.h,
		lock:        sync.Mutex{},
	}
}

//func CreateWithState(state *proto.GameState) *GameStruct {}

func (g *GameStruct) GetWidth() *int32 {
	tmp := int32(g.w)
	return &tmp
}
func (g *GameStruct) GetHeight() *int32 {
	tmp := int32(g.h)
	return &tmp
}
func (g *GameStruct) GetFood() *int32 {
	tmp := int32(g.static_food)
	return &tmp
}

func abs(x int) int {
	if x >= 0 {
		return x
	} else {
		return -x
	}
}

func (g *GameStruct) GetSnakes() []*proto.GameState_Snake {
	snakes := make([]*proto.GameState_Snake, 0, len(g.snakes))

	for _, s := range g.snakes {
		snake := make([]*proto.GameState_Coord, 0)
		prev := coord{}
		for _, e := range s.coords {
			if len(snake) == 0 {
				x, y := int32(e.X), int32(e.Y)
				snake = append(snake, &proto.GameState_Coord{
					X: &x,
					Y: &y,
				})
			} else {

				tmp := coord{
					X: (e.X - prev.X + g.w) % g.w,
					Y: (e.Y - prev.Y + g.h) % g.h,
				}

				if abs(tmp.X) != 1 {
					tmp.X = -tmp.X / (g.w - 1)
				}
				if abs(tmp.Y) != 1 {
					tmp.Y = -tmp.Y / (g.h - 1)
				}
				x, y := int32(tmp.X), int32(tmp.Y)
				snake = append(snake, &proto.GameState_Coord{
					X: &x,
					Y: &y,
				})

			}
			prev = e

		}
		tmpID, tmpDir := s.id, s.dir
		snakes = append(snakes, &proto.GameState_Snake{
			PlayerId:      &tmpID,
			Points:        snake,
			State:         proto.GameState_Snake_ALIVE.Enum(),
			HeadDirection: &tmpDir,
		})
	}

	return snakes
}

func (g *GameStruct) GetFoods() []*proto.GameState_Coord {
	foods := make([]*proto.GameState_Coord, 0, len(g.foods))
	for _, f := range g.foods {
		x, y := int32(f.c.X), int32(f.c.Y)
		foods = append(foods, &proto.GameState_Coord{
			X: &x,
			Y: &y,
		})
	}
	return foods
}

func (g *GameStruct) GetPlayers() *proto.GamePlayers {
	players := make([]*proto.GamePlayer, 0, len(g.players))

	for i := range g.players {
		tmpAddr := g.players[i].addr.IP.String()
		tmpPort := int32(g.players[i].addr.Port)
		players = append(players, &proto.GamePlayer{
			Name:      &g.players[i].name,
			Id:        &g.players[i].id,
			IpAddress: &tmpAddr,
			Port:      &tmpPort,
			Role:      (*proto.NodeRole)(g.players[i].nrole),
			Type:      proto.PlayerType_HUMAN.Enum(),
			Score:     &g.players[i].score,
		})
	}
	return &proto.GamePlayers{
		Players: players,
	}
}

func (g *GameStruct) GetOrder() *int32 {
	return &g.stateOrder
}

func equalCoords(x coord, y coord) bool {
	return (x.X == y.X && x.Y == y.Y)
}

func (g *GameStruct) NextStep() {
	g.lock.Lock()
	defer g.lock.Unlock()

	//двигаем змейку и проверяем съеденную еду
	for i := range g.snakes {
		var newHead coord
		head := g.snakes[i].coords[0]
		//fmt.Println(g.w, g.h)
		switch g.snakes[i].dir {
		case proto.Direction_DOWN:
			{
				newHead = coord{
					X: head.X,
					Y: (head.Y + 1 + g.h) % g.h,
				}
			}
		case proto.Direction_LEFT:
			{
				newHead = coord{
					X: (head.X - 1 + g.w) % g.w,
					Y: head.Y,
				}
			}
		case proto.Direction_RIGHT:
			{
				newHead = coord{
					X: (head.X + 1 + g.w) % g.w,
					Y: head.Y,
				}
			}
		case proto.Direction_UP:
			{
				newHead = coord{
					X: head.X,
					Y: (head.Y - 1 + g.h) % g.h,
				}
			}
		}
		eated := false
		for ifood := range g.foods {
			if equalCoords(newHead, g.foods[ifood].c) {
				//добавляем очки
				for ind := range g.players {
					if g.snakes[i].id == int32(g.players[ind].id) {
						g.players[ind].score++
					}
				}

				g.foods[ifood].eated = true
				g.snakes[i].coords = append([]coord{newHead}, g.snakes[i].coords...)
				eated = true
				break
			}
			// append([]type{el}, slice[:len-1]...)
		}
		if !eated {
			g.snakes[i].coords = append([]coord{newHead}, g.snakes[i].coords[:len(g.snakes[i].coords)-1]...)
			//g.snakes[i].coords = append(g.snakes[i].coords[1:], newHead)

		}
	}

	//удаляем еду

	newFoods := []food{}
	for i := range g.foods {
		if !g.foods[i].eated {
			newFoods = append(newFoods, g.foods[i])
		}
	}
	g.foods = newFoods

	//удаляем самоубийц
	newSnakes := []snake{}
	for s1i := range g.snakes {
		head := g.snakes[s1i].coords[0]
		death := false
		for s2i, s2 := range g.snakes {
			if death {
				break
			}
			for bi, b := range s2.coords {
				if s1i == s2i && bi == 0 {
					continue
				}
				if equalCoords(b, head) {
					death = true
					// еда вместо туши
					for _, coord := range g.snakes[s1i].coords {
						if mrand.Intn(2) == 1 {
							g.foods = append(g.foods, food{
								c:     coord,
								eated: false,
							})
						}

					}
					break
				}
			}
		}
		if !death {
			newSnakes = append(newSnakes, g.snakes[s1i])
		}
	}
	g.snakes = newSnakes

	//добавляем еду
	for len(g.foods) < g.static_food+len(g.snakes) {
		f := coord{
			X: mrand.Intn(g.w),
			Y: mrand.Intn(g.h),
		}
		clear := true
		for _, s := range g.snakes {
			if !clear {
				break
			}
			for _, b := range s.coords {
				if equalCoords(f, b) {
					clear = false
					break
				}
			}

		}
		for _, food := range g.foods {
			if equalCoords(f, food.c) {
				clear = false
				break
			}

		}

		if clear {
			g.foods = append(g.foods, food{
				c:     f,
				eated: false,
			})
		}
	}

	g.stateOrder++

}

func (g *GameStruct) UpdateDir(prevG *GameStruct, id int, dir *proto.Direction) {
	g.lock.Lock()
	defer g.lock.Unlock()

	pl := -1
	for i, p := range prevG.snakes {
		if p.id == int32(id) {
			pl = i
			break
		}
	}
	if pl == -1 {
		return
	}
	prevDir := prevG.snakes[pl].dir

	sum := *dir + prevDir
	if sum == 3 || sum == 7 {
		return
	}

	g.snakes[pl].dir = *dir
}

func (g *GameStruct) AddPlayer(pname string, nrole *proto.NodeRole, addr *net.UDPAddr, cid *int32) int32 {
	pid := int32(0)
	if cid == nil {
		for _, p := range g.players {
			pid = max(pid, p.id) + 1
		}
	} else {
		pid = *cid
	}
	g.players = append(g.players, player{
		name:  pname,
		id:    pid,
		addr:  addr,
		prole: proto.PlayerType_HUMAN.Enum(),
		nrole: nrole,
		score: 0,
	})

	fmt.Println("ИГрок добавлен")

	return pid

	//if (prole != proto.)
}

func (g *GameStruct) AddSnake(sid int32) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	cx, cy := 2, 2
	done := false
	if len(g.snakes) == 0 {
		cx = mrand.Intn(g.w-1) + 1
		cy = mrand.Intn(g.h-1) + 1
		done = true
	}
	for ; cx < g.w-2 && !done; cx++ {
		for ; cy < g.h-2 && !done; cy++ {
			for _, s := range g.snakes {
				if done {
					break
				}
				for _, b := range s.coords {
					if int(math.Abs(float64(b.X-cx))) > 2 && int(math.Abs(float64(b.Y-cy))) > 2 {
						done = true
						break
					}
				}
			}
		}
	}

	if !done {
		fmt.Println("НЕТ МЕСТА")
		return err.New("Нет места")
	}
	newDir := mrand.Intn(4) + 1
	secCoord := coord{
		X: cx,
		Y: cy,
	}
	switch newDir {
	case 1:
		{
			secCoord.X--
		}
	case 2:
		{
			secCoord.X++
		}
	case 3:
		{
			secCoord.Y++
		}
	case 4:
		{
			secCoord.Y--
		}
	}
	g.snakes = append(g.snakes, snake{
		status: proto.GameState_Snake_ALIVE.Enum(),
		id:     int32(sid),
		coords: []coord{{
			X: cx,
			Y: cy,
		}, secCoord},
		dir: proto.Direction(newDir),
	})

	fmt.Println("Змейка добавлена")
	return nil
}
