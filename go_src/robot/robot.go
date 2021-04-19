//package about robot status
package robot

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	grid "github.com/fukurin00/grid_server/grid"
	msg "github.com/fukurin00/grid_server/msg"
	"github.com/fukurin00/grid_server/synerex"
	tools "github.com/fukurin00/grid_server/tools"
	sxmqtt "github.com/synerex/proto_mqtt"
)

//RobotStatus robot information
type RobotStatus struct {
	Id        uint32                `json:"id"`
	GridReso  float64               `json:"gridResolution"`
	PoseStamp msg.ROS_PoseStamped   `json:"pose"`
	Path      []msg.ROS_PoseStamped `json:"path"`

	Radius    float64 `json:"radius"`
	Velocity  float64 `json:"velocity"`
	RotateVel float64 `json:"rotateVelocity"`

	EstPose   []GridPath `json:"estPose"`
	PathGrids []int      `json:"pathGrids"`

	RGrid *grid.Grid
}

type GridPath struct {
	Pose  msg.Pose      `json:"pose"`
	Grids []int         `json:"grids"`
	Stamp msg.TimeStamp `json:"stamp"`
}

// robotstatus constructor
func NewRobot(id uint32, radius, vel, rotVel, resolution float64, mapFile, yamlFile string) *RobotStatus {
	s := new(RobotStatus)

	s.Id = id
	s.Radius = radius
	s.Velocity = vel
	s.RotateVel = rotVel
	s.GridReso = resolution

	s.RGrid = grid.NewGrid(resolution)
	err := s.RGrid.ReadMapImage(yamlFile, mapFile)
	if err != nil {
		log.Print("readmap error", err)
	}
	s.RGrid.CalcObjMap(radius)
	return s
}

func (r *RobotStatus) UpdateVel(vel float64) {
	r.Velocity = vel
	r.calcPathMore()
}

func (r *RobotStatus) UpdateRadius(radius float64) {
	r.Radius = radius
}

//UpdatePose update robot pose
func (r *RobotStatus) UpdatePose(rcd *sxmqtt.MQTTRecord) {
	var pose msg.ROS_PoseStamped
	var id uint32

	err := json.Unmarshal(rcd.Record, &pose)
	if err != nil {
		log.Print(err)
	}
	fmt.Sscanf(rcd.Topic, "robot/pose/%d", &id)

	r.PoseStamp = pose

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

	r.Path = path.Poses
	if r.RGrid.Nodes != nil {
		r.calcPathMore()
	}
}

// a地点からb地点へ向かうときの回転角の計算
func calcRotateYaw(current msg.Pose, dest msg.Pose) float64 {
	destYaw := math.Atan2(dest.Position.Y-current.Position.Y, dest.Position.X-current.Position.X)
	currentYaw := 2 * math.Acos(current.Orientation.W)
	diffYaw := math.Mod(destYaw-currentYaw, math.Pi/2.0)
	return diffYaw
}

func (r *RobotStatus) calcPathMore() {
	r.EstPose = nil

	//add current_pose
	currentStamp := r.PoseStamp.Header.Stamp.ToF()
	firstPose := GridPath{
		Pose:  r.PoseStamp.Pose,
		Stamp: msg.FtoStamp(currentStamp),
		Grids: r.RGrid.CalcRobotGrid(r.PoseStamp.Pose.Position.X, r.PoseStamp.Pose.Position.Y, r.Radius),
	}
	r.EstPose = append(r.EstPose, firstPose)

	// search nearest pose
	minDis := 999.9
	minIndex := 0
	for i, pose := range r.Path {
		dis := r.PoseStamp.Pose.Position.Distance(pose.Pose.Position)
		if dis < minDis {
			minIndex = i
			minDis = dis
		}
	}

	prevPose := r.PoseStamp.Pose
	var allGrids []int //grids list of trajectory

	// calculate estimated passing time in all poses from nearest pose to last pose
	for _, pose := range r.Path[minIndex:] {
		// 回転分の時間を加える
		diffYaw := calcRotateYaw(prevPose, r.Path[minIndex].Pose)
		rotateTime := diffYaw / r.RotateVel
		currentStamp += rotateTime

		dis := prevPose.Position.Distance(pose.Pose.Position)
		elaptime := dis / r.Velocity
		currentStamp += elaptime
		grid := r.RGrid.CalcRobotGrid(pose.Pose.Position.X, pose.Pose.Position.Y, r.Radius)
		allGrids = append(allGrids, grid...)
		pos := msg.Pose{
			Position:    pose.Pose.Position,
			Orientation: pose.Pose.Orientation,
		}

		estPose := GridPath{
			Pose:  pos,
			Stamp: msg.FtoStamp(currentStamp),
			Grids: grid,
		}
		r.EstPose = append(r.EstPose, estPose)
		prevPose = pos

		//経路をさらにgridResolution間隔で細かくする
		/*
			dis := pose.Pose.Position.Distance(prevPose.Position)
			splitNum := int(math.Round(dis / r.GridReso))
			log.Print(j, "distance:", dis, " resolution:", r.GridReso, " splitNum:", splitNum)

			diffX := pose.Pose.Position.X - prevPose.Position.X
			diffY := pose.Pose.Position.Y - prevPose.Position.Y

			for i := 1; i <= splitNum; i++ {
				nx := prevPose.Position.X + diffX*float64(i)/float64(splitNum)
				ny := prevPose.Position.Y + diffY*float64(i)/float64(splitNum)
				nPosition := msg.Point{X: nx, Y: ny, Z: 0.0}
				splitDis := prevPose.Position.Distance(nPosition)
				splitTime := splitDis / r.Velocity
				currentStamp += splitTime
				grid := r.RGrid.CalcRobotGrid(nx, ny, r.Radius)
				allGrids = append(allGrids, grid...)
				pos := msg.Pose{
					Position:    nPosition,
					Orientation: pose.Pose.Orientation,
				}

				estPose := GridPath{
					Pose:  pos,
					Stamp: msg.FtoStamp(currentStamp),
					Grids: grid,
				}
				r.EstPose = append(r.EstPose, estPose)
				prevPose = pos
			}
		*/
	}

	pathGrid := tools.RemoveDuplicate(allGrids)
	r.PathGrids = pathGrid

	// デバッグ用：ファイルに書き込む
	f := fmt.Sprintf("test/robot%dtest.json", r.Id)
	m, err := json.MarshalIndent(r.EstPose, "", " ")
	if err != nil {
		log.Print(err)
	}
	tools.WriteFile(f, m)
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
	topic := fmt.Sprintf("/robot/cmd/stop/%d", r.Id)
	out := RobotMsg{
		Topic:   topic,
		Content: jm,
		Stamp:   from,
	}
	return out, nil
}

func (r RobotStatus) MakePathCmd(rx, ry []float64) (RobotMsg, error) {
	topic := fmt.Sprintf("robot/cmd/path/%d", r.Id)
	var poses []msg.ROS_PoseStamped

	for i := 0; i < len(rx); i++ {
		x := rx[i]
		y := ry[i]
		pos := msg.ROS_PoseStamped{
			Header: msg.ROS_header{Seq: uint32(i)},
			Pose: msg.Pose{
				Position: msg.Point{X: x, Y: y, Z: 0.0},
			},
		}
		poses = append(poses, pos)
	}

	planm := msg.Path{
		Header: msg.ROS_header{},
		Poses:  poses,
	}

	jm, err := json.MarshalIndent(planm, "", " ")
	if err != nil {
		return RobotMsg{}, err
	}
	out := RobotMsg{
		Topic:   topic,
		Content: jm,
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
	dups := tools.CheckDuplicate(a.PathGrids, b.PathGrids)
	if len(dups) > 0 {
		for _, pa := range a.EstPose {
			for _, pb := range b.EstPose {
				from := pa.Stamp.ToF() - span/2
				to := pa.Stamp.ToF() + span/2
				//time check
				if tools.CheckSpan(from, to, pb.Stamp.ToF()) {
					aOvers := pa.Grids
					bOvers := pb.Grids
					// log.Print(aOvers, bOvers)
					overs := tools.CheckDuplicate(aOvers, bOvers)

					pinfo := PathInfo{}
					if overs == nil {
						pinfo.Check = false
						return pinfo
					}
					if len(overs) > 0 {
						pinfo.Check = true
						pinfo.Grids = overs
						pinfo.From = msg.FtoStamp(from)
						pinfo.To = msg.FtoStamp(to)
						return pinfo
					}
				}
			}
		}
	}
	pinfo := PathInfo{}
	pinfo.Check = false
	return pinfo
}
