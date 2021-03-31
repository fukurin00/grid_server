package path

import (
	"log"

	robot "github.com/fukurin00/grid_server/robot"
	tools "github.com/fukurin00/grid_server/tools"
)

// true: possible crush, false: no danger
func CheckRobotPath(a, b robot.RobotStatus, span float64) bool {
	for _, pa := range a.EstPose {
		for _, pb := range b.EstPose {
			//time check
			if tools.CheckSpan(pa.Unix-span/2, pa.Unix+span/2, pb.Unix) {
				aOvers := a.RGrid.CalcRobotGrid(pa.Pose.Position.X, pa.Pose.Position.Y, a.Radius)
				bOvers := b.RGrid.CalcRobotGrid(pb.Pose.Position.X, pb.Pose.Position.Y, b.Radius)
				log.Print(aOvers, bOvers)
				overs := tools.CheckDuplicate(aOvers, bOvers)

				if overs == nil {
					return false
				}
				if len(overs) > 0 {
					return true
				}
			}
		}
	}
	return false
}
