//package about robot status
package robot

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	grid "github.com/fukurin00/grid_server/grid"
	msg "github.com/fukurin00/grid_server/msg"
	"github.com/fukurin00/grid_server/synerex"
	sxmqtt "github.com/synerex/proto_mqtt"
)

var (
	gridReso float64 = 0.5
)

//RobotStatus robot information
type RobotStatus struct {
	Id   uint32
	Pose msg.ROS_Pose
	Path msg.Path

	Radius   float64
	Velocity float64

	EstPose []PoseStamp

	RGrid *grid.Grid
}

type PoseStamp struct {
	Pose  msg.Pose
	Stamp msg.TimeStamp
}

type GridPath struct {
	Pose  msg.Pose
	Grids []int
	Stamp msg.TimeStamp
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
	var pose msg.ROS_Pose
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
		uni := pose.Header.Stamp.ToF() + elap

		estPose := PoseStamp{
			Pose:  pose.Pose,
			Stamp: msg.FtoStamp(uni),
		}

		r.EstPose = append(r.EstPose, estPose)
		currentPose = pose.Pose
	}
}

//send stop command
func (r RobotStatus) SendStopCmd(from, to time.Time) error {
	m := msg.Stop{
		Header: msg.ROS_header{
			Stamp:    msg.CalcStamp(time.Now()),
			Frame_id: fmt.Sprint(r.Id),
		},
		From: msg.CalcStamp(from),
		To:   msg.CalcStamp(to),
	}
	jm, err := json.Marshal(m)
	if err != nil {
		return err
	}
	topic := fmt.Sprintf("/robot/stop/%d", r.Id)
	opt := synerex.GeneMqttSupply(topic, jm)
	_, merr := synerex.Mqttclient.NotifySupply(opt)
	if merr != nil {
		return merr
	}
	return nil
}
