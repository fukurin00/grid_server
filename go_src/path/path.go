package path

import (
	"log"

	msg "github.com/fukurin00/grid_server/msg"
	robot "github.com/fukurin00/grid_server/robot"
	tools "github.com/fukurin00/grid_server/tools"
)

type PathInfo struct {
	Check bool //可能性あるか
	From  msg.TimeStamp
	To    msg.TimeStamp
	Grids []int //index list
}

// true: possible crush, false: no danger
func CheckRobotPath(a, b robot.RobotStatus, span float64) PathInfo {
	for _, pa := range a.EstPose {
		for _, pb := range b.EstPose {
			from := pa.Stamp.ToF() - span/2
			to := pa.Stamp.ToF() + span/2
			//time check
			if tools.CheckSpan(from, to, pb.Stamp.ToF()) {
				aOvers := a.RGrid.CalcRobotGrid(pa.Pose.Position.X, pa.Pose.Position.Y, a.Radius)
				bOvers := b.RGrid.CalcRobotGrid(pb.Pose.Position.X, pb.Pose.Position.Y, b.Radius)
				log.Print(aOvers, bOvers)
				overs := tools.CheckDuplicate(aOvers, bOvers)

				pinfo := PathInfo{}
				if overs == nil {
					pinfo.Check = false
					return pinfo
				}
				if len(overs) > 0 {

					pinfo.Grids = overs
					pinfo.From = msg.FtoStamp(from)
					pinfo.To = msg.FtoStamp(to)
					return pinfo
				}
			}
		}
	}
	pinfo := PathInfo{}
	pinfo.Check = false
	return pinfo
}
