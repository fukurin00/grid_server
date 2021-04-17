package main

import (
	"log"

	robot "github.com/fukurin00/grid_server/robot"
)

func main() {
	var yamlFile string = "../../map/willow_garage.yaml"
	var mapFile string = "../../map/willow_garage.pgm"

	rob := robot.NewRobot(1, 0.1, 1.5, 1.5, 0.5, mapFile, yamlFile)

	var over []int
	rx, ry, ok := rob.RGrid.AstarPlan(30, -10, 5, 5, over)

	log.Print(rx[0:5])
	log.Print(ry[0:5])
	log.Print(ok)
}
