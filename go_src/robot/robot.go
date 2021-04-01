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
	tools "github.com/fukurin00/grid_server/tools"
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

	EstPose []GridPath

	RGrid *grid.Grid
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

		estPose := GridPath{
			Pose:  pose.Pose,
			Stamp: msg.FtoStamp(uni),
			Grids: r.RGrid.CalcRobotGrid(r.Pose.Pose.Position.X, r.Pose.Pose.Position.Y, r.Radius),
		}

		r.EstPose = append(r.EstPose, estPose)
		currentPose = pose.Pose
	}
}

//send stop command
func (r RobotStatus) MakeStopCmd(from, to msg.TimeStamp) (RobotMsg, error) {
	m := msg.Stop{
		Header: msg.ROS_header{
			Stamp:    msg.CalcStamp(time.Now()),
			Frame_id: fmt.Sprint(r.Id),
		},
		From: from,
		To:   to,
	}
	jm, err := json.Marshal(m)
	if err != nil {
		return RobotMsg{}, err
	}
	topic := fmt.Sprintf("/robot/stop/%d", r.Id)
	out := RobotMsg{
		Topic:   topic,
		Content: jm,
		Stamp:   from,
	}
	return out, nil
}

type RobotMsg struct {
	Stamp   msg.TimeStamp //time to sending
	Topic   string        //topic
	Content []byte        //json content
}

func SendCmdRobot(m RobotMsg) error {
	opt := synerex.GeneMqttSupply(m.Topic, m.Content)
	_, merr := synerex.Mqttclient.NotifySupply(opt)
	if merr != nil {
		return merr
	}
	return nil
}

type PathInfo struct {
	Check bool //可能性あるか
	From  msg.TimeStamp
	To    msg.TimeStamp
	Grids []int //index list
}

// true: possible crush, false: no danger
func CheckRobotPath(a, b RobotStatus, span float64) PathInfo {
	for _, pa := range a.EstPose {
		for _, pb := range b.EstPose {
			from := pa.Stamp.ToF() - span/2
			to := pa.Stamp.ToF() + span/2
			//time check
			if tools.CheckSpan(from, to, pb.Stamp.ToF()) {
				aOvers := pa.Grids
				bOvers := pb.Grids
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
