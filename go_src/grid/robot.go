package grid

import (
	"encoding/json"
	"fmt"
	"log"

	msg "example.com/msg"
	sxmqtt "github.com/synerex/proto_mqtt"
)

var ()

type RobotStatus struct {
	Id   uint32
	Pose msg.PoseStamp
	Path msg.Path

	Radius   float64
	Velocity float64

	EstPose []PoseUnix

	Ox []float64
	Oy []float64
}

type PoseUnix struct {
	Pose msg.Pose
	Unix float64
}

func NewRobot(id uint32, radius, vel float64) *RobotStatus {
	s := new(RobotStatus)

	s.Id = id
	s.Radius = radius
	s.Velocity = vel
	return s
}

func (r *RobotStatus) UpdateVel(vel float64) {
	r.Velocity = vel
	r.calcPathTime()
}

func (r *RobotStatus) UpdateRadius(radius float64) {
	r.Radius = radius
}

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
	for _, pose := range r.Path.Poses {
		//distance from current pose
		dis := pose.Pose.Position.Distance(r.Pose.Pose.Position)
		elap := dis / r.Velocity //est elaps time

		estPose := PoseUnix{
			Pose: pose.Pose,
			Unix: pose.Header.Stamp.CalcUnix() + elap,
		}

		r.EstPose = append(r.EstPose, estPose)
	}
}

// func (r *RobotStatus) CalcGridArea(yamlFile string) {
// 	mapConfig := msg.ReadImageYaml(yamlFile)
// 	reso := mapConfig.Resolution
// 	origins := mapConfig.Origin

// }
