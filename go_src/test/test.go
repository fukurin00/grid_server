package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	grid "github.com/fukurin00/grid_server/grid"
	robot "github.com/fukurin00/grid_server/robot"
	"github.com/fukurin00/grid_server/synerex"
	sxutil "github.com/synerex/synerex_sxutil"
)

func publish(rob *robot.RobotStatus, rx, ry []float64) {
	for {
		// 経路が成功すれば新しいpathを送る
		m, err := rob.MakePathCmd(rx, ry)
		if err != nil {
			log.Print(err)
		}
		log.Print("send new path command")
		err2 := robot.SendCmdRobot(m)
		if err2 != nil {
			log.Print(err2)
		}
		time.Sleep(time.Second)
	}
}

func main() {
	var yamlFile string = "../../map/willow_garage.yaml"
	var mapFile string = "../../map/willow_garage.pgm"

	sx := 13.0
	sy := -28.0
	gx := -4.0
	gy := 5.0

	rob := robot.NewRobot(1, 0.4, 1.5, 1.5, 0.5, mapFile, yamlFile)

	var over []int
	rx, ry, _ := grid.AstarPlan(rob.RGrid, sx, sy, gx, gy, over)

	sxi := rob.RGrid.XyIndex(sx, rob.RGrid.MinX)
	syi := rob.RGrid.XyIndex(sy, rob.RGrid.MinY)
	sind := syi*rob.RGrid.XWidth + sxi

	gxi := rob.RGrid.XyIndex(gx, rob.RGrid.MinX)
	gyi := rob.RGrid.XyIndex(gy, rob.RGrid.MinY)
	gind := gyi*rob.RGrid.XWidth + gxi

	for i := 0; i < rob.RGrid.MaxIndex; i++ {
		if i == sind {
			fmt.Print("S")
		} else if i == gind {
			fmt.Print("G")
		} else {
			x, y := rob.RGrid.CalcPosition(i)
			ook := false
			for j := 0; j < len(rx); j++ {
				if x == rx[j] && y == ry[j] {
					fmt.Print("+")
					ook = true
				}
			}
			if !ook {
				if rob.RGrid.Nodes[i].Obj {
					fmt.Print("*")
				} else {
					fmt.Print(".")
				}
				if i%rob.RGrid.XWidth == rob.RGrid.XWidth-1 {
					fmt.Println()
				}
			}
		}
	}
	fmt.Println()

	wg := sync.WaitGroup{} //wait exit for gorouting
	wg.Add(1)
	synerex.RunSynerex()
	publish(rob, rx, ry)
	wg.Wait()
	sxutil.CallDeferFunctions()
}
