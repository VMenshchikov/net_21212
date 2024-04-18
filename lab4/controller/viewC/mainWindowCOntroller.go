package view

import (
	"fmt"
	stats "snake/controller/stats"
	"snake/model/anouncement"
	snaketable "snake/model/snakeTable"
	"snake/view"

	"time"
)

var (
	closeCh = make(chan struct{})
)

func updateAnounces() {
	ans := anouncement.GetAllAnounces()
	var vans []view.Anounce
	for _, v := range ans {
		vans = append(vans, view.Anounce{
			Can_join:     v.An.Can_join,
			Name:         v.An.Name,
			W:            v.An.W,
			H:            v.An.H,
			Food_static:  v.An.Food_static,
			Food_dynamic: v.An.Food_dynamic,
			RemoteAddres: &v.Addr,
			Duration:     v.An.Duration,
		})
	}

	view.UpdateAnounces(vans)

}

func StartUpdateAnounces() {
	for {
		select {
		case <-closeCh:
			break
		case <-time.NewTimer(time.Second).C:
			{
				updateAnounces()
			}
		}
	}
}

func StopUpdateAnounces() {
	close(closeCh)
}

func PaintTable(ch chan [][]byte) {
	for {
		tableModel, ok := <-ch
		if ok == false {
			break
		}
		for i := range tableModel {
			for j := range tableModel[i] {
				switch tableModel[i][j] {
				case snaketable.CLEAR:
					{
						view.SetClear(i, j)
					}
				case snaketable.ME:
					{
						view.SetMe(i, j)
					}
				case snaketable.ENEMY:
					{
						view.SetEnemy(i, j)
					}
				case snaketable.EAT:
					{
						view.SetEat(i, j)
					}
				}

			}
		}
	}
}

func UpdateStats(ch chan []stats.Stats) {
	for {
		newStats, ok := <-ch
		if !ok {
			return
		}
		var raitStats []stats.Stats

		for _, v := range newStats {
			raitStats = append(raitStats, v)
			/* if len(raitStats) == 0 {
				raitStats = append(raitStats, v)
			} else {
				for i := 0; i < len(newStats); i++ {
					if v.Score > raitStats[i].Score {
						raitStats = append(raitStats[0:i], append([]stats.Stats{v}, raitStats[i+1:]...)...)
						break
					}
					if i == len(raitStats)-1 {
						raitStats = append(raitStats, v)
					}
				}
			} */
		}

		resString := ""
		for _, v := range raitStats {
			resString += fmt.Sprintln(v.Name, v.Score)
		}
		view.UpdateStats(resString)
	}
}
