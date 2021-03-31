//package about robot status
package robot

import (
	"encoding/json"
	"fmt"
	"log"

	grid "github.com/fukurin00/grid_server/grid"
	msg "github.com/fukurin00/grid_server/msg"
	sxmqtt "github.com/synerex/proto_mqtt"
)

var (
	gridReso float64 = 0.5
)

//RobotStatus robot information
type RobotStatus struct {
	Id   uint32
	Pose msg.PoseStamp
	Path msg.Path

	Radius   float64
	Velocity float64

	EstPose []PoseUnix

	RGrid *grid.Grid
}

type PoseUnix struct {
	Pose msg.Pose `json:"pose"`
	Unix float64  `json:"unix"`
}

// robotstatus constructor
func NewRobot(id uint32, radius, vel float64, mapFile, yamlFile string) *RobotStatus {
	s := new(RobotStatus)

	s.Id = id
	s.Radius = radius
	s.Velocity = vel

	s.RGrid = grid.NewGrid(gridReso)
	err := s.RGrid.ReadMapImage(yamlFile, mapFile)
	if err != nil {
		log.Print("readmap error", err)
	}
	s.RGrid.CalcObjMap(radius)
	return s
}

func (r *RobotStatus) UpdateVel(vel float64) {
	r.Velocity = vel
	r.calcPathTime()
}

func (r *RobotStatus) UpdateRadius(radius float64) {
	r.Radius = radius
}

//UpdatePose update robot pose
func (r *RobotStatus) UpdatePose(rcd *sxmqtt.MQTTRecord) {
	var pose msg.PoseStamp
	var id uint32

	err := json.Unmarshal(rcd.Record, &pose)
	if err != nil {
		log.Print(err)
	}
	fmt.Sscanf(rcd.Topic, "robot/pose/%d", &id)

	r.Pose = pose

}

//UpdatePath update robot path
func (r *RobotStatus) UpdatePath(rcd *sxmqtt.MQTTRecord) {
	var path msg.Path
	var id uint32

	err := json.Unmarshal(rcd.Record, &path)
	if err != nil {
		log.Print(err)
	}
	fmt.Sscanf(rcd.Topic, "robot/path/%d", &id)

	r.Path = path
	r.calcPathTime()
}

func (r *RobotStatus) calcPathTime() {
	r.EstPose = nil
	currentPose := r.Pose.Pose
	for _, pose := range r.Path.Poses {
		//distance from current pose
		dis := pose.Pose.Position.Distance(currentPose.Position)
		elap := dis / r.Velocity //est elaps time

		estPose := PoseUnix{
			Pose: pose.Pose,
			Unix: pose.Header.Stamp.CalcUnix() + elap,
		}

		r.EstPose = append(r.EstPose, estPose)
		currentPose = pose.Pose
	}
}

// func (r *RobotStatus) CalcGridArea(yamlFile string) {
// 	mapConfig := msg.ReadImageYaml(yamlFile)
// 	reso := mapConfig.Resolution
// 	origins := mapConfig.Origin

// }
