package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	grid "github.com/fukurin00/grid_server/grid"
	msg "github.com/fukurin00/grid_server/msg"
	robot "github.com/fukurin00/grid_server/robot"
	synerex "github.com/fukurin00/grid_server/synerex"
	sxmqtt "github.com/synerex/proto_mqtt"
	api "github.com/synerex/synerex_api"
	sxutil "github.com/synerex/synerex_sxutil"
	"google.golang.org/protobuf/proto"
)

var (
	robotList map[int]*robot.RobotStatus // robot list
	yamlFile  string                     = "../map/willow_garage.yaml"
	mapFile   string                     = "../map/willow_garage.pgm"
	span      float64                    = 10 //crush check span
	//mode      string                     = "reroute"
)

func supplyMQTTCallback(clt *sxutil.SXServiceClient, sp *api.Supply) {
	//from MQTT broker
	if sp.SenderId == uint64(clt.ClientID) {
		// ignore my message.
		return
	}

	rcd := &sxmqtt.MQTTRecord{}
	err := proto.Unmarshal(sp.Cdata.Entity, rcd)
	if err == nil {
		if strings.HasPrefix(rcd.Topic, "robot/") {
			if strings.HasPrefix(rcd.Topic, "robot/path") {
				var p msg.Path
				var id int

				err := json.Unmarshal(rcd.Record, &p)
				if err != nil {
					log.Print(err)
				}
				fmt.Sscanf(rcd.Topic, "robot/path/%d", &id)
				// log.Print("get ", rcd.Topic)
				// log.Print(p)

				if rob, ok := robotList[id]; ok {
					rob.UpdatePath(rcd)
					for key, val := range robotList {
						if key != id && len(val.EstPose) > 0 { //multi robot already had path
							// log.Print("check path ", id, " and ", key)
							out := robot.CheckRobotPath(*rob, *val, span)

							//衝突の可能性がある場合は再計算
							if out.Check {
								// かぶる場所をふさいで再計算
								log.Print("start new path planning", key, " and ", id)
								rx, ry, ok := grid.AstarPlan(rob.RGrid, rob.PoseStamp.Pose.Position.X, rob.PoseStamp.Pose.Position.Y, rob.Path[len(rob.Path)-1].Pose.Position.X, rob.Path[len(rob.Path)-1].Pose.Position.Y, out.Grids)

								if ok {
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
								} else {
									//失敗すれば停止司令を送る
									m, err := rob.MakeStopCmd(out.From, out.To)
									if err != nil {
										log.Print(err)
									}
									log.Print("send stop command")
									err2 := robot.SendCmdRobot(m)
									if err2 != nil {
										log.Print(err2)
									}
								}
							}
						}
					}

				}

			} else if strings.HasPrefix(rcd.Topic, "robot/pose") {
				var pose msg.ROS_PoseStamped
				var id int

				err := json.Unmarshal(rcd.Record, &pose)
				if err != nil {
					log.Print(err)
				}
				fmt.Sscanf(rcd.Topic, "robot/pose/%d", &id)

				// log.Print(id, pose)
				if rob, ok := robotList[id]; ok {
					rob.UpdatePose(rcd)
				}
			}
		}
	}
}

type PubRobState struct {
	Pose  msg.Pose `json:"pose"`
	Grids []int    `json:"pathGrids"`
}

func publishState() {
	timer := time.NewTicker(time.Second / 5)
	defer timer.Stop()

	for _ = range timer.C {
		for key, val := range robotList {
			s := PubRobState{
				Pose:  val.PoseStamp.Pose,
				Grids: val.PathGrids,
			}
			message, err := json.Marshal(&s)
			// log.Print(string(message))
			if err != nil {
				log.Print("json marshal error", err)
			}
			top := fmt.Sprintf("/robot/status/%d", key)
			serr := synerex.StatePublish(top, message)
			if serr != nil {
				log.Print("state publish failure", serr)
			}
		}
	}
}

func main() {
	fn, _ := os.Executable()
	log.Print("starting", fn)

	wg := sync.WaitGroup{} //wait exit for gorouting

	robotList = make(map[int]*robot.RobotStatus)
	robotList[1] = robot.NewRobot(1, 0.5, 1.5, 1.5, 0.5, mapFile, yamlFile)
	robotList[2] = robot.NewRobot(2, 1, 1.5, 1.5, 0.5, mapFile, yamlFile)
	robotList[3] = robot.NewRobot(3, 1, 1.5, 1.5, 0.5, mapFile, yamlFile)

	wg.Add(1)
	synerex.RunSynerex()
	go synerex.SubscribeMQTTSupply(synerex.Mqttclient, supplyMQTTCallback)
	go publishState()

	wg.Wait()
	sxutil.CallDeferFunctions()
}
