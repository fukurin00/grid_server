package main

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

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
	yamlFile  string                     = "../map/trusco_map_edited.yaml"
	mapFile   string                     = "../map/trusco_map_edited.pgm"
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
				var path msg.Path
				var id int

				err := json.Unmarshal(rcd.Record, &path)
				if err != nil {
					log.Print(err)
				}
				fmt.Sscanf(rcd.Topic, "robot/path/%d", &id)

				log.Print(id, path)

				if rob, ok := robotList[id]; ok {
					rob.UpdatePath(rcd)
				}

			} else if strings.HasPrefix(rcd.Topic, "robot/pose") {
				var pose msg.ROS_Pose
				var id int

				err := json.Unmarshal(rcd.Record, &pose)
				if err != nil {
					log.Print(err)
				}
				fmt.Sscanf(rcd.Topic, "robot/pose/%d", &id)

				log.Print(id, pose)
				if rob, ok := robotList[id]; ok {
					rob.UpdatePose(rcd)
				}
			}

		}
	}
}

func publishState() {
	timer := time.NewTicker(time.Second / 5)
	defer timer.Stop()

	for _ = range timer.C {
		for key, val := range robotList {
			message, err := json.Marshal(val.EstPose)
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
	_, fn, _, _ := runtime.Caller(1)
	log.Print("starting", fn)

	wg := sync.WaitGroup{} //wait exit for gorouting

	robotList = make(map[int]*robot.RobotStatus)
	robotList[1] = robot.NewRobot(1, 0.5, 3.0, mapFile, yamlFile)
	robotList[2] = robot.NewRobot(2, 1, 2.0, mapFile, yamlFile)

	wg.Add(1)
	synerex.RunSynerex()
	go synerex.SubscribeMQTTSupply(synerex.Mqttclient, supplyMQTTCallback)
	go publishState()

	wg.Wait()
	sxutil.CallDeferFunctions()
}
