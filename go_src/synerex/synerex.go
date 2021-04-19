//synerex package for using with mqtt-gateway provider
package synerex

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	sxmqtt "github.com/synerex/proto_mqtt"
	api "github.com/synerex/synerex_api"
	pbase "github.com/synerex/synerex_proto"
	sxutil "github.com/synerex/synerex_sxutil"
	"google.golang.org/protobuf/proto"
)

var (
	Nodesrv         = flag.String("nodesrv", "127.0.0.1:9990", "Node ID Server")
	SxServerAddress string
	Mu              sync.Mutex
	Mqttclient      *sxutil.SXServiceClient
)

// Setup synerex
func RunSynerex() {
	flag.Parse()
	go sxutil.HandleSigInt() //exit by Ctrl + C
	sxutil.RegisterDeferFunction(sxutil.UnRegisterNode)

	channelTypes := []uint32{pbase.MQTT_GATEWAY_SVC}
	// obtain synerex server address from nodeserv
	srv, err := sxutil.RegisterNode(*Nodesrv, "Grid", channelTypes, nil)
	if err != nil {
		log.Fatal("Can't register node...")
	}
	log.Printf("Connecting Server [%s]\n", srv)

	SxServerAddress = srv

	client := sxutil.GrpcConnectServer(srv)
	argJSON1 := fmt.Sprintf("{Client:GRID_MQTT}")
	Mqttclient = sxutil.NewSXServiceClient(client, pbase.MQTT_GATEWAY_SVC, argJSON1)

	log.Print("running synerex")
}

func reconnectClient(client *sxutil.SXServiceClient) {
	Mu.Lock()
	if client.SXClient != nil {
		client.SXClient = nil
		log.Printf("Client reset \n")
	}
	Mu.Unlock()
	time.Sleep(5 * time.Second) // wait 5 seconds to reconnect
	Mu.Lock()
	if client.SXClient == nil {
		newClt := sxutil.GrpcConnectServer(SxServerAddress)
		if newClt != nil {
			log.Printf("Reconnect server [%s]\n", SxServerAddress)
			client.SXClient = newClt
		}
	} else { // someone may connect!
		log.Print("Use reconnected server\n", SxServerAddress)
	}
	Mu.Unlock()
}

// start subscribe from synerex's node server
func SubscribeMQTTSupply(client *sxutil.SXServiceClient, callback func(clt *sxutil.SXServiceClient, sp *api.Supply)) {
	//Goroutine! wait message from CLI
	ctx := context.Background()
	for { // make it continuously working..
		err := client.SubscribeSupply(ctx, callback)
		if err != nil {
			log.Print("Error on subscribe MQTT")
			reconnectClient(client)
		}
	}
}

func StatePublish(topic string, content []byte) error {
	smo := GeneMqttSupply(topic, content)
	_, err := Mqttclient.NotifySupply(smo)
	if err != nil {
		return err
	}
	return nil
}

func GeneMqttSupply(topic string, content []byte) *sxutil.SupplyOpts {
	if len(content) > 268435455 {
		log.Print("message size is overflow!!", len(content))
	}
	rcd := sxmqtt.MQTTRecord{
		Topic:  topic,
		Record: content,
	}
	out, err := proto.Marshal(&rcd)
	if err != nil {
		log.Print(err)
	}
	cont := api.Content{Entity: out}
	smo := sxutil.SupplyOpts{
		Name:  "GridMQTTPublish",
		Cdata: &cont,
	}
	return &smo
}
